package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpFile := "test_config.yaml"
	configContent := `
bind_ip: "0.0.0.0"
https_port: 8443
domain: "test.local"
cert_path: "./certs/cert.pem"
key_path: "./certs/key.pem"
database_path: "./test.db"
database_type: "sqlite"
phishlets_path: "./configs/phishlets"
api_enabled: true
api_port: 8080
api_key: "test-api-key"
debug: false
`

	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.BindIP != "0.0.0.0" {
		t.Errorf("Expected BindIP 0.0.0.0, got %s", cfg.BindIP)
	}

	if cfg.HTTPSPort != 8443 {
		t.Errorf("Expected HTTPSPort 8443, got %d", cfg.HTTPSPort)
	}

	if cfg.Domain != "test.local" {
		t.Errorf("Expected Domain test.local, got %s", cfg.Domain)
	}

	if !cfg.APIEnabled {
		t.Error("Expected APIEnabled to be true")
	}

	if cfg.APIPort != 8080 {
		t.Errorf("Expected APIPort 8080, got %d", cfg.APIPort)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	tmpFile := "test_config_minimal.yaml"
	configContent := `
bind_ip: "0.0.0.0"
https_port: 443
domain: "example.com"
`

	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if cfg.BindIP != "0.0.0.0" {
		t.Errorf("Expected BindIP 0.0.0.0, got %s", cfg.BindIP)
	}

	if cfg.Debug != false {
		t.Error("Expected Debug to be false by default")
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	_, err := Load("nonexistent_config.yaml")
	if err == nil {
		t.Error("Expected error when loading nonexistent config")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpFile := "test_save_config.yaml"
	defer os.Remove(tmpFile)

	cfg := &Config{
		BindIP:       "127.0.0.1",
		HTTPSPort:    9443,
		Domain:       "save-test.local",
		CertPath:     "./certs/cert.pem",
		KeyPath:      "./certs/key.pem",
		DatabasePath: "./test.db",
		APIEnabled:   true,
		APIPort:      9090,
		APIKey:       "save-test-key",
		Debug:        true,
	}

	err := cfg.Save(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load and verify
	loaded, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.BindIP != cfg.BindIP {
		t.Errorf("Expected BindIP %s, got %s", cfg.BindIP, loaded.BindIP)
	}

	if loaded.HTTPSPort != cfg.HTTPSPort {
		t.Errorf("Expected HTTPSPort %d, got %d", cfg.HTTPSPort, loaded.HTTPSPort)
	}

	if loaded.Domain != cfg.Domain {
		t.Errorf("Expected Domain %s, got %s", cfg.Domain, loaded.Domain)
	}

	if loaded.Debug != cfg.Debug {
		t.Errorf("Expected Debug %v, got %v", cfg.Debug, loaded.Debug)
	}
}

func TestConfigValidation(t *testing.T) {
	tmpFile := "test_invalid_config.yaml"
	configContent := `
bind_ip: "invalid-ip"
https_port: -1
domain: ""
`

	err := os.WriteFile(tmpFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Should either fail to load or have validation errors
	cfg, err := Load(tmpFile)
	if err == nil {
		// If it loaded, check if values are invalid
		if cfg.BindIP == "invalid-ip" {
			t.Error("Expected validation to reject invalid IP")
		}
	}
}
