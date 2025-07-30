package model

import (
	"context"
	"testing"

	"github.com/ValerySidorin/gigago/client"
	"github.com/tmc/langchaingo/llms"
)

func TestNew(t *testing.T) {
	gigaClient := &client.Client{}
	modelName := "GigaChat:latest"

	llm := New(gigaClient, modelName)

	if llm == nil {
		t.Fatal("New returned nil")
	}

	if llm.gigaClient != gigaClient {
		t.Errorf("Expected gigaClient to be %v, got %v", gigaClient, llm.gigaClient)
	}

	if llm.model != modelName {
		t.Errorf("Expected model to be '%s', got '%s'", modelName, llm.model)
	}
}

func TestCallRequestStructure(t *testing.T) {
	gigaClient := client.NewClient("Basic invalid_auth")
	llm := New(gigaClient, "GigaChat:latest")

	ctx := context.Background()
	prompt := "Test prompt"

	_, err := llm.Call(ctx, prompt)
	if err == nil {
		t.Error("Expected error with invalid credentials")
	}
}

func TestGenerateContentRequestStructure(t *testing.T) {
	gigaClient := client.NewClient("Basic invalid_auth")
	llm := New(gigaClient, "GigaChat:latest")

	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello"),
			},
		},
	}

	_, err := llm.GenerateContent(ctx, messages)
	if err == nil {
		t.Error("Expected error with invalid credentials")
	}
}
