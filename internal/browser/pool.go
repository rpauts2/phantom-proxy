// Package browser provides browser automation for phishing simulations
//go:build ignore
// +build ignore

package browser

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/publicsuffix"
)

// BrowserPool manages browser sessions
type BrowserPool struct {
	mu       sync.RWMutex
	config   *Config
	logger   *zap.Logger
	client   *http.Client
	sessions map[string]*BrowserSession
}

// BrowserSession represents a browser session
type BrowserSession struct {
	ID        string
	Cookies   []*http.Cookie
	Headers   map[string]string
	CreatedAt time.Time
	LastUsed  time.Time
	UserAgent string
}

// Config browser pool configuration
type Config struct {
	Headless       bool
	PoolSize       int
	Timeout        time.Duration
	UserAgent      string
	MaxRedirects   int
	SkipTLSVerify  bool
}

// Response HTTP response
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	URL        string
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Headless:      true,
		PoolSize:      5,
		Timeout:       30 * time.Second,
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		MaxRedirects:  10,
		SkipTLSVerify: false,
	}
}

// NewPool creates new browser pool
func NewPool(config *Config, logger *zap.Logger) (*BrowserPool, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create cookie jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipTLSVerify,
			MinVersion:         tls.VersionTLS12,
		},
		MaxIdleConns:        config.PoolSize,
		MaxIdleConnsPerHost: config.PoolSize,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		},
	}

	p := &BrowserPool{
		config:   config,
		logger:   logger,
		client:   client,
		sessions: make(map[string]*BrowserSession),
	}

	logger.Info("Browser pool initialized",
		zap.Int("pool_size", config.PoolSize),
		zap.Bool("headless", config.Headless))

	return p, nil
}

// Execute performs HTTP request
func (p *BrowserPool) Execute(ctx context.Context, urlStr string, method string, headers map[string]string, body string) (*Response, error) {
	p.mu.Lock()

	// Get or create session
	sessionID := "default"
	session, exists := p.sessions[sessionID]
	if !exists {
		session = &BrowserSession{
			ID:        sessionID,
			Cookies:   make([]*http.Cookie, 0),
			Headers:   make(map[string]string),
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
			UserAgent: p.config.UserAgent,
		}
		p.sessions[sessionID] = session
	}
	session.LastUsed = time.Now()

	p.mu.Unlock()

	// Create request
	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequestWithContext(ctx, method, urlStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, urlStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	// Set headers
	req.Header.Set("User-Agent", session.UserAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for k, v := range session.Headers {
		if _, exists := headers[k]; !exists {
			req.Header.Set(k, v)
		}
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	buf := make([]byte, 1024*1024) // 1MB max
	n, _ := resp.Body.Read(buf)
	bodyStr := string(buf[:n])

	// Save cookies
	cookies := p.client.Jar.Cookies(req.URL)
	session.Cookies = append(session.Cookies, cookies...)

	// Build response headers
	respHeaders := make(map[string]string)
	for k := range resp.Header {
		respHeaders[k] = resp.Header.Get(k)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       bodyStr,
		URL:        resp.Request.URL.String(),
	}, nil
}

// GetSession returns session by ID
func (p *BrowserPool) GetSession(id string) (*BrowserSession, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	session, exists := p.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

// DeleteSession removes session
func (p *BrowserPool) DeleteSession(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}
	delete(p.sessions, id)
	return nil
}

// Cleanup removes old sessions
func (p *BrowserPool) Cleanup(maxAge time.Duration) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	count := 0
	now := time.Now()
	for id, session := range p.sessions {
		if now.Sub(session.LastUsed) > maxAge {
			delete(p.sessions, id)
			count++
		}
	}

	if count > 0 {
		p.logger.Debug("Cleaned up old sessions", zap.Int("count", count))
	}

	return count
}

// Close closes browser pool
func (p *BrowserPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sessions = make(map[string]*BrowserSession)
	p.logger.Info("Browser pool closed")
	return nil
}

// GetStats returns pool statistics
func (p *BrowserPool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"pool_size":     p.config.PoolSize,
		"active_sessions": len(p.sessions),
		"headless":      p.config.Headless,
		"timeout":       p.config.Timeout.String(),
	}
}
