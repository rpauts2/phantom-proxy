package simulation

import (
	"errors"
	"testing"
	"time"
)

func TestStoreValidatesAndBoundsEvents(t *testing.T) {
	store := NewStore(2, "campaign-1")
	if _, err := store.Add("campaign-1", "participant", "password_captured"); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected unsupported event type error, got %v", err)
	}
	if _, err := store.Add("../campaign", "participant", "click"); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected invalid id error, got %v", err)
	}
	if _, err := store.Add("outside-scope", "participant", "click"); !errors.Is(err, ErrCampaignNotInScope) {
		t.Fatalf("expected out-of-scope error, got %v", err)
	}

	for _, eventType := range []string{"visit", "click", "training_complete"} {
		if _, err := store.Add("campaign-1", "user_1", eventType); err != nil {
			t.Fatalf("Add(%q) error = %v", eventType, err)
		}
	}

	got := store.List(10)
	if len(got) != 2 {
		t.Fatalf("List returned %d events, want 2", len(got))
	}
	if got[0].Type != "training_complete" || got[1].Type != "click" {
		t.Fatalf("expected newest-first retained events, got %#v", got)
	}
	if got[0].Sequence <= got[1].Sequence {
		t.Fatalf("sequence should increase, got %#v", got)
	}
}

func TestStoreListReturnsIndependentSlice(t *testing.T) {
	store := NewStore(3, "campaign")
	_, _ = store.Add("campaign", "participant", "visit")

	first := store.List(1)
	first[0].Type = "changed"
	second := store.List(1)
	if second[0].Type != "visit" {
		t.Fatal("mutating returned slice changed stored event")
	}
}

func TestStoreRequiresExplicitAllowlistAndSupportsEmergencyStop(t *testing.T) {
	unscoped := NewStore(10)
	if _, err := unscoped.Add("demo", "person", "visit"); !errors.Is(err, ErrCampaignNotInScope) {
		t.Fatalf("empty allowlist should deny ingestion, got %v", err)
	}

	store := NewStore(10, "demo")
	if _, err := store.Add("demo", "person", "visit"); err != nil {
		t.Fatalf("in-scope event rejected: %v", err)
	}
	store.Stop()
	if !store.IsStopped() {
		t.Fatal("stop state was not set")
	}
	if _, err := store.Add("demo", "person", "click"); !errors.Is(err, ErrLabStopped) {
		t.Fatalf("expected emergency stop to reject event, got %v", err)
	}
}

func TestStoreRejectsExpiredScope(t *testing.T) {
	store := NewStoreWithExpiry(10, time.Now().Add(-time.Second), "demo")
	if _, err := store.Add("demo", "person", "visit"); !errors.Is(err, ErrScopeExpired) {
		t.Fatalf("expected expired scope error, got %v", err)
	}
	if !store.IsScopeExpired() {
		t.Fatal("expired scope should report expired state")
	}
}
