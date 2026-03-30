import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createHelionxExpressMiddleware,
  type HelionxExpressRequest,
  type HelionxExpressResponse,
  type HelionxRequestContext,
} from "./express.js";

interface MockRequest {
  headers: Record<string, string | string[] | undefined>;
  method?: string;
  path?: string;
  originalUrl?: string;
  helionx?: HelionxRequestContext;
}

type MockResponse = HelionxExpressResponse & {
  onMock: ReturnType<typeof vi.fn>;
};

describe("createHelionxExpressMiddleware", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("attaches helionx context and preserves x-request-id", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => "",
    });

    vi.stubGlobal("fetch", fetchMock);

    const middleware = createHelionxExpressMiddleware({
      endpoint: "http://localhost:8080",
      service: "express-service",
      autoTrackRequestLifecycle: false,
    });

    const req: HelionxExpressRequest = {
      headers: {
        "x-request-id": "req-123",
      },
      method: "POST",
      path: "/test",
      originalUrl: "/test",
    };

    const onMock = vi.fn();
    const res: MockResponse = {
      statusCode: 200,
      on: (event, listener) => onMock(event, listener),
      onMock,
    };

    const next = vi.fn();

    middleware(req, res, next);

    expect(next).toHaveBeenCalledOnce();
    expect(req.helionx).toBeDefined();
    expect(req.helionx?.requestId).toBe("req-123");
    expect(req.helionx?.source).toBe("x-request-id");

    await req.helionx?.success("handler.completed", {
      path: "/test",
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);

    const [, requestInit] = fetchMock.mock.calls[0] as [string, RequestInit];
    const body = JSON.parse(String(requestInit.body));

    expect(body).toMatchObject({
      requestId: "req-123",
      service: "express-service",
      eventType: "handler.completed",
      status: "SUCCESS",
      metadata: {
        path: "/test",
      },
    });
  });

  it("emits started and finished lifecycle events for a successful request", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => "",
    });

    vi.stubGlobal("fetch", fetchMock);

    const middleware = createHelionxExpressMiddleware({
      endpoint: "http://localhost:8080",
      service: "express-service",
    });

    let finishListener: (() => void) | undefined;

    const req: HelionxExpressRequest = {
      headers: {
        "x-request-id": "req-456",
      },
      method: "POST",
      path: "/test",
      originalUrl: "/test",
    };

    const onMock = vi.fn((event: "finish" | "close", listener: () => void) => {
      if (event === "finish") {
        finishListener = listener;
      }
    });

    const res: MockResponse = {
      statusCode: 200,
      on: (event, listener) => onMock(event, listener),
      onMock,
    };

    const next = vi.fn();

    middleware(req, res, next);

    expect(next).toHaveBeenCalledOnce();
    expect(fetchMock).toHaveBeenCalledTimes(1);

    expect(finishListener).toBeDefined();
    finishListener?.();

    await Promise.resolve();

    expect(fetchMock).toHaveBeenCalledTimes(2);

    const startedBody = JSON.parse(String((fetchMock.mock.calls[0] as [string, RequestInit])[1].body));
    const finishedBody = JSON.parse(String((fetchMock.mock.calls[1] as [string, RequestInit])[1].body));

    expect(startedBody).toMatchObject({
      requestId: "req-456",
      service: "express-service",
      eventType: "http.request.started",
      status: "STARTED",
      metadata: {
        method: "POST",
        path: "/test",
        originalUrl: "/test",
        requestIdSource: "x-request-id",
      },
    });

    expect(finishedBody).toMatchObject({
      requestId: "req-456",
      service: "express-service",
      eventType: "http.request.finished",
      status: "SUCCESS",
      metadata: {
        method: "POST",
        path: "/test",
        originalUrl: "/test",
        requestIdSource: "x-request-id",
        statusCode: 200,
      },
    });
  });

  it("emits a failed finished event for 5xx responses", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => "",
    });

    vi.stubGlobal("fetch", fetchMock);

    const middleware = createHelionxExpressMiddleware({
      endpoint: "http://localhost:8080",
      service: "express-service",
    });

    let finishListener: (() => void) | undefined;

    const req: HelionxExpressRequest = {
      headers: {
        "x-request-id": "req-789",
      },
      method: "GET",
      path: "/boom",
      originalUrl: "/boom",
    };

    const onMock = vi.fn((event: "finish" | "close", listener: () => void) => {
      if (event === "finish") {
        finishListener = listener;
      }
    });

    const res: MockResponse = {
      statusCode: 500,
      on: (event, listener) => onMock(event, listener),
      onMock,
    };

    const next = vi.fn();

    middleware(req, res, next);
    finishListener?.();

    await Promise.resolve();

    expect(fetchMock).toHaveBeenCalledTimes(2);

    const finishedBody = JSON.parse(String((fetchMock.mock.calls[1] as [string, RequestInit])[1].body));

    expect(finishedBody).toMatchObject({
      requestId: "req-789",
      service: "express-service",
      eventType: "http.request.finished",
      status: "FAILED",
      metadata: {
        method: "GET",
        path: "/boom",
        originalUrl: "/boom",
        requestIdSource: "x-request-id",
        statusCode: 500,
      },
    });
  });
});