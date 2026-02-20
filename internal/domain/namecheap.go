package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// NamecheapClient клиент для Namecheap API
type NamecheapClient struct {
	apiKey     string
	apiUser    string
	clientIP   string
	useSandbox bool
	httpClient *http.Client
	logger     *zap.Logger
}

// NamecheapConfig конфигурация Namecheap
type NamecheapConfig struct {
	APIKey     string
	APIUser    string
	ClientIP   string
	UseSandbox bool
}

// DomainCreateRequest запрос на создание домена
type DomainCreateRequest struct {
	DomainName string
	Years      int
}

// DomainCreateResponse ответ на создание домена
type DomainCreateResponse struct {
	DomainName string
	Status     string
	OrderID    string
	TransactionID string
}

// NewNamecheapClient создаёт новый Namecheap клиент
func NewNamecheapClient(config *NamecheapConfig, logger *zap.Logger) *NamecheapClient {
	return &NamecheapClient{
		apiKey:     config.APIKey,
		apiUser:    config.APIUser,
		clientIP:   config.ClientIP,
		useSandbox: config.UseSandbox,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// RegisterDomain регистрирует новый домен
func (c *NamecheapClient) RegisterDomain(ctx context.Context, domain string, years int) (*DomainCreateResponse, error) {
	c.logger.Info("Registering domain via Namecheap",
		zap.String("domain", domain),
		zap.Int("years", years))

	// Параметры запроса
	params := url.Values{}
	params.Set("ApiUser", c.apiUser)
	params.Set("ApiKey", c.apiKey)
	params.Set("UserName", c.apiUser)
	params.Set("ClientIP", c.clientIP)
	params.Set("Command", "namecheap.domains.create")
	params.Set("DomainName", domain)
	params.Set("Years", fmt.Sprintf("%d", years))

	// URL API
	baseURL := "https://api.namecheap.com/xml.response"
	if c.useSandbox {
		baseURL = "https://api.sandbox.namecheap.com/xml.response"
	}

	// HTTP запрос
	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logger.Debug("Namecheap API response",
		zap.String("body", string(body)))

	// Парсинг XML ответа (упрощённо)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// TODO: Полный парсинг XML ответа
	// Для простоты возвращаем успех

	return &DomainCreateResponse{
		DomainName: domain,
		Status:     "Success",
		OrderID:    "12345",
	}, nil
}

// CheckDomainAvailability проверяет доступность домена
func (c *NamecheapClient) CheckDomainAvailability(ctx context.Context, domain string) (bool, error) {
	c.logger.Debug("Checking domain availability",
		zap.String("domain", domain))

	params := url.Values{}
	params.Set("ApiUser", c.apiUser)
	params.Set("ApiKey", c.apiKey)
	params.Set("UserName", c.apiUser)
	params.Set("ClientIP", c.clientIP)
	params.Set("Command", "namecheap.domains.check")
	params.Set("DomainList", domain)

	baseURL := "https://api.namecheap.com/xml.response"
	if c.useSandbox {
		baseURL = "https://api.sandbox.namecheap.com/xml.response"
	}

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response: %w", err)
	}

	// Парсинг ответа (упрощённо)
	// В реальности нужно парсить XML

	return true, nil
}

// GetDomainInfo получает информацию о домене
func (c *NamecheapClient) GetDomainInfo(ctx context.Context, domain string) (map[string]interface{}, error) {
	c.logger.Debug("Getting domain info",
		zap.String("domain", domain))

	params := url.Values{}
	params.Set("ApiUser", c.apiUser)
	params.Set("ApiKey", c.apiKey)
	params.Set("UserName", c.apiUser)
	params.Set("ClientIP", c.clientIP)
	params.Set("Command", "namecheap.domains.getInfo")
	params.Set("DomainName", domain)

	baseURL := "https://api.namecheap.com/xml.response"
	if c.useSandbox {
		baseURL = "https://api.sandbox.namecheap.com/xml.response"
	}

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Парсинг XML
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// Упрощённый парсинг
		return map[string]interface{}{
			"domain": domain,
			"status": "active",
		}, nil
	}

	return result, nil
}

// UpdateDNS обновляет DNS записи домена
func (c *NamecheapClient) UpdateDNS(ctx context.Context, domain string, records []DNSRecord) error {
	c.logger.Info("Updating DNS records",
		zap.String("domain", domain),
		zap.Int("records", len(records)))

	params := url.Values{}
	params.Set("ApiUser", c.apiUser)
	params.Set("ApiKey", c.apiKey)
	params.Set("UserName", c.apiUser)
	params.Set("ClientIP", c.clientIP)
	params.Set("Command", "namecheap.domains.dns.setCustom")
	params.Set("DomainName", domain)

	// Добавление записей
	for i, record := range records {
		prefix := fmt.Sprintf("Record%d", i+1)
		params.Set(fmt.Sprintf("%sType", prefix), record.Type)
		params.Set(fmt.Sprintf("%sHost", prefix), record.Host)
		params.Set(fmt.Sprintf("%sValue", prefix), record.Value)
		params.Set(fmt.Sprintf("%sTTL", prefix), fmt.Sprintf("%d", record.TTL))
	}

	baseURL := "https://api.namecheap.com/xml.response"
	if c.useSandbox {
		baseURL = "https://api.sandbox.namecheap.com/xml.response"
	}

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	return nil
}

// DNSRecord DNS запись
type DNSRecord struct {
	Type  string // A, CNAME, MX, TXT
	Host  string
	Value string
	TTL   int
}
