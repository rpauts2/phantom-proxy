package credentialstuffing

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewEngine(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      true,
		RateLimit:    10,
		DelayBetween: 100 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)
	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.config.Enabled != true {
		t.Error("Expected engine to be enabled")
	}
}

func TestEngineDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      false,
		DelayBetween: 10 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)
	ctx := context.Background()

	service := TargetService{
		ID:       "test-service",
		Name:     "Test Service",
		LoginURL: "https://example.com/login",
		Enabled:  true,
	}

	// Тест CheckCredential с отключенным движком
	result, err := engine.CheckCredential(ctx, nil, service)
	if result != nil {
		t.Error("Expected nil result when engine is disabled")
	}
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
}

func TestCheckAllCredentialsDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      false,
		DelayBetween: 10 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)
	ctx := context.Background()

	results, err := engine.CheckAllCredentials(ctx)
	if results != nil {
		t.Error("Expected nil results when engine is disabled")
	}
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}
}

func TestTargetService(t *testing.T) {
	service := TargetService{
		ID:       "test-id",
		Name:     "Test Service",
		LoginURL: "https://example.com/login",
		Enabled:  true,
	}

	if service.ID != "test-id" {
		t.Errorf("Expected ID test-id, got %s", service.ID)
	}

	if service.Name != "Test Service" {
		t.Errorf("Expected name Test Service, got %s", service.Name)
	}
}

func TestStuffingResult(t *testing.T) {
	result := StuffingResult{
		CredentialID: "cred-123",
		ServiceID:    "service-456",
		Success:      true,
		StatusCode:   200,
		Error:        "",
		CheckedAt:    time.Now(),
	}

	if result.CredentialID != "cred-123" {
		t.Errorf("Expected CredentialID cred-123, got %s", result.CredentialID)
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.StatusCode != 200 {
		t.Errorf("Expected StatusCode 200, got %d", result.StatusCode)
	}
}

func TestAttackParams(t *testing.T) {
	params := AttackParams{
		UsernameField:   "email",
		PasswordField:   "password",
		SuccessIndicators: []string{"Dashboard", "Welcome", "success"},
		FailureIndicators: []string{"Invalid", "incorrect", "failed"},
	}

	if params.UsernameField != "email" {
		t.Errorf("Expected UsernameField email, got %s", params.UsernameField)
	}

	if params.PasswordField != "password" {
		t.Errorf("Expected PasswordField password, got %s", params.PasswordField)
	}

	if len(params.SuccessIndicators) != 3 {
		t.Errorf("Expected 3 success indicators, got %d", len(params.SuccessIndicators))
	}

	if len(params.FailureIndicators) != 3 {
		t.Errorf("Expected 3 failure indicators, got %d", len(params.FailureIndicators))
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		Enabled:        true,
		TargetServices: []TargetService{},
		RateLimit:      10,
		DelayBetween:   100 * time.Millisecond,
	}

	if !config.Enabled {
		t.Error("Expected config to be enabled")
	}

	if config.RateLimit != 10 {
		t.Errorf("Expected RateLimit 10, got %d", config.RateLimit)
	}

	if config.DelayBetween <= 0 {
		t.Error("Expected positive DelayBetween")
	}
}

func TestEngineHTTPClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      true,
		DelayBetween: 10 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)

	if engine.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if engine.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", engine.httpClient.Timeout)
	}
}

func TestEngineLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      true,
		DelayBetween: 10 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)

	if engine.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestRunAttackEmptyCredentials(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      true,
		DelayBetween: 1 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)
	ctx := context.Background()

	service := TargetService{
		ID:       "test",
		Name:     "Test",
		LoginURL: "https://example.com/login",
	}

	params := AttackParams{
		UsernameField: "email",
		PasswordField: "password",
	}

	// Пустые списки credentials
	results, err := engine.RunAttack(ctx, service, []string{}, []string{}, params)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestTestSingleCredential(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:      true,
		DelayBetween: 1 * time.Millisecond,
	}

	engine := NewEngine(config, nil, logger)
	ctx := context.Background()

	service := TargetService{
		ID:       "test",
		Name:     "Test",
		LoginURL: "https://example.com/login",
		Enabled:  true,
	}

	result, err := engine.TestSingleCredential(ctx, service, "test@example.com", "password123")
	
	// Ожидается что запрос будет выполнен (может вернуть ошибку соединения)
	if result == nil {
		t.Error("Expected result to be returned")
	}
	
	_ = err // Ошибка допустима (нет реального сервера)
}
