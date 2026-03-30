CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    request_id TEXT NOT NULL,
    service TEXT NOT NULL,
    event_type TEXT NOT NULL,
    status TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_events_request_id ON events(request_id);
CREATE INDEX IF NOT EXISTS idx_events_request_id_timestamp ON events(request_id, timestamp);