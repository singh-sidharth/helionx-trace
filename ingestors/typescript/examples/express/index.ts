import express from "express";
import { createHelionxClient } from "../../dist/index.js";

const app = express();
app.use(express.json());

const client = createHelionxClient({
  endpoint: "http://localhost:8080",
});

app.post("/test", async (req, res) => {
  try {
    await client.ingest({
      requestId: "req-1",
      service: "express-service",
      eventType: "http.request",
      status: "SUCCESS",
      metadata: {
        path: "/test",
      },
    });

    res.json({ ok: true });
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: "failed to ingest" });
  }
});

app.listen(3000, () => {
  console.log("Example app running on http://localhost:3000");
});