package store

import (
	"errors"
	"sync"

	"github.com/singh-sidharth/helionx-trace/internal/model"
)

var ErrEmptyRequestID = errors.New("requestId is required")

type EventStore interface {
	Add(event model.Event) error
	GetByRequestID(requestID string) ([]model.Event, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	events map[string][]model.Event
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		events: make(map[string][]model.Event),
	}
}

func (s *MemoryStore) Add(event model.Event) error {
	if event.RequestID == "" {
		return ErrEmptyRequestID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events[event.RequestID] = append(s.events[event.RequestID], event)
	return nil
}

func (s *MemoryStore) GetByRequestID(requestID string) ([]model.Event, error) {
	if requestID == "" {
		return nil, ErrEmptyRequestID
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[requestID]
	out := make([]model.Event, len(events))
	copy(out, events)

	return out, nil
}
