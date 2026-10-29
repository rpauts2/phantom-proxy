// Package simulation provides a safe, synthetic-event-only awareness lab.
package simulation

import (
	"errors"
	"regexp"
	"sync"
	"time"
)

var (
	ErrInvalidID    = errors.New("invalid campaign or participant id")
	ErrInvalidEvent = errors.New("unsupported simulation event type")
	idPattern       = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	allowedEvents   = map[string]struct{}{
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
	mu     sync.RWMutex
	max    int
	nextID uint64
	events []Event
}

func NewStore(maxEvents int) *Store {
	if maxEvents < 1 {
		maxEvents = 500
	}
	return &Store{max: maxEvents, events: make([]Event, 0, maxEvents)}
}

func (s *Store) Add(campaignID, participantID, eventType string) (Event, error) {
	if !idPattern.MatchString(campaignID) || !idPattern.MatchString(participantID) {
		return Event{}, ErrInvalidID
	}
	if _, ok := allowedEvents[eventType]; !ok {
		return Event{}, ErrInvalidEvent
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
