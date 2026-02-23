package domain

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestNewCloudflareProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Без credentials - mock режим
	provider, err := NewCloudflareProvider("", "", logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
}

func TestCloudflareProviderAddRecord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewCloudflareProvider("", "", logger)
	ctx := context.Background()

	err := provider.AddRecord(ctx, "example.com", "A", "192.168.1.1")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}
}

func TestCloudflareProviderDeleteRecord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewCloudflareProvider("", "", logger)
	ctx := context.Background()

	err := provider.DeleteRecord(ctx, "example.com", "A")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}
}

func TestCloudflareProviderValidate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewCloudflareProvider("", "", logger)
	ctx := context.Background()

	valid := provider.Validate(ctx, "example.com")
	if !valid {
		t.Error("Expected validation to return true")
	}
}

func TestNewNamecheapProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Без credentials - mock режим
	provider, err := NewNamecheapProvider("", "", "", logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
}

func TestNamecheapProviderAddRecord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewNamecheapProvider("", "", "", logger)
	ctx := context.Background()

	err := provider.AddRecord(ctx, "example.com", "A", "192.168.1.1")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}
}

func TestNamecheapProviderDeleteRecord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewNamecheapProvider("", "", "", logger)
	ctx := context.Background()

	err := provider.DeleteRecord(ctx, "example.com", "A")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}
}

func TestNamecheapProviderValidate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewNamecheapProvider("", "", "", logger)
	ctx := context.Background()

	valid := provider.Validate(ctx, "example.com")
	if !valid {
		t.Error("Expected validation to return true")
	}
}

func TestNewRoute53Provider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Без credentials - mock режим
	provider, err := NewRoute53Provider("", "", "", logger)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
}

func TestRoute53ProviderAddRecord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewRoute53Provider("", "", "", logger)
	ctx := context.Background()

	err := provider.AddRecord(ctx, "example.com", "A", "192.168.1.1")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}
}

func TestRoute53ProviderDeleteRecord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewRoute53Provider("", "", "", logger)
	ctx := context.Background()

	err := provider.DeleteRecord(ctx, "example.com", "A")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}
}

func TestRoute53ProviderValidate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	provider, _ := NewRoute53Provider("", "", "", logger)
	ctx := context.Background()

	valid := provider.Validate(ctx, "example.com")
	if !valid {
		t.Error("Expected validation to return true")
	}
}

func TestGetRootDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"example.com", "example.com"},
		{"sub.example.com", "example.com"},
		{"a.b.c.example.com", "example.com"},
		{"test.co.uk", "co.uk"},
		{"localhost", "localhost"},
	}

	for _, tt := range tests {
		result := getRootDomain(tt.domain)
		if result != tt.expected {
			t.Errorf("getRootDomain(%s) = %s, expected %s", tt.domain, result, tt.expected)
		}
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	config := LoadConfigFromEnv()

	expectedKeys := []string{
		"cloudflare_api_key",
		"cloudflare_email",
		"namecheap_api_key",
		"namecheap_api_user",
		"namecheap_client_ip",
		"aws_access_key",
		"aws_secret_key",
		"aws_region",
	}

	for _, key := range expectedKeys {
		if _, ok := config[key]; !ok {
			t.Errorf("Expected key %s in config", key)
		}
	}
}

func TestDebugProviders(t *testing.T) {
	debugInfo := DebugProviders()

	// Должна возвращать строку с информацией о провайдерах
	if debugInfo == "" {
		t.Error("Expected debug info to be returned")
	}
}
