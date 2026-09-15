/**
 * useRealtime — subscribe to a realtime topic from a React component.
 *
 * Phase 4 exposes the wire; Phase 5 lands the publishers. Until then this
 * hook is safe to call from any page: if the client cannot be constructed
 * (no token, or `VITE_REALTIME_ENABLED` off), it returns a disconnected
 * status and the handler is never called.
 */

import { useEffect, useRef, useState } from "react";
import {
  getRealtimeClient,
  type RealtimeEnvelope,
  type RealtimeStatus,
} from "@/lib/realtime";

function readToken(): string | null {
  try {
    return localStorage.getItem("access_token") || localStorage.getItem("token") || null;
  } catch {
    return null;
  }
}

function buildUrl(): string {
  // Same-origin WS. Uses wss:// on HTTPS pages.
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${window.location.host}/api/realtime/ws`;
}

/**
 * Subscribe to `topic` while the component is mounted. The handler receives
 * the raw envelope; consumers narrow `payload` themselves.
 *
 * Returns the current connection status so components can render a "live"
 * indicator or fall back to polling.
 */
export function useRealtime(
  topic: string,
  handler: (env: RealtimeEnvelope) => void,
  enabled = true,
): RealtimeStatus {
  const [status, setStatus] = useState<RealtimeStatus>("disconnected");
  // Stash the latest handler so re-subscribes are cheap when the caller
  // passes an inline function.
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    if (!enabled) {
      setStatus("disconnected");
      return;
    }
    const client = getRealtimeClient(() => {
      const token = readToken();
      if (!token) return null;
      return { url: buildUrl(), token };
    });
    if (!client) return;
    const offStatus = client.onStatus(setStatus);
    client.connect();
    const unsubscribe = client.subscribe(topic, (env) => handlerRef.current(env));
    return () => {
      unsubscribe();
      offStatus();
    };
    // topic string identity + enabled toggle are the only inputs that
    // trigger a resubscribe. Handler churn is absorbed by handlerRef above.
  }, [topic, enabled]);

  return status;
}
