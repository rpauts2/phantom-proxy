package simulation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const snapshotVersion = 1
const maxSnapshotBytes = 2 << 20

var ErrInvalidSnapshot = errors.New("invalid simulation snapshot")

type storeSnapshot struct {
	Version        int       `json:"version"`
	NextSequence   uint64    `json:"next_sequence"`
	ScopeExpiresAt time.Time `json:"scope_expires_at,omitempty"`
	Stopped        bool      `json:"stopped"`
	CampaignIDs    []string  `json:"campaign_ids"`
	Events         []Event   `json:"events"`
}

// OpenPersistentStore loads an existing bounded synthetic-event history or creates a new one.
// A saved scope deadline is authoritative on restart and cannot be extended by changing flags.
func OpenPersistentStore(path string, maxEvents int, initialExpiry time.Time, campaignIDs ...string) (*Store, error) {
	if path == "" {
		return nil, errors.New("persistence path is required")
	}
	if initialExpiry.IsZero() {
		return nil, errors.New("simulation scope expiry is required")
	}
	store := NewStoreWithExpiry(maxEvents, initialExpiry, campaignIDs...)
	store.persistencePath = path

	info, statErr := os.Stat(path)
	if statErr == nil && info.Size() > maxSnapshotBytes {
		return nil, fmt.Errorf("%w: snapshot exceeds %d bytes", ErrInvalidSnapshot, maxSnapshotBytes)
	}
	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect simulation snapshot: %w", statErr)
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		store.mu.Lock()
		defer store.mu.Unlock()
		if err := store.persistLocked(); err != nil {
			return nil, fmt.Errorf("create initial snapshot: %w", err)
		}
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read simulation snapshot: %w", err)
	}

	var snapshot storeSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSnapshot, err)
	}
	if snapshot.Version != snapshotVersion || len(snapshot.Events) > store.max {
		return nil, ErrInvalidSnapshot
	}
	expected := sortedCampaignIDs(store.allowlist)
	if !equalStrings(expected, snapshot.CampaignIDs) {
		return nil, errors.New("saved campaign allowlist differs from requested allowlist; use the original IDs or choose a new data path")
	}
	if snapshot.ScopeExpiresAt.IsZero() {
		return nil, fmt.Errorf("%w: missing scope expiry", ErrInvalidSnapshot)
	}

	var lastSequence uint64
	for _, event := range snapshot.Events {
		if !ValidID(event.CampaignID) || !ValidID(event.ParticipantID) {
			return nil, ErrInvalidSnapshot
		}
		if _, ok := store.allowlist[event.CampaignID]; !ok {
			return nil, ErrInvalidSnapshot
		}
		if _, ok := allowedEvents[event.Type]; !ok || event.Sequence <= lastSequence || event.OccurredAt.IsZero() {
			return nil, ErrInvalidSnapshot
		}
		lastSequence = event.Sequence
	}
	if snapshot.NextSequence < lastSequence {
		return nil, ErrInvalidSnapshot
	}

	store.mu.Lock()
	store.scopeExpiresAt = snapshot.ScopeExpiresAt
	store.stopped = snapshot.Stopped
	store.nextID = snapshot.NextSequence
	// Persisted events are stored oldest-first; List returns newest-first.
	store.events = append(store.events[:0], snapshot.Events...)
	store.mu.Unlock()
	return store, nil
}

func (s *Store) persistLocked() error {
	if s.persistencePath == "" {
		return nil
	}
	snapshot := storeSnapshot{
		Version:        snapshotVersion,
		NextSequence:   s.nextID,
		ScopeExpiresAt: s.scopeExpiresAt,
		Stopped:        s.stopped,
		CampaignIDs:    sortedCampaignIDs(s.allowlist),
		Events:         append([]Event(nil), s.events...),
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.persistencePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".simulation-snapshot-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, s.persistencePath); err != nil {
		return err
	}
	return nil
}

func sortedCampaignIDs(allowlist map[string]struct{}) []string {
	ids := make([]string, 0, len(allowlist))
	for id := range allowlist {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
