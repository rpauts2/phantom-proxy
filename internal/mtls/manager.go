// Package mtls - Zero-Trust mTLS Manager
// Implements SPIFFE-style mutual TLS between services
package mtls

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ============================================================================
// Configuration
// ============================================================================

type Config struct {
	// Server configuration
	CertPath       string `yaml:"cert_path" env:"MTLS_CERT_PATH"`
	KeyPath        string `yaml:"key_path" env:"MTLS_KEY_PATH"`
	CAcertPath     string `yaml:"ca_cert_path" env:"MTLS_CA_CERT_PATH"`
	CAkeyPath      string `yaml:"ca_key_path" env:"MTLS_CA_KEY_PATH"`
	
	// Auto-generation
	AutoGenerate   bool   `yaml:"auto_generate" env:"MTLS_AUTO_GENERATE"`
	CertValidity   time.Duration `yaml:"cert_validity" env:"MTLS_CERT_VALIDITY"`
	
	// Verification
	VerifyClient  bool   `yaml:"verify_client" env:"MTLS_VERIFY_CLIENT"`
	MinTLSVersion uint16 `yaml:"min_tls_version" env:"MTLS_MIN_TLS_VERSION"`
	
	// SPIFFE
	SpiffeEnabled bool   `yaml:"spiffe_enabled" env:"MTLS_SPIFFE_ENABLED"`
	TrustDomain   string `yaml:"trust_domain" env:"MTLS_TRUST_DOMAIN"`
	ServiceName   string `yaml:"service_name" env:"MTLS_SERVICE_NAME"`
}

func DefaultConfig() *Config {
	return &Config{
		AutoGenerate:   true,
		CertValidity:   24 * time.Hour,
		VerifyClient:  true,
		MinTLSVersion:  tls.VersionTLS12,
		SpiffeEnabled:  true,
		TrustDomain:   "phantom.local",
	}
}

// ============================================================================
// Certificate Authority
// ============================================================================

type CA struct {
	mu         sync.RWMutex
	cert       *x509.Certificate
	privateKey *rsa.PrivateKey
	certPEM    []byte
	keyPEM     []byte
	config     *Config
	logger     *zap.Logger
}

func NewCA(config *Config, logger *zap.Logger) (*CA, error) {
	ca := &CA{
		config: config,
		logger: logger,
	}

	if config.AutoGenerate {
		if err := ca.generateCA(); err != nil {
			return nil, fmt.Errorf("failed to generate CA: %w", err)
		}
	} else {
		if err := ca.loadCA(); err != nil {
			return nil, fmt.Errorf("failed to load CA: %w", err)
		}
	}

	return ca, nil
}

func (ca *CA) generateCA() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"PhantomProxy CA"},
			CommonName:   "PhantomProxy Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return err
	}

	ca.cert = cert
	ca.privateKey = privateKey
	ca.certPEM = encodeCertPEM(certDER)
	ca.keyPEM = encodeKeyPEM(privateKey)

	ca.logger.Info("CA generated", zap.String("subject", cert.Subject.CommonName))

	// Save to files
	if ca.config.CAcertPath != "" {
		os.WriteFile(ca.config.CAcertPath, ca.certPEM, 0644)
	}
	if ca.config.CAkeyPath != "" {
		os.WriteFile(ca.config.CAkeyPath, ca.keyPEM, 0600)
	}

	return nil
}

func (ca *CA) loadCA() error {
	certPEM, err := os.ReadFile(ca.config.CAcertPath)
	if err != nil {
		return err
	}

	keyPEM, err := os.ReadFile(ca.config.CAkeyPath)
	if err != nil {
		return err
	}

	cert, err := parseCertPEM(certPEM)
	if err != nil {
		return err
	}

	key, err := parseKeyPEM(keyPEM)
	if err != nil {
		return err
	}

	ca.cert = cert
	ca.privateKey = key
	ca.certPEM = certPEM
	ca.keyPEM = keyPEM

	return nil
}

// ============================================================================
// Certificate Generation
// ============================================================================

func (ca *CA) GenerateServiceCert(serviceName string, sans []string) ([]byte, []byte, error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumber := big.NewInt(time.Now().UnixNano())

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"PhantomProxy"},
			CommonName:   serviceName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(ca.config.CertValidity),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:                 sans,
	}

	// Add SPIFFE SAN if enabled
	if ca.config.SpiffeEnabled {
		spiffeID := fmt.Sprintf("spiffe://%s/ns/default/sa/%s", ca.config.TrustDomain, serviceName)
		template.DNSNames = append(template.DNSNames, spiffeID)
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, ca.cert, &privateKey.PublicKey, ca.privateKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := encodeCertPEM(certDER)
	keyPEM := encodeKeyPEM(privateKey)

	ca.logger.Info("Service certificate generated",
		zap.String("service", serviceName),
		zap.Strings("sans", sans))

	return certPEM, keyPEM, nil
}

// ============================================================================
// TLS Config Generation
// ============================================================================

func (ca *CA) GetServerTLSConfig() (*tls.Config, error) {
	certPEM, keyPEM, err := ca.GenerateServiceCert(ca.config.ServiceName, []string{
		ca.config.ServiceName,
		"localhost",
		"127.0.0.1",
	})
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca.certPEM)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequestClientCert,
		MinVersion:   ca.config.MinTLSVersion,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}, nil
}

func (ca *CA) GetClientTLSConfig() (*tls.Config, error) {
	certPEM, keyPEM, err := ca.GenerateServiceCert(ca.config.ServiceName, nil)
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca.certPEM)

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:           pool,
		MinVersion:        ca.config.MinTLSVersion,
		InsecureSkipVerify: false,
	}, nil
}

// ============================================================================
// SPIFFE Workload API (Simplified)
// ============================================================================

type WorkloadAPI struct {
	ca      *CA
	logger  *zap.Logger
	entries map[string]*SpiffeEntry
	mu      sync.RWMutex
}

type SpiffeEntry struct {
	SpiffeID     string
	ServiceName  string
	TrustDomain  string
	CertPEM      []byte
	KeyPEM       []byte
	CertChainPEM []byte
	ExpiresAt    time.Time
}

func NewWorkloadAPI(ca *CA, logger *zap.Logger) *WorkloadAPI {
	return &WorkloadAPI{
		ca:      ca,
		logger:  logger,
		entries: make(map[string]*SpiffeEntry),
	}
}

func (w *WorkloadAPI) FetchX509SVID(ctx context.Context, spiffeID string) (*SpiffeEntry, error) {
	w.mu.RLock()
	if entry, ok := w.entries[spiffeID]; ok {
		w.mu.RUnlock()
		if time.Now().Before(entry.ExpiresAt) {
			return entry, nil
		}
	}
	w.mu.RUnlock()

	// Generate new SVID
	serviceName := extractServiceName(spiffeID)
	certPEM, keyPEM, err := w.ca.GenerateServiceCert(serviceName, []string{spiffeID})
	if err != nil {
		return nil, err
	}

	entry := &SpiffeEntry{
		SpiffeID:     spiffeID,
		ServiceName:  serviceName,
		TrustDomain:  w.ca.config.TrustDomain,
		CertPEM:      certPEM,
		KeyPEM:       keyPEM,
		CertChainPEM: w.ca.certPEM,
		ExpiresAt:    time.Now().Add(w.ca.config.CertValidity),
	}

	w.mu.Lock()
	w.entries[spiffeID] = entry
	w.mu.Unlock()

	w.logger.Info("X.509 SVID issued", zap.String("spiffe_id", spiffeID))

	return entry, nil
}

func extractServiceName(spiffeID string) string {
	// spiffe://trust-domain/ns/default/sa/service-name
	parts := split(spiffeID, "/")
	if len(parts) >= 5 {
		return parts[4]
	}
	return "unknown"
}

func split(s, sep string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if string(c) == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// ============================================================================
// Manager
// ============================================================================

type Manager struct {
	config   *Config
	logger   *zap.Logger
	ca       *CA
	workload *WorkloadAPI
}

func NewManager(config *Config, logger *zap.Logger) (*Manager, error) {
	ca, err := NewCA(config, logger)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		config:   config,
		logger:   logger,
		ca:       ca,
		workload: NewWorkloadAPI(ca, logger),
	}

	logger.Info("mTLS Manager initialized",
		zap.Bool("spiffe", config.SpiffeEnabled),
		zap.String("trust_domain", config.TrustDomain))

	return m, nil
}

func (m *Manager) GetServerTLSConfig() (*tls.Config, error) {
	return m.ca.GetServerTLSConfig()
}

func (m *Manager) GetClientTLSConfig() (*tls.Config, error) {
	return m.ca.GetClientTLSConfig()
}

func (m *Manager) GetWorkloadAPI() *WorkloadAPI {
	return m.workload
}

func (m *Manager) VerifyPeerCert(connState tls.ConnectionState) error {
	if len(connState.PeerCertificates) == 0 {
		return fmt.Errorf("no peer certificates")
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		Roots:         x509.NewCertPool(),
		Intermediates: x509.NewCertPool(),
	}

	opts.Roots.AppendCertsFromPEM(m.ca.certPEM)

	for i, cert := range connState.PeerCertificates {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}

	_, err := connState.PeerCertificates[0].Verify(opts)
	return err
}

// ============================================================================
// Helper Functions
// ============================================================================

func encodeCertPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func encodeKeyPEM(key *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func parseCertPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM data found")
	}
	return x509.ParseCertificate(block.Bytes)
}

func parseKeyPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM data found")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// Ensure directory exists
func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// LoadOrGenerate loads existing certificates or generates new ones
func LoadOrGenerate(config *Config, logger *zap.Logger, serviceName string) (*tls.Config, *tls.Config, error) {
	mgr, err := NewManager(config, logger)
	if err != nil {
		return nil, nil, err
	}

	serverConfig, err := mgr.GetServerTLSConfig()
	if err != nil {
		return nil, nil, err
	}

	clientConfig, err := mgr.GetClientTLSConfig()
	if err != nil {
		return nil, nil, err
	}

	// Save certs
	if config.CertPath != "" {
		ensureDir(filepath.Dir(config.CertPath))
	}

	return serverConfig, clientConfig, nil
}
