package client

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	authKey := "test_auth_key"
	client := NewClient("Basic " + authKey)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.authorization != "Basic "+authKey {
		t.Errorf("Expected authorization to be 'Basic %s', got '%s'", authKey, client.authorization)
	}

	if client.baseURL != "https://gigachat.devices.sberbank.ru/api/v1" {
		t.Errorf("Expected baseURL to be 'https://gigachat.devices.sberbank.ru/api/v1', got '%s'", client.baseURL)
	}

	if client.authURL != "https://ngw.devices.sberbank.ru:9443/api/v2/oauth" {
		t.Errorf("Expected authURL to be 'https://ngw.devices.sberbank.ru:9443/api/v2/oauth', got '%s'", client.authURL)
	}
}

func TestChatRequest(t *testing.T) {
	req := &ChatRequest{
		Model: "GigaChat:latest",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Hello, world!",
			},
		},
	}

	if req.Model != "GigaChat:latest" {
		t.Errorf("Expected model to be 'GigaChat:latest', got '%s'", req.Model)
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if req.Messages[0].Role != "user" {
		t.Errorf("Expected role to be 'user', got '%s'", req.Messages[0].Role)
	}

	if req.Messages[0].Content != "Hello, world!" {
		t.Errorf("Expected content to be 'Hello, world!', got '%s'", req.Messages[0].Content)
	}
}

func TestEmbeddingRequest(t *testing.T) {
	req := &EmbeddingRequest{
		Model: "Embeddings",
		Input: []string{"Hello", "World"},
	}

	if req.Model != "Embeddings" {
		t.Errorf("Expected model to be 'Embeddings', got '%s'", req.Model)
	}

	if len(req.Input) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(req.Input))
	}

	if req.Input[0] != "Hello" {
		t.Errorf("Expected first input to be 'Hello', got '%s'", req.Input[0])
	}

	if req.Input[1] != "World" {
		t.Errorf("Expected second input to be 'World', got '%s'", req.Input[1])
	}
}

func TestFunction(t *testing.T) {
	function := Function{
		Name:        "test_function",
		Description: "A test function",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	if function.Name != "test_function" {
		t.Errorf("Expected name to be 'test_function', got '%s'", function.Name)
	}

	if function.Description != "A test function" {
		t.Errorf("Expected description to be 'A test function', got '%s'", function.Description)
	}

	if function.Parameters["type"] != "object" {
		t.Errorf("Expected type to be 'object', got '%v'", function.Parameters["type"])
	}
}

// Mock тест для проверки структуры ответов
func TestResponseStructures(t *testing.T) {
	// Тест структуры ChatResponse
	chatResp := &ChatResponse{
		ID:      "test_id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "GigaChat:latest",
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	if chatResp.ID != "test_id" {
		t.Errorf("Expected ID to be 'test_id', got '%s'", chatResp.ID)
	}

	if len(chatResp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(chatResp.Choices))
	}

	if chatResp.Usage.TotalTokens != 15 {
		t.Errorf("Expected total tokens to be 15, got %d", chatResp.Usage.TotalTokens)
	}

	// Тест структуры EmbeddingResponse
	embedResp := &EmbeddingResponse{
		Object: "list",
		Data: []Embedding{
			{
				Object:    "embedding",
				Embedding: []float64{0.1, 0.2, 0.3},
				Index:     0,
			},
		},
		Usage: Usage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}

	if embedResp.Object != "list" {
		t.Errorf("Expected object to be 'list', got '%s'", embedResp.Object)
	}

	if len(embedResp.Data) != 1 {
		t.Errorf("Expected 1 embedding, got %d", len(embedResp.Data))
	}

	if len(embedResp.Data[0].Embedding) != 3 {
		t.Errorf("Expected embedding to have 3 dimensions, got %d", len(embedResp.Data[0].Embedding))
	}
}

// Интеграционный тест (требует реальных credentials)
func TestIntegrationWithMock(t *testing.T) {
	// Этот тест можно запускать только с реальными credentials
	// Для CI/CD можно использовать mock сервер
	t.Skip("Skipping integration test - requires real credentials")

	authKey := "test_auth_key"
	client := NewClient("Basic " + authKey)
	ctx := context.Background()

	// Тест получения токена (будет fail без реальных credentials)
	err := client.GetAccessToken(ctx, GIGACHAT_API_PERS)
	if err == nil {
		t.Log("Successfully obtained access token")
	} else {
		t.Logf("Expected error without real credentials: %v", err)
	}
}
