// Package ai provides AI orchestration for phishing campaigns
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AIOrchestrator управляет AI моделями для генерации контента
type AIOrchestrator struct {
	mu           sync.RWMutex
	config       *Config
	client       *http.Client
	logger       *zap.Logger
	endpoint     string
	apiKey       string
	model        string
}

// Config конфигурация AI оркестратора
type Config struct {
	Endpoint       string        `json:"endpoint"`
	APIKey         string        `json:"api_key"`
	Model          string        `json:"model"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	Temperature    float64       `json:"temperature"`
	MaxTokens      int           `json:"max_tokens"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Endpoint:    "http://localhost:8081",
		Timeout:     60 * time.Second,
		MaxRetries:  3,
		Temperature: 0.7,
		MaxTokens:   2048,
		Model:       "llama-3.1-70b",
	}
}

// NewAIOrchestrator создает новый AI оркестратор
func NewAIOrchestrator(endpoint string, logger *zap.Logger) *AIOrchestrator {
	config := DefaultConfig()
	if endpoint != "" {
		config.Endpoint = endpoint
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	orchestrator := &AIOrchestrator{
		config:   config,
		client:   client,
		logger:   logger,
		endpoint: endpoint,
		apiKey:   config.APIKey,
		model:    config.Model,
	}

	logger.Info("AI Orchestrator initialized",
		zap.String("endpoint", endpoint),
		zap.String("model", config.Model))

	return orchestrator
}

// GenerateEmail генерирует фишинговое письмо
func (a *AIOrchestrator) GenerateEmail(ctx context.Context, targetData map[string]interface{}, template string) (string, error) {
	a.logger.Debug("Generating phishing email",
		zap.Any("target_data", targetData),
		zap.String("template", template))

	prompt := buildEmailPrompt(targetData, template)

	response, err := a.generateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate email: %w", err)
	}

	a.logger.Info("Email generated", zap.Int("length", len(response)))
	return response, nil
}

// GenerateSubject генерирует тему письма
func (a *AIOrchestrator) GenerateSubject(ctx context.Context, emailBody string) (string, error) {
	a.logger.Debug("Generating email subject")

	prompt := fmt.Sprintf(`Generate a compelling email subject line for this email:

%s

Subject:`, emailBody)

	response, err := a.generateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate subject: %w", err)
	}

	return response, nil
}

// PersonalizeContent персонализирует контент под цель
func (a *AIOrchestrator) PersonalizeContent(ctx context.Context, content string, targetProfile map[string]interface{}) (string, error) {
	a.logger.Debug("Personalizing content")

	prompt := fmt.Sprintf(`Personalize this content for the target:

Target Profile:
- Name: %v
- Company: %v
- Role: %v
- Interests: %v

Original Content:
%s

Personalized Content:`,
		targetProfile["name"],
		targetProfile["company"],
		targetProfile["role"],
		targetProfile["interests"],
		content)

	response, err := a.generateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to personalize content: %w", err)
	}

	return response, nil
}

// AnalyzeCredential анализирует перехваченные креденшалы
func (a *AIOrchestrator) AnalyzeCredential(ctx context.Context, username, password string, metadata map[string]interface{}) (map[string]interface{}, error) {
	a.logger.Debug("Analyzing credential")

	// Заглушка - в реальности будет анализ на утечки
	analysis := map[string]interface{}{
		"username":      username,
		"password_strength": "unknown",
		"reuse_risk":    "unknown",
		"recommendations": []string{
			"Change password immediately",
			"Enable 2FA",
			"Check for breaches",
		},
	}

	return analysis, nil
}

// GenerateReport генерирует отчет по кампании
func (a *AIOrchestrator) GenerateReport(ctx context.Context, campaignData map[string]interface{}) (string, error) {
	a.logger.Debug("Generating campaign report")

	prompt := fmt.Sprintf(`Generate a comprehensive security assessment report based on this campaign data:

%v

Report:`, campaignData)

	response, err := a.generateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate report: %w", err)
	}

	return response, nil
}

// generateText отправляет запрос к AI модели
func (a *AIOrchestrator) generateText(ctx context.Context, prompt string) (string, error) {
	// Пробуем внешний AI сервис
	if a.endpoint != "" && a.endpoint != "http://localhost:8081" {
		text, err := a.callExternalAI(ctx, prompt)
		if err == nil {
			return text, nil
		}
		a.logger.Warn("External AI failed, using local generation", zap.Error(err))
	}

	// Локальная генерация на основе шаблонов
	return a.localGenerate(prompt), nil
}

// callExternalAI вызывает внешний AI сервис
func (a *AIOrchestrator) callExternalAI(ctx context.Context, prompt string) (string, error) {
	request := map[string]interface{}{
		"model":       a.model,
		"prompt":      prompt,
		"temperature": a.config.Temperature,
		"max_tokens":  a.config.MaxTokens,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	var lastErr error
	for i := 0; i < a.config.MaxRetries; i++ {
		resp, err := a.client.Post(
			fmt.Sprintf("%s/v1/generate", a.endpoint),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			lastErr = err
			a.logger.Warn("AI request failed", zap.Error(err), zap.Int("retry", i+1))
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = err
			continue
		}

		if text, ok := result["text"].(string); ok {
			return text, nil
		}

		if error_msg, ok := result["error"].(string); ok {
			return "", fmt.Errorf("AI API error: %s", error_msg)
		}
	}

	return "", fmt.Errorf("all retries failed: %w", lastErr)
}

// localGenerate локальная генерация на основе шаблонов
func (a *AIOrchestrator) localGenerate(prompt string) string {
	promptLower := strings.ToLower(prompt)

	// Определяем тип по ключевым словам
	if strings.Contains(promptLower, "phishing") || strings.Contains(promptLower, "verify") {
		return `Dear Valued Customer,

We have detected unusual activity on your account. For your security, please verify your information immediately.

Click here to verify: [LINK]

If you did not request this, please ignore this message.

Security Team`
	}

	if strings.Contains(promptLower, "invoice") || strings.Contains(promptLower, "payment") {
		return `INVOICE NOTIFICATION

Your invoice #INV-2024-001 is ready for review.

Amount: $1,247.50
Due Date: Within 7 days

Please download and review the attached invoice.

Accounting Department`
	}

	if strings.Contains(promptLower, "meeting") || strings.Contains(promptLower, "invite") {
		return `Meeting Invitation

You have been invited to a meeting.

Topic: Security Update
Time: Tomorrow at 10:00 AM
Location: Conference Room A / Zoom

Please confirm your attendance.

Best regards`
	}

	if strings.Contains(promptLower, "password") || strings.Contains(promptLower, "reset") {
		return `Password Reset Request

We received a request to reset your password.

If you made this request, click here: [RESET_LINK]

If not, your account may be at risk. Please contact support immediately.

Support Team`
	}

	// Default template
	return `Dear User,

This is an important notification regarding your account.

Please review the information and take appropriate action.

Thank you,
Support Team`
}

// HealthCheck проверяет доступность AI сервиса
func (a *AIOrchestrator) HealthCheck(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := a.client.Get(fmt.Sprintf("%s/health", a.endpoint))
	if err != nil {
		a.logger.Warn("AI health check failed", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetModels получает доступные модели
func (a *AIOrchestrator) GetModels(ctx context.Context) ([]string, error) {
	resp, err := a.client.Get(fmt.Sprintf("%s/v1/models", a.endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models, ok := result["models"].([]interface{})
	if !ok {
		return []string{a.model}, nil
	}

	modelNames := make([]string, len(models))
	for i, m := range models {
		if model, ok := m.(string); ok {
			modelNames[i] = model
		}
	}

	return modelNames, nil
}

// SetModel устанавливает модель для генерации
func (a *AIOrchestrator) SetModel(model string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.model = model
	a.logger.Info("AI model changed", zap.String("model", model))
}

// GetStats возвращает статистику использования
func (a *AIOrchestrator) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"endpoint":      a.endpoint,
		"model":         a.model,
		"timeout":       a.config.Timeout.String(),
		"max_retries":   a.config.MaxRetries,
		"temperature":   a.config.Temperature,
		"max_tokens":    a.config.MaxTokens,
	}
}

// Close закрывает оркестратор
func (a *AIOrchestrator) Close() error {
	a.logger.Info("AI Orchestrator closed")
	return nil
}

// buildEmailPrompt строит промпт для генерации письма
func buildEmailPrompt(targetData map[string]interface{}, template string) string {
	return fmt.Sprintf(`Write a professional phishing email using this template:

Template: %s

Target Information:
- Name: %v
- Company: %v
- Position: %v
- Email: %v

Make it convincing and personalized. Do not mention that this is a simulation.

Email:`,
		template,
		targetData["name"],
		targetData["company"],
		targetData["position"],
		targetData["email"])
}
