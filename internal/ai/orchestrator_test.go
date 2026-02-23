package ai

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Endpoint != "http://localhost:8081" {
		t.Errorf("Expected endpoint http://localhost:8081, got %s", config.Endpoint)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}

	if config.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", config.Temperature)
	}

	if config.MaxTokens != 2048 {
		t.Errorf("Expected max tokens 2048, got %d", config.MaxTokens)
	}

	if config.Model != "llama-3.1-70b" {
		t.Errorf("Expected model llama-3.1-70b, got %s", config.Model)
	}
}

func TestNewAIOrchestrator(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("http://test-ai:8081", logger)

	if orchestrator == nil {
		t.Fatal("Expected orchestrator to be created")
	}

	if orchestrator.endpoint != "http://test-ai:8081" {
		t.Errorf("Expected endpoint http://test-ai:8081, got %s", orchestrator.endpoint)
	}
}

func TestNewAIOrchestratorDefault(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	if orchestrator == nil {
		t.Fatal("Expected orchestrator to be created")
	}

	// Пустой endpoint остается пустым, default используется только для config
	if orchestrator.endpoint != "" {
		t.Errorf("Expected empty endpoint, got %s", orchestrator.endpoint)
	}

	// Но config должен иметь default
	if orchestrator.config.Endpoint != "http://localhost:8081" {
		t.Errorf("Expected default endpoint in config http://localhost:8081, got %s", orchestrator.config.Endpoint)
	}
}

func TestLocalGeneratePhishing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	prompt := "Generate a phishing email to verify account"
	result := orchestrator.localGenerate(prompt)

	if result == "" {
		t.Error("Expected generated text")
	}

	// Проверка что содержит ключевые слова phishing шаблона
	if !containsAny(result, []string{"verify", "security", "account", "unusual activity"}) {
		t.Error("Expected phishing template content")
	}
}

func TestLocalGenerateInvoice(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	prompt := "Generate an invoice notification email"
	result := orchestrator.localGenerate(prompt)

	if result == "" {
		t.Error("Expected generated text")
	}

	if !containsAny(result, []string{"invoice", "amount", "payment", "review"}) {
		t.Error("Expected invoice template content")
	}
}

func TestLocalGenerateMeeting(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	prompt := "Generate a meeting invitation"
	result := orchestrator.localGenerate(prompt)

	if result == "" {
		t.Error("Expected generated text")
	}

	if !containsAny(result, []string{"meeting", "invited", "topic", "time"}) {
		t.Error("Expected meeting template content")
	}
}

func TestLocalGeneratePasswordReset(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	prompt := "Generate a password reset email"
	result := orchestrator.localGenerate(prompt)

	if result == "" {
		t.Error("Expected generated text")
	}

	if !containsAny(result, []string{"password", "reset", "account", "support"}) {
		t.Error("Expected password reset template content")
	}
}

func TestLocalGenerateDefault(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	prompt := "Some random prompt without keywords"
	result := orchestrator.localGenerate(prompt)

	if result == "" {
		t.Error("Expected generated text")
	}

	if !containsAny(result, []string{"Dear User", "important notification", "action"}) {
		t.Error("Expected default template content")
	}
}

func TestGenerateEmail(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)
	ctx := context.Background()

	targetData := map[string]interface{}{
		"name":     "John Doe",
		"company":  "Acme Corp",
		"position": "Manager",
		"email":    "john@example.com",
	}

	template := "phishing"
	result, err := orchestrator.GenerateEmail(ctx, targetData, template)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == "" {
		t.Error("Expected generated email")
	}
}

func TestGenerateSubject(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)
	ctx := context.Background()

	emailBody := "Dear customer, please verify your account immediately"
	result, err := orchestrator.GenerateSubject(ctx, emailBody)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == "" {
		t.Error("Expected generated subject")
	}
}

func TestGeneratePersonalizedContent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)
	ctx := context.Background()

	targetData := map[string]interface{}{
		"name":    "Jane Smith",
		"company": "Tech Inc",
	}

	// Используем GenerateEmail вместо несуществующего метода
	result, err := orchestrator.GenerateEmail(ctx, targetData, "greeting")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == "" {
		t.Error("Expected generated content")
	}
}

func TestHealthCheck(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Тест с недоступным endpoint - должен вернуть false
	orchestrator := NewAIOrchestrator("http://invalid-host:9999", logger)
	ctx := context.Background()

	healthy := orchestrator.HealthCheck(ctx)
	if healthy {
		t.Error("Expected health check to fail for invalid host")
	}
}

func TestGetModels(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Используем пустой endpoint
	orchestrator := NewAIOrchestrator("", logger)
	ctx := context.Background()

	// С пустым endpoint должна вернуться ошибка
	_, err := orchestrator.GetModels(ctx)

	// Ожидаем ошибку (нет сервера)
	if err == nil {
		t.Error("Expected error when endpoint is empty")
	}
}

func TestSetModel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	orchestrator.SetModel("test-model-123")

	if orchestrator.model != "test-model-123" {
		t.Errorf("Expected model test-model-123, got %s", orchestrator.model)
	}
}

func TestGetStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	stats := orchestrator.GetStats()

	if stats == nil {
		t.Fatal("Expected stats to be returned")
	}

	expectedKeys := []string{"endpoint", "model", "timeout", "max_retries", "temperature", "max_tokens"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Expected key %s in stats", key)
		}
	}
}

func TestClose(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("", logger)

	err := orchestrator.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}
}

func TestGenerateTextUsesLocal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// orchestrator с localhost endpoint должен использовать local генерацию
	orchestrator := NewAIOrchestrator("http://localhost:8081", logger)
	ctx := context.Background()

	result, err := orchestrator.generateText(ctx, "test prompt")

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == "" {
		t.Error("Expected generated text")
	}
}

func TestCallExternalAIFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	orchestrator := NewAIOrchestrator("http://invalid-host:9999", logger)
	ctx := context.Background()

	_, err := orchestrator.callExternalAI(ctx, "test prompt")

	if err == nil {
		t.Error("Expected error for invalid host")
	}
}

// Helper functions
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if contains(text, keyword) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
