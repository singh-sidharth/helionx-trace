import { describe, expect, it } from "vitest";

import { resolveRequestId } from "./request-id.js";

describe("resolveRequestId", () => {
  it("uses x-request-id when present", () => {
    const resolved = resolveRequestId({
      "x-request-id": "req-123",
    });

    expect(resolved).toEqual({
      requestId: "req-123",
      source: "x-request-id",
    });
  });

  it("falls back to x-correlation-id when x-request-id is missing", () => {
    const resolved = resolveRequestId({
      "x-correlation-id": "corr-456",
    });

    expect(resolved).toEqual({
      requestId: "corr-456",
      source: "x-correlation-id",
    });
  });

  it("extracts trace id from traceparent", () => {
    const resolved = resolveRequestId({
      traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
    });

    expect(resolved).toEqual({
      requestId: "4bf92f3577b34da6a3ce929d0e0e4736",
      source: "traceparent",
    });
  });

  it("extracts root trace id from x-amzn-trace-id", () => {
    const resolved = resolveRequestId({
      "x-amzn-trace-id": "Root=1-67891233-abcdef012345678912345678;Parent=53995c3f42cd8ad8;Sampled=1",
    });

    expect(resolved).toEqual({
      requestId: "1-67891233-abcdef012345678912345678",
      source: "x-amzn-trace-id",
    });
  });

  it("respects precedence order across supported headers", () => {
    const resolved = resolveRequestId({
      "x-request-id": "req-999",
      "x-correlation-id": "corr-999",
      traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
      "x-amzn-trace-id": "Root=1-67891233-abcdef012345678912345678;Parent=53995c3f42cd8ad8;Sampled=1",
    });

    expect(resolved).toEqual({
      requestId: "req-999",
      source: "x-request-id",
    });
  });

  it("ignores empty header values and falls back correctly", () => {
    const resolved = resolveRequestId({
      "x-request-id": "   ",
      "x-correlation-id": "corr-123",
    });

    expect(resolved).toEqual({
      requestId: "corr-123",
      source: "x-correlation-id",
    });
  });

  it("generates a request id when no supported headers exist", () => {
    const resolved = resolveRequestId({});

    expect(resolved.source).toBe("generated");
    expect(resolved.requestId).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i,
    );
  });
});