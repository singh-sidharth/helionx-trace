

# Helionx Trace Backend (Go Service)

This service is the core backend for Helionx Trace.

It is responsible for:
- event ingestion
- event storage (in-memory or Postgres)
- timeline reconstruction
- debugging APIs for request flows

## Project Structure

- `cmd/server` → entrypoint for the service
- `internal/api` → HTTP handlers
- `internal/service` → core business logic (timeline reconstruction)
- `internal/store` → storage implementations (memory, Postgres)
- `internal/model` → domain models
- `db/init.sql` → database schema
- `scripts/test.sh` → quick test script

## Running the Service

### In-memory mode (default)

```bash
go mod tidy
STORE_BACKEND=memory go run ./cmd/server
```

### Postgres mode

Start Postgres:

```bash
docker-compose up -d
```

Run the server:

```bash
STORE_BACKEND=postgres go run ./cmd/server
```

## Environment Variables

- `STORE_BACKEND` → `memory` or `postgres`
- `PORT` → server port (default: 8080)

## Quick Test

After starting the server:

```bash
./scripts/test.sh add
./scripts/test.sh summary
```

## Notes

- In-memory mode is useful for local development and quick testing.
- Postgres mode enables persistence across restarts.
- This service is part of a larger system that includes ingestion clients (TypeScript, Python) and debugging tooling.