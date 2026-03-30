# Helionx Trace – TypeScript SDK

Helionx Trace lets you capture and debug request flows across services by sending structured events to a backend.

This SDK provides:

- a core ingestion client
- request ID resolution (X-Ray, OTel, API Gateway, etc.)
- an Express middleware adapter

---

## Quick Start (Express)

```ts
import express from "express";
import { createHelionxExpressMiddleware } from "helionx-trace";

const app = express();

app.use(
  createHelionxExpressMiddleware({
    endpoint: "http://localhost:8080",
    service: "my-service",
  }),
);

app.post("/test", async (req, res) => {
  await req.helionx?.success("handler.completed", {
    path: "/test",
  });

  res.json({ ok: true });
});

app.listen(3000);
```

---

## What Happens Automatically

For each request:

- resolves request ID from headers
- generates one if missing
- emits `STARTED` event
- emits final `SUCCESS` / `FAILED`
- attaches `req.helionx`

---

## Request Context API

```ts
req.helionx?.event({
  eventType: "db.query",
  status: "SUCCESS",
});

req.helionx?.success("handler.completed");
req.helionx?.fail("handler.failed");
```

---

## Core Client (Advanced)

```ts
import { createHelionxClient } from "helionx-trace";

const client = createHelionxClient({
  endpoint: "http://localhost:8080",
});

await client.ingest({
  requestId: "req-1",
  service: "my-service",
  eventType: "custom.event",
  status: "SUCCESS",
});
```

---

## Request ID Resolution

Priority order:

1. `x-request-id`
2. `x-correlation-id`
3. `traceparent`
4. `x-amzn-trace-id`
5. generated UUID

---

## Configuration

```ts
createHelionxExpressMiddleware({
  endpoint: string;
  service: string;

  apiKey?: string;
  timeoutMs?: number;

  autoTrackRequestLifecycle?: boolean;
  eventType?: string;
  captureRequestMetadata?: boolean;

  onError?: (err) => void;
});
```

---

## Structure

```text
client.ts        → core ingestion
request-id.ts    → identity resolution
adapters/        → framework integrations
examples/        → usage
```

---

## Status

Early version.
