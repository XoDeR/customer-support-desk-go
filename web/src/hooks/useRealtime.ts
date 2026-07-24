import { useEffect, useRef } from "react";
import { API_URL } from "../api";
import { useAuthStore } from "../store";

export type RealtimeEvent = {
  type?: string;
  ticket_id?: string;
  payload?: unknown;
};

/**
 * Connects to GET /api/v1/ws. Auth uses `?token=` because browsers cannot set
 * Authorization on the WebSocket handshake (backend accepts this for /ws only).
 * Reconnects with backoff when the socket closes.
 */
export function useRealtime(onEvent: (event: RealtimeEvent) => void) {
  const token = useAuthStore((state) => state.tokens?.access_token);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  useEffect(() => {
    if (!token) return;

    let socket: WebSocket | undefined;
    let retryTimer: number | undefined;
    let attempt = 0;
    let cancelled = false;

    const connect = () => {
      const base = API_URL.replace(/^http/, "ws");
      const url = `${base}/ws?token=${encodeURIComponent(token)}`;
      socket = new WebSocket(url);

      socket.onopen = () => {
        attempt = 0;
      };

      socket.onmessage = (message) => {
        try {
          onEventRef.current(JSON.parse(message.data) as RealtimeEvent);
        } catch {
          // Ignore malformed events.
        }
      };

      socket.onclose = () => {
        if (cancelled) return;
        const delay = Math.min(10_000, 1_000 * 2 ** attempt);
        attempt += 1;
        retryTimer = window.setTimeout(connect, delay);
      };

      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();

    return () => {
      cancelled = true;
      if (retryTimer) window.clearTimeout(retryTimer);
      socket?.close();
    };
  }, [token]);
}
