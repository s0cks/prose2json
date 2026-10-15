package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "embed"

	"github.com/ollama/ollama/api"
	"github.com/s0cks/prose2json/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCommand *cobra.Command
var client *api.Client

const (
	DefaultOllamaHost = "http://localhost:11434"
)

func connectToOllama() *api.Client {
	ollamaUrl := viper.GetString("host")
	u, err := url.Parse(ollamaUrl)
	if err != nil {
		log.Fatalf("Invalid Ollama URL: %v", err)
	}

	log.Printf("connecting to ollama at %s...", ollamaUrl)
	return api.NewClient(u, http.DefaultClient)
}

func getSchema() (map[string]any, []byte, error) {
	schemaPath := viper.GetString("SCHEMA")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read schema from path `%s`: %v", schemaPath, err)
	}

	var schema map[string]any
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal schema from path `%s`: %v", schemaPath, err)
	}

	schemaData, err := json.Marshal(schema)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal schema %s: %v", string(schemaBytes), err)
	}

	return schema, schemaData, nil
}

//go:embed system_prompt.txt
var systemPrompt string

func createRequest(prose string, schema []byte) *api.ChatRequest {
	model := viper.GetString("model")
	log.Printf("system prompt: %s", systemPrompt)
	log.Printf("schema: %s", string(schema))
	return &api.ChatRequest{
		Model:  model,
		Format: json.RawMessage(schema),
		Messages: []api.Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "odm",
				Content: prose,
			},
			{
				Role:    "user",
				Content: prose,
			},
		},
		Options: map[string]any{
			"temperature": 0,
		},
	}
}

func submit(ctx context.Context, req *api.ChatRequest) ([]byte, error) {
	var result []byte
	respFunc := func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)
		result = append(result, resp.Message.Content...)
		return nil
	}

	log.Println("sending request....")
	err := client.Chat(ctx, req, respFunc)
	if err != nil {
		return nil, err
	}
	log.Println("received response")
	return result, nil
}

func init() {
	RootCommand = &cobra.Command{
		Use:  "prose2json",
		Args: cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.SetEnvPrefix("PROSE2JSON")
			viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
			viper.AutomaticEnv()
			return viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			client = connectToOllama()
			_, schemaBytes, err := getSchema()
			if err != nil {
				log.Fatalf("failed to get schema: %v", err)
			}

			req := createRequest(args[0], schemaBytes)
			result, err := submit(ctx, req)
			if err != nil {
				log.Fatalf("failed to submit request: %v", err)
			}

			var summary internal.BookSummary
			if err := json.Unmarshal(result, &summary); err != nil {
				log.Fatalf("Failed to parse JSON response: %v", err)
			}

			out := viper.GetString("output")
			if out != "" {
				if err := os.WriteFile(out, result, 0644); err != nil {
					log.Fatalf("failed to write output file `%s`: %v", out, err)
				}
			}

			fmt.Printf("\nParsed Struct: %+v\n", summary)
			return nil
		},
	}
	RootCommand.Flags().String("host", DefaultOllamaHost, "The ollama host endpoint")
	RootCommand.Flags().StringP("schema", "s", "", "The schema to convert to")
	RootCommand.Flags().StringP("model", "m", "nuextract:3.8b", "The ollama model to use")
	RootCommand.Flags().StringP("output", "o", "", "The output file to write the JSON to")
}

func stopModel() func() error {
	return func() error {
		if client == nil {
			return nil
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		model := viper.GetString("model")
		req := &api.GenerateRequest{
			Model:     model,
			KeepAlive: &api.Duration{Duration: 0},
		}
		respFunc := func(resp api.GenerateResponse) error {
			return nil
		}

		return client.Generate(stopCtx, req, respFunc)
	}
}

func Exec(ctx context.Context) (error, func() error) {
	return RootCommand.ExecuteContext(ctx), stopModel()
}
