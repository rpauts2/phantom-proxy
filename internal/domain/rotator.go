// Package domain provides domain rotation and SSL certificate management
package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DomainRotator автоматическая ротация доменов
type DomainRotator struct {
	mu              sync.RWMutex
	logger          *zap.Logger
	config          *RotatorConfig
	domains         []string
	currentDomain   string
	lastRotation    time.Time
	certificates    map[string]*CertificateInfo
	dnsProviders    map[string]DNSProvider
	autoRenewBefore int // часов до истечения
}

// CertificateInfo информация о SSL сертификате
type CertificateInfo struct {
	Domain     string    `json:"domain"`
	IssuedAt   time.Time `json:"issued_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	AutoRenew  bool      `json:"auto_renew"`
	DNSStatus  string    `json:"dns_status"`
	SSLStatus  string    `json:"ssl_status"`
	CertPath   string    `json:"cert_path"`
	KeyPath    string    `json:"key_path"`
}

// RotatorConfig конфигурация ротатора
type RotatorConfig struct {
	AutoRenew        bool          `json:"auto_renew"`
	AutoRenewBefore  int           `json:"auto_renew_before"` // часов
	RotationInterval time.Duration `json:"rotation_interval"`
	DNSProvider      string        `json:"dns_provider"`
	Email            string        `json:"email"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *RotatorConfig {
	return &RotatorConfig{
		AutoRenew:       true,
		AutoRenewBefore: 48, // 48 часов до истечения
		RotationInterval: 24 * time.Hour,
		DNSProvider:     "cloudflare",
		Email:          "admin@example.com",
	}
}

// NewDomainRotator создает новый ротатор доменов
func NewDomainRotator(config *RotatorConfig, logger *zap.Logger) (*DomainRotator, error) {
	if config == nil {
		config = DefaultConfig()
	}

	r := &DomainRotator{
		logger:          logger,
		config:          config,
		domains:         make([]string, 0),
		certificates:    make(map[string]*CertificateInfo),
		dnsProviders:    make(map[string]DNSProvider),
		autoRenewBefore: config.AutoRenewBefore,
	}

	// Регистрация DNS провайдеров
	r.dnsProviders["cloudflare"] = &CloudflareProvider{}
	r.dnsProviders["namecheap"] = &NamecheapProvider{}
	r.dnsProviders["route53"] = &Route53Provider{}

	logger.Info("Domain rotator initialized",
		zap.Bool("auto_renew", config.AutoRenew),
		zap.Duration("rotation_interval", config.RotationInterval))

	return r, nil
}

// AddDomain добавляет домен
func (r *DomainRotator) AddDomain(ctx context.Context, domain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверка на дубликаты
	for _, d := range r.domains {
		if d == domain {
			return fmt.Errorf("domain already exists: %s", domain)
		}
	}

	r.domains = append(r.domains, domain)

	// Создание информации о сертификате
	certInfo := &CertificateInfo{
		Domain:    domain,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(90 * 24 * time.Hour), // 90 дней
		AutoRenew: r.config.AutoRenew,
		DNSStatus: "pending",
		SSLStatus: "pending",
	}

	r.certificates[domain] = certInfo

	// Валидация DNS
	if provider, ok := r.dnsProviders[r.config.DNSProvider]; ok {
		if provider.Validate(ctx, domain) {
			certInfo.DNSStatus = "valid"
		} else {
			certInfo.DNSStatus = "invalid"
		}
	}

	// Если это первый домен, делаем его текущим
	if len(r.domains) == 1 {
		r.currentDomain = domain
		certInfo.SSLStatus = "active"
	} else {
		certInfo.SSLStatus = "standby"
	}

	r.logger.Info("Domain added",
		zap.String("domain", domain),
		zap.String("dns_status", certInfo.DNSStatus),
		zap.String("ssl_status", certInfo.SSLStatus))

	return nil
}

// RemoveDomain удаляет домен
func (r *DomainRotator) RemoveDomain(ctx context.Context, domain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Поиск и удаление домена
	found := false
	for i, d := range r.domains {
		if d == domain {
			r.domains = append(r.domains[:i], r.domains[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("domain not found: %s", domain)
	}

	// Удаление сертификата
	delete(r.certificates, domain)

	// Если удалили текущий домен, выбираем новый
	if r.currentDomain == domain {
		if len(r.domains) > 0 {
			r.currentDomain = r.domains[0]
			if cert, ok := r.certificates[r.currentDomain]; ok {
				cert.SSLStatus = "active"
			}
		} else {
			r.currentDomain = ""
		}
	}

	r.logger.Info("Domain removed", zap.String("domain", domain))
	return nil
}

// Rotate переключается на следующий домен
func (r *DomainRotator) Rotate() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.domains) == 0 {
		return "", fmt.Errorf("no domains configured")
	}

	// Поиск текущего индекса
	currentIdx := -1
	for i, d := range r.domains {
		if d == r.currentDomain {
			currentIdx = i
			break
		}
	}

	// Переключение на следующий
	nextIdx := (currentIdx + 1) % len(r.domains)
	r.currentDomain = r.domains[nextIdx]
	r.lastRotation = time.Now()

	// Обновление статусов
	for _, d := range r.domains {
		if cert, ok := r.certificates[d]; ok {
			if d == r.currentDomain {
				cert.SSLStatus = "active"
			} else {
				cert.SSLStatus = "standby"
			}
		}
	}

	r.logger.Info("Domain rotated",
		zap.String("new_domain", r.currentDomain),
		zap.Int("total_domains", len(r.domains)))

	return r.currentDomain, nil
}

// GetCurrentDomain возвращает текущий домен
func (r *DomainRotator) GetCurrentDomain() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentDomain
}

// GetNextDomain возвращает следующий домен без переключения
func (r *DomainRotator) GetNextDomain() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.domains) == 0 {
		return ""
	}

	currentIdx := -1
	for i, d := range r.domains {
		if d == r.currentDomain {
			currentIdx = i
			break
		}
	}

	nextIdx := (currentIdx + 1) % len(r.domains)
	return r.domains[nextIdx]
}

// GetAllDomains возвращает все домены
func (r *DomainRotator) GetAllDomains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domains := make([]string, len(r.domains))
	copy(domains, r.domains)
	return domains
}

// GetDomainInfo получает информацию о домене
func (r *DomainRotator) GetDomainInfo(domain string) (*CertificateInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cert, ok := r.certificates[domain]
	if !ok {
		return nil, fmt.Errorf("domain not found: %s", domain)
	}

	// Проверка SSL
	now := time.Now()
	if now.Before(cert.ExpiresAt.Add(-time.Duration(r.autoRenewBefore) * time.Hour)) {
		cert.SSLStatus = "valid"
	} else if now.Before(cert.ExpiresAt) {
		cert.SSLStatus = "expiring"
	} else {
		cert.SSLStatus = "expired"
	}

	info := *cert // Копия
	return &info, nil
}

// GetAllDomainInfo получает информацию о всех доменах
func (r *DomainRotator) GetAllDomainInfo() []*CertificateInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]*CertificateInfo, 0, len(r.certificates))
	for _, cert := range r.certificates {
		info := *cert // Копия
		infos = append(infos, &info)
	}
	return infos
}

// RenewDomain продлевает сертификат домена
func (r *DomainRotator) RenewDomain(ctx context.Context, domain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cert, ok := r.certificates[domain]
	if !ok {
		return fmt.Errorf("domain not found: %s", domain)
	}

	// Обновление информации о сертификате
	cert.IssuedAt = time.Now()
	cert.ExpiresAt = time.Now().Add(90 * 24 * time.Hour)
	cert.SSLStatus = "valid"

	r.logger.Info("Certificate renewed",
		zap.String("domain", domain),
		zap.Time("expires_at", cert.ExpiresAt))

	return nil
}

// AutoRenew автоматически продлевает истекающие сертификаты
func (r *DomainRotator) AutoRenew(ctx context.Context) error {
	r.mu.RLock()
	domainsToRenew := make([]string, 0)

	for domain, cert := range r.certificates {
		if !cert.AutoRenew {
			continue
		}

		timeUntilExpiry := time.Until(cert.ExpiresAt)
		if timeUntilExpiry < time.Duration(r.autoRenewBefore)*time.Hour {
			domainsToRenew = append(domainsToRenew, domain)
		}
	}
	r.mu.RUnlock()

	if len(domainsToRenew) == 0 {
		return nil
	}

	r.logger.Info("Auto-renewing certificates",
		zap.Int("count", len(domainsToRenew)))

	for _, domain := range domainsToRenew {
		if err := r.RenewDomain(ctx, domain); err != nil {
			r.logger.Error("Failed to renew certificate",
				zap.String("domain", domain),
				zap.Error(err))
		}
	}

	return nil
}

// generateRandomSubdomain генерирует случайный поддомен
func (r *DomainRotator) generateRandomSubdomain() string {
	// Генерация случайной строки
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)[:12]

	// Добавление префикса
	prefixes := []string{"app", "web", "cloud", "secure", "login"}
	prefix := prefixes[r.randomInt(0, len(prefixes))]

	return fmt.Sprintf("%s-%s", prefix, randomStr)
}

func (r *DomainRotator) randomInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(n.Int64()) + min
}

// Start запускает фоновые задачи
func (r *DomainRotator) Start(ctx context.Context) error {
	go r.autoRenewLoop(ctx)
	return nil
}

func (r *DomainRotator) autoRenewLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.AutoRenew(ctx); err != nil {
				r.logger.Error("Auto-renew failed", zap.Error(err))
			}
		}
	}
}

// Close закрывает ротатор
func (r *DomainRotator) Close() error {
	r.logger.Info("Domain rotator closed")
	return nil
}

// GetStats возвращает статистику
func (r *DomainRotator) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	valid := 0
	expiring := 0
	expired := 0

	for _, cert := range r.certificates {
		now := time.Now()
		if now.After(cert.ExpiresAt) {
			expired++
		} else if time.Until(cert.ExpiresAt) < time.Duration(r.autoRenewBefore)*time.Hour {
			expiring++
		} else {
			valid++
		}
	}

	return map[string]interface{}{
		"total_domains":   len(r.domains),
		"current_domain":  r.currentDomain,
		"valid_certs":     valid,
		"expiring_certs":  expiring,
		"expired_certs":   expired,
		"last_rotation":   r.lastRotation,
		"auto_renew":      r.config.AutoRenew,
	}
}
