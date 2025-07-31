package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ValerySidorin/gigago/client"
	"github.com/ValerySidorin/gigago/model"
	"github.com/tmc/langchaingo/embeddings"
)

func main() {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	authKey := os.Getenv("GIGACHAT_AUTH_KEY")
	if authKey == "" {
		fmt.Println("Set env var GIGACHAT_AUTH_KEY")
		fmt.Println("Example: export GIGACHAT_AUTH_KEY=\"$(echo -n 'your_auth_key')\"")
		os.Exit(1)
	}

	gigaClient := client.NewClient(authKey, client.WithHTTPClient(httpClient))
	ctx := context.Background()

	fmt.Println("=== GigaGo - GigaChat API Go SDK ===")
	fmt.Println()

	fmt.Println("1. Get available llm models:")
	models, err := gigaClient.GetModels(ctx)
	if err != nil {
		log.Printf("Error getting llm models: %v", err)
	} else {
		for _, model := range models.Data {
			fmt.Printf("  - %s (ID: %s)\n", model.Name, model.ID)
		}
	}

	fmt.Println("\n2. Simple chat:")
	chatReq := &client.ChatRequest{
		Model: "GigaChat:latest",
		Messages: []client.ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
	}

	chatResp, err := gigaClient.Chat(ctx, chatReq)
	if err != nil {
		log.Printf("Chat error: %v", err)
	} else {
		if len(chatResp.Choices) > 0 {
			fmt.Printf("Response: %s\n", chatResp.Choices[0].Message.Content)
			fmt.Printf("Tokens used: %d\n", chatResp.Usage.TotalTokens)
		}
	}

	fmt.Println("\n3. Create embeddings:")
	embedReq := &client.EmbeddingRequest{
		Model: "Embeddings",
		Input: []string{"Go", "Golang", "Programming"},
	}

	embedResp, err := gigaClient.CreateEmbeddings(ctx, embedReq)
	if err != nil {
		log.Printf("Error creating embeddings: %v", err)
	} else {
		fmt.Printf("Created embeddings: %d\n", len(embedResp.Data))
		for i, embedding := range embedResp.Data {
			fmt.Printf("  Embedding %d: %d dimensions\n", i+1, len(embedding.Embedding))
		}
	}

	fmt.Println("\n4. Chat with functions:")
	function := client.Function{
		Name:        "get_weather",
		Description: "Get the weather in a specified city",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"city": map[string]interface{}{
					"type":        "string",
					"description": "City name",
				},
			},
			"required": []string{"city"},
		},
	}

	funcChatReq := &client.ChatRequest{
		Model: "GigaChat:latest",
		Messages: []client.ChatMessage{
			{
				Role:    client.RoleUser,
				Content: "What's the weather in Moscow?",
			},
		},
		Functions: []client.Function{function},
		FunctionCall: map[string]string{
			"name": "get_weather",
		},
	}

	funcChatResp, err := gigaClient.Chat(ctx, funcChatReq)
	if err != nil {
		log.Printf("Error chat with functions: %v", err)
	} else {
		if len(funcChatResp.Choices) > 0 {
			choice := funcChatResp.Choices[0]
			if choice.Message.FunctionCall != nil {
				fmt.Printf("Function called: %s\n", choice.Message.FunctionCall.Name)
				fmt.Printf("Args: %v\n", choice.Message.FunctionCall.Arguments)
			} else {
				fmt.Printf("Response: %s\n", choice.Message.Content)
			}
		}
	}

	fmt.Println("\n5. Upload files:")

	tempFile := "temp_example.txt"
	err = os.WriteFile(tempFile, []byte("This is a test file for GigaChat"), 0644)
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return
	}
	defer os.Remove(tempFile)

	file, err := gigaClient.UploadFile(ctx, tempFile, client.General)
	if err != nil {
		log.Printf("Error uploading file: %v", err)
	} else {
		fmt.Printf("File uploaded: %s (ID: %s)\n", file.Filename, file.ID)
	}

	fmt.Println("\n5. List files:")
	files, err := gigaClient.GetFiles(ctx)
	if err != nil {
		log.Printf("Error getting files: %v", err)
	} else {
		fmt.Printf("Found files: %d\n", len(files.Data))
		for _, f := range files.Data {
			fmt.Printf("- %s (ID: %s, size: %d bytes)\n", f.Filename, f.ID, f.Bytes)
		}
	}

	fmt.Println("\n6. Usage with langchaingo:")
	llm := model.New(gigaClient, "GigaChat:latest")

	response, err := llm.Call(ctx, "Hello! How are you?")
	if err != nil {
		log.Printf("Langchaingo error: %v", err)
	} else {
		fmt.Printf("Langchaingo response: %s\n", response)
	}

	embeddingsLLM := model.New(gigaClient, "Embeddings")
	embedder, err := embeddings.NewEmbedder(embeddingsLLM)
	if err != nil {
		log.Fatalf("Error creating Langchaingo embedder: %v", err)
	}

	texts := []string{"Hello, world!", "GigaChat is awesome!"}
	vectors, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		log.Printf("Error creating embeddings: %v", err)
	}
	for i, vec := range vectors {
		fmt.Printf("Embedding %d: %d dimensions\n", i+1, len(vec))
	}

	fmt.Println("\n=== Example finished ===")
}
