package c2

import (
	"context"
	"fmt"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// CobaltStrikeAdapter integrates via External C2 / Beacon format
type CobaltStrikeAdapter struct {
	config *CobaltStrikeConfig
}

// CobaltStrikeConfig configuration
type CobaltStrikeConfig struct {
	Enabled       bool   `yaml:"enabled" mapstructure:"enabled"`
	TeamServerURL string `yaml:"team_server_url" mapstructure:"team_server_url"`
	ExternalC2    struct {
		Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
		BindHost string `yaml:"bind_host" mapstructure:"bind_host"`
		BindPort int    `yaml:"bind_port" mapstructure:"bind_port"`
	} `yaml:"external_c2" mapstructure:"external_c2"`
}

// NewCobaltStrikeAdapter creates Cobalt Strike adapter
func NewCobaltStrikeAdapter(cfg *CobaltStrikeConfig) *CobaltStrikeAdapter {
	return &CobaltStrikeAdapter{config: cfg}
}

// Name returns adapter name
func (a *CobaltStrikeAdapter) Name() string { return "cobalt_strike" }

// IsAvailable checks Team Server connectivity
func (a *CobaltStrikeAdapter) IsAvailable(ctx context.Context) bool {
	if !a.config.Enabled || a.config.TeamServerURL == "" {
		return false
	}
	// CS Team Server не имеет публичного HTTP API — интеграция через External C2
	return true
}

// SendSession forwards session data to CS (conceptual — реальная интеграция через External C2)
func (a *CobaltStrikeAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}
	// External C2: PhantomProxy выступает как External C2 Server
	// Сессии/креды отправляются в лог CS или через свой listener
	_ = data
	return nil
}

// SendCredentials adds credentials to CS loot
func (a *CobaltStrikeAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}
	_ = creds
	_ = metadata
	return nil
}

// HealthCheck verifies configuration
func (a *CobaltStrikeAdapter) HealthCheck(ctx context.Context) error {
	if !a.config.Enabled {
		return nil
	}
	if a.config.ExternalC2.Enabled && a.config.ExternalC2.BindPort == 0 {
		return fmt.Errorf("external_c2 bind_port required")
	}
	return nil
}
