package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"

	"github.com/phantom-proxy/phantom-proxy/internal/config"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// HTTP3Proxy HTTP/3 QUIC прокси
type HTTP3Proxy struct {
	cfg        *config.Config
	db         *database.Database
	tlsManager TLSDialer
	logger     *zap.Logger
	server     *http3.Server
	handler    http.Handler
}

// NewHTTP3Proxy создаёт новый HTTP/3 прокси
func NewHTTP3Proxy(
	cfg *config.Config,
	db *database.Database,
	tlsManager TLSDialer,
	logger *zap.Logger,
) (*HTTP3Proxy, error) {
	
	// Создание HTTP/3 сервера
	server := &http3.Server{
		Addr: fmt.Sprintf("%s:%d", cfg.BindIP, cfg.HTTP3Port),
		Handler: nil, // Будет установлен позже
		QUICConfig: &quic.Config{
			MaxIdleTimeout:      30 * time.Second,
			KeepAlivePeriod:     10 * time.Second,
			MaxIncomingStreams:  1000,
			EnableDatagrams:     true,
			DisablePathMTUDiscovery: false,
		},
	}
	
	p := &HTTP3Proxy{
		cfg:        cfg,
		db:         db,
		tlsManager: tlsManager,
		logger:     logger,
		server:     server,
	}
	
	// Установка обработчика
	server.Handler = p
	
	return p, nil
}

// ServeHTTP обрабатывает HTTP запросы (реализация http.Handler)
func (p *HTTP3Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Логирование
	p.logger.Debug("HTTP/3 request",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("host", r.Host),
	)
	
	// Обработка через основной HTTP прокси
	// В реальной реализации здесь будет отдельная логика для HTTP/3
	// Пока делегируем основному прокси
	
	// Получение HTTP прокси из контекста или создание нового
	httpProxy, err := p.getHTTPProxy()
	if err != nil {
		p.logger.Error("Failed to get HTTP proxy", zap.Error(err))
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	
	httpProxy.ServeHTTP(w, r)
}

// getHTTPProxy получает или создаёт HTTP прокси для обработки
func (p *HTTP3Proxy) getHTTPProxy() (*HTTPProxy, error) {
	// В реальной реализации здесь будет пул HTTP прокси
	// Или общая логика между HTTP/2 и HTTP/3
	
	// Пока создаём новый (это неэффективно, но для MVP подойдёт)
	return NewHTTPProxy(p.cfg, p.db, p.tlsManager, p.logger)
}

// Start запускает HTTP/3 сервер
func (p *HTTP3Proxy) Start(ctx context.Context) error {
	// Graceful shutdown
	go func() {
		<-ctx.Done()
		p.server.Close()
	}()
	
	p.logger.Info("HTTP/3 Proxy starting", 
		zap.String("addr", p.server.Addr),
		zap.Bool("QUIC", true))
	
	// Запуск сервера
	// Для HTTP/3 нужен TLS сертификат
	if p.cfg.AutoCert {
		// Использование Let's Encrypt через certmagic
		return p.startWithAutoCert()
	} else {
		// Использование предоставленных сертификатов
		return p.server.ListenAndServeTLS(
			p.cfg.CertPath,
			p.cfg.KeyPath,
		)
	}
}

// startWithAutoCert запускает сервер с автоматическими сертификатами
func (p *HTTP3Proxy) startWithAutoCert() error {
	// TODO: Интеграция с certmagic для автоматических сертификатов
	// Пока заглушка
	p.logger.Warn("AutoCert not implemented, using provided certificates")
	
	return p.server.ListenAndServeTLS(
		p.cfg.CertPath,
		p.cfg.KeyPath,
	)
}

// GetStats возвращает статистику HTTP/3 прокси
func (p *HTTP3Proxy) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"protocol": "HTTP/3",
		"addr":     p.server.Addr,
		"quic":     true,
	}
}
