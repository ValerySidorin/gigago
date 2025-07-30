package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Scope string

const (
	GIGACHAT_API_PERS Scope = "GIGACHAT_API_PERS"
	GIGACHAT_API_B2B  Scope = "GIGACHAT_API_B2B"
	GIGACHAT_API_CORP Scope = "GIGACHAT_API_CORP"
)

type Purpose string

const (
	General Purpose = "general"
)

// Client представляет клиент для работы с GigaChat API
type Client struct {
	httpClient    *http.Client
	baseURL       string
	authURL       string
	authorization string
	accessToken   string
	tokenExpiry   time.Time
}

// NewClient создает новый клиент GigaChat
func NewClient(authKey string, opts ...Option) *Client {
	cl := &Client{
		httpClient:    http.DefaultClient,
		baseURL:       "https://gigachat.devices.sberbank.ru/api/v1",
		authURL:       "https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
		authorization: "Basic " + authKey,
	}

	for _, opt := range opts {
		opt(cl)
	}

	return cl
}

// TokenResponse представляет ответ на запрос токена
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

// Model представляет модель GigaChat
type Model struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelsResponse представляет ответ со списком моделей
type ModelsResponse struct {
	Data []Model `json:"data"`
}

// Message представляет сообщение в чате
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Function представляет функцию для вызова
type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

// FunctionCall представляет вызов функции
type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ChatMessage представляет сообщение в чате с возможными функциями
type ChatMessage struct {
	Role         string        `json:"role"`
	Content      string        `json:"content,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

// ChatRequest представляет запрос на чат
type ChatRequest struct {
	Model        string        `json:"model"`
	Messages     []ChatMessage `json:"messages"`
	Temperature  *float64      `json:"temperature,omitempty"`
	TopP         *float64      `json:"top_p,omitempty"`
	N            *int          `json:"n,omitempty"`
	Stream       *bool         `json:"stream,omitempty"`
	MaxTokens    *int          `json:"max_tokens,omitempty"`
	Functions    []Function    `json:"functions,omitempty"`
	FunctionCall any           `json:"function_call,omitempty"`
}

// ChatResponse представляет ответ от чата
type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

// ChatChoice представляет выбор модели
type ChatChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	Delta   ChatMessage `json:"delta,omitempty"`
}

// Usage представляет использование токенов
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// File представляет файл в хранилище
type File struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

// FilesResponse представляет ответ со списком файлов
type FilesResponse struct {
	Data []File `json:"data"`
}

// EmbeddingRequest представляет запрос на создание эмбеддингов
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse представляет ответ с эмбеддингами
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Usage  Usage       `json:"usage"`
}

// Embedding представляет эмбеддинг
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// GetAccessToken получает токен доступа
func (c *Client) GetAccessToken(ctx context.Context, scope Scope) error {
	data := fmt.Sprintf("scope=%s", scope)
	req, err := http.NewRequestWithContext(ctx, "POST", c.authURL, bytes.NewBufferString(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", uuid.New().String())
	req.Header.Set("Authorization", c.authorization)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Unix(tokenResp.ExpiresAt, 0)

	return nil
}

// ensureToken проверяет и обновляет токен при необходимости
func (c *Client) ensureToken(ctx context.Context) error {
	if c.accessToken == "" || time.Now().After(c.tokenExpiry.Add(-5*time.Minute)) {
		return c.GetAccessToken(ctx, GIGACHAT_API_PERS)
	}
	return nil
}

// makeRequest выполняет HTTP запрос с автоматическим обновлением токена
func (c *Client) makeRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Попробуем обновить токен и повторить запрос
		if err := c.GetAccessToken(ctx, GIGACHAT_API_PERS); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to retry request: %w", err)
		}
	}

	return resp, nil
}

// GetModels получает список доступных моделей
func (c *Client) GetModels(ctx context.Context) (*ModelsResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/models", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get models with status %d: %s", resp.StatusCode, string(body))
	}

	var models ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	return &models, nil
}

// Chat выполняет запрос к чату
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/chat/completions", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to chat with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode chat response: %w", err)
	}

	return &chatResp, nil
}

// CreateEmbeddings создает эмбеддинги для текста
func (c *Client) CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	resp, err := c.makeRequest(ctx, "POST", "/embeddings", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create embeddings with status %d: %s", resp.StatusCode, string(body))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embeddings response: %w", err)
	}

	return &embeddingResp, nil
}

// UploadFile загружает файл в хранилище
func (c *Client) UploadFile(
	ctx context.Context, filePath string, purpose Purpose,
) (*File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var contentType string
	ext := filepath.Ext(filePath)
	if ext != "" {
		contentType = mime.TypeByExtension(ext)
	}

	if contentType == "" {
		return nil, fmt.Errorf("failed to determine content type of file: %s", filePath)
	}

	return c.UploadFileReader(
		ctx, file, filepath.Base(filePath), contentType, purpose,
	)
}

func (c *Client) UploadFileReader(
	ctx context.Context,
	r io.Reader, fileName string, contentType string,
	purpose Purpose,
) (*File, error) {
	if contentType == "" || contentType == "application/octet-stream" {
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, r); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.WriteField("purpose", string(purpose)); err != nil {
		return nil, fmt.Errorf("failed to write purpose field: %w", err)
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/files", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to upload file with status %d: %s", resp.StatusCode, string(body))
	}

	var uploadedFile File
	if err := json.NewDecoder(resp.Body).Decode(&uploadedFile); err != nil {
		return nil, fmt.Errorf("failed to decode file response: %w", err)
	}

	return &uploadedFile, nil
}

// GetFiles получает список файлов
func (c *Client) GetFiles(ctx context.Context) (*FilesResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/files", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get files with status %d: %s", resp.StatusCode, string(body))
	}

	var files FilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to decode files response: %w", err)
	}

	return &files, nil
}

// GetFile получает информацию о файле
func (c *Client) GetFile(ctx context.Context, fileID string) (*File, error) {
	resp, err := c.makeRequest(ctx, "GET", "/files/"+fileID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get file with status %d: %s", resp.StatusCode, string(body))
	}

	var file File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("failed to decode file response: %w", err)
	}

	return &file, nil
}

// DeleteFile удаляет файл
func (c *Client) DeleteFile(ctx context.Context, fileID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", "/files/"+fileID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete file with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DownloadFile скачивает файл
func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	resp, err := c.makeRequest(ctx, "GET", "/files/"+fileID+"/content", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download file with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
