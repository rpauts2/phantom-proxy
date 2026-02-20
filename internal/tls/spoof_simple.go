package tls_spoof

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"
)

// SpoofManager управляет TLS соединениями
type SpoofManager struct {
	mu        sync.RWMutex
	cert      tls.Certificate
	templates map[string]*tls.Config
}

// Profile представляет TLS профиль
type Profile struct {
	ID       string
	Priority int
}

// NewSpoofManager создаёт новый SpoofManager
func NewSpoofManager() *SpoofManager {
	return &SpoofManager{
		templates: make(map[string]*tls.Config),
	}
}

// GenerateSelfSignedCert генерирует самоподписанный сертификат
func (sm *SpoofManager) GenerateSelfSignedCert(host string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"PhantomProxy"},
			CommonName:   host,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	sm.cert = tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}

	return nil
}

// LoadCertFromFile загружает сертификат из файла
func (sm *SpoofManager) LoadCertFromFile(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	sm.cert = cert
	return nil
}

// GetTLSConfig возвращает TLS конфигурацию
func (sm *SpoofManager) GetTLSConfig() *tls.Config {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return &tls.Config{
		Certificates: []tls.Certificate{sm.cert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2", "http/1.1"},
	}
}

// Dial создаёт TLS соединение
func (sm *SpoofManager) Dial(network, addr string) (net.Conn, error) {
	config := &tls.Config{
		ServerName: getServerName(addr),
		MinVersion: tls.VersionTLS12,
	}

	return tls.Dial(network, addr, config)
}

// GetJA3Fingerprint возвращает JA3 отпечаток
func (sm *SpoofManager) GetJA3Fingerprint() string {
	// Упрощённая версия
	return "simple_tls"
}

// GetStats возвращает статистику
func (sm *SpoofManager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"type": "simple",
	}
}

func getServerName(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

// CalculateJA3 вычисляет JA3 fingerprint
func CalculateJA3(data []byte) string {
	return hex.EncodeToString(data[:16])
}
