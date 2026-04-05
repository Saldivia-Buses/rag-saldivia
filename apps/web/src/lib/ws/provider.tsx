"use client";

import { createContext, useContext, useEffect, useCallback, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { wsManager, type WsMessageHandler, type WsMessage } from "./manager";
import { useAuthStore } from "@/lib/auth/store";

interface WsContextValue {
  subscribe: (channel: string, handler: WsMessageHandler) => () => void;
  send: (msg: WsMessage) => void;
  state: string;
}

const WsContext = createContext<WsContextValue | null>(null);

export function WsProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const stateRef = useRef(wsManager.state);

  // Connect/disconnect based on auth state
  useEffect(() => {
    if (isAuthenticated) {
      wsManager.connect();
    } else {
      wsManager.disconnect();
    }

    return () => wsManager.disconnect();
  }, [isAuthenticated]);

  // Invalidate TanStack Query caches when WS events arrive
  useEffect(() => {
    const unsubs: (() => void)[] = [];

    // Module changes → re-fetch enabled modules
    unsubs.push(
      wsManager.subscribe("modules", () => {
        queryClient.invalidateQueries({ queryKey: ["modules", "enabled"] });
      }),
    );

    // Notification events → re-fetch notifications + chat invalidation
    unsubs.push(
      wsManager.subscribe("notifications", (data) => {
        queryClient.invalidateQueries({ queryKey: ["notifications"] });

        // data is already a parsed object from the WS JSON message (not a string)
        const evt = data as { type?: string; data?: { session_id?: string } | string };
        if (evt.type === "chat.new_message" && evt.data) {
          // evt.data may be an object (already parsed) or a string (needs parsing)
          const payload = typeof evt.data === "string"
            ? (() => { try { return JSON.parse(evt.data as string); } catch { return null; } })()
            : evt.data;
          if (payload?.session_id) {
            queryClient.invalidateQueries({
              queryKey: ["chat", "messages", payload.session_id],
            });
          }
        }
        queryClient.invalidateQueries({ queryKey: ["chat", "sessions"] });
      }),
    );

    // Preload event → hydrate cache directly
    unsubs.push(
      wsManager.subscribe("preload", (data) => {
        const preload = data as Record<string, unknown>;
        if (preload.modules) {
          queryClient.setQueryData(["modules", "enabled"], preload.modules);
        }
        if (preload.notificationCount !== undefined) {
          queryClient.setQueryData(
            ["notifications", "count"],
            preload.notificationCount,
          );
        }
      }),
    );

    return () => unsubs.forEach((fn) => fn());
  }, [queryClient]);

  const subscribe = useCallback(
    (channel: string, handler: WsMessageHandler) =>
      wsManager.subscribe(channel, handler),
    [],
  );

  const send = useCallback((msg: WsMessage) => wsManager.send(msg), []);

  const value: WsContextValue = {
    subscribe,
    send,
    state: stateRef.current,
  };

  return <WsContext.Provider value={value}>{children}</WsContext.Provider>;
}

export function useWs() {
  const ctx = useContext(WsContext);
  if (!ctx) throw new Error("useWs must be used within WsProvider");
  return ctx;
}

/**
 * Subscribe to a WebSocket channel with automatic cleanup.
 */
export function useWsChannel(channel: string, handler: WsMessageHandler) {
  const { subscribe } = useWs();

  useEffect(() => {
    return subscribe(channel, handler);
  }, [channel, handler, subscribe]);
}
