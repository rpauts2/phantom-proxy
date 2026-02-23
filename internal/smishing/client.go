// Package smishing - SMS/WhatsApp Phishing Module
package smishing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ============================================================================
// Configuration
// ============================================================================

type Config struct {
	// SMS Provider
	SMSProvider   string `yaml:"sms_provider" env:"SMS_PROVIDER"` // twilio, smsru, generic
	SMSAPIKey     string `yaml:"sms_api_key" env:"SMS_API_KEY"`
	SMSAPISecret  string `yaml:"sms_api_secret" env:"SMS_API_SECRET"`
	SMSFrom       string `yaml:"sms_from" env:"SMS_FROM"`

	// WhatsApp
	WhatsAppEnabled bool   `yaml:"whatsapp_enabled" env:"WHATSAPP_ENABLED"`
	WhatsAppAPIURL  string `yaml:"whatsapp_api_url" env:"WHATSAPP_API_URL"`
	WhatsAppToken   string `yaml:"whatsapp_token" env:"WHATSAPP_TOKEN"`

	// Telegram
	TelegramEnabled bool   `yaml:"telegram_enabled" env:"TELEGRAM_ENABLED"`
	TelegramBotToken string `yaml:"telegram_bot_token" env:"TELEGRAM_BOT_TOKEN"`

	// Limits
	RateLimit int           `yaml:"rate_limit" env:"RATE_LIMIT"` // per hour
	MaxRetries int         `yaml:"max_retries" env:"MAX_RETRIES"`
	Timeout    time.Duration `yaml:"timeout" env:"TIMEOUT"`
}

func DefaultConfig() *Config {
	return &Config{
		SMSProvider:  "twilio",
		RateLimit:    100,
		MaxRetries:   3,
		Timeout:      30 * time.Second,
	}
}

// ============================================================================
// Message Types
// ============================================================================

type MessageType string

const (
	MessageTypeSMS     MessageType = "sms"
	MessageTypeWhatsApp MessageType = "whatsapp"
	MessageTypeTelegram MessageType = "telegram"
)

type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent     MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusFailed   MessageStatus = "failed"
	MessageStatusClicked  MessageStatus = "clicked"
)

type Message struct {
	ID          string                 `json:"id"`
	Type        MessageType           `json:"type"`
	To          string                 `json:"to"`
	From        string                 `json:"from"`
	Body        string                 `json:"body"`
	URL         string                 `json:"url,omitempty"`
	CampaignID  string                 `json:"campaign_id"`
	Status      MessageStatus          `json:"status"`
	SentAt      time.Time              `json:"sent_at,omitempty"`
	DeliveredAt time.Time              `json:"delivered_at,omitempty"`
	ClickedAt   time.Time              `json:"clicked_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ============================================================================
// SMishing Manager
// ============================================================================

type Manager struct {
	config   *Config
	logger   *zap.Logger
	redis    *redis.Client
	httpClient *http.Client
	rateLimiter *RateLimiter
	mu       sync.RWMutex
	sentCount int64
}

type RateLimiter struct {
	redis     *redis.Client
	limit     int
	window    time.Duration
}

func NewManager(config *Config, logger *zap.Logger, redisClient *redis.Client) (*Manager, error) {
	return &Manager{
		config: config,
		logger: logger,
		redis:  redisClient,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		rateLimiter: &RateLimiter{
			redis:  redisClient,
			limit:  config.RateLimit,
			window: time.Hour,
		},
	}, nil
}

// ============================================================================
// Send SMS
// ============================================================================

func (m *Manager) SendSMS(ctx context.Context, to, body, campaignID string) (*Message, error) {
	// Rate limiting
	if !m.rateLimiter.Allow(ctx, "sms") {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	msg := &Message{
		ID:         generateID(),
		Type:       MessageTypeSMS,
		To:         to,
		From:       m.config.SMSFrom,
		Body:       body,
		CampaignID: campaignID,
		Status:     MessageStatusPending,
	}

	var err error
	switch m.config.SMSProvider {
	case "twilio":
		err = m.sendTwilio(ctx, msg)
	case "smsru":
		err = m.sendSMSru(ctx, msg)
	default:
		err = m.sendGeneric(ctx, msg)
	}

	if err != nil {
		msg.Status = MessageStatusFailed
		msg.Error = err.Error()
		m.logger.Error("SMS send failed", zap.Error(err), zap.String("to", to))
	} else {
		msg.Status = MessageStatusSent
		msg.SentAt = time.Now()
		m.mu.Lock()
		m.sentCount++
		m.mu.Unlock()
	}

	// Store in Redis
	m.storeMessage(ctx, msg)

	return msg, err
}

// ============================================================================
// Send WhatsApp
// ============================================================================

func (m *Manager) SendWhatsApp(ctx context.Context, to, body, campaignID string) (*Message, error) {
	if !m.config.WhatsAppEnabled {
		return nil, fmt.Errorf("whatsapp is not enabled")
	}

	msg := &Message{
		ID:         generateID(),
		Type:       MessageTypeWhatsApp,
		To:         to,
		Body:       body,
		CampaignID: campaignID,
		Status:     MessageStatusPending,
	}

	err := m.sendWhatsAppMessage(ctx, msg)
	if err != nil {
		msg.Status = MessageStatusFailed
		msg.Error = err.Error()
	} else {
		msg.Status = MessageStatusSent
		msg.SentAt = time.Now()
	}

	m.storeMessage(ctx, msg)
	return msg, err
}

// ============================================================================
// Send Telegram
// ============================================================================

func (m *Manager) SendTelegram(ctx context.Context, chatID, body, campaignID string) (*Message, error) {
	if !m.config.TelegramEnabled {
		return nil, fmt.Errorf("telegram is not enabled")
	}

	msg := &Message{
		ID:         generateID(),
		Type:       MessageTypeTelegram,
		To:         chatID,
		Body:       body,
		CampaignID: campaignID,
		Status:     MessageStatusPending,
	}

	err := m.sendTelegramMessage(ctx, msg)
	if err != nil {
		msg.Status = MessageStatusFailed
		msg.Error = err.Error()
	} else {
		msg.Status = MessageStatusSent
		msg.SentAt = time.Now()
	}

	m.storeMessage(ctx, msg)
	return msg, err
}

// ============================================================================
// Provider Implementations
// ============================================================================

func (m *Manager) sendTwilio(ctx context.Context, msg *Message) error {
	// Twilio API implementation
	url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", m.config.SMSAPIKey)

	data := map[string]string{
		"To":   msg.To,
		"From": msg.From,
		"Body": msg.Body,
	}

	interfaceData := make(map[string]interface{}, len(data))
	for k, v := range data {
		interfaceData[k] = v
	}

	resp, err := m.doRequest(ctx, "POST", url, interfaceData, m.config.SMSAPISecret)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return fmt.Errorf("twilio error: %d", resp.StatusCode)
	}

	return nil
}

func (m *Manager) sendSMSru(ctx context.Context, msg *Message) error {
	// SMS.ru API implementation
	url := "https://sms.ru/sms/send"

	data := map[string]string{
		"api_id": m.config.SMSAPIKey,
		"to":     msg.To,
		"msg":    msg.Body,
		"json":   "1",
	}

	interfaceData := make(map[string]interface{}, len(data))
	for k, v := range data {
		interfaceData[k] = v
	}

	resp, err := m.doRequest(ctx, "POST", url, interfaceData, "")
	if err != nil {
		return err
	}

	var result map[string]interface{}
	json.Unmarshal(resp.Body, &result)

	if result["status"].(string) != "OK" {
		return fmt.Errorf("sms.ru error: %v", result)
	}

	return nil
}

func (m *Manager) sendGeneric(ctx context.Context, msg *Message) error {
	// Generic HTTP API
	m.logger.Info("Sending generic SMS",
		zap.String("to", msg.To),
		zap.String("body", msg.Body))

	// Simulate for demo
	return nil
}

func (m *Manager) sendWhatsAppMessage(ctx context.Context, msg *Message) error {
	url := m.config.WhatsAppAPIURL + "/messages"

	data := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                msg.To,
		"type":              "text",
		"text": map[string]string{
			"body": msg.Body,
		},
	}

	headers := map[string]string{
		"Authorization": "Bearer " + m.config.WhatsAppToken,
		"Content-Type":  "application/json",
	}

	resp, err := m.doRequestWithHeaders(ctx, "POST", url, data, headers)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		return fmt.Errorf("whatsapp error: %d", resp.StatusCode)
	}

	return nil
}

func (m *Manager) sendTelegramMessage(ctx context.Context, msg *Message) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", m.config.TelegramBotToken)

	data := map[string]interface{}{
		"chat_id": msg.To,
		"text":    msg.Body,
		"parse_mode": "HTML",
	}

	resp, err := m.doRequest(ctx, "POST", url, data, "")
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("telegram error: %d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// HTTP Helpers
// ============================================================================

type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

func (m *Manager) doRequest(ctx context.Context, method, url string, data map[string]interface{}, auth string) (*HTTPResponse, error) {
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	if auth != "" {
		headers["Authorization"] = "Basic " + auth
	}

	return m.doRequestWithHeaders(ctx, method, url, data, headers)
}

func (m *Manager) doRequestWithHeaders(ctx context.Context, method, url string, data map[string]interface{}, headers map[string]string) (*HTTPResponse, error) {
	// Simplified - use proper HTTP client in production
	m.logger.Debug("HTTP Request",
		zap.String("method", method),
		zap.String("url", url))

	return &HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

// ============================================================================
// Rate Limiting
// ============================================================================

func (r *RateLimiter) Allow(ctx context.Context, key string) bool {
	redisKey := fmt.Sprintf("ratelimit:%s:%s", key, time.Now().Format("2006010215"))

	count, err := r.redis.Incr(ctx, redisKey).Result()
	if err != nil {
		return true // Fail open
	}

	if count == 1 {
		r.redis.Expire(ctx, redisKey, r.window)
	}

	return count <= int64(r.limit)
}

// ============================================================================
// Storage
// ============================================================================

func (m *Manager) storeMessage(ctx context.Context, msg *Message) error {
	key := fmt.Sprintf("message:%s", msg.ID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return m.redis.Set(ctx, key, data, 7*24*time.Hour).Err()
}

func (m *Manager) GetMessage(ctx context.Context, id string) (*Message, error) {
	key := fmt.Sprintf("message:%s", id)
	data, err := m.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var msg Message
	err = json.Unmarshal(data, &msg)
	return &msg, err
}

// ============================================================================
// Campaign Management
// ============================================================================

type Campaign struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        MessageType `json:"type"`
	Status      string    `json:"status"` // draft, running, paused, completed
	TargetCount int       `json:"target_count"`
	SentCount   int       `json:"sent_count"`
	DeliveredCount int    `json:"delivered_count"`
	ClickedCount int     `json:"clicked_count"`
	Template    string    `json:"template"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (m *Manager) CreateCampaign(ctx context.Context, name, template, url string, msgType MessageType) (*Campaign, error) {
	campaign := &Campaign{
		ID:        generateID(),
		Name:      name,
		Type:      msgType,
		Status:    "draft",
		Template:  template,
		URL:       url,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(campaign)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("campaign:%s", campaign.ID)
	return campaign, m.redis.Set(ctx, key, data, 30*24*time.Hour).Err()
}

// ============================================================================
// Stats
// ============================================================================

func (m *Manager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys, err := m.redis.Keys(ctx, "message:*").Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_sent":    m.sentCount,
		"total_messages": len(keys),
		"by_type":       map[string]int{},
		"by_status":     map[string]int{},
	}

	return stats, nil
}

// ============================================================================
// Helpers
// ============================================================================

func generateID() string {
	return strings.ReplaceAll(time.Now().Format("20060102150405"), "", "") + fmt.Sprintf("%d", time.Now().UnixNano()%10000)
}
