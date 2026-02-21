// Package vishing - Vishing & Smishing Manager
package vishing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CallStatus статус звонка
type CallStatus string

const (
	StatusQueued     CallStatus = "queued"
	StatusCalling    CallStatus = "calling"
	StatusInProgress CallStatus = "in_progress"
	StatusCompleted  CallStatus = "completed"
	StatusFailed     CallStatus = "failed"
	StatusNoAnswer   CallStatus = "no_answer"
)

// Call звонок
type Call struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	TargetPhone  string            `json:"target_phone"`
	ScriptID     string            `json:"script_id"`
	VoiceID      string            `json:"voice_id"`
	Status       CallStatus        `json:"status"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Duration     int               `json:"duration"`
	RecordingURL string            `json:"recording_url"`
	Transcript   string            `json:"transcript"`
	KeyPresses   string            `json:"key_presses"`
	CustomData   map[string]string `json:"custom_data"`
	Error        string            `json:"error"`
}

// Script сценарий
type Script struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	Greeting    string   `json:"greeting"`
	Prompts     []Prompt `json:"prompts"`
}

// Prompt реплика
type Prompt struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Timeout    int    `json:"timeout"`
	MaxRetries int    `json:"max_retries"`
}

// Manager менеджер vishing
type Manager struct {
	mu       sync.RWMutex
	logger   *zap.Logger
	config   *Config
	calls    map[string]*Call
	scripts  map[string]*Script
	provider CallProvider
}

// Config конфигурация
type Config struct {
	Enabled          bool   `json:"enabled"`
	DefaultProvider  string `json:"default_provider"`
	TwilioAccountSID string `json:"twilio_account_sid"`
	TwilioAuthToken  string `json:"twilio_auth_token"`
	TwilioPhoneNumber string `json:"twilio_phone_number"`
	ElevenLabsAPIKey string `json:"elevenlabs_api_key"`
	SMSRuAPIKey      string `json:"sms_ru_api_key"`
}

// CallProvider интерфейс провайдера
type CallProvider interface {
	MakeCall(ctx context.Context, config *Config, call *Call) error
	GetStatus(ctx context.Context, callID string) (CallStatus, error)
	EndCall(ctx context.Context, callID string) error
}

// NewManager создает менеджер
func NewManager(logger *zap.Logger, config *Config) *Manager {
	if config == nil {
		config = &Config{
			Enabled:         false,
			DefaultProvider: "twilio",
		}
	}

	m := &Manager{
		logger:   logger,
		config:   config,
		calls:    make(map[string]*Call),
		scripts:  make(map[string]*Script),
		provider: NewTwilioProvider(config, logger),
	}

	// Загрузить стандартные сценарии
	m.loadDefaultScripts()

	return m
}

// StartCall начинает звонок
func (m *Manager) StartCall(ctx context.Context, tenantID string, config *CallConfig) (*Call, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Enabled {
		return nil, fmt.Errorf("vishing is disabled")
	}

	call := &Call{
		ID:          generateCallID(),
		TenantID:    tenantID,
		TargetPhone: config.TargetPhone,
		ScriptID:    config.ScriptID,
		VoiceID:     config.VoiceID,
		Status:      StatusQueued,
		StartTime:   time.Now(),
		CustomData:  config.CustomData,
	}

	m.calls[call.ID] = call

	// Запустить в горутине
	go m.executeCall(ctx, call)

	m.logger.Info("Vishing call started",
		zap.String("call_id", call.ID),
		zap.String("target", config.TargetPhone))

	return call, nil
}

// GetCall возвращает звонок
func (m *Manager) GetCall(callID string) (*Call, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	call, ok := m.calls[callID]
	if !ok {
		return nil, fmt.Errorf("call not found: %s", callID)
	}

	return call.clone(), nil
}

// ListCalls список звонков
func (m *Manager) ListCalls(tenantID string, limit int) []*Call {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var calls []*Call
	for _, call := range m.calls {
		if tenantID != "" && call.TenantID != tenantID {
			continue
		}
		calls = append(calls, call.clone())
	}

	if len(calls) > limit {
		calls = calls[:limit]
	}

	return calls
}

// EndCall завершает звонок
func (m *Manager) EndCall(callID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	call, ok := m.calls[callID]
	if !ok {
		return fmt.Errorf("call not found")
	}

	call.Status = StatusCompleted
	call.EndTime = time.Now()
	call.Duration = int(call.EndTime.Sub(call.StartTime).Seconds())

	m.logger.Info("Vishing call ended",
		zap.String("call_id", callID),
		zap.Int("duration", call.Duration))

	return nil
}

// AddScript добавляет сценарий
func (m *Manager) AddScript(script *Script) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.scripts[script.ID] = script
	m.logger.Info("Vishing script added", zap.String("id", script.ID))
	return nil
}

// GetScript возвращает сценарий
func (m *Manager) GetScript(scriptID string) (*Script, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	script, ok := m.scripts[scriptID]
	if !ok {
		return nil, fmt.Errorf("script not found: %s", scriptID)
	}

	return script, nil
}

// ListScripts список сценариев
func (m *Manager) ListScripts() []*Script {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var scripts []*Script
	for _, s := range m.scripts {
		scripts = append(scripts, s)
	}
	return scripts
}

// GetStats статистика
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"total_calls":   len(m.calls),
		"total_scripts": len(m.scripts),
		"provider":      m.config.DefaultProvider,
	}

	statusCount := make(map[string]int)
	for _, call := range m.calls {
		statusCount[string(call.Status)]++
	}
	stats["by_status"] = statusCount

	return stats
}

// executeCall выполняет звонок
func (m *Manager) executeCall(ctx context.Context, call *Call) {
	m.updateCallStatus(call.ID, StatusCalling)

	// Получить сценарий
	script, err := m.GetScript(call.ScriptID)
	if err != nil {
		m.updateCallStatus(call.ID, StatusFailed)
		call.Error = fmt.Sprintf("Script not found: %v", err)
		return
	}

	// Начать звонок через провайдера
	err = m.provider.MakeCall(ctx, m.config, call)
	if err != nil {
		m.updateCallStatus(call.ID, StatusFailed)
		call.Error = fmt.Sprintf("Call failed: %v", err)
		return
	}

	// Мониторинг
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := m.provider.GetStatus(ctx, call.ID)
			if err != nil {
				continue
			}

			m.updateCallStatus(call.ID, status)

			if status == StatusCompleted || status == StatusFailed {
				m.mu.Lock()
				call.EndTime = time.Now()
				call.Duration = int(call.EndTime.Sub(call.StartTime).Seconds())
				m.mu.Unlock()
				return
			}
		}
	}
}

func (m *Manager) updateCallStatus(callID string, status CallStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if call, ok := m.calls[callID]; ok {
		call.Status = status
	}
}

func (m *Manager) loadDefaultScripts() {
	// IT Support
	itSupport := &Script{
		ID:          "it_support",
		Name:        "IT Support Scam",
		Description: "Classic IT support scam",
		Language:    "ru",
		Greeting:    "Здравствуйте, это техническая поддержка Microsoft.",
		Prompts: []Prompt{
			{ID: "greeting", Text: "Здравствуйте, это техподдержка Microsoft. Обнаружена подозрительная активность.", Timeout: 10},
			{ID: "confirm", Text: "Нажмите 1 для подтверждения.", Timeout: 15},
			{ID: "explain", Text: "Нужно проверить ваш компьютер.", Timeout: 20},
		},
	}

	// Bank Security
	bankSecurity := &Script{
		ID:          "bank_security",
		Name:        "Bank Security Alert",
		Description: "Bank fraud prevention",
		Language:    "ru",
		Greeting:    "Здравствуйте, это служба безопасности Сбербанка.",
		Prompts: []Prompt{
			{ID: "greeting", Text: "Обнаружена подозрительная операция по вашей карте.", Timeout: 10},
			{ID: "verify", Text: "Назовите последние 4 цифры карты.", Timeout: 15},
			{ID: "block", Text: "Хотите заблокировать карту?", Timeout: 20},
		},
	}

	m.AddScript(itSupport)
	m.AddScript(bankSecurity)
}

func generateCallID() string {
	return fmt.Sprintf("call_%d", time.Now().UnixNano())
}

func (c *Call) clone() *Call {
	if c == nil {
		return nil
	}
	clone := *c
	clone.CustomData = make(map[string]string)
	for k, v := range c.CustomData {
		clone.CustomData[k] = v
	}
	return &clone
}

// CallConfig конфигурация звонка
type CallConfig struct {
	TargetPhone string            `json:"target_phone"`
	ScriptID    string            `json:"script_id"`
	VoiceID     string            `json:"voice_id"`
	MaxDuration int               `json:"max_duration"`
	CustomData  map[string]string `json:"custom_data"`
}

// TwilioProvider провайдер Twilio
type TwilioProvider struct {
	config   *Config
	logger   *zap.Logger
	httpClient *http.Client
}

func NewTwilioProvider(config *Config, logger *zap.Logger) *TwilioProvider {
	return &TwilioProvider{
		config:   config,
		logger:   logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *TwilioProvider) MakeCall(ctx context.Context, config *Config, call *Call) error {
	// Twilio API: POST /Accounts/{AccountSid}/Calls.json
	// Заглушка для демонстрации
	p.logger.Info("Twilio call initiated", zap.String("call_id", call.ID))
	return nil
}

func (p *TwilioProvider) GetStatus(ctx context.Context, callID string) (CallStatus, error) {
	return StatusInProgress, nil
}

func (p *TwilioProvider) EndCall(ctx context.Context, callID string) error {
	return nil
}

// SmishingManager менеджер SMS
type SmishingManager struct {
	config *Config
	logger *zap.Logger
}

// NewSmishingManager создает SMS менеджер
func NewSmishingManager(config *Config, logger *zap.Logger) *SmishingManager {
	return &SmishingManager{
		config: config,
		logger: logger,
	}
}

// SendSMS отправляет SMS
func (m *SmishingManager) SendSMS(ctx context.Context, phoneNumber, message string) (string, error) {
	if !m.config.Enabled {
		return "", fmt.Errorf("smishing is disabled")
	}

	// SMS.ru API
	messageID := fmt.Sprintf("sms_%d", time.Now().UnixNano())
	m.logger.Info("SMS sent", zap.String("to", phoneNumber), zap.String("id", messageID))

	return messageID, nil
}

// SendBulkSMS массовая рассылка
func (m *SmishingManager) SendBulkSMS(ctx context.Context, phoneNumbers []string, message string) map[string]string {
	results := make(map[string]string)

	for _, phone := range phoneNumbers {
		id, err := m.SendSMS(ctx, phone, message)
		if err != nil {
			results[phone] = fmt.Sprintf("error: %v", err)
		} else {
			results[phone] = id
		}
	}

	return results
}
