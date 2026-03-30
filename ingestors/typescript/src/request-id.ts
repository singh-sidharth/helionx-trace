import { randomUUID } from "node:crypto";

export type HeaderValue = string | string[] | undefined;
export type HeaderMap = Record<string, HeaderValue>;

export type RequestIdSource =
  | "x-request-id"
  | "x-correlation-id"
  | "traceparent"
  | "x-amzn-trace-id"
  | "generated";

export interface ResolvedRequestId {
  requestId: string;
  source: RequestIdSource;
}

const TRACEPARENT_REGEX =
  /^[\da-f]{2}-([\da-f]{32})-([\da-f]{16})-[\da-f]{2}$/i;
const X_AMZN_TRACE_ID_ROOT_REGEX = /(?:^|;)\s*Root=([^;]+)/i;

export function generateRequestId(): string {
  return randomUUID();
}

export function resolveRequestId(headers: HeaderMap): ResolvedRequestId {
  const xRequestId = getHeaderValue(headers, "x-request-id");
  if (xRequestId) {
    return {
      requestId: xRequestId,
      source: "x-request-id",
    };
  }

  const xCorrelationId = getHeaderValue(headers, "x-correlation-id");
  if (xCorrelationId) {
    return {
      requestId: xCorrelationId,
      source: "x-correlation-id",
    };
  }

  const traceparent = getHeaderValue(headers, "traceparent");
  if (traceparent) {
    const traceparentTraceId = extractTraceparentTraceId(traceparent);
    if (traceparentTraceId) {
      return {
        requestId: traceparentTraceId,
        source: "traceparent",
      };
    }
  }

  const xAmznTraceId = getHeaderValue(headers, "x-amzn-trace-id");
  if (xAmznTraceId) {
    const amazonRootTraceId = extractAmazonRootTraceId(xAmznTraceId);
    if (amazonRootTraceId) {
      return {
        requestId: amazonRootTraceId,
        source: "x-amzn-trace-id",
      };
    }

    return {
      requestId: xAmznTraceId,
      source: "x-amzn-trace-id",
    };
  }

  return {
    requestId: generateRequestId(),
    source: "generated",
  };
}

function getHeaderValue(headers: HeaderMap, headerName: string): string | undefined {
  const headerValue = headers[headerName] ?? headers[headerName.toLowerCase()] ?? headers[headerName.toUpperCase()];

  if (Array.isArray(headerValue)) {
    const firstValue = headerValue.find((value) => value.trim().length > 0);
    return firstValue?.trim();
  }

  if (typeof headerValue === "string") {
    const trimmedValue = headerValue.trim();
    return trimmedValue.length > 0 ? trimmedValue : undefined;
  }

  return undefined;
}

function extractTraceparentTraceId(traceparent: string): string | undefined {
  const match = traceparent.match(TRACEPARENT_REGEX);
  return match?.[1];
}

function extractAmazonRootTraceId(xAmznTraceId: string): string | undefined {
  const match = xAmznTraceId.match(X_AMZN_TRACE_ID_ROOT_REGEX);
  return match?.[1]?.trim();
}