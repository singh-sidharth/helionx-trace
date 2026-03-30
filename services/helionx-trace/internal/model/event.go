package model

import "time"

type EventStatus string

const (
	StatusSuccess EventStatus = "SUCCESS"
	StatusFailed  EventStatus = "FAILED"
	StatusRetry   EventStatus = "RETRY"
)

type Event struct {
	RequestID string                 `json:"requestId"`
	Service   string                 `json:"service"`
	EventType string                 `json:"eventType"`
	Status    EventStatus            `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type TimelineStep struct {
	Service   string                 `json:"service"`
	EventType string                 `json:"eventType"`
	Status    EventStatus            `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	DeltaMs   int64                  `json:"deltaMsFromPrevious"`
	IsRetry   bool                   `json:"isRetry"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type TimelineResponse struct {
	RequestID       string         `json:"requestId"`
	Status          string         `json:"status"`
	TotalEvents     int            `json:"totalEvents"`
	RetryCount      int            `json:"retryCount"`
	TotalDurationMs int64          `json:"totalDurationMs"`
	FailurePoint    string         `json:"failurePoint,omitempty"`
	FirstFailed     *TimelineStep  `json:"firstFailed,omitempty"`
	Timeline        []TimelineStep `json:"timeline"`
}
