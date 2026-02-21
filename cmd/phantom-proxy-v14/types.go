// Package main - PhantomProxy v14.0 Core Types
package main

import (
	"time"
)

// Config - основная конфигурация
type Config struct {
	// Сеть
	BindIP      string `yaml:"bind_ip" env:"BIND_IP"`
	HTTPSPort   int    `yaml:"https_port" env:"HTTPS_PORT"`
	HTTP3Port   int    `yaml:"http3_port" env:"HTTP3_PORT"`
	Domain      string `yaml:"domain" env:"DOMAIN"`

	// TLS
	AutoCert   bool   `yaml:"auto_cert" env:"AUTO_CERT"`
	CertPath   string `yaml:"cert_path" env:"CERT_PATH"`
	KeyPath    string `yaml:"key_path" env:"KEY_PATH"`

	// База данных
	DatabaseType    string `yaml:"database_type" env:"DATABASE_TYPE"`
	DatabasePath    string `yaml:"database_path" env:"DATABASE_PATH"`
	PostgresURL     string `yaml:"postgres_url" env:"POSTGRES_URL"`

	// Redis
	RedisAddr       string `yaml:"redis_addr" env:"REDIS_ADDR"`
	RedisPassword   string `yaml:"redis_password" env:"REDIS_PASSWORD"`
	RedisDB         int    `yaml:"redis_db" env:"REDIS_DB"`
	SessionTTL      time.Duration `yaml:"session_ttl" env:"SESSION_TTL"`

	// Phishlets
	PhishletsPath   string `yaml:"phishlets_path" env:"PHISHLETS_PATH"`

	// Event Bus
	EventBusType    string `yaml:"event_bus_type" env:"EVENT_BUS_TYPE"` // redis, nats
	NATSURL         string `yaml:"nats_url" env:"NATS_URL"`

	// API
	APIEnabled      bool   `yaml:"api_enabled" env:"API_ENABLED"`
	APIPort         int    `yaml:"api_port" env:"API_PORT"`
	APIKey          string `yaml:"api_key" env:"API_KEY"`

	// Безопасность
	JA3Enabled      bool   `yaml:"ja3_enabled" env:"JA3_ENABLED"`
	MLDetection     bool   `yaml:"ml_detection" env:"ML_DETECTION"`
	MLThreshold     float64 `yaml:"ml_threshold" env:"ML_THRESHOLD"`

	// Polymorphic
	PolymorphicEnabled bool   `yaml:"polymorphic_enabled" env:"POLYMORPHIC_ENABLED"`
	PolymorphicLevel   string `yaml:"polymorphic_level" env:"POLYMORPHIC_LEVEL"`

	// Service Worker
	ServiceWorkerEnabled bool `yaml:"serviceworker_enabled" env:"SERVICEWORKER_ENABLED"`

	// WebSocket
	WebSocketEnabled bool `yaml:"websocket_enabled" env:"WEBSOCKET_ENABLED"`

	// Multi-tenant
	MultiTenantEnabled bool `yaml:"multi_tenant_enabled" env:"MULTI_TENANT_ENABLED"`

	// Risk Score
	RiskScoreEnabled bool    `yaml:"risk_score_enabled" env:"RISK_SCORE_ENABLED"`
	RiskThresholdHigh float64 `yaml:"risk_threshold_high" env:"RISK_THRESHOLD_HIGH"`
	RiskThresholdCritical float64 `yaml:"risk_threshold_critical" env:"RISK_THRESHOLD_CRITICAL"`

	// Vishing
	VishingEnabled   bool   `yaml:"vishing_enabled" env:"VISHING_ENABLED"`
	VishingProvider  string `yaml:"vishing_provider" env:"VISHING_PROVIDER"`
	TwilioAccountSID string `yaml:"twilio_account_sid" env:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `yaml:"twilio_auth_token" env:"TWILIO_AUTH_TOKEN"`
	TwilioPhoneNumber string `yaml:"twilio_phone_number" env:"TWILIO_PHONE_NUMBER"`
	ElevenLabsAPIKey string `yaml:"elevenlabs_api_key" env:"ELEVENLABS_API_KEY"`

	// FSTEC
	FSTEDEnabled      bool   `yaml:"fstec_enabled" env:"FSTEC_ENABLED"`
	FSTEDEncryptLogs  bool   `yaml:"fstec_encrypt_logs" env:"FSTEC_ENCRYPT_LOGS"`
	FSTECCategory     string `yaml:"fstec_category" env:"FSTEC_CATEGORY"`

	// AI Service
	AIServiceEnabled bool   `yaml:"ai_service_enabled" env:"AI_SERVICE_ENABLED"`
	AIServiceURL     string `yaml:"ai_service_url" env:"AI_SERVICE_URL"`
	AIEnabled        bool   `yaml:"ai_enabled" env:"AI_ENABLED"`

	// Логирование
	Debug    bool   `yaml:"debug" env:"DEBUG"`
	LogPath  string `yaml:"log_path" env:"LOG_PATH"`
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		BindIP:               "0.0.0.0",
		HTTPSPort:            8443,
		HTTP3Port:            8443,
		Domain:               "phantom.local",
		AutoCert:             false,
		DatabaseType:         "sqlite",
		DatabasePath:         "./phantom.db",
		RedisAddr:            "localhost:6379",
		RedisDB:              0,
		SessionTTL:           24 * time.Hour,
		PhishletsPath:        "./configs/phishlets",
		EventBusType:         "redis",
		APIEnabled:           true,
		APIPort:              8080,
		JA3Enabled:           true,
		MLDetection:          false,
		PolymorphicEnabled:   true,
		PolymorphicLevel:     "high",
		ServiceWorkerEnabled: true,
		WebSocketEnabled:     true,
		MultiTenantEnabled:   false,
		RiskScoreEnabled:     true,
		RiskThresholdHigh:    80,
		RiskThresholdCritical: 95,
		VishingEnabled:       false,
		VishingProvider:      "twilio",
		FSTEDEnabled:         false,
		FSTEDEncryptLogs:     true,
		FSTECCategory:        "УЗ-2",
		AIServiceEnabled:     true,
		AIServiceURL:         "http://localhost:8081",
		Debug:                false,
		LogPath:              "./logs/phantom.log",
		LogLevel:             "info",
	}
}

// Database представляет подключение к БД
type Database struct {
	// SQLite/PostgreSQL connection
	// TODO: Implement
}

// EventBus - шина событий
type EventBus struct {
	// Redis/NATS client
	// TODO: Implement
}

// SessionManager - менеджер сессий
type SessionManager struct {
	// Redis client
	// TODO: Implement
}

// PhishletLoader - загрузчик фишлетов
type PhishletLoader struct {
	// Phishlets map
	// TODO: Implement
}

// AiTMProxy - основной proxy
type AiTMProxy struct {
	// Fiber app, TLS config, etc
	// TODO: Implement
}
