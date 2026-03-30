import { afterEach, describe, expect, it, vi } from "vitest";

import { createHelionxClient } from "./client.js";

describe("createHelionxClient", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it("sends a POST request to the normalized /events endpoint", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => "",
    });

    vi.stubGlobal("fetch", fetchMock);

    const client = createHelionxClient({
      endpoint: "http://localhost:8080///",
    });

    await client.ingest({
      requestId: "req-1",
      service: "service-a",
      eventType: "handler.completed",
      status: "SUCCESS",
      metadata: {
        path: "/test",
      },
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);

    const [url, requestInit] = fetchMock.mock.calls[0] as [string, RequestInit];
    const body = JSON.parse(String(requestInit.body));

    expect(url).toBe("http://localhost:8080/events");
    expect(requestInit.method).toBe("POST");
    expect(requestInit.headers).toEqual({
      "Content-Type": "application/json",
    });
    expect(body).toMatchObject({
      requestId: "req-1",
      service: "service-a",
      eventType: "handler.completed",
      status: "SUCCESS",
      metadata: {
        path: "/test",
      },
    });
  });

  it("includes the Authorization header when apiKey is provided", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => "",
    });

    vi.stubGlobal("fetch", fetchMock);

    const client = createHelionxClient({
      endpoint: "http://localhost:8080",
      apiKey: "secret-key",
    });

    await client.ingest({
      requestId: "req-2",
      service: "service-a",
      eventType: "http.request.finished",
      status: "SUCCESS",
    });

    const [, requestInit] = fetchMock.mock.calls[0] as [string, RequestInit];

    expect(requestInit.headers).toEqual({
      "Content-Type": "application/json",
      Authorization: "Bearer secret-key",
    });
  });

  it("throws a descriptive error when the backend responds with non-2xx", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      text: async () => "internal error",
    });

    vi.stubGlobal("fetch", fetchMock);

    const client = createHelionxClient({
      endpoint: "http://localhost:8080",
    });

    await expect(
      client.ingest({
        requestId: "req-3",
        service: "service-a",
        eventType: "http.request.finished",
        status: "FAILED",
      }),
    ).rejects.toThrow("Helionx ingest failed: 500 internal error");
  });

  it("times out when fetch does not complete within timeoutMs", async () => {
    vi.useFakeTimers();

    const fetchMock = vi.fn((_: string, requestInit?: RequestInit) => {
      return new Promise((_, reject) => {
        requestInit?.signal?.addEventListener("abort", () => {
          const abortError = new Error("aborted");
          abortError.name = "AbortError";
          reject(abortError);
        });
      });
    });

    vi.stubGlobal("fetch", fetchMock);

    const client = createHelionxClient({
      endpoint: "http://localhost:8080",
      timeoutMs: 25,
    });

    const ingestPromise = client.ingest({
      requestId: "req-4",
      service: "service-a",
      eventType: "http.request.started",
      status: "STARTED",
    });

    const assertion = expect(ingestPromise).rejects.toThrow(
      "Helionx ingest request timed out",
    );

    await vi.advanceTimersByTimeAsync(25);
    await assertion;
  });

  it("rejects invalid events before calling fetch", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    const client = createHelionxClient({
      endpoint: "http://localhost:8080",
    });

    await expect(
      client.ingest({
        requestId: "",
        service: "service-a",
        eventType: "http.request.started",
        status: "STARTED",
      }),
    ).rejects.toThrow();

    expect(fetchMock).not.toHaveBeenCalled();
  });
});