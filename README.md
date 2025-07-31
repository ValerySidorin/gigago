# GigaGo - Go client for GigaChat API

This project provides a Go client for working with the [GigaChat API](https://developers.sber.ru/docs/files/openapi/gigachat/api.yml) by Sber.

## Features

- ✅ Automatic access token management
- ✅ Chat support with GigaChat models
- ✅ Embeddings creation
- ✅ Function calling support
- ✅ File upload and management
- ✅ Integration with langchaingo
- ✅ Support for all main GigaChat models

## Installation

```bash
go get github.com/ValerySidorin/gigago
```

## Quick Start

### 1. Get your authorization credentials

To use the GigaChat API, you need to obtain authorization credentials:

1. Register at [developers.sber.ru](https://developers.sber.ru/)
2. Create an application in the GigaChat section
3. Obtain your Client ID and Client Secret

### 2. Create a client

```go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "log"
    "os"

    "github.com/ValerySidorin/gigago/client"
)

func main() {
    // Create the authorization key
    authKey := "your_auth_key"
    
    // Create the client
    gigaClient := client.NewClient(authKey)
    
    ctx := context.Background()
    
    // Get the list of models
    models, err := gigaClient.GetModels(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, model := range models.Data {
        fmt.Printf("Model: %s (ID: %s)\n", model.Name, model.ID)
    }
}
```

### 3. Simple chat

```go
// Create a chat request
chatReq := &client.ChatRequest{
    Model: "GigaChat:latest",
    Messages: []client.ChatMessage{
        {
            Role:    "user",
            Content: "Hello! How are you?",
        },
    },
}

// Send the request
resp, err := gigaClient.Chat(ctx, chatReq)
if err != nil {
    log.Fatal(err)
}

// Get the response
if len(resp.Choices) > 0 {
    fmt.Printf("Response: %s\n", resp.Choices[0].Message.Content)
    fmt.Printf("Tokens used: %d\n", resp.Usage.TotalTokens)
}
```

### 4. Creating embeddings

```go
embedReq := &client.EmbeddingRequest{
    Model: "Embeddings",
    Input: []string{"Hello, world!", "Привет, мир!"},
}

embedResp, err := gigaClient.CreateEmbeddings(ctx, embedReq)
if err != nil {
    log.Fatal(err)
}

for i, embedding := range embedResp.Data {
    fmt.Printf("Embedding %d: %d dimensions\n", i+1, len(embedding.Embedding))
}
```

### 4a. Creating embeddings via langchaingo (langchain-go)

```go
import (
    "context"
    "fmt"
    "github.com/ValerySidorin/gigago/client"
    "github.com/ValerySidorin/gigago/model"
    "github.com/tmc/langchaingo/embeddings"
)

func main() {
    gigaClient := client.NewClient("your_auth_key")
    llm := model.New(gigaClient, "Embeddings")
    embedder := embeddings.NewEmbedder(llm)
    ctx := context.Background()

    texts := []string{"Hello, world!", "GigaChat is awesome!"}
    vectors, err := embedder.EmbedDocuments(ctx, texts)
    if err != nil {
        panic(err)
    }
    for i, vec := range vectors {
        fmt.Printf("Embedding %d: %d dimensions\n", i+1, len(vec))
    }
}
```

### 5. Function calling

```go
// Define a function
gigaFunc := client.Function{
    Name:        "get_weather",
    Description: "Get the weather in a specified city",
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "city": map[string]any{
                "type":        "string",
                "description": "City name",
            },
        },
        "required": []string{"city"},
    },
}

// Create a request with a function
funcChatReq := &client.ChatRequest{
    Model: "GigaChat:latest",
    Messages: []client.ChatMessage{
        {
            Role:    "user",
            Content: "What's the weather in Moscow?",
        },
    },
    Functions: []client.Function{gigaFunc},
    FunctionCall: map[string]string{
        "name": "get_weather",
    },
}

funcChatResp, err := gigaClient.Chat(ctx, funcChatReq)
if err != nil {
    log.Fatal(err)
}

if len(funcChatResp.Choices) > 0 {
    choice := funcChatResp.Choices[0]
    if choice.Message.FunctionCall != nil {
        fmt.Printf("Function called: %s\n", choice.Message.FunctionCall.Name)
        fmt.Printf("Arguments: %v\n", choice.Message.FunctionCall.Arguments)
    }
}
```

### 6. File operations

```go
// Upload a file
file, err := gigaClient.UploadFile(ctx, "path/to/file.txt", client.General)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("File uploaded: %s (ID: %s)\n", file.Filename, file.ID)

// Get the list of files
files, err := gigaClient.GetFiles(ctx)
if err != nil {
    log.Fatal(err)
}

for _, f := range files.Data {
    fmt.Printf("- %s (ID: %s, size: %d bytes)\n", f.Filename, f.ID, f.Bytes)
}

// Download a file
content, err := gigaClient.DownloadFile(ctx, file.ID)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("File content: %s\n", string(content))

// Delete a file
err = gigaClient.DeleteFile(ctx, file.ID)
if err != nil {
    log.Fatal(err)
}
```

## Integration with langchaingo

The client is integrated with the [langchaingo](https://github.com/tmc/langchaingo) library:

```go
import (
    "github.com/ValerySidorin/gigago/client"
    "github.com/ValerySidorin/gigago/model"
    "github.com/tmc/langchaingo/llms"
)

// Create a GigaChat client
authKey := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
gigaClient := client.NewClient(authKey)

// Create a langchaingo model
llm := model.New(gigaClient, "GigaChat:latest")

// Use as a regular langchaingo model
response, err := llm.Call(ctx, "Hello! How are you?")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Response: %s\n", response)
```

## Configuration via environment variables

```bash
# Create the authorization key
export GIGACHAT_AUTH_KEY="$(echo -n 'your_auth_key' | base64)"

# Run the example
go run client/example.go
```

## Supported models

- `GigaChat:latest` - latest model version
- `GigaChat:latest-16k` - model with extended context
- `GigaChat:latest-128k` - model with large context
- `Embeddings` - embeddings model
- `EmbeddingsGigaR` - advanced embeddings model

## Error handling

The client automatically handles:
- Access token expiration
- Network errors
- API errors with detailed messages

## License

MIT License

## Links

- [GigaChat API Documentation](https://developers.sber.ru/docs/files/openapi/gigachat/api.yml)
- [langchaingo](https://github.com/tmc/langchaingo)
- [Sber Developers](https://developers.sber.ru/) 