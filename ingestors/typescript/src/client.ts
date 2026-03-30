

import { z } from "zod";
import type { HelionxClientOptions, HelionxEvent } from "./types.js";

const HelionxEventSchema = z.object({
  requestId: z.string().min(1),
  service: z.string().min(1),
  eventType: z.string().min(1),
  status: z.enum(["STARTED", "SUCCESS", "FAILED"]),
  timestamp: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

export function createHelionxClient(options: HelionxClientOptions) {
  const baseUrl = options.endpoint.replace(/\/+$/, "");

  return {
    async ingest(event: HelionxEvent): Promise<void> {
      const parsed = HelionxEventSchema.parse(event);

      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), options.timeoutMs ?? 5000);

      try {
        const res = await fetch(`${baseUrl}/events`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            ...(options.apiKey ? { Authorization: `Bearer ${options.apiKey}` } : {}),
          },
          body: JSON.stringify(parsed),
          signal: controller.signal,
        });

        if (!res.ok) {
          const text = await res.text();
          throw new Error(`Helionx ingest failed: ${res.status} ${text}`);
        }
      } catch (err) {
        if ((err as Error).name === "AbortError") {
          throw new Error("Helionx ingest request timed out");
        }
        throw err;
      } finally {
        clearTimeout(timeout);
      }
    },
  };
}