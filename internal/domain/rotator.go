package domain

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
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
	certificates    map[string]*certificate.Resource
}

// RotatorConfig конфигурация ротатора
type RotatorConfig struct {
	// Регистратор доменов
	RegistrarName     string // namecheap, godaddy
	RegistrarAPIKey   string
	RegistrarAPISecret string
	RegistrarAccount  string
	
	// DNS провайдер
	DNSProvider string // cloudflare, route53, etc
	
	// SSL настройки
	SSLProvider     string // letsencrypt
	SSLEmail        string
	SSLStoragePath  string
	
	// Ротация
	MinDomainAge      int // минут
	MaxDomainAge      int // минут
	AutoRenewBefore   int // часов до истечения
	
	// Лимиты
	MaxDomains        int
}

// DomainInfo информация о домене
type DomainInfo struct {
	Domain      string    `json:"domain"`
	Status      string    `json:"status"` // active, expired, blocked
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	SSLStatus   string    `json:"ssl_status"` // valid, expiring, expired
	DNSStatus   string    `json:"dns_status"` // configured, pending
}

// NewDomainRotator создаёт новый ротатор
func NewDomainRotator(config *RotatorConfig, logger *zap.Logger) (*DomainRotator, error) {
	if config.MinDomainAge == 0 {
		config.MinDomainAge = 60 // 1 час
	}
	if config.MaxDomainAge == 0 {
		config.MaxDomainAge = 1440 // 24 часа
	}
	if config.AutoRenewBefore == 0 {
		config.AutoRenewBefore = 24 // 24 часа
	}
	if config.MaxDomains == 0 {
		config.MaxDomains = 10
	}
	
	r := &DomainRotator{
		logger:       logger,
		config:       config,
		domains:      make([]string, 0),
		certificates: make(map[string]*certificate.Resource),
	}
	
	// Загрузка существующих доменов
	if err := r.loadDomains(); err != nil {
		logger.Warn("Failed to load existing domains", zap.Error(err))
	}
	
	return r, nil
}

// Start запускает фоновые задачи ротации
func (r *DomainRotator) Start(ctx context.Context) error {
	r.logger.Info("Starting domain rotator",
		zap.Int("min_age", r.config.MinDomainAge),
		zap.Int("max_age", r.config.MaxDomainAge))
	
	// Фоновая задача проверки возраста доменов
	go r.rotationWorker(ctx)
	
	// Фоновая задача обновления SSL
	go r.sslRenewalWorker(ctx)
	
	return nil
}

// RegisterDomain регистрирует новый домен
func (r *DomainRotator) RegisterDomain(ctx context.Context, baseDomain string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Проверка лимита
	if len(r.domains) >= r.config.MaxDomains {
		return "", fmt.Errorf("max domains limit reached: %d", r.config.MaxDomains)
	}
	
	// Генерация случайного поддомена
	subdomain := r.generateRandomSubdomain()
	newDomain := fmt.Sprintf("%s.%s", subdomain, baseDomain)
	
	r.logger.Info("Registering new domain",
		zap.String("domain", newDomain),
		zap.String("registrar", r.config.RegistrarName))
	
	// Регистрация через API регистратора
	if err := r.registerViaAPI(newDomain); err != nil {
		return "", fmt.Errorf("failed to register domain: %w", err)
	}
	
	// Настройка DNS
	if err := r.configureDNS(newDomain); err != nil {
		return "", fmt.Errorf("failed to configure DNS: %w", err)
	}
	
	// Получение SSL сертификата
	if err := r.obtainSSL(ctx, newDomain); err != nil {
		return "", fmt.Errorf("failed to obtain SSL: %w", err)
	}
	
	// Добавление в список
	r.domains = append(r.domains, newDomain)
	r.currentDomain = newDomain
	r.lastRotation = time.Now()
	
	r.logger.Info("Domain registered successfully",
		zap.String("domain", newDomain))
	
	return newDomain, nil
}

// RotateDomain переключается на новый домен
func (r *DomainRotator) RotateDomain(ctx context.Context, baseDomain string) (string, error) {
	r.logger.Info("Rotating domain")
	
	newDomain, err := r.RegisterDomain(ctx, baseDomain)
	if err != nil {
		return "", err
	}
	
	// Уведомление о смене домена (можно добавить callback)
	r.logger.Info("Domain rotated",
		zap.String("new_domain", newDomain),
		zap.String("old_domain", r.currentDomain))
	
	return newDomain, nil
}

// GetCurrentDomain возвращает текущий активный домен
func (r *DomainRotator) GetCurrentDomain() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentDomain
}

// GetDomains возвращает список всех доменов
func (r *DomainRotator) GetDomains() []DomainInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	infos := make([]DomainInfo, 0, len(r.domains))
	for _, domain := range r.domains {
		info := DomainInfo{
			Domain:    domain,
			Status:    "active",
			DNSStatus: "configured",
		}
		
		// Проверка SSL
		if cert, ok := r.certificates[domain]; ok {
			if time.Now().Before(cert.NotAfter.Add(-time.Duration(r.config.AutoRenewBefore) * time.Hour)) {
				info.SSLStatus = "valid"
			} else {
				info.SSLStatus = "expiring"
			}
		} else {
			info.SSLStatus = "missing"
		}
		
		infos = append(infos, info)
	}
	
	return infos
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
	return min + int(n.Int64())
}

// registerViaAPI регистрирует домен через API регистратора
func (r *DomainRotator) registerViaAPI(domain string) error {
	// TODO: Реализация для Namecheap
	// TODO: Реализация для GoDaddy
	
	r.logger.Debug("Domain registration simulated",
		zap.String("domain", domain))
	
	return nil
}

// configureDNS настраивает DNS записи
func (r *DomainRotator) configureDNS(domain string) error {
	// TODO: Интеграция с Cloudflare DNS
	// TODO: Интеграция с Route53
	
	r.logger.Debug("DNS configuration simulated",
		zap.String("domain", domain))
	
	return nil
}

// obtainSSL получает SSL сертификат через Let's Encrypt
func (r *DomainRotator) obtainSSL(ctx context.Context, domain string) error {
	r.logger.Info("Obtaining SSL certificate",
		zap.String("domain", domain),
		zap.String("provider", "letsencrypt"))
	
	// Создание пользователя
	user := newLegoUser(r.config.SSLEmail)
	
	// Конфигурация lego
	config := lego.NewConfig()
	config.CADirURL = lego.LEDirectoryProduction // Production
	// config.CADirURL = lego.LEDirectoryStaging   // Staging для тестов
	
	config.Certificate.KeyType = certcrypto.RSA2048
	
	// Создание клиента
	client, err := lego.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create lego client: %w", err)
	}
	
	// Регистрация пользователя
	var reg registration.Registration
	if r.config.RegistrarAccount != "" {
		// Восстановление существующего пользователя
		reg.Registration.Body.AccountID = r.config.RegistrarAccount
	} else {
		// Новая регистрация
		reg, err = client.Registration.Register(registration.RegisterOptions{
			TermsOfServiceAgreed: true,
		})
		if err != nil {
			return fmt.Errorf("failed to register user: %w", err)
		}
	}
	user.Registration = reg
	
	client.User = user
	
	// HTTP challenge
solver := http01.NewProviderServer("", "80")
	if err := client.Challenge.SetHTTP01Provider(solver); err != nil {
		return fmt.Errorf("failed to set HTTP provider: %w", err)
	}
	
	// Заказ сертификата
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %w", err)
	}
	
	// Сохранение сертификата
	if err := r.saveCertificate(domain, cert); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}
	
	r.certificates[domain] = cert
	
	r.logger.Info("SSL certificate obtained",
		zap.String("domain", domain),
		zap.Time("expires", cert.NotAfter))
	
	return nil
}

// saveCertificate сохраняет сертификат в файл
func (r *DomainRotator) saveCertificate(domain string, cert *certificate.Resource) error {
	if r.config.SSLStoragePath == "" {
		r.config.SSLStoragePath = "./certs"
	}
	
	// Создание директории
	if err := os.MkdirAll(r.config.SSLStoragePath, 0755); err != nil {
		return err
	}
	
	// Сохранение
	certFile := fmt.Sprintf("%s/%s.crt", r.config.SSLStoragePath, strings.ReplaceAll(domain, ".", "_"))
	keyFile := fmt.Sprintf("%s/%s.key", r.config.SSLStoragePath, strings.ReplaceAll(domain, ".", "_"))
	
	if err := os.WriteFile(certFile, cert.Certificate, 0644); err != nil {
		return err
	}
	
	if err := os.WriteFile(keyFile, cert.PrivateKey, 0600); err != nil {
		return err
	}
	
	r.logger.Debug("Certificate saved",
		zap.String("cert_file", certFile),
		zap.String("key_file", keyFile))
	
	return nil
}

// loadDomains загружает существующие домены
func (r *DomainRotator) loadDomains() error {
	// TODO: Загрузка из БД или конфига
	return nil
}

// rotationWorker фоновая задача ротации
func (r *DomainRotator) rotationWorker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.checkRotation()
		}
	}
}

// checkRotation проверяет необходимость ротации
func (r *DomainRotator) checkRotation() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.lastRotation.IsZero() {
		return
	}
	
	age := time.Since(r.lastRotation).Minutes()
	
	// Проверка на необходимость ротации
	if age >= float64(r.config.MinDomainAge) {
		// Рандомная проверка чтобы не все домены ротировались одновременно
		if r.randomInt(0, 100) < 30 { // 30% шанс
			r.logger.Info("Domain rotation triggered",
				zap.Float64("age_minutes", age))
			// Здесь можно вызвать RotateDomain
		}
	}
}

// sslRenewalWorker фоновая задача обновления SSL
func (r *DomainRotator) sslRenewalWorker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.checkSSLRenewal()
		}
	}
}

// checkSSLRenewal проверяет необходимость обновления SSL
func (r *DomainRotator) checkSSLRenewal() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for domain, cert := range r.certificates {
		timeToExpiry := time.Until(cert.NotAfter)
		
		if timeToExpiry.Hours() < float64(r.config.AutoRenewBefore) {
			r.logger.Info("SSL certificate expiring soon",
				zap.String("domain", domain),
				zap.Duration("time_to_expiry", timeToExpiry))
			
			// TODO: Запуск перевыпуска сертификата
		}
	}
}

// legoUser реализация registration.User
type legoUser struct {
	Email        string
	Registration *registration.Registration
	Key          []byte
}

func newLegoUser(email string) *legoUser {
	return &legoUser{
		Email: email,
	}
}

func (u *legoUser) GetEmail() string {
	return u.Email
}

func (u *legoUser) GetRegistration() *registration.Registration {
	return u.Registration
}

func (u *legoUser) GetPrivateKey() []byte {
	return u.Key
}
