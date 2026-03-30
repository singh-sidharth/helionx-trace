# Helionx Trace (MVP)

## 🚀 Debug distributed requests in one call

Most systems require digging through logs across multiple services.

Helionx Trace gives you:

> A single timeline showing what happened, where it failed, and how it recovered.

---

## ⚡ Quick Demo

```bash
go mod tidy
STORE_BACKEND=memory go run ./cmd/server
./scripts/test.sh add
./scripts/test.sh summary
```

---

## 🧾 Example Output (Human Readable)

```text
Request ID: req-2
Status: SUCCESS_AFTER_RETRY
Failure Point: payment.charge
Retry Count: 1
Total Duration: 8.093s

Timeline
--------
1. [api] request.received -> SUCCESS
2. [order] order.created -> SUCCESS (+1.015s)
3. [payment] charge -> FAILED (+2.024s) error=stripe timeout
4. [payment] charge -> RETRY (+3.028s) [retry]
5. [payment] charge -> SUCCESS (+2.025s)
```

---

## 🧩 How it works

Client → API → EventStore → Timeline Engine → Summary Output

---

## 🚀 What this MVP does

- Ingest events via HTTP
- Group events by `requestId`
- Reconstruct ordered timeline
- Detect failures and retries
- Compute latency between steps
- Provide a summarized debugging view

---

## 📦 API Endpoints

### Health

```
GET /health
```

---

### Ingest Event

```
POST /events
```

Example:

```json
{
  "requestId": "req-1",
  "service": "payment",
  "eventType": "charge",
  "status": "FAILED",
  "metadata": {
    "error": "stripe timeout"
  }
}
```

---

### Debug Timeline

```
GET /debug/{requestId}
```

Example Response:

```json
{
  "requestId": "req-2",
  "status": "SUCCESS_AFTER_RETRY",
  "totalEvents": 5,
  "retryCount": 1,
  "totalDurationMs": 8122,
  "failurePoint": "payment.charge",
  "timeline": [ ... ]
}
```

---

### Debug Summary (Human Readable)

```
GET /debug/{requestId}/summary
```

Example Output:

```text
Request ID: req-2
Status: SUCCESS_AFTER_RETRY
Failure Point: payment.charge
Retry Count: 1
Total Duration: 8.093s

Timeline
--------
1. [api] request.received -> SUCCESS
2. [order] order.created -> SUCCESS (+1.015s)
3. [payment] charge -> FAILED (+2.024s) error=stripe timeout
4. [payment] charge -> RETRY (+3.028s) [retry]
5. [payment] charge -> SUCCESS (+2.025s)
```

---

## 🧪 Local Setup

Default: in-memory (no persistence)

```bash
go mod tidy
STORE_BACKEND=memory go run ./cmd/server
```

---

## 🐳 Docker (Postgres)

Start Postgres locally with Docker Compose:

```bash
make up
```

Run the server using Postgres:

```bash
STORE_BACKEND=postgres go run ./cmd/server
```

Check status and logs:

```bash
make ps
make logs
```

Connect to the database:

```bash
make psql
```

Reset database (re-runs init.sql on next start):

```bash
make db-reset
make up
```

Default DB config:

- host: `localhost`
- port: `5432`
- database: `helionx`
- user: `helionx`
- password: `helionx`

> Note: `db/init.sql` is auto-applied on first startup via Docker.

---

## 🧪 Testing

Use the provided script to simulate flows:

```bash
./scripts/test.sh add
./scripts/test.sh summary
./scripts/test.sh timeline
```

Or run everything:

```bash
./scripts/test.sh all
```

You can also override the request ID:

```bash
REQUEST_ID=req-7 ./scripts/test.sh add
```

---

## 🧩 Current Architecture (MVP)

- Go HTTP server
- EventStore interface
  - InMemoryStore (default)
  - PostgresStore (persistent storage)
- Timeline reconstruction service
- Optional Postgres via Docker Compose

⚠️ In-memory mode loses data on restart. Use Postgres for persistence.

---

## 💡 Why this matters

Modern distributed systems are hard to debug.

Helionx Trace reduces debugging from:

> "search logs across services"

To:

> "fetch one request timeline"

---

## 📌 What Helionx Trace Answers

- Where did the request fail?
- Did it recover?
- How long did each step take?
- How many retries occurred?

---

## 🛠️ Next Improvements

- [x] Docker Compose (Postgres) + init.sql
- [x] Human-readable summary output
- [ ] Better retry detection logic
- [ ] UI for timeline visualization
- [ ] Integration with Kafka / event streams

---

## 🧪 Example Debug Story

Scenario:
Payment fails due to a Stripe timeout, system retries and eventually succeeds.

Helionx Trace Output:
- Failure detected at `payment.charge`
- Retry triggered after ~3 seconds
- Successful recovery after retry
- Total request duration ~8 seconds

This reduces debugging from minutes of log searching to a single API call.

---

## 🧱 Status

MVP working:
- ingestion ✅
- timeline reconstruction ✅
- failure detection ✅
- retry tracking ✅

Next focus: usability + persistence

---