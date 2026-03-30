package store

import (
	"testing"
	"time"

	"github.com/singh-sidharth/helionx-trace/internal/model"
)

func TestMemoryStoreAddAppendsEventsForSameRequestID(t *testing.T) {
	store := NewMemoryStore()

	firstEvent := model.Event{
		RequestID: "req-1",
		Service:   "service-a",
		EventType: "http.request.started",
		Status:    "STARTED",
		Timestamp: time.Date(2026, 3, 30, 18, 24, 17, 0, time.UTC),
	}

	secondEvent := model.Event{
		RequestID: "req-1",
		Service:   "service-a",
		EventType: "http.request.finished",
		Status:    "SUCCESS",
		Timestamp: time.Date(2026, 3, 30, 18, 24, 18, 0, time.UTC),
	}

	if err := store.Add(firstEvent); err != nil {
		t.Fatalf("Add(firstEvent) returned error: %v", err)
	}

	if err := store.Add(secondEvent); err != nil {
		t.Fatalf("Add(secondEvent) returned error: %v", err)
	}

	events, err := store.GetByRequestID("req-1")
	if err != nil {
		t.Fatalf("GetByRequestID(req-1) returned error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].EventType != "http.request.started" {
		t.Fatalf("expected first event type %q, got %q", "http.request.started", events[0].EventType)
	}

	if events[1].EventType != "http.request.finished" {
		t.Fatalf("expected second event type %q, got %q", "http.request.finished", events[1].EventType)
	}
}

func TestMemoryStoreAddReturnsErrorWhenRequestIDIsEmpty(t *testing.T) {
	store := NewMemoryStore()

	err := store.Add(model.Event{})
	if err == nil {
		t.Fatal("expected error when requestId is empty, got nil")
	}

	if err != ErrEmptyRequestID {
		t.Fatalf("expected ErrEmptyRequestID, got %v", err)
	}
}

func TestMemoryStoreGetByRequestIDReturnsCopy(t *testing.T) {
	store := NewMemoryStore()

	event := model.Event{
		RequestID: "req-2",
		Service:   "service-a",
		EventType: "handler.completed",
		Status:    "SUCCESS",
		Timestamp: time.Date(2026, 3, 30, 18, 24, 19, 0, time.UTC),
	}

	if err := store.Add(event); err != nil {
		t.Fatalf("Add(event) returned error: %v", err)
	}

	firstRead, err := store.GetByRequestID("req-2")
	if err != nil {
		t.Fatalf("GetByRequestID(req-2) first read returned error: %v", err)
	}

	firstRead[0].EventType = "mutated.event"

	secondRead, err := store.GetByRequestID("req-2")
	if err != nil {
		t.Fatalf("GetByRequestID(req-2) second read returned error: %v", err)
	}

	if secondRead[0].EventType != "handler.completed" {
		t.Fatalf("expected stored event type to remain %q, got %q", "handler.completed", secondRead[0].EventType)
	}
}
