package service

import (
	"sort"

	"github.com/singh-sidharth/helionx-trace/internal/model"
	"github.com/singh-sidharth/helionx-trace/internal/store"
)

type TimelineService struct {
	store store.EventStore
}

func NewTimelineService(store store.EventStore) *TimelineService {
	return &TimelineService{store: store}
}

func (s *TimelineService) Build(requestID string) (model.TimelineResponse, error) {
	events, err := s.store.GetByRequestID(requestID)
	if err != nil {
		return model.TimelineResponse{}, err
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	steps := make([]model.TimelineStep, 0, len(events))
	seen := make(map[string]int)

	var prev *model.Event
	var firstFailed *model.TimelineStep
	var retryCount int
	var failurePoint string
	hasFailure := false
	finalSucceeded := false

	for _, e := range events {
		key := e.Service + "|" + e.EventType
		seen[key]++

		isRetry := e.Status == model.StatusRetry
		if isRetry {
			retryCount++
		}

		step := model.TimelineStep{
			Service:   e.Service,
			EventType: e.EventType,
			Status:    e.Status,
			Timestamp: e.Timestamp,
			IsRetry:   isRetry,
			Metadata:  e.Metadata,
		}

		if prev != nil {
			step.DeltaMs = e.Timestamp.Sub(prev.Timestamp).Milliseconds()
		}

		if firstFailed == nil && e.Status == model.StatusFailed {
			copyStep := step
			firstFailed = &copyStep
			failurePoint = e.Service + "." + e.EventType
			hasFailure = true
		}

		if e.Status == model.StatusSuccess {
			finalSucceeded = true
		}

		steps = append(steps, step)
		prev = &e
	}

	var totalDurationMs int64
	if len(events) > 1 {
		totalDurationMs = events[len(events)-1].Timestamp.Sub(events[0].Timestamp).Milliseconds()
	}

	status := deriveOverallStatus(hasFailure, finalSucceeded, retryCount)

	return model.TimelineResponse{
		RequestID:       requestID,
		Status:          status,
		TotalEvents:     len(steps),
		RetryCount:      retryCount,
		TotalDurationMs: totalDurationMs,
		FailurePoint:    failurePoint,
		FirstFailed:     firstFailed,
		Timeline:        steps,
	}, nil
}

func deriveOverallStatus(hasFailure, finalSucceeded bool, retryCount int) string {
	switch {
	case hasFailure && finalSucceeded && retryCount > 0:
		return "SUCCESS_AFTER_RETRY"
	case hasFailure && !finalSucceeded:
		return "FAILED"
	case finalSucceeded:
		return "SUCCESS"
	default:
		return "UNKNOWN"
	}
}
