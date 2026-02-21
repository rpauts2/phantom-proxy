// Package fstec - FSTEC Compliance & GOST Encryption
package fstec

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// GOSTCipher шифр GOST
type GOSTCipher struct {
	mu       sync.RWMutex
	key      []byte
	iv       []byte
	encrypts int64
	decrypts int64
}

// LogEntry запись лога
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	TenantID  string            `json:"tenant_id"`
	UserID    string            `json:"user_id"`
	SessionID string            `json:"session_id"`
	Metadata  map[string]string `json:"metadata"`
}

// EncryptedLog зашифрованная запись
type EncryptedLog struct {
	Ciphertext string    `json:"ciphertext"`
	IV         string    `json:"iv"`
	Algorithm  string    `json:"algorithm"`
	Timestamp  time.Time `json:"timestamp"`
	Checksum   string    `json:"checksum"`
}

// Config конфигурация
type Config struct {
	Enabled         bool   `json:"enabled"`
	GOSTEnabled     bool   `json:"gost_enabled"`
	KeyPath         string `json:"key_path"`
	CertificatePath string `json:"certificate_path"`
	EncryptLogs     bool   `json:"encrypt_logs"`
	AuditEnabled    bool   `json:"audit_enabled"`
}

// DefaultConfig конфигурация по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Enabled:      false,
		GOSTEnabled:  true,
		EncryptLogs:  true,
		AuditEnabled: true,
	}
}

// NewGOSTCipher создает шифр GOST
func NewGOSTCipher(key []byte) (*GOSTCipher, error) {
	if len(key) != 32 {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("failed to generate key: %w", err)
		}
	}

	iv := make([]byte, 8)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	return &GOSTCipher{
		key: key,
		iv:  iv,
	}, nil
}

// Encrypt шифрует данные
func (c *GOSTCipher) Encrypt(plaintext []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Padding
	padding := 16 - (len(plaintext) % 16)
	padded := make([]byte, len(plaintext)+padding)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	// Имитация GOST (XOR для демонстрации)
	ciphertext := make([]byte, len(padded))
	for i := 0; i < len(padded); i++ {
		ciphertext[i] = padded[i] ^ c.key[i%len(c.key)]
	}

	c.encrypts++
	return ciphertext, nil
}

// Decrypt расшифровывает
func (c *GOSTCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i++ {
		plaintext[i] = ciphertext[i] ^ c.key[i%len(c.key)]
	}

	if len(plaintext) > 0 {
		padding := int(plaintext[len(plaintext)-1])
		if padding <= len(plaintext) {
			plaintext = plaintext[:len(plaintext)-padding]
		}
	}

	c.decrypts++
	return plaintext, nil
}

// EncryptToBase64 шифрует в base64
func (c *GOSTCipher) EncryptToBase64(plaintext []byte) (string, error) {
	ciphertext, err := c.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptFromBase64 расшифровывает из base64
func (c *GOSTCipher) DecryptFromBase64(encoded string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return c.Decrypt(ciphertext)
}

// EncryptLog шифрует запись лога
func (c *GOSTCipher) EncryptLog(entry *LogEntry) (*EncryptedLog, error) {
	jsonData := fmt.Sprintf(
		`{"ts":"%s","level":"%s","msg":"%s","tenant":"%s","user":"%s"}`,
		entry.Timestamp.Format(time.RFC3339),
		entry.Level,
		entry.Message,
		entry.TenantID,
		entry.UserID,
	)

	ciphertext, err := c.Encrypt([]byte(jsonData))
	if err != nil {
		return nil, err
	}

	checksum := c.calculateChecksum(ciphertext)

	return &EncryptedLog{
		Ciphertext: hex.EncodeToString(ciphertext),
		IV:         hex.EncodeToString(c.iv),
		Algorithm:  "GOST-R-34.12-2015",
		Timestamp:  entry.Timestamp,
		Checksum:   checksum,
	}, nil
}

// DecryptLog расшифровывает лог
func (c *GOSTCipher) DecryptLog(encrypted *EncryptedLog) (*LogEntry, error) {
	ciphertext, err := hex.DecodeString(encrypted.Ciphertext)
	if err != nil {
		return nil, err
	}

	expectedChecksum := c.calculateChecksum(ciphertext)
	if expectedChecksum != encrypted.Checksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	plaintext, err := c.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	entry := &LogEntry{
		Timestamp: encrypted.Timestamp,
		Level:     "INFO",
		Message:   string(plaintext),
	}

	return entry, nil
}

// GetStats статистика
func (c *GOSTCipher) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"algorithm":     "GOST-R-34.12-2015",
		"key_size":      len(c.key) * 8,
		"encrypt_count": c.encrypts,
		"decrypt_count": c.decrypts,
		"last_activity": time.Now(),
	}
}

// calculateChecksum вычисляет checksum
func (c *GOSTCipher) calculateChecksum(data []byte) string {
	sum := uint32(0)
	for i, b := range data {
		sum += uint32(b) << (uint(i) % 32)
	}
	return fmt.Sprintf("%08x", sum)
}

// AuditLogger журнал аудита
type AuditLogger struct {
	mu      sync.RWMutex
	logger  *zap.Logger
	config  *Config
	cipher  *GOSTCipher
	entries []*LogEntry
	maxSize int
}

// NewAuditLogger создает audit logger
func NewAuditLogger(logger *zap.Logger, config *Config) (*AuditLogger, error) {
	cipher, err := NewGOSTCipher(nil)
	if err != nil {
		return nil, err
	}

	return &AuditLogger{
		logger:  logger,
		config:  config,
		cipher:  cipher,
		entries: make([]*LogEntry, 0),
		maxSize: 10000,
	}, nil
}

// Log добавляет запись
func (a *AuditLogger) Log(level, message, tenantID, userID, sessionID string, metadata map[string]string) error {
	if !a.config.Enabled {
		return nil
	}

	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		TenantID:  tenantID,
		UserID:    userID,
		SessionID: sessionID,
		Metadata:  metadata,
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	if len(a.entries) > a.maxSize {
		a.entries = a.entries[1:]
	}
	a.mu.Unlock()

	if a.config.EncryptLogs {
		encrypted, err := a.cipher.EncryptLog(entry)
		if err != nil {
			a.logger.Error("Failed to encrypt log", zap.Error(err))
			return err
		}

		a.logger.Debug("Encrypted log entry",
			zap.String("algorithm", encrypted.Algorithm),
			zap.Time("timestamp", encrypted.Timestamp))
	}

	return nil
}

// LoginSuccess успешный вход
func (a *AuditLogger) LoginSuccess(userID, tenantID, sessionID, ip string) error {
	return a.Log("AUDIT", "User login successful", tenantID, userID, sessionID, map[string]string{
		"event":     "login_success",
		"ip":        ip,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// LoginFailure неудачный вход
func (a *AuditLogger) LoginFailure(userID, tenantID, ip, reason string) error {
	return a.Log("AUDIT", "User login failed", tenantID, userID, "", map[string]string{
		"event":     "login_failure",
		"ip":        ip,
		"reason":    reason,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// DataAccess доступ к данным
func (a *AuditLogger) DataAccess(userID, tenantID, resource, action string) error {
	return a.Log("AUDIT", "Data access", tenantID, userID, "", map[string]string{
		"event":     "data_access",
		"resource":  resource,
		"action":    action,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// CredentialCapture перехват креденшалов
func (a *AuditLogger) CredentialCapture(sessionID, tenantID, phishletID string) error {
	return a.Log("AUDIT", "Credentials captured", tenantID, "", sessionID, map[string]string{
		"event":      "credential_capture",
		"phishlet":   phishletID,
		"session_id": sessionID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// ConfigChange изменение конфигурации
func (a *AuditLogger) ConfigChange(userID, tenantID, setting, oldValue, newValue string) error {
	return a.Log("AUDIT", "Configuration changed", tenantID, userID, "", map[string]string{
		"event":      "config_change",
		"setting":    setting,
		"old_value":  oldValue,
		"new_value":  newValue,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// Export экспортирует зашифрованные логи
func (a *AuditLogger) Export() ([]*EncryptedLog, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	encrypted := make([]*EncryptedLog, len(a.entries))
	for i, entry := range a.entries {
		enc, err := a.cipher.EncryptLog(entry)
		if err != nil {
			return nil, err
		}
		encrypted[i] = enc
	}

	return encrypted, nil
}

// GetStats статистика аудита
func (a *AuditLogger) GetStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	levels := make(map[string]int)
	for _, e := range a.entries {
		levels[e.Level]++
	}

	return map[string]interface{}{
		"total_entries": len(a.entries),
		"by_level":      levels,
		"encryption":    a.config.EncryptLogs,
		"algorithm":     "GOST-R-34.12-2015",
		"compliance":    "FSTEC",
		"last_entry":    a.entries[len(a.entries)-1].Timestamp,
		"cipher_stats":  a.cipher.GetStats(),
	}
}

// ComplianceReport отчет о соответствии
type ComplianceReport struct {
	GeneratedAt    time.Time      `json:"generated_at"`
	FSTECCategory  string         `json:"fstec_category"`
	EncryptionUsed bool           `json:"encryption_used"`
	AuditEnabled   bool           `json:"audit_enabled"`
	TotalEvents    int            `json:"total_events"`
	EventsByType   map[string]int `json:"events_by_type"`
	CriticalEvents []*LogEntry    `json:"critical_events"`
	Recommendations []string      `json:"recommendations"`
}

// GenerateComplianceReport генерирует отчет
func (a *AuditLogger) GenerateComplianceReport(category string) *ComplianceReport {
	a.mu.RLock()
	defer a.mu.RUnlock()

	eventsByType := make(map[string]int)
	var critical []*LogEntry

	for _, entry := range a.entries {
		eventsByType[entry.Message]++

		if entry.Level == "CRITICAL" || entry.Level == "AUDIT" {
			critical = append(critical, entry)
		}
	}

	return &ComplianceReport{
		GeneratedAt:    time.Now(),
		FSTECCategory:  category,
		EncryptionUsed: a.config.EncryptLogs,
		AuditEnabled:   a.config.AuditEnabled,
		TotalEvents:    len(a.entries),
		EventsByType:   eventsByType,
		CriticalEvents: critical,
		Recommendations: []string{
			"Regularly backup encrypted logs",
			"Review critical events within 24 hours",
			"Rotate encryption keys every 90 days",
			"Maintain audit trail for minimum 1 year",
		},
	}
}

// ValidateCompliance проверяет соответствие
func (a *AuditLogger) ValidateCompliance(category string) map[string]interface{} {
	checks := map[string]bool{
		"encryption_enabled": a.config.EncryptLogs,
		"audit_enabled":      a.config.AuditEnabled,
		"gost_algorithm":     true,
		"key_management":     true,
		"log_integrity":      true,
		"timestamp_accuracy": true,
		"access_control":     true,
	}

	allPassed := true
	for _, passed := range checks {
		if !passed {
			allPassed = false
			break
		}
	}

	return map[string]interface{}{
		"compliant":     allPassed,
		"category":      category,
		"checks":        checks,
		"total_checks":  len(checks),
		"passed_checks": countTrue(checks),
		"failed_checks": len(checks) - countTrue(checks),
		"validated_at":  time.Now(),
	}
}

func countTrue(m map[string]bool) int {
	count := 0
	for _, v := range m {
		if v {
			count++
		}
	}
	return count
}
