package model

import (
	"context"
	"fmt"

	"github.com/ValerySidorin/gigago/client"
	"github.com/tmc/langchaingo/llms"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type LLM struct {
	gigaClient *client.Client
	model      string
}

var _ llms.Model = (*LLM)(nil)

func New(gigaClient *client.Client, model string) *LLM {
	return &LLM{
		gigaClient: gigaClient,
		model:      model,
	}
}

func (o *LLM) Call(
	ctx context.Context, prompt string, options ...llms.CallOption,
) (string, error) {
	chatReq := &client.ChatRequest{
		Model: o.model,
		Messages: []client.ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.Temperature > 0 {
		temp := opts.Temperature
		chatReq.Temperature = &temp
	}
	if opts.MaxTokens > 0 {
		maxTokens := opts.MaxTokens
		chatReq.MaxTokens = &maxTokens
	}

	resp, err := o.gigaClient.Chat(ctx, chatReq)
	if err != nil {
		return "", fmt.Errorf("failed to call GigaChat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from GigaChat")
	}

	return resp.Choices[0].Message.Content, nil
}

func (o *LLM) GenerateContent(
	ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	chatMessages := make([]client.ChatMessage, len(messages))
	for i, msg := range messages {
		var content string
		for _, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				content = textPart.Text
				break
			}
		}

		var role string
		switch msg.Role {
		case llms.ChatMessageTypeAI:
			role = RoleAssistant
		case llms.ChatMessageTypeHuman, llms.ChatMessageTypeGeneric:
			role = RoleUser
		default:
			role = string(msg.Role)
		}

		chatMessages[i] = client.ChatMessage{
			Role:    role,
			Content: content,
		}
	}

	chatReq := &client.ChatRequest{
		Model:    o.model,
		Messages: chatMessages,
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.Temperature > 0 {
		temp := opts.Temperature
		chatReq.Temperature = &temp
	}
	if opts.MaxTokens > 0 {
		maxTokens := opts.MaxTokens
		chatReq.MaxTokens = &maxTokens
	}

	resp, err := o.gigaClient.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from GigaChat")
	}

	content := resp.Choices[0].Message.Content
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: content,
			},
		},
	}, nil
}
