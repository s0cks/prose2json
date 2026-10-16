package prose

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	_ "embed"

	"github.com/ollama/ollama/api"
)

type ProseClient struct {
	client       *api.Client
	model        string
	schema       []byte
	systemPrompt string
}

func NewProseClient(ollamaUrl string, model string, schema []byte, systemPrompt string) (*ProseClient, error) {
	u, err := url.Parse(ollamaUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid ollama url `%s`: %v", ollamaUrl, err)
	}

	log.Printf("connecting to ollama at %s...", ollamaUrl)
	client := &ProseClient{
		client:       api.NewClient(u, http.DefaultClient),
		model:        model,
		systemPrompt: systemPrompt,
		schema:       schema,
	}

	return client, nil

}

func (client *ProseClient) StopModel() error {
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &api.GenerateRequest{
		Model:     client.model,
		KeepAlive: &api.Duration{Duration: 0},
	}
	respFunc := func(resp api.GenerateResponse) error {
		return nil
	}

	return client.client.Generate(stopCtx, req, respFunc)
}

func (client *ProseClient) newRequest(prose string) *api.ChatRequest {
	return &api.ChatRequest{
		Model:  client.model,
		Format: json.RawMessage(client.schema),
		Messages: []api.Message{
			{
				Role:    "system",
				Content: client.systemPrompt,
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

func (client *ProseClient) Submit(ctx context.Context, prose string) ([]byte, error) {
	var result []byte
	respFunc := func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)
		result = append(result, resp.Message.Content...)
		return nil
	}

	log.Println("sending request....")
	err := client.client.Chat(ctx, client.newRequest(prose), respFunc)
	if err != nil {
		return nil, err
	}

	log.Println("received response")
	return result, nil
}
