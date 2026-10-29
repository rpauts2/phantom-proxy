package simulation

import (
	"errors"
	"testing"
)

func TestStoreValidatesAndBoundsEvents(t *testing.T) {
	store := NewStore(2)
	if _, err := store.Add("campaign", "participant", "password_captured"); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected unsupported event type error, got %v", err)
	}
	if _, err := store.Add("../campaign", "participant", "click"); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected invalid id error, got %v", err)
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
	store := NewStore(3)
	_, _ = store.Add("campaign", "participant", "visit")

	first := store.List(1)
	first[0].Type = "changed"
	second := store.List(1)
	if second[0].Type != "visit" {
		t.Fatal("mutating returned slice changed stored event")
	}
}
