import { createHelionxClient } from "../client.js";
import { resolveRequestId } from "../request-id.js";
import type { HelionxClientOptions, HelionxEvent } from "../types.js";

export interface HelionxExpressRequest {
  headers: Record<string, string | string[] | undefined>;
  method?: string;
  path?: string;
  originalUrl?: string;
  helionx?: HelionxRequestContext;
}

export interface HelionxExpressResponse {
  statusCode?: number;
  on(event: "finish" | "close", listener: () => void): unknown;
}

export type HelionxExpressNextFunction = () => void;

export interface HelionxRequestContext {
  requestId: string;
  source: string;
  event(event: Omit<HelionxEvent, "requestId" | "service">): Promise<void>;
  started(eventType: string, metadata?: Record<string, unknown>): Promise<void>;
  success(eventType: string, metadata?: Record<string, unknown>): Promise<void>;
  fail(eventType: string, metadata?: Record<string, unknown>): Promise<void>;
}

export interface HelionxExpressAdapterOptions extends HelionxClientOptions {
  service: string;
  autoTrackRequestLifecycle?: boolean;
  eventType?: string;
  captureRequestMetadata?: boolean;
  onError?: (error: unknown) => void;
}

export function createHelionxExpressMiddleware(options: HelionxExpressAdapterOptions) {
  const client = createHelionxClient(options);
  const eventType = options.eventType ?? "http.request";
  const startedEventType = `${eventType}.started`;
  const finishedEventType = `${eventType}.finished`;
  const autoTrackRequestLifecycle = options.autoTrackRequestLifecycle ?? true;
  const captureRequestMetadata = options.captureRequestMetadata ?? true;

  return function helionxExpressMiddleware(
    req: HelionxExpressRequest,
    res: HelionxExpressResponse,
    next: HelionxExpressNextFunction,
  ): void {
    const resolvedRequestId = resolveRequestId(req.headers ?? {});

    req.helionx = {
      requestId: resolvedRequestId.requestId,
      source: resolvedRequestId.source,
      event: (event) =>
        client.ingest({
          ...event,
          requestId: resolvedRequestId.requestId,
          service: options.service,
        }),
      started: (trackedEventType, metadata) =>
        sendLifecycleEvent(
          client,
          createLifecycleEvent(
            resolvedRequestId.requestId,
            options.service,
            trackedEventType,
            "STARTED",
            metadata,
          ),
        ),
      success: (trackedEventType, metadata) =>
        sendLifecycleEvent(
          client,
          createLifecycleEvent(
            resolvedRequestId.requestId,
            options.service,
            trackedEventType,
            "SUCCESS",
            metadata,
          ),
        ),
      fail: (trackedEventType, metadata) =>
        sendLifecycleEvent(
          client,
          createLifecycleEvent(
            resolvedRequestId.requestId,
            options.service,
            trackedEventType,
            "FAILED",
            metadata,
          ),
        ),
    };

    if (autoTrackRequestLifecycle) {
      void sendLifecycleEvent(
        client,
        createLifecycleEvent(
          resolvedRequestId.requestId,
          options.service,
          startedEventType,
          "STARTED",
          captureRequestMetadata
            ? buildRequestMetadata(req, resolvedRequestId.source)
            : { requestIdSource: resolvedRequestId.source },
        ),
        options.onError,
      );

      res.on("finish", () => {
        const status = (res.statusCode ?? 500) >= 500 ? "FAILED" : "SUCCESS";

        void sendLifecycleEvent(
          client,
          createLifecycleEvent(
            resolvedRequestId.requestId,
            options.service,
            finishedEventType,
            status,
            captureRequestMetadata
              ? buildResponseMetadata(req, res, resolvedRequestId.source)
              : {
                  requestIdSource: resolvedRequestId.source,
                  statusCode: res.statusCode,
                },
          ),
          options.onError,
        );
      });
    }

    next();
  };
}

function createLifecycleEvent(
  requestId: string,
  service: string,
  eventType: string,
  status: HelionxEvent["status"],
  metadata?: Record<string, unknown>,
): HelionxEvent {
  return {
    requestId,
    service,
    eventType,
    status,
    ...(metadata !== undefined ? { metadata } : {}),
  };
}

async function sendLifecycleEvent(
  client: ReturnType<typeof createHelionxClient>,
  event: HelionxEvent,
  onError?: (error: unknown) => void,
): Promise<void> {
  try {
    await client.ingest(event);
  } catch (error) {
    onError?.(error);
  }
}

function buildRequestMetadata(
  req: HelionxExpressRequest,
  requestIdSource: string,
): Record<string, unknown> {
  return {
    method: req.method,
    path: req.path,
    originalUrl: req.originalUrl,
    requestIdSource,
  };
}

function buildResponseMetadata(
  req: HelionxExpressRequest,
  res: HelionxExpressResponse,
  requestIdSource: string,
): Record<string, unknown> {
  return {
    method: req.method,
    path: req.path,
    originalUrl: req.originalUrl,
    requestIdSource,
    statusCode: res.statusCode,
  };
}