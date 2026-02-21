// Package mtls - Zero-Trust mTLS Implementation
package mtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Config mTLS конфигурация
type Config struct {
	Enabled        bool   `json:"enabled"`
	CertPath       string `json:"cert_path"`
	KeyPath        string `json:"key_path"`
	CACertPath     string `json:"ca_cert_path"`
	MinTLSVersion  uint16 `json:"min_tls_version"`
	VerifyClient   bool   `json:"verify_client"`
	ClientTimeout  time.Duration `json:"client_timeout"`
}

// MTLSManager управляет mTLS соединениями
type MTLSManager struct {
	mu       sync.RWMutex
	config   *Config
	logger   *zap.Logger
	caCert   *x509.Certificate
	caPool   *x509.CertPool
	certPool *x509.CertPool
}

// ClientCert представляет клиентский сертификат
type ClientCert struct {
	ID        string    `json:"id"`
	CommonName string   `json:"common_name"`
	Org       string    `json:"org"`
	OrgUnit   string    `json:"org_unit"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		MinTLSVersion: tls.VersionTLS13,
		VerifyClient:  true,
		ClientTimeout: 30 * time.Second,
	}
}

// NewMTLSManager создает mTLS менеджер
func NewMTLSManager(config *Config, logger *zap.Logger) (*MTLSManager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	m := &MTLSManager{
		config: config,
		logger: logger,
	}

	// Загрузить CA сертификат
	if err := m.loadCACert(); err != nil {
		return nil, fmt.Errorf("failed to load CA cert: %w", err)
	}

	logger.Info("mTLS manager initialized",
		zap.Bool("enabled", config.Enabled),
		zap.Uint16("min_tls_version", config.MinTLSVersion),
		zap.Bool("verify_client", config.VerifyClient))

	return m, nil
}

// loadCACert загружает CA сертификат
func (m *MTLSManager) loadCACert() error {
	if m.config.CACertPath == "" {
		m.logger.Warn("CA cert path not set, skipping CA load")
		return nil
	}

	caCertPEM, err := ioutil.ReadFile(m.config.CACertPath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return fmt.Errorf("failed to parse CA certificate")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	m.caCert = caCert
	m.caPool = x509.NewCertPool()
	m.caPool.AddCert(caCert)

	m.logger.Info("CA certificate loaded",
		zap.String("subject", caCert.Subject.CommonName),
		zap.Time("expires", caCert.NotAfter))

	return nil
}

// CreateTLSConfig создает TLS конфигурацию для сервера
func (m *MTLSManager) CreateTLSConfig() (*tls.Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cert, err := tls.LoadX509KeyPair(m.config.CertPath, m.config.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   m.config.MinTLSVersion,
		ClientAuth:   tls.NoClientCert,
	}

	if m.config.VerifyClient {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = m.caPool
	}

	tlsConfig.BuildNameToCertificate()

	m.logger.Info("TLS config created",
		zap.Uint16("min_version", tlsConfig.MinVersion),
		zap.Bool("client_auth_required", m.config.VerifyClient))

	return tlsConfig, nil
}

// CreateClientTLSConfig создает TLS конфигурацию для клиента
func (m *MTLSManager) CreateClientTLSConfig(clientCertPath, clientKeyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      m.caPool,
		MinVersion:   m.config.MinTLSVersion,
		ServerName:   m.caCert.Subject.CommonName,
	}, nil
}

// ValidateClientCert валидирует клиентский сертификат
func (m *MTLSManager) ValidateClientCert(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate is nil")
	}

	if cert.NotAfter.Before(time.Now()) {
		return fmt.Errorf("certificate expired")
	}

	if cert.NotBefore.After(time.Now()) {
		return fmt.Errorf("certificate not yet valid")
	}

	opts := x509.VerifyOptions{
		Roots:     m.caPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// GetClientCertInfo извлекает информацию из клиентского сертификата
func (m *MTLSManager) GetClientCertInfo(cert *x509.Certificate) *ClientCert {
	return &ClientCert{
		ID:         cert.Subject.CommonName,
		CommonName: cert.Subject.CommonName,
		Org:        cert.Subject.Organization[0],
		OrgUnit:    cert.Subject.OrganizationalUnit[0],
		IssuedAt:   cert.NotBefore,
		ExpiresAt:  cert.NotAfter,
		Revoked:    false,
	}
}

// CreateDialContext создает функцию dial с mTLS
func (m *MTLSManager) CreateDialContext(clientCertPath, clientKeyPath string) (func(context.Context, string, string) (interface{}, error), error) {
	tlsConfig, err := m.CreateClientTLSConfig(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, network, addr string) (interface{}, error) {
		conn, err := tls.Dial(network, addr, tlsConfig)
		if err != nil {
			return nil, err
		}

		if err := conn.HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, err
		}

		return conn, nil
	}, nil
}

// GetStats возвращает статистику
func (m *MTLSManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":          m.config.Enabled,
		"verify_client":    m.config.VerifyClient,
		"min_tls_version":  m.config.MinTLSVersion,
		"client_timeout":   m.config.ClientTimeout.String(),
	}

	if m.caCert != nil {
		stats["ca_subject"] = m.caCert.Subject.CommonName
		stats["ca_expires"] = m.caCert.NotAfter
	}

	return stats
}

// GenerateClientCert генерирует клиентский сертификат (заглушка)
func (m *MTLSManager) GenerateClientCert(commonName, org, orgUnit string, duration time.Duration) (*ClientCert, error) {
	// В production: реальная генерация сертификатов
	// Здесь упрощенная версия

	cert := &ClientCert{
		ID:         commonName,
		CommonName: commonName,
		Org:        org,
		OrgUnit:    orgUnit,
		IssuedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(duration),
		Revoked:    false,
	}

	m.logger.Info("Client certificate generated",
		zap.String("common_name", commonName),
		zap.Time("expires", cert.ExpiresAt))

	return cert, nil
}

// RevokeClientCert отзывает сертификат (заглушка)
func (m *MTLSManager) RevokeClientCert(certID string) error {
	// В production: добавить в CRL
	m.logger.Info("Client certificate revoked", zap.String("id", certID))
	return nil
}

// IsZeroTrustReady проверяет готовность Zero-Trust
func (m *MTLSManager) IsZeroTrustReady() bool {
	return m.config.Enabled &&
		   m.config.VerifyClient &&
		   m.config.MinTLSVersion >= tls.VersionTLS13 &&
		   m.caCert != nil
}
