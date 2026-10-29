package simulation

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPersistentStoreRestoresEventsScopeAndStopState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "events.json")
	expiry := time.Now().Add(3 * time.Hour).UTC().Truncate(time.Second)
	store, err := OpenPersistentStore(path, 3, expiry, "demo", "training")
	if err != nil {
		t.Fatalf("OpenPersistentStore() error = %v", err)
	}
	if _, err := store.Add("demo", "person_01", "visit"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if _, err := store.Add("training", "person_02", "training_complete"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if err := store.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("snapshot missing: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("snapshot permissions = %o, want 600", info.Mode().Perm())
	}

	// A later requested expiry must not extend the previously approved scope.
	reopened, err := OpenPersistentStore(path, 3, time.Now().Add(24*time.Hour), "demo", "training")
	if err != nil {
		t.Fatalf("reopen persistent store: %v", err)
	}
	if !reopened.ScopeExpiresAt().Equal(expiry) {
		t.Fatalf("scope expiry changed across restart: got %s, want %s", reopened.ScopeExpiresAt(), expiry)
	}
	events := reopened.List(10)
	if len(events) != 2 || events[0].Type != "training_complete" || events[1].Type != "visit" {
		t.Fatalf("events not restored in expected newest-first order: %#v", events)
	}
	if !reopened.IsStopped() {
		t.Fatal("emergency stop state was not restored")
	}
	if _, err := reopened.Add("demo", "person_03", "click"); !errors.Is(err, ErrLabStopped) {
		t.Fatalf("reopened stopped store accepted event: %v", err)
	}
}

func TestPersistentStoreRejectsChangedAllowlist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	expiry := time.Now().Add(time.Hour)
	if _, err := OpenPersistentStore(path, 10, expiry, "demo"); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenPersistentStore(path, 10, expiry, "other"); err == nil {
		t.Fatal("expected allowlist mismatch to be rejected")
	}
}

func TestPersistentStoreRejectsCorruptSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenPersistentStore(path, 10, time.Now().Add(time.Hour), "demo"); !errors.Is(err, ErrInvalidSnapshot) {
		t.Fatalf("expected invalid snapshot error, got %v", err)
	}
}

func TestPersistentStoreEnforcesRetentionLimitAfterRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	store, err := OpenPersistentStore(path, 2, time.Now().Add(time.Hour), "demo")
	if err != nil {
		t.Fatal(err)
	}
	for _, participant := range []string{"person_01", "person_02", "person_03"} {
		if _, err := store.Add("demo", participant, "visit"); err != nil {
			t.Fatal(err)
		}
	}
	reopened, err := OpenPersistentStore(path, 2, time.Now().Add(time.Hour), "demo")
	if err != nil {
		t.Fatal(err)
	}
	events := reopened.List(10)
	if len(events) != 2 || events[0].ParticipantID != "person_03" || events[1].ParticipantID != "person_02" {
		t.Fatalf("retention limit not respected after restart: %#v", events)
	}
	if reopened.nextID != 3 {
		t.Fatalf("sequence counter = %d, want 3", reopened.nextID)
	}
}

func TestPersistentStoreRejectsOversizedSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	if err := os.WriteFile(path, make([]byte, maxSnapshotBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := OpenPersistentStore(path, 10, time.Now().Add(time.Hour), "demo")
	if !errors.Is(err, ErrInvalidSnapshot) {
		t.Fatalf("expected oversized snapshot to be rejected, got %v", err)
	}
}

func TestPersistentStoreRollsBackEventWhenSnapshotWriteFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.json")
	store, err := OpenPersistentStore(path, 10, time.Now().Add(time.Hour), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Add("demo", "person_01", "visit"); err == nil {
		t.Fatal("expected snapshot write failure")
	}
	if got := len(store.List(10)); got != 0 {
		t.Fatalf("failed durable write left %d in-memory events, want 0", got)
	}
	if store.nextID != 0 {
		t.Fatalf("sequence counter advanced after failed write: %d", store.nextID)
	}
}
