package service

import (
	"testing"
	"time"

	"github.com/singh-sidharth/helionx-trace/internal/model"
	"github.com/singh-sidharth/helionx-trace/internal/store"
)

func TestTimelineServiceBuildOrdersEventsAndCalculatesDeltas(t *testing.T) {
	memoryStore := store.NewMemoryStore()

	earlierEvent := model.Event{
		RequestID: "req-1",
		Service:   "service-a",
		EventType: "event-1",
		Status:    "STARTED",
		Timestamp: time.Date(2026, 3, 30, 18, 0, 1, 0, time.UTC),
	}

	laterEvent := model.Event{
		RequestID: "req-1",
		Service:   "service-a",
		EventType: "event-2",
		Status:    "SUCCESS",
		Timestamp: time.Date(2026, 3, 30, 18, 0, 2, 0, time.UTC),
	}

	// Add out of order so the service has to sort by timestamp.
	if err := memoryStore.Add(laterEvent); err != nil {
		t.Fatalf("Add(laterEvent) returned error: %v", err)
	}

	if err := memoryStore.Add(earlierEvent); err != nil {
		t.Fatalf("Add(earlierEvent) returned error: %v", err)
	}

	timelineService := NewTimelineService(memoryStore)

	resp, err := timelineService.Build("req-1")
	if err != nil {
		t.Fatalf("Build(req-1) returned error: %v", err)
	}

	if len(resp.Timeline) != 2 {
		t.Fatalf("expected 2 timeline steps, got %d", len(resp.Timeline))
	}

	if resp.Timeline[0].EventType != "event-1" {
		t.Fatalf("expected first event type %q, got %q", "event-1", resp.Timeline[0].EventType)
	}

	if resp.Timeline[1].EventType != "event-2" {
		t.Fatalf("expected second event type %q, got %q", "event-2", resp.Timeline[1].EventType)
	}

	if resp.Timeline[0].DeltaMs != 0 {
		t.Fatalf("expected first delta to be 0, got %d", resp.Timeline[0].DeltaMs)
	}

	if resp.Timeline[1].DeltaMs != 1000 {
		t.Fatalf("expected second delta to be 1000, got %d", resp.Timeline[1].DeltaMs)
	}
}

func TestTimelineServiceBuildCalculatesTotalDurationAndOverallStatus(t *testing.T) {
	memoryStore := store.NewMemoryStore()

	startEvent := model.Event{
		RequestID: "req-2",
		Service:   "service-a",
		EventType: "start",
		Status:    "STARTED",
		Timestamp: time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC),
	}

	endEvent := model.Event{
		RequestID: "req-2",
		Service:   "service-a",
		EventType: "end",
		Status:    "SUCCESS",
		Timestamp: time.Date(2026, 3, 30, 18, 0, 5, 0, time.UTC),
	}

	if err := memoryStore.Add(startEvent); err != nil {
		t.Fatalf("Add(startEvent) returned error: %v", err)
	}

	if err := memoryStore.Add(endEvent); err != nil {
		t.Fatalf("Add(endEvent) returned error: %v", err)
	}

	timelineService := NewTimelineService(memoryStore)

	resp, err := timelineService.Build("req-2")
	if err != nil {
		t.Fatalf("Build(req-2) returned error: %v", err)
	}

	if resp.TotalDurationMs != 5000 {
		t.Fatalf("expected total duration 5000ms, got %d", resp.TotalDurationMs)
	}

	if resp.Status != "SUCCESS" {
		t.Fatalf("expected overall status %q, got %q", "SUCCESS", resp.Status)
	}
}

func TestTimelineServiceBuildCountsRetriesAndReturnsSuccessAfterRetry(t *testing.T) {
	memoryStore := store.NewMemoryStore()

	retryEvent := model.Event{
		RequestID: "req-3",
		Service:   "service-a",
		EventType: "db.write",
		Status:    "RETRY",
		Timestamp: time.Date(2026, 3, 30, 18, 0, 1, 0, time.UTC),
	}

	successEvent := model.Event{
		RequestID: "req-3",
		Service:   "service-a",
		EventType: "db.write",
		Status:    "SUCCESS",
		Timestamp: time.Date(2026, 3, 30, 18, 0, 2, 0, time.UTC),
	}

	if err := memoryStore.Add(retryEvent); err != nil {
		t.Fatalf("Add(retryEvent) returned error: %v", err)
	}

	if err := memoryStore.Add(successEvent); err != nil {
		t.Fatalf("Add(successEvent) returned error: %v", err)
	}

	timelineService := NewTimelineService(memoryStore)

	resp, err := timelineService.Build("req-3")
	if err != nil {
		t.Fatalf("Build(req-3) returned error: %v", err)
	}

	if resp.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", resp.RetryCount)
	}

	if resp.Status != "SUCCESS" {
		t.Fatalf("expected overall status %q, got %q", "SUCCESS", resp.Status)
	}
}
