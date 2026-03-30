import { createHelionxClient } from "./dist/index.js";

const client = createHelionxClient({
  endpoint: "http://localhost:8080",
});

await client.ingest({
  requestId: "req-1",
  service: "test-service",
  eventType: "test.event",
  status: "SUCCESS",
  metadata: { hello: "world" },
});

console.log("sent");