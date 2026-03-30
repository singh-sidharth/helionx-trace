export type HelionxEventStatus = "STARTED" | "SUCCESS" | "FAILED";

export interface HelionxEvent {
  requestId: string;
  service: string;
  eventType: string;
  status: HelionxEventStatus;
  timestamp?: string;
  metadata?: Record<string, unknown>;
}

export interface HelionxIngestResponse {
  ok: boolean;
}

export interface HelionxClientOptions {
  endpoint: string;
  apiKey?: string;
  timeoutMs?: number;
}