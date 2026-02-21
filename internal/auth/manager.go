// Package auth - Authentication & Authorization (Keycloak/Zitadel integration)
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Config конфигурация аутентификации
type Config struct {
	Enabled       bool   `json:"enabled"`
	Provider      string `json:"provider"` // keycloak, zitadel, internal
	ServerURL     string `json:"server_url"`
	Realm         string `json:"realm"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	RedirectURL   string `json:"redirect_url"`
	Scopes        []string `json:"scopes"`
	IssuerURL     string `json:"issuer_url"`
}

// User представляет пользователя
type User struct {
	ID            string   `json:"id"`
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Roles         []string `json:"roles"`
	Groups        []string `json:"groups"`
	Enabled       bool     `json:"enabled"`
	EmailVerified bool     `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	LastLogin     time.Time `json:"last_login"`
}

// Token представляет access токен
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// AuthManager менеджер аутентификации
type AuthManager struct {
	mu      sync.RWMutex
	config  *Config
	logger  *zap.Logger
	client  *http.Client
	cache   map[string]*Token
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Enabled:  false,
		Provider: "internal",
		Scopes:   []string{"openid", "profile", "email"},
	}
}

// NewAuthManager создает менеджер аутентификации
func NewAuthManager(config *Config, logger *zap.Logger) *AuthManager {
	if config == nil {
		config = DefaultConfig()
	}

	return &AuthManager{
		config: config,
		logger: logger,
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  make(map[string]*Token),
	}
}

// Login аутентифицирует пользователя
func (m *AuthManager) Login(ctx context.Context, username, password string) (*Token, error) {
	if !m.config.Enabled {
		return &Token{
			AccessToken:  "internal_token_" + username,
			TokenType:    "Bearer",
			RefreshToken: "refresh_" + username,
			ExpiresIn:    3600,
		}, nil
	}

	switch m.config.Provider {
	case "keycloak":
		return m.loginKeycloak(ctx, username, password)
	case "zitadel":
		return m.loginZitadel(ctx, username, password)
	default:
		return m.loginInternal(ctx, username, password)
	}
}

// loginKeycloak аутентификация через Keycloak
func (m *AuthManager) loginKeycloak(ctx context.Context, username, password string) (*Token, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		m.config.ServerURL, m.config.Realm)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", m.config.ClientID)
	data.Set("client_secret", m.config.ClientSecret)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", joinStrings(m.config.Scopes))

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = ioutil.NopCloser(nil)
	req.PostForm = data

	resp, err := m.client.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("keycloak login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("keycloak returned status %d", resp.StatusCode)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	m.cache[username] = &token

	m.logger.Info("Keycloak login successful",
		zap.String("username", username),
		zap.Int("expires_in", token.ExpiresIn))

	return &token, nil
}

// loginZitadel аутентификация через Zitadel
func (m *AuthManager) loginZitadel(ctx context.Context, username, password string) (*Token, error) {
	// Zitadel использует OIDC flow
	tokenURL := fmt.Sprintf("%s/oauth/v2/token", m.config.ServerURL)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", m.config.ClientID)
	data.Set("client_secret", m.config.ClientSecret)
	data.Set("username", username)
	data.Set("password", password)
	data.Set("scope", joinStrings(m.config.Scopes))

	resp, err := m.client.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("zitadel login failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zitadel returned status %d", resp.StatusCode)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	m.cache[username] = &token

	m.logger.Info("Zitadel login successful",
		zap.String("username", username))

	return &token, nil
}

// loginInternal внутренняя аутентификация (fallback)
func (m *AuthManager) loginInternal(ctx context.Context, username, password string) (*Token, error) {
	// Простая внутренняя аутентификация
	// В production: использовать bcrypt для паролей
	token := &Token{
		AccessToken:  fmt.Sprintf("internal_%s_%d", username, time.Now().Unix()),
		TokenType:    "Bearer",
		RefreshToken: fmt.Sprintf("refresh_%s", username),
		ExpiresIn:    3600,
		Scope:        "read write",
	}

	m.cache[username] = token

	m.logger.Info("Internal login successful",
		zap.String("username", username))

	return token, nil
}

// RefreshToken обновляет токен
func (m *AuthManager) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	if !m.config.Enabled {
		return &Token{
			AccessToken:  "refreshed_token",
			TokenType:    "Bearer",
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		}, nil
	}

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		m.config.ServerURL, m.config.Realm)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", m.config.ClientID)
	data.Set("client_secret", m.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	resp, err := m.client.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// ValidateToken валидирует токен
func (m *AuthManager) ValidateToken(ctx context.Context, token string) (bool, error) {
	if !m.config.Enabled {
		return true, nil
	}

	// Проверка в кэше
	for _, cached := range m.cache {
		if cached.AccessToken == token {
			return true, nil
		}
	}

	// Интроспекция токена
	introspectURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect",
		m.config.ServerURL, m.config.Realm)

	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", m.config.ClientID)
	data.Set("client_secret", m.config.ClientSecret)

	resp, err := m.client.PostForm(introspectURL, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	active, ok := result["active"].(bool)
	return ok && active, nil
}

// GetUser получает информацию о пользователе
func (m *AuthManager) GetUser(ctx context.Context, userID string) (*User, error) {
	if !m.config.Enabled {
		return &User{
			ID:       userID,
			Username: userID,
			Email:    userID + "@local",
			Roles:    []string{"user"},
			Enabled:  true,
		}, nil
	}

	userURL := fmt.Sprintf("%s/admin/realms/%s/users/%s",
		m.config.ServerURL, m.config.Realm, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// HasRole проверяет наличие роли
func (m *AuthManager) HasRole(ctx context.Context, userID, role string) (bool, error) {
	user, err := m.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, r := range user.Roles {
		if r == role {
			return true, nil
		}
	}

	return false, nil
}

// Logout разлогинивает пользователя
func (m *AuthManager) Logout(ctx context.Context, refreshToken string) error {
	if !m.config.Enabled {
		return nil
	}

	logoutURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout",
		m.config.ServerURL, m.config.Realm)

	data := url.Values{}
	data.Set("client_id", m.config.ClientID)
	data.Set("client_secret", m.config.ClientSecret)
	data.Set("refresh_token", refreshToken)

	_, err := m.client.PostForm(logoutURL, data)
	return err
}

// GetAuthURL возвращает URL для OAuth flow
func (m *AuthManager) GetAuthURL(state string) string {
	authURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth",
		m.config.ServerURL, m.config.Realm)

	params := url.Values{}
	params.Set("client_id", m.config.ClientID)
	params.Set("redirect_uri", m.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", joinStrings(m.config.Scopes))
	params.Set("state", state)

	return authURL + "?" + params.Encode()
}

// IsZeroTrustReady проверяет готовность Zero-Trust
func (m *AuthManager) IsZeroTrustReady() bool {
	return m.config.Enabled &&
		   (m.config.Provider == "keycloak" || m.config.Provider == "zitadel") &&
		   m.config.ServerURL != "" &&
		   m.config.ClientID != ""
}

// GetStats возвращает статистику
func (m *AuthManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"enabled":      m.config.Enabled,
		"provider":     m.config.Provider,
		"server_url":   m.config.ServerURL,
		"realm":        m.config.Realm,
		"cached_tokens": len(m.cache),
		"zero_trust_ready": m.IsZeroTrustReady(),
	}
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}
