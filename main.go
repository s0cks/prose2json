package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

type BookSummary struct {
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	KeyThemes []string `json:"key_themes"`
	WordCount int      `json:"word_count"`
}

func main() {
	ctx := context.Background()

	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	u, err := url.Parse(ollamaURL)
	if err != nil {
		log.Fatalf("Invalid Ollama URL: %v", err)
	}

	log.Println("connecting to ollama")
	client := api.NewClient(u, http.DefaultClient)
	proseText := `In 1949, George Orwell published 1984, a dystopian social science fiction novel that follows Winston Smith as he navigates a totalitarian society under the constant surveillance of Big Brother.`

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title":  map[string]string{"type": "string"},
			"author": map[string]string{"type": "string"},
			"key_themes": map[string]any{
				"type":  "array",
				"items": map[string]string{"type": "string"},
			},
			"word_count": map[string]string{"type": "integer"},
		},
		"required": []string{"title", "author", "key_themes", "word_count"},
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}

	req := &api.ChatRequest{
		Model: "nuextract:3.8b",
		Messages: []api.Message{
			{
				Role:    "system",
				Content: "Extract information from the prose and return valid JSON matching the schema.",
			},
			{
				Role:    "odm",
				Content: proseText,
			},
			{
				Role:    "user",
				Content: proseText,
			},
		},
		Format: json.RawMessage(schemaBytes),
		Options: map[string]any{
			"temperature": 0,
		},
	}

	var resultJSON []byte
	respFunc := func(resp api.ChatResponse) error {
		fmt.Print(resp.Message.Content)
		resultJSON = append(resultJSON, resp.Message.Content...)
		return nil
	}

	log.Println("sending request....")
	err = client.Chat(ctx, req, respFunc)
	log.Println("received response")
	if err != nil {
		log.Fatalf("Chat error: %v", err)
	}

	var summary BookSummary
	if err := json.Unmarshal(resultJSON, &summary); err != nil {
		log.Fatalf("Failed to parse JSON response: %v", err)
	}

	fmt.Printf("\nParsed Struct: %+v\n", summary)
}
