// Package simulation provides a safe, synthetic-event-only awareness lab.
package simulation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidID          = errors.New("invalid campaign or participant id")
	ErrInvalidEvent       = errors.New("unsupported simulation event type")
	ErrCampaignNotInScope = errors.New("campaign is not in the explicit simulation allowlist")
	ErrLabStopped         = errors.New("simulation lab is stopped")
	ErrScopeExpired       = errors.New("simulation scope has expired")
	idPattern             = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	allowedEvents         = map[string]struct{}{
		"visit":             {},
		"click":             {},
		"training_complete": {},
	}
)

// Event intentionally has no arbitrary payload, password, cookie, or token field.
type Event struct {
	Sequence      uint64    `json:"sequence"`
	CampaignID    string    `json:"campaign_id"`
	ParticipantID string    `json:"participant_id"`
	Type          string    `json:"type"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// Store keeps a bounded in-memory history of synthetic training events.
type Store struct {
	mu             sync.RWMutex
	max            int
	nextID         uint64
	events         []Event
	allowlist      map[string]struct{}
	scopeExpiresAt  time.Time
	stopped         bool
	persistencePath string
}

// ValidID reports whether an identifier is safe for use as a synthetic campaign or participant ID.
func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

// NewStore requires an explicit campaign allowlist. With no IDs, all ingestion is denied.
func NewStore(maxEvents int, campaignIDs ...string) *Store {
	return NewStoreWithExpiry(maxEvents, time.Time{}, campaignIDs...)
}

// NewStoreWithExpiry configures a fixed campaign allowlist and an optional scope deadline.
func NewStoreWithExpiry(maxEvents int, scopeExpiresAt time.Time, campaignIDs ...string) *Store {
	if maxEvents < 1 {
		maxEvents = 500
	}
	allowlist := make(map[string]struct{}, len(campaignIDs))
	for _, id := range campaignIDs {
		id = strings.TrimSpace(id)
		if idPattern.MatchString(id) {
			allowlist[id] = struct{}{}
		}
	}
	return &Store{
		max:            maxEvents,
		events:         make([]Event, 0, maxEvents),
		allowlist:      allowlist,
		scopeExpiresAt: scopeExpiresAt,
	}
}

// Stop is an irreversible emergency stop for this process lifetime.
func (s *Store) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
	_ = s.persistLocked()
}

func (s *Store) IsStopped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopped
}

func (s *Store) ScopeExpiresAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scopeExpiresAt
}

func (s *Store) IsScopeExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return !s.scopeExpiresAt.IsZero() && !time.Now().Before(s.scopeExpiresAt)
}

func (s *Store) Add(campaignID, participantID, eventType string) (Event, error) {
	if !ValidID(campaignID) || !ValidID(participantID) {
		return Event{}, ErrInvalidID
	}
	if _, ok := allowedEvents[eventType]; !ok {
		return Event{}, ErrInvalidEvent
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return Event{}, ErrLabStopped
	}
	if !s.scopeExpiresAt.IsZero() && !time.Now().Before(s.scopeExpiresAt) {
		return Event{}, ErrScopeExpired
	}
	if _, ok := s.allowlist[campaignID]; !ok {
		return Event{}, ErrCampaignNotInScope
	}

	previousEvents := append([]Event(nil), s.events...)
	previousID := s.nextID
	s.nextID++
	event := Event{
		Sequence:      s.nextID,
		CampaignID:    campaignID,
		ParticipantID: participantID,
		Type:          eventType,
		OccurredAt:    time.Now().UTC(),
	}
	if len(s.events) == s.max {
		copy(s.events, s.events[1:])
		s.events[len(s.events)-1] = event
	} else {
		s.events = append(s.events, event)
	}
	if err := s.persistLocked(); err != nil {
		s.events = previousEvents
		s.nextID = previousID
		return Event{}, fmt.Errorf("persist simulation event: %w", err)
	}
	return event, nil
}

// List returns newest-first events. A non-positive limit returns an empty slice.
func (s *Store) List(limit int) []Event {
	if limit <= 0 {
		return []Event{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit > len(s.events) {
		limit = len(s.events)
	}
	result := make([]Event, 0, limit)
	for i := len(s.events) - 1; i >= len(s.events)-limit; i-- {
		result = append(result, s.events[i])
	}
	return result
}
