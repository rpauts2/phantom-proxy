package events

import (
	"context"
	"testing"
	"time"
)

func TestNewBus(t *testing.T) {
	bus := NewBus()
	if bus == nil {
		t.Fatal("Expected event bus to be created")
	}
}

func TestPublishSubscribe(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	received := make(chan string, 10)

	// Subscribe to test events
	bus.Subscribe("test.event", func(ctx context.Context, eventType string, payload EventPayload) error {
		received <- eventType
		return nil
	})

	// Publish event
	bus.Publish(ctx, "test.event", map[string]interface{}{"key": "value"})

	// Wait for event
	select {
	case eventType := <-received:
		if eventType != "test.event" {
			t.Errorf("Expected event type test.event, got %s", eventType)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	received1 := make(chan string, 10)
	received2 := make(chan string, 10)

	// Subscribe twice
	bus.Subscribe("test.event", func(ctx context.Context, eventType string, payload EventPayload) error {
		received1 <- eventType
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, eventType string, payload EventPayload) error {
		received2 <- eventType
		return nil
	})

	// Publish event
	bus.Publish(ctx, "test.event", map[string]interface{}{"test": "data"})

	// Both subscribers should receive event
	select {
	case <-received1:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("First subscriber didn't receive event")
	}

	select {
	case <-received2:
		// OK
	case <-time.After(1 * time.Second):
		t.Fatal("Second subscriber didn't receive event")
	}
}

func TestDifferentEventTypes(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	received := make(chan string, 10)

	// Subscribe to specific event type
	bus.Subscribe("credential.captured", func(ctx context.Context, eventType string, payload EventPayload) error {
		received <- eventType
		return nil
	})

	// Publish different event types
	bus.Publish(ctx, "session.created", nil)
	bus.Publish(ctx, "credential.captured", map[string]interface{}{"user": "test"})
	bus.Publish(ctx, "session.closed", nil)

	// Should only receive credential.captured
	select {
	case eventType := <-received:
		if eventType != "credential.captured" {
			t.Errorf("Expected credential.captured, got %s", eventType)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for event")
	}

	// No more events
	select {
	case eventType := <-received:
		t.Errorf("Unexpected event: %s", eventType)
	case <-time.After(100 * time.Millisecond):
		// OK - no more events
	}
}

func TestEventConstants(t *testing.T) {
	// Test that event constants are defined
	if EventCredentialCaptured == "" {
		t.Error("EventCredentialCaptured should be defined")
	}
	if EventSessionCreated == "" {
		t.Error("EventSessionCreated should be defined")
	}
	if EventSessionCaptured == "" {
		t.Error("EventSessionCaptured should be defined")
	}
}

func TestCredentialEvent(t *testing.T) {
	event := CredentialEvent{
		SessionID: "test-1",
		Username:  "testuser",
		Password:  "testpass",
		VictimIP:  "192.168.1.1",
	}

	if event.SessionID != "test-1" {
		t.Errorf("Expected SessionID test-1, got %s", event.SessionID)
	}
	if event.Username != "testuser" {
		t.Errorf("Expected Username testuser, got %s", event.Username)
	}
}

func TestSessionEvent(t *testing.T) {
	event := SessionEvent{
		SessionID: "test-1",
		VictimIP:  "192.168.1.1",
		TargetURL: "https://example.com",
		State:     "active",
	}

	if event.TargetURL != "https://example.com" {
		t.Errorf("Expected TargetURL https://example.com, got %s", event.TargetURL)
	}
}
