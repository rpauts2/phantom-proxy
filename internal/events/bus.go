package events

import (
	"context"
	"sync"
	"time"
)

// Event types for PhantomProxy v13
const (
	EventCredentialCaptured = "credential.captured"
	EventSessionCreated     = "session.created"
	EventSessionCaptured    = "session.captured"
	EventCookieCaptured     = "cookie.captured"
	EventVictimLanded       = "victim.landed"
	EventExfilStarted       = "exfil.started"
	EventExfilCompleted     = "exfil.completed"
	EventPayloadGenerated   = "payload.generated"
	EventC2BeaconReceived   = "c2.beacon"
)

// Event payload interface
type EventPayload interface{}

// CredentialEvent payload
type CredentialEvent struct {
	SessionID   string            `json:"session_id"`
	Username    string            `json:"username"`
	Password    string            `json:"password"`
	CustomFields map[string]string `json:"custom_fields"`
	PhishletID  string            `json:"phishlet_id"`
	VictimIP    string            `json:"victim_ip"`
	Timestamp   time.Time         `json:"timestamp"`
}

// SessionEvent payload
type SessionEvent struct {
	SessionID  string    `json:"session_id"`
	VictimIP   string    `json:"victim_ip"`
	TargetURL  string    `json:"target_url"`
	UserAgent  string    `json:"user_agent"`
	PhishletID string    `json:"phishlet_id"`
	State      string    `json:"state"`
	Timestamp  time.Time `json:"timestamp"`
}

// Handler function type
type Handler func(ctx context.Context, eventType string, payload EventPayload) error

// Bus implements event bus for module communication
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewBus creates new event bus
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe adds handler for event type
func (b *Bus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// Publish sends event to all subscribers
func (b *Bus) Publish(ctx context.Context, eventType string, payload EventPayload) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers[eventType]))
	copy(handlers, b.handlers[eventType])
	b.mu.RUnlock()

	for _, h := range handlers {
		go func(handler Handler) {
			_ = handler(ctx, eventType, payload)
		}(h)
	}
}

// Module interface for v13 plugins
type Module interface {
	Name() string
	Init(ctx context.Context, bus *Bus) error
	Shutdown(ctx context.Context) error
}
