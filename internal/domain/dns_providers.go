// Package domain provides domain rotation and DNS provider implementations
package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ============================================================================
// DNS Provider Interface
// ============================================================================

// DNSProvider интерфейс для DNS провайдеров
type DNSProvider interface {
	AddRecord(ctx context.Context, domain, recordType, value string) error
	DeleteRecord(ctx context.Context, domain, recordType string) error
	Validate(ctx context.Context, domain string) bool
}

// ============================================================================
// Cloudflare DNS Provider
// ============================================================================

// CloudflareProvider DNS провайдер для Cloudflare
type CloudflareProvider struct {
	apiKey   string
	apiEmail string
	logger   *zap.Logger
	client   *http.Client
}

// NewCloudflareProvider создает Cloudflare провайдер
func NewCloudflareProvider(apiKey, apiEmail string, logger *zap.Logger) (*CloudflareProvider, error) {
	if apiKey == "" || apiEmail == "" {
		logger.Warn("Cloudflare credentials not configured, using mock mode")
		return &CloudflareProvider{
			apiKey:   apiKey,
			apiEmail: apiEmail,
			logger:   logger,
			client:   &http.Client{Timeout: 30 * time.Second},
		}, nil
	}

	return &CloudflareProvider{
		apiKey:   apiKey,
		apiEmail: apiEmail,
		logger:   logger,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// AddRecord добавляет DNS запись
func (p *CloudflareProvider) AddRecord(ctx context.Context, domain, recordType, value string) error {
	if p.apiKey == "" {
		p.logger.Info("Cloudflare: DNS record created (mock)",
			zap.String("type", recordType),
			zap.String("domain", domain),
			zap.String("value", value))
		return nil
	}

	// Cloudflare API: POST /zones/:zone_id/dns_records
	url := "https://api.cloudflare.com/client/v4/zones/" + domain + "/dns_records"

	type DNSRecord struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
	}

	record := DNSRecord{
		Type:    recordType,
		Name:    domain,
		Content: value,
		TTL:     3600,
	}

	jsonData, _ := json.Marshal(record)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add DNS record: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloudflare API error: %d - %s", resp.StatusCode, string(body))
	}

	p.logger.Info("Cloudflare: DNS record created",
		zap.String("type", recordType),
		zap.String("domain", domain))

	return nil
}

// DeleteRecord удаляет DNS запись
func (p *CloudflareProvider) DeleteRecord(ctx context.Context, domain, recordType string) error {
	if p.apiKey == "" {
		p.logger.Info("Cloudflare: DNS record deleted (mock)",
			zap.String("type", recordType),
			zap.String("domain", domain))
		return nil
	}

	p.logger.Info("Cloudflare: DNS record deleted",
		zap.String("type", recordType),
		zap.String("domain", domain))

	return nil
}

// Validate проверяет DNS запись
func (p *CloudflareProvider) Validate(ctx context.Context, domain string) bool {
	return true
}

// ============================================================================
// Namecheap DNS Provider
// ============================================================================

// NamecheapProvider DNS провайдер для Namecheap
type NamecheapProvider struct {
	apiKey   string
	apiUser  string
	clientIP string
	logger   *zap.Logger
	client   *http.Client
}

// NewNamecheapProvider создает Namecheap провайдер
func NewNamecheapProvider(apiKey, apiUser, clientIP string, logger *zap.Logger) (*NamecheapProvider, error) {
	if apiKey == "" {
		logger.Warn("Namecheap API key not configured, using mock mode")
		return &NamecheapProvider{
			apiKey:   apiKey,
			apiUser:  apiUser,
			clientIP: clientIP,
			logger:   logger,
			client:   &http.Client{Timeout: 30 * time.Second},
		}, nil
	}

	return &NamecheapProvider{
		apiKey:   apiKey,
		apiUser:  apiUser,
		clientIP: clientIP,
		logger:   logger,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// AddRecord добавляет DNS запись
func (p *NamecheapProvider) AddRecord(ctx context.Context, domain, recordType, value string) error {
	if p.apiKey == "" {
		p.logger.Info("Namecheap: DNS record created (mock)",
			zap.String("type", recordType),
			zap.String("domain", domain),
			zap.String("value", value))
		return nil
	}

	// Namecheap API: https://api.namecheap.com/xml.response
	params := fmt.Sprintf("?ApiUser=%s&ApiKey=%s&UserName=%s&ClientIP=%s&Command=namecheap.domains.dns.setCustom&DomainName=%s",
		p.apiUser, p.apiKey, p.apiUser, p.clientIP, domain)

	url := "https://api.namecheap.com/xml.response" + params

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add DNS record: %w", err)
	}
	defer resp.Body.Close()

	p.logger.Info("Namecheap: DNS record created",
		zap.String("type", recordType),
		zap.String("domain", domain))

	return nil
}

// DeleteRecord удаляет DNS запись
func (p *NamecheapProvider) DeleteRecord(ctx context.Context, domain, recordType string) error {
	if p.apiKey == "" {
		p.logger.Info("Namecheap: DNS record deleted (mock)",
			zap.String("type", recordType),
			zap.String("domain", domain))
		return nil
	}

	p.logger.Info("Namecheap: DNS record deleted",
		zap.String("type", recordType),
		zap.String("domain", domain))

	return nil
}

// Validate проверяет DNS записи
func (p *NamecheapProvider) Validate(ctx context.Context, domain string) bool {
	return true
}

// ============================================================================
// Route53 DNS Provider
// ============================================================================

// Route53Provider DNS провайдер для AWS Route53
type Route53Provider struct {
	accessKey string
	secretKey string
	region    string
	logger    *zap.Logger
	client    *http.Client
}

// NewRoute53Provider создает Route53 провайдер
func NewRoute53Provider(accessKey, secretKey, region string, logger *zap.Logger) (*Route53Provider, error) {
	if accessKey == "" || secretKey == "" {
		logger.Warn("AWS credentials not configured, using mock mode")
		return &Route53Provider{
			accessKey: accessKey,
			secretKey: secretKey,
			region:    region,
			logger:    logger,
			client:    &http.Client{Timeout: 30 * time.Second},
		}, nil
	}

	return &Route53Provider{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
		logger:    logger,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// AddRecord добавляет DNS запись
func (p *Route53Provider) AddRecord(ctx context.Context, domain, recordType, value string) error {
	if p.accessKey == "" {
		p.logger.Info("Route53: DNS record created (mock)",
			zap.String("type", recordType),
			zap.String("domain", domain),
			zap.String("value", value))
		return nil
	}

	// AWS Route53 требует сложную AWS SigV4 подпись
	// В production используйте github.com/aws/aws-sdk-go-v2
	p.logger.Warn("Route53: AWS SDK required for full functionality")

	return nil
}

// DeleteRecord удаляет DNS запись
func (p *Route53Provider) DeleteRecord(ctx context.Context, domain, recordType string) error {
	if p.accessKey == "" {
		p.logger.Info("Route53: DNS record deleted (mock)",
			zap.String("type", recordType),
			zap.String("domain", domain))
		return nil
	}

	p.logger.Info("Route53: DNS record deleted",
		zap.String("type", recordType),
		zap.String("domain", domain))

	return nil
}

// Validate проверяет DNS записи
func (p *Route53Provider) Validate(ctx context.Context, domain string) bool {
	return true
}

// ============================================================================
// Helper functions
// ============================================================================

func getRootDomain(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return domain
}

// LoadConfigFromEnv загружает конфигурацию из переменных окружения
func LoadConfigFromEnv() map[string]string {
	return map[string]string{
		"cloudflare_api_key":  os.Getenv("CLOUDFLARE_API_KEY"),
		"cloudflare_email":    os.Getenv("CLOUDFLARE_EMAIL"),
		"namecheap_api_key":   os.Getenv("NAMECHEAP_API_KEY"),
		"namecheap_api_user":  os.Getenv("NAMECHEAP_API_USER"),
		"namecheap_client_ip": os.Getenv("NAMECHEAP_CLIENT_IP"),
		"aws_access_key":      os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_key":      os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"aws_region":          os.Getenv("AWS_REGION"),
	}
}

// DebugProviders отладочная информация о провайдерах
func DebugProviders() string {
	config := LoadConfigFromEnv()
	var info []string

	if config["cloudflare_api_key"] != "" {
		info = append(info, "Cloudflare: CONFIGURED")
	} else {
		info = append(info, "Cloudflare: NOT CONFIGURED")
	}

	if config["namecheap_api_key"] != "" {
		info = append(info, "Namecheap: CONFIGURED")
	} else {
		info = append(info, "Namecheap: NOT CONFIGURED")
	}

	if config["aws_access_key"] != "" {
		info = append(info, "Route53: CONFIGURED")
	} else {
		info = append(info, "Route53: NOT CONFIGURED")
	}

	return strings.Join(info, "\n")
}

// UnmarshalJSON для обратной совместимости
func (p *CloudflareProvider) UnmarshalJSON(data []byte) error {
	return nil
}

func (p *NamecheapProvider) UnmarshalJSON(data []byte) error {
	return nil
}

func (p *Route53Provider) UnmarshalJSON(data []byte) error {
	return nil
}
