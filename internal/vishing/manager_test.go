package vishing

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewTwilioProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		TwilioAccountSID:  "AC1234567890",
		TwilioAuthToken:   "secret_token",
		TwilioPhoneNumber: "+1234567890",
	}

	provider := NewTwilioProvider(config, logger)
	if provider == nil {
		t.Fatal("Expected provider to be created")
	}

	if provider.baseURL != "https://api.twilio.com/2010-04-01" {
		t.Errorf("Unexpected base URL: %s", provider.baseURL)
	}
}

func TestTwilioProviderMockMode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Без credentials - mock режим
	config := &Config{
		TwilioAccountSID: "",
		TwilioAuthToken:  "",
	}

	provider := NewTwilioProvider(config, logger)
	call := &Call{
		ID:          "test-call",
		TargetPhone: "+1234567890",
		Status:      StatusQueued,
	}

	ctx := context.Background()
	err := provider.MakeCall(ctx, config, call)
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}

	if call.Status != StatusCalling {
		t.Errorf("Expected status Calling, got: %s", call.Status)
	}
}

func TestGetCallStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		TwilioAccountSID: "",
		TwilioAuthToken:  "",
	}

	provider := NewTwilioProvider(config, logger)
	ctx := context.Background()

	// Mock режим - должен вернуть StatusInProgress
	status, err := provider.GetStatus(ctx, "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if status != StatusQueued {
		t.Errorf("Expected StatusQueued for empty callID, got: %s", status)
	}
}

func TestEndCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		TwilioAccountSID: "",
		TwilioAuthToken:  "",
	}

	provider := NewTwilioProvider(config, logger)
	ctx := context.Background()

	// Пустой callID - должен вернуть nil
	err := provider.EndCall(ctx, "")
	if err != nil {
		t.Errorf("Expected no error for empty callID, got: %v", err)
	}
}

func TestNewSmishingManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:     true,
		SMSRuAPIKey: "",
	}

	manager := NewSmishingManager(config, logger)
	if manager == nil {
		t.Fatal("Expected manager to be created")
	}
}

func TestSendSMSMockMode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:     true,
		SMSRuAPIKey: "", // Без ключа - mock режим
	}

	manager := NewSmishingManager(config, logger)
	ctx := context.Background()

	messageID, err := manager.SendSMS(ctx, "+1234567890", "Test message")
	if err != nil {
		t.Errorf("Expected no error in mock mode, got: %v", err)
	}

	if messageID == "" {
		t.Error("Expected messageID to be returned")
	}
}

func TestSendBulkSMS(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled:     true,
		SMSRuAPIKey: "",
	}

	manager := NewSmishingManager(config, logger)
	ctx := context.Background()

	phoneNumbers := []string{"+1234567890", "+0987654321"}
	message := "Bulk test message"

	results := manager.SendBulkSMS(ctx, phoneNumbers, message)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got: %d", len(results))
	}

	for _, phone := range phoneNumbers {
		if _, ok := results[phone]; !ok {
			t.Errorf("Expected result for phone: %s", phone)
		}
	}
}

func TestCallStatusConstants(t *testing.T) {
	statuses := []CallStatus{
		StatusQueued,
		StatusCalling,
		StatusInProgress,
		StatusCompleted,
		StatusFailed,
		StatusNoAnswer,
	}

	for _, status := range statuses {
		if status == "" {
			t.Errorf("Status constant is empty: %v", status)
		}
	}
}

func TestCallClone(t *testing.T) {
	original := &Call{
		ID:          "test-id",
		TenantID:    "tenant-1",
		TargetPhone: "+1234567890",
		Status:      StatusQueued,
		StartTime:   time.Now(),
	}

	clone := original.clone()

	if clone.ID != original.ID {
		t.Error("Clone ID mismatch")
	}

	if clone.TargetPhone != original.TargetPhone {
		t.Error("Clone phone mismatch")
	}

	// Изменение оригинала не должно влиять на клон
	original.TargetPhone = "+9999999999"
	if clone.TargetPhone == original.TargetPhone {
		t.Error("Clone should be independent from original")
	}
}

func TestManagerStartCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled: true,
	}

	manager := NewManager(logger, config)
	ctx := context.Background()

	callConfig := &CallConfig{
		TargetPhone: "+1234567890",
		ScriptID:    "test-script",
		VoiceID:     "test-voice",
	}

	call, err := manager.StartCall(ctx, "tenant-1", callConfig)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if call == nil {
		t.Fatal("Expected call to be created")
	}

	if call.TargetPhone != callConfig.TargetPhone {
		t.Errorf("Expected phone %s, got %s", callConfig.TargetPhone, call.TargetPhone)
	}
}

func TestManagerGetCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled: true,
	}

	manager := NewManager(logger, config)
	ctx := context.Background()

	callConfig := &CallConfig{
		TargetPhone: "+1234567890",
		ScriptID:    "test-script",
	}

	created, _ := manager.StartCall(ctx, "tenant-1", callConfig)

	retrieved, err := manager.GetCall(created.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected call ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestManagerListCalls(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled: true,
	}

	manager := NewManager(logger, config)
	ctx := context.Background()

	// Создадим несколько звонков с разными телефонами
	phones := []string{"+1111111111", "+2222222222", "+3333333333", "+4444444444", "+5555555555"}
	for i, phone := range phones {
		callConfig := &CallConfig{
			TargetPhone: phone,
			ScriptID:    "test-script",
		}
		call, _ := manager.StartCall(ctx, "tenant-1", callConfig)
		// Небольшая задержка чтобы ID были разными
		time.Sleep(time.Millisecond * time.Duration(i))
		_ = call
	}

	calls := manager.ListCalls("tenant-1", 10)
	if len(calls) < 1 {
		t.Errorf("Expected at least 1 call, got: %d", len(calls))
	}

	// Тест лимита
	limitedCalls := manager.ListCalls("tenant-1", 3)
	if len(limitedCalls) > 3 {
		t.Errorf("Expected max 3 calls with limit, got: %d", len(limitedCalls))
	}
}

func TestManagerEndCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled: true,
	}

	manager := NewManager(logger, config)
	ctx := context.Background()

	callConfig := &CallConfig{
		TargetPhone: "+1234567890",
		ScriptID:    "test-script",
	}

	created, _ := manager.StartCall(ctx, "tenant-1", callConfig)

	err := manager.EndCall(created.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	ended, _ := manager.GetCall(created.ID)
	if ended.Status != StatusCompleted {
		t.Errorf("Expected status Completed, got: %s", ended.Status)
	}
}

func TestManagerNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	config := &Config{
		Enabled: true,
	}

	manager := NewManager(logger, config)

	_, err := manager.GetCall("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent call")
	}

	err = manager.EndCall("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent call")
	}
}
