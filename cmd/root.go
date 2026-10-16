package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "embed"

	"github.com/s0cks/prose2json/internal/prose"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed system_prompt.txt
var systemPrompt string

const (
	DefaultModel      = "nuextract:3.8b"
	DefaultOllamaHost = "http://localhost:11434"
)

var RootCommand *cobra.Command
var client *prose.ProseClient

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

func init() {
	RootCommand = &cobra.Command{
		Use:   "prose2json [prose segment]",
		Short: "Convert prose segments to structured JSON",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.SetEnvPrefix("PROSE2JSON")
			viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
			viper.AutomaticEnv()
			return viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			host := viper.GetString("host")
			model := viper.GetString("model")
			_, schemaBytes, err := getSchema()
			if err != nil {
				return fmt.Errorf("failed to get schema bytes: %v", err)
			}

			client, err = prose.NewProseClient(host, model, schemaBytes, systemPrompt)

			if err != nil {
				return fmt.Errorf("failed to create new ProseClient: %v", err)
			}

			result, err := client.Submit(ctx, args[0])
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}

				return fmt.Errorf("failed to submit request: %v", err)
			}

			out := viper.GetString("output")
			if out != "" {
				if err := os.WriteFile(out, result, 0644); err != nil {
					return fmt.Errorf("failed to write output file `%s`: %v", out, err)
				}
			}

			return nil
		},
	}
	RootCommand.Flags().String("host", DefaultOllamaHost, "The ollama host endpoint")
	RootCommand.Flags().StringP("schema", "s", "", "The schema to convert to")
	RootCommand.Flags().StringP("model", "m", DefaultModel, "The ollama model to use")
	RootCommand.Flags().StringP("output", "o", "", "The output file to write the JSON to")
}

func Exec(ctx context.Context) (error, func() error) {
	return RootCommand.ExecuteContext(ctx), func() error {
		if client == nil {
			return nil
		}

		return client.StopModel()
	}
}
