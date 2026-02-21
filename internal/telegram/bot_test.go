package telegram

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewBot(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &BotConfig{
		Token:  "test-token",
		ChatID: 123456,
	}

	bot, err := NewBot(config, logger)
	// Bot creation might fail due to invalid token, but should not panic
	if err == nil && bot == nil {
		t.Fatal("Expected bot or error")
	}
}

func TestSendMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &BotConfig{
		Token:  "invalid-token-for-testing",
		ChatID: 123456,
	}

	bot, _ := NewBot(config, logger)
	if bot == nil {
		t.Skip("Bot not created, skipping test")
	}
	defer bot.Close()

	// This will fail due to invalid token, but tests the method exists
	err := bot.SendMessage("Test message")
	if err == nil {
		t.Error("Expected error with invalid token")
	}
}

func TestSendSessionAlert(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &BotConfig{
		Token:  "invalid-token-for-testing",
		ChatID: 123456,
	}

	bot, _ := NewBot(config, logger)
	if bot == nil {
		t.Skip("Bot not created, skipping test")
	}
	defer bot.Close()

	session := &SessionInfo{
		ID:        "test-session",
		VictimIP:  "192.168.1.1",
		TargetURL: "https://example.com",
		UserAgent: "Mozilla/5.0",
	}

	err := bot.SendSessionAlert(session)
	// Will fail with invalid token, but tests the method
	if err == nil {
		t.Error("Expected error with invalid token")
	}
}

func TestSendCredentialAlert(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &BotConfig{
		Token:  "invalid-token-for-testing",
		ChatID: 123456,
	}

	bot, _ := NewBot(config, logger)
	if bot == nil {
		t.Skip("Bot not created, skipping test")
	}
	defer bot.Close()

	cred := &CredentialInfo{
		SessionID: "test-session",
		Username:  "testuser",
		Password:  "testpass",
		Target:    "example.com",
	}

	err := bot.SendCredentialAlert(cred)
	// Will fail with invalid token, but tests the method
	if err == nil {
		t.Error("Expected error with invalid token")
	}
}

func TestGetStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &BotConfig{
		Token:  "invalid-token-for-testing",
		ChatID: 123456,
	}

	bot, _ := NewBot(config, logger)
	if bot == nil {
		t.Skip("Bot not created, skipping test")
	}
	defer bot.Close()

	stats := bot.GetStats()
	if stats == nil {
		t.Error("Expected stats to be returned")
	}
}

func TestClose(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &BotConfig{
		Token:  "invalid-token-for-testing",
		ChatID: 123456,
	}

	bot, _ := NewBot(config, logger)
	if bot == nil {
		t.Skip("Bot not created, skipping test")
	}

	// Close should not panic
	bot.Close()
}
