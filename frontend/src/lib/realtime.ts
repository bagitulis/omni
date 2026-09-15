/**
 * Realtime WebSocket client — Phase 4.
 *
 * Wraps a single WebSocket connection to `/api/realtime/ws` with:
 *  - JWT auth via `?token=` query param (same channel the backend Auth
 *    middleware accepts for SSE and now WS).
 *  - Topic subscribe / unsubscribe.
 *  - Exponential-backoff reconnect (1s → 30s cap) so a lost network segment
 *    heals without extra plumbing.
 *  - App-level ping every 20s (server pings every 30s; we're a little
 *    faster so the server's `pongWait` deadline is refreshed in time).
 *
 * This module does NOT bind to React state. Consumers use `useRealtime` in
 * `hooks/useRealtime.ts`, which layers subscribe/unsubscribe semantics on
 * top of the singleton client.
 */

import { logger } from "@/lib/logger";

export type RealtimeEnvelope = {
  type: "auth" | "ack" | "error" | "subscribe" | "unsubscribe" | "publish" | "ping" | "pong";
  topic?: string;
  tenant_id?: string;
  id?: string;
  payload?: unknown;
  error?: string;
};

type Handler = (env: RealtimeEnvelope) => void;

export type RealtimeStatus =
  | "disconnected"
  | "connecting"
  | "connected"
  | "reconnecting"
  | "auth_failed";

const RECONNECT_BASE_MS = 1_000;
const RECONNECT_CAP_MS = 30_000;
const APP_PING_MS = 20_000;

export interface RealtimeClientOptions {
  /** ws:// or wss:// full URL to the realtime endpoint. */
  url: string;
  /** JWT that Auth middleware / realtime hub will resolve. */
  token: string;
}

export class RealtimeClient {
  private ws: WebSocket | null = null;
  private status: RealtimeStatus = "disconnected";
  private handlers = new Map<string, Set<Handler>>();
  private statusListeners = new Set<(s: RealtimeStatus) => void>();
  private wantOpen = false;
  private reconnectAttempt = 0;
  private pingTimer: ReturnType<typeof setInterval> | null = null;

  constructor(private opts: RealtimeClientOptions) {}

  connect(): void {
    if (this.status === "connected" || this.status === "connecting") return;
    this.wantOpen = true;
    this.open();
  }

  disconnect(): void {
    this.wantOpen = false;
    this.clearPing();
    if (this.ws) {
      try {
        this.ws.close();
      } catch {
        /* ignore */
      }
      this.ws = null;
    }
    this.setStatus("disconnected");
  }

  subscribe(topic: string, handler: Handler): () => void {
    let set = this.handlers.get(topic);
    if (!set) {
      set = new Set();
      this.handlers.set(topic, set);
      this.send({ type: "subscribe", topic });
    }
    set.add(handler);
    return () => {
      set?.delete(handler);
      if (set && set.size === 0) {
        this.handlers.delete(topic);
        this.send({ type: "unsubscribe", topic });
      }
    };
  }

  onStatus(listener: (s: RealtimeStatus) => void): () => void {
    this.statusListeners.add(listener);
    listener(this.status);
    return () => {
      this.statusListeners.delete(listener);
    };
  }

  getStatus(): RealtimeStatus {
    return this.status;
  }

  private open(): void {
    this.setStatus(this.reconnectAttempt === 0 ? "connecting" : "reconnecting");
    const url = new URL(this.opts.url);
    url.searchParams.set("token", this.opts.token);
    let ws: WebSocket;
    try {
      ws = new WebSocket(url.toString());
    } catch (e) {
      logger.error("realtime: WebSocket constructor threw", { error: String(e) });
      this.scheduleReconnect();
      return;
    }
    this.ws = ws;
    ws.onopen = () => {
      this.reconnectAttempt = 0;
      this.setStatus("connected");
      // Re-subscribe to every topic the caller had before the drop.
      for (const topic of this.handlers.keys()) {
        this.send({ type: "subscribe", topic });
      }
      this.startPing();
    };
    ws.onmessage = (ev) => this.onMessage(ev);
    ws.onerror = () => {
      // onclose handles the reconnect path.
    };
    ws.onclose = (ev) => {
      this.clearPing();
      this.ws = null;
      // Code 4401 is our convention for auth failures. The backend emits
      // {"type":"error","error":"unauthorized"} and closes the socket; we
      // treat any close after zero successful pings as auth_failed if the
      // very first frame was an error.
      if (this.lastError === "unauthorized") {
        this.setStatus("auth_failed");
        this.wantOpen = false;
        return;
      }
      if (!this.wantOpen) {
        this.setStatus("disconnected");
        return;
      }
      logger.warn("realtime: socket closed", { code: ev.code, reason: ev.reason });
      this.scheduleReconnect();
    };
  }

  private lastError: string | null = null;

  private onMessage(ev: MessageEvent): void {
    let env: RealtimeEnvelope;
    try {
      env = JSON.parse(String(ev.data));
    } catch {
      logger.warn("realtime: dropped non-JSON frame");
      return;
    }
    if (env.type === "error" && env.error) {
      this.lastError = env.error;
      logger.warn("realtime: server error frame", { error: env.error });
      return;
    }
    if (env.type === "publish" && env.topic) {
      const set = this.handlers.get(env.topic);
      if (set) {
        for (const h of set) {
          try {
            h(env);
          } catch (err) {
            logger.error("realtime: handler threw", { error: String(err) });
          }
        }
      }
    }
    // "ack" and "pong" are ignored; their purpose is to keep the write pipe
    // alive so onerror/onclose fire at the right time.
  }

  private send(env: Partial<RealtimeEnvelope>): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    try {
      this.ws.send(JSON.stringify(env));
    } catch (err) {
      logger.warn("realtime: send failed", { error: String(err) });
    }
  }

  private startPing(): void {
    this.clearPing();
    this.pingTimer = setInterval(() => {
      this.send({ type: "ping" });
    }, APP_PING_MS);
  }

  private clearPing(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
  }

  private scheduleReconnect(): void {
    if (!this.wantOpen) {
      this.setStatus("disconnected");
      return;
    }
    const backoff = Math.min(
      RECONNECT_CAP_MS,
      RECONNECT_BASE_MS * 2 ** this.reconnectAttempt,
    );
    this.reconnectAttempt++;
    this.setStatus("reconnecting");
    setTimeout(() => {
      if (this.wantOpen) this.open();
    }, backoff);
  }

  private setStatus(s: RealtimeStatus): void {
    if (s === this.status) return;
    this.status = s;
    for (const l of this.statusListeners) {
      try {
        l(s);
      } catch {
        /* ignore */
      }
    }
  }
}

// ---- Singleton wiring for the app ----

let singleton: RealtimeClient | null = null;

/**
 * Returns the app-wide RealtimeClient, creating it lazily. `opts` are only
 * read the first time. Subsequent calls with different opts are ignored so
 * the singleton is stable across React re-renders.
 */
export function getRealtimeClient(
  optsFactory: () => RealtimeClientOptions | null,
): RealtimeClient | null {
  if (singleton) return singleton;
  const opts = optsFactory();
  if (!opts) return null;
  singleton = new RealtimeClient(opts);
  return singleton;
}

/**
 * Reset the singleton — test-only. Not exported from index.
 */
export function __resetRealtimeSingletonForTests(): void {
  if (singleton) {
    singleton.disconnect();
  }
  singleton = null;
}
