// Package playwright provides captcha solving functionality
//go:build ignore
// +build ignore

package playwright

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CaptchaSolver решатель капч через внешние сервисы
type CaptchaSolver struct {
	mu            sync.RWMutex
	config        *Config
	client        *http.Client
	logger        *zap.Logger
	apiKey        string
	serviceURL    string
}

// Config конфигурация решателя капч
type Config struct {
	ServiceName  string        // 2captcha, anticaptcha, etc.
	APIKey       string        // API ключ сервиса
	ServiceURL   string        // URL сервиса
	Timeout      time.Duration // Таймаут запроса
	PollInterval time.Duration // Интервал опроса результата
	MaxAttempts  int           // Максимальное количество попыток
}

// CaptchaResult результат решения капчи
type CaptchaResult struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   error  `json:"error,omitempty"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		ServiceName:  "2captcha",
		Timeout:      60 * time.Second,
		PollInterval: 5 * time.Second,
		MaxAttempts:  12,
	}
}

// NewCaptchaSolver создает новый решатель капч
func NewCaptchaSolver(config *Config, logger *zap.Logger) (*CaptchaSolver, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	s := &CaptchaSolver{
		config:     config,
		client:     client,
		logger:     logger,
		apiKey:     config.APIKey,
		serviceURL: config.ServiceURL,
	}

	// Set default service URL if not provided
	if s.serviceURL == "" {
		switch config.ServiceName {
		case "2captcha":
			s.serviceURL = "https://2captcha.com"
		case "anticaptcha":
			s.serviceURL = "https://api.anti-captcha.com"
		default:
			s.serviceURL = "https://2captcha.com"
		}
	}

	logger.Info("CaptchaSolver initialized",
		zap.String("service", config.ServiceName),
		zap.String("url", s.serviceURL))

	return s, nil
}

// SolveReCAPTCHA решает reCAPTCHA v2/v3
func (s *CaptchaSolver) SolveReCAPTCHA(pageURL, siteKey string) (*CaptchaResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	s.logger.Debug("Solving reCAPTCHA",
		zap.String("page_url", pageURL),
		zap.String("site_key", siteKey))

	// Отправка капчи на решение
	taskID, err := s.submitReCAPTCHA(ctx, pageURL, siteKey)
	if err != nil {
		return &CaptchaResult{
			Success: false,
			Error:   fmt.Errorf("failed to submit captcha: %w", err),
		}, err
	}

	s.logger.Debug("reCAPTCHA submitted", zap.String("task_id", taskID))

	// Ожидание результата
	for i := 0; i < s.config.MaxAttempts; i++ {
		select {
		case <-ctx.Done():
			return &CaptchaResult{
				Success: false,
				Error:   ctx.Err(),
			}, ctx.Err()
		case <-time.After(s.config.PollInterval):
			result, err := s.getReCAPTCHAResult(ctx, taskID)
			if err != nil {
				s.logger.Warn("Failed to get captcha result", zap.Error(err))
				continue
			}

			if result != "" {
				s.logger.Info("reCAPTCHA solved", zap.String("token", result[:10]+"..."))
				return &CaptchaResult{
					Success: true,
					Token:   result,
				}, nil
			}
		}
	}

	return &CaptchaResult{
		Success: false,
		Error:   fmt.Errorf("max attempts reached"),
	}, fmt.Errorf("max attempts reached")
}

// SolveHCaptcha решает hCaptcha
func (s *CaptchaSolver) SolveHCaptcha(pageURL, siteKey string) (*CaptchaResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Timeout)
	defer cancel()

	s.logger.Debug("Solving hCaptcha",
		zap.String("page_url", pageURL),
		zap.String("site_key", siteKey))

	// Отправка капчи на решение
	taskID, err := s.submitHCaptcha(ctx, pageURL, siteKey)
	if err != nil {
		return &CaptchaResult{
			Success: false,
			Error:   fmt.Errorf("failed to submit captcha: %w", err),
		}, err
	}

	s.logger.Debug("hCaptcha submitted", zap.String("task_id", taskID))

	// Ожидание результата
	for i := 0; i < s.config.MaxAttempts; i++ {
		select {
		case <-ctx.Done():
			return &CaptchaResult{
				Success: false,
				Error:   ctx.Err(),
			}, ctx.Err()
		case <-time.After(s.config.PollInterval):
			result, err := s.getHCaptchaResult(ctx, taskID)
			if err != nil {
				s.logger.Warn("Failed to get captcha result", zap.Error(err))
				continue
			}

			if result != "" {
				s.logger.Info("hCaptcha solved", zap.String("token", result[:10]+"..."))
				return &CaptchaResult{
					Success: true,
					Token:   result,
				}, nil
			}
		}
	}

	return &CaptchaResult{
		Success: false,
		Error:   fmt.Errorf("max attempts reached"),
	}, fmt.Errorf("max attempts reached")
}

// submitReCAPTCHA отправляет reCAPTCHA на решение
func (s *CaptchaSolver) submitReCAPTCHA(ctx context.Context, pageURL, siteKey string) (string, error) {
	payload := map[string]interface{}{
		"key":      s.apiKey,
		"method":   "userrecaptcha",
		"googlekey": siteKey,
		"pageurl":  pageURL,
		"json":     1,
	}

	return s.submitTask(payload)
}

// submitHCaptcha отправляет hCaptcha на решение
func (s *CaptchaSolver) submitHCaptcha(ctx context.Context, pageURL, siteKey string) (string, error) {
	payload := map[string]interface{}{
		"key":       s.apiKey,
		"method":    "hcaptcha",
		"sitekey":   siteKey,
		"pageurl":   pageURL,
		"json":      1,
	}

	return s.submitTask(payload)
}

// submitTask отправляет задачу на решение
func (s *CaptchaSolver) submitTask(payload map[string]interface{}) (string, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := s.client.Post(
		fmt.Sprintf("%s/in.php", s.serviceURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if status, ok := result["status"].(float64); !ok || status != 1 {
		errMsg, _ := result["request"].(string)
		return "", fmt.Errorf("api error: %s", errMsg)
	}

	taskID, ok := result["request"].(string)
	if !ok {
		return "", fmt.Errorf("invalid task_id in response")
	}

	return taskID, nil
}

// getReCAPTCHAResult получает результат reCAPTCHA
func (s *CaptchaSolver) getReCAPTCHAResult(ctx context.Context, taskID string) (string, error) {
	return s.getResult(taskID)
}

// getHCaptchaResult получает результат hCaptcha
func (s *CaptchaSolver) getHCaptchaResult(ctx context.Context, taskID string) (string, error) {
	return s.getResult(taskID)
}

// getResult получает результат задачи
func (s *CaptchaSolver) getResult(taskID string) (string, error) {
	url := fmt.Sprintf("%s/res.php?key=%s&action=get&id=%s&json=1",
		s.serviceURL, s.apiKey, taskID)

	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if status, ok := result["status"].(float64); !ok || status != 1 {
		if errMsg, ok := result["request"].(string); ok && errMsg == "CAPCHA_NOT_READY" {
			return "", nil // Ещё не готово
		}
		return "", fmt.Errorf("api error: %s", result["request"])
	}

	token, ok := result["request"].(string)
	if !ok {
		return "", fmt.Errorf("invalid token in response")
	}

	return token, nil
}

// ReportBad сообщает о плохом решении капчи
func (s *CaptchaSolver) ReportBad(taskID string) error {
	payload := map[string]interface{}{
		"key":    s.apiKey,
		"action": "reportbad",
		"id":     taskID,
		"json":   1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := s.client.Post(
		fmt.Sprintf("%s/res.php", s.serviceURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if status, ok := result["status"].(float64); !ok || status != 1 {
		return fmt.Errorf("api error: %s", result["request"])
	}

	s.logger.Info("Bad captcha reported", zap.String("task_id", taskID))
	return nil
}

// GetBalance получает баланс сервиса
func (s *CaptchaSolver) GetBalance() (float64, error) {
	url := fmt.Sprintf("%s/res.php?key=%s&action=getbalance&json=1", s.serviceURL, s.apiKey)

	resp, err := s.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if status, ok := result["status"].(float64); !ok || status != 1 {
		return 0, fmt.Errorf("api error: %s", result["request"])
	}

	balance, ok := result["request"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid balance in response")
	}

	return balance, nil
}

// Close закрывает решатель капч
func (s *CaptchaSolver) Close() error {
	s.logger.Info("CaptchaSolver closed")
	return nil
}
