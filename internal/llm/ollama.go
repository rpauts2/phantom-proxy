package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ============================================================================
// LLM Client for Ollama (local LLM inference)
// ============================================================================

// Client for local LLM inference via Ollama
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// Config LLM configuration
type Config struct {
	BaseURL    string // http://localhost:11434
	Model      string // llama3.2, mistral, etc.
	Timeout    time.Duration
	MaxTokens  int
	Temperature float64
}

// NewClient creates new LLM client
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}
	if config.Model == "" {
		config.Model = "llama3.2"
	}
	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2048
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	return &Client{
		baseURL: config.BaseURL,
		model:   config.Model,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GenerateRequest request for text generation
type GenerateRequest struct {
	Model    string  `json:"model"`
	Prompt   string  `json:"prompt"`
	Stream   bool    `json:"stream"`
	Options  Options `json:"options,omitempty"`
}

// Options generation options
type Options struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
	MaxTokens   int     `json:"num_predict,omitempty"`
	Stop       []string `json:"stop,omitempty"`
}

// GenerateResponse response from LLM
type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Context   []int  `json:"context,omitempty"`
	TotalDur  int64  `json:"total_duration,omitempty"`
	LoadDur   int64  `json:"load_duration,omitempty"`
	PromptDur int64  `json:"prompt_duration,omitempty"`
	EvalDur   int64  `json:"eval_duration,omitempty"`
}

// Generate generates text from prompt
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	req := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		Options: Options{
			Temperature: 0.7,
			MaxTokens:   2048,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %d - %s", resp.StatusCode, string(body))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

// GenerateWithOptions generates with custom options
func (c *Client) GenerateWithOptions(ctx context.Context, prompt string, opts Options) (string, error) {
	req := GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		Options: opts,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %d - %s", resp.StatusCode, string(body))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Response, nil
}

// ChatRequest request for chat completion
type ChatRequest struct {
	Model    string   `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool    `json:"stream"`
}

// Message chat message
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// ChatResponse response from chat
type ChatResponse struct {
	Model      string    `json:"model"`
	Message    Message   `json:"message"`
	Done       bool      `json:"done"`
	TotalDur   int64     `json:"total_duration,omitempty"`
	EvalDur    int64     `json:"eval_duration,omitempty"`
}

// Chat generates response in chat format
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %d - %s", resp.StatusCode, string(body))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Message.Content, nil
}

// ListModels lists available models
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// EmbeddingRequest request for embeddings
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse response with embeddings
type EmbeddingResponse struct {
	Model      string    `json:"model"`
	Embedding  []float64 `json:"embedding"`
}

// GetEmbedding generates embedding for text
func (c *Client) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	req := EmbeddingRequest{
		Model:  c.model,
		Prompt: text,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/embeddings", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error: %d", resp.StatusCode)
	}

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Embedding, nil
}

// HealthCheck checks if Ollama is available
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama not available: %d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// Phishing Content Generation
// ============================================================================

// GeneratePhishingEmail generates phishing email content
func (c *Client) GeneratePhishingEmail(ctx context.Context, target, templateType string) (string, error) {
	prompt := fmt.Sprintf(`Generate a professional phishing email for target: %s
Template type: %s

Requirements:
- Convincing and realistic
- Uses social engineering
- No explicit markers
- Professional tone

Email content:`, target, templateType)

	return c.Generate(ctx, prompt)
}

// AnalyzeTarget analyzes target for vulnerabilities
func (c *Client) AnalyzeTarget(ctx context.Context, targetInfo string) (map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Analyze this target information and provide:
1. Potential attack vectors
2. Likely security awareness level
3. Recommended phishing approach
4. Key information to leverage

Target: %s

Provide analysis in JSON format.`, targetInfo)

	result, err := c.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse JSON from response
	result = strings.TrimSpace(result)
	result = strings.Trim(result, "```json")
	result = strings.Trim(result, "```")

	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(result), &analysis); err != nil {
		// Return as text if not valid JSON
		return map[string]interface{}{
			"analysis": result,
		}, nil
	}

	return analysis, nil
}
