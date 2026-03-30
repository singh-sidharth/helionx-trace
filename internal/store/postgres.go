package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/singh-sidharth/helionx-trace/internal/model"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Add(event model.Event) error {
	if event.RequestID == "" {
		return ErrEmptyRequestID
	}

	metadataJSON, err := marshalMetadata(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO events (request_id, service, event_type, status, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = s.db.ExecContext(
		ctx,
		query,
		event.RequestID,
		event.Service,
		event.EventType,
		string(event.Status),
		event.Timestamp,
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetByRequestID(requestID string) ([]model.Event, error) {
	if requestID == "" {
		return nil, ErrEmptyRequestID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT request_id, service, event_type, status, timestamp, metadata
		FROM events
		WHERE request_id = $1
		ORDER BY timestamp ASC, id ASC
	`

	rows, err := s.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("query events by request id: %w", err)
	}
	defer rows.Close()

	events := make([]model.Event, 0)
	for rows.Next() {
		var event model.Event
		var status string
		var metadataBytes []byte

		if err := rows.Scan(
			&event.RequestID,
			&event.Service,
			&event.EventType,
			&status,
			&event.Timestamp,
			&metadataBytes,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		event.Status = model.EventStatus(status)

		metadata, err := unmarshalMetadata(metadataBytes)
		if err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
		event.Metadata = metadata

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}

func marshalMetadata(metadata map[string]interface{}) ([]byte, error) {
	if metadata == nil {
		return []byte("{}"), nil
	}

	return json.Marshal(metadata)
}

func unmarshalMetadata(data []byte) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
