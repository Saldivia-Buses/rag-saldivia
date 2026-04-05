/**
 * WebSocket manager for SDA Framework.
 *
 * Connects to the WS Hub service with JWT auth.
 * Auto-reconnects with exponential backoff.
 * Routes incoming messages to registered handlers.
 */

import { useAuthStore } from "@/lib/auth/store";
import { getApiBaseUrl } from "@/lib/api/client";

export type WsMessageHandler = (data: unknown) => void;

export interface WsMessage {
  type: string;
  channel?: string;
  data?: unknown;
}

type ConnectionState = "connecting" | "connected" | "disconnected" | "reconnecting";

class WebSocketManager {
  private ws: WebSocket | null = null;
  private handlers = new Map<string, Set<WsMessageHandler>>();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempt = 0;
  private maxReconnectDelay = 30_000;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private intentionalClose = false;
  private _state: ConnectionState = "disconnected";
  private stateListeners = new Set<(state: ConnectionState) => void>();

  get state(): ConnectionState {
    return this._state;
  }

  private setState(state: ConnectionState) {
    this._state = state;
    this.stateListeners.forEach((fn) => fn(state));
  }

  onStateChange(fn: (state: ConnectionState) => void): () => void {
    this.stateListeners.add(fn);
    return () => this.stateListeners.delete(fn);
  }

  connect() {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    const token = useAuthStore.getState().accessToken;
    if (!token) return;

    this.intentionalClose = false;
    this.setState("connecting");

    const baseUrl = getApiBaseUrl();
    // Convert http(s) to ws(s) protocol
    const wsBase = baseUrl
      ? baseUrl.replace(/^http/, "ws")
      : `ws://${typeof window !== "undefined" ? window.location.host : "localhost"}`;
    const url = `${wsBase}/ws`;

    this.ws = new WebSocket(url, ["bearer", token]);

    this.ws.onopen = () => {
      this.reconnectAttempt = 0;
      this.setState("connected");
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data);
        this.dispatch(msg);
      } catch {
        // Non-JSON message (heartbeat pong, etc.)
      }
    };

    this.ws.onclose = () => {
      this.stopHeartbeat();
      if (!this.intentionalClose) {
        this.setState("reconnecting");
        this.scheduleReconnect();
      } else {
        this.setState("disconnected");
      }
    };

    this.ws.onerror = () => {
      // onclose will fire after onerror
    };
  }

  disconnect() {
    this.intentionalClose = true;
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.setState("disconnected");
  }

  /**
   * Subscribe to a channel. Returns an unsubscribe function.
   */
  subscribe(channel: string, handler: WsMessageHandler): () => void {
    let set = this.handlers.get(channel);
    if (!set) {
      set = new Set();
      this.handlers.set(channel, set);
    }
    set.add(handler);

    // Tell the Hub we want this channel
    this.send({ type: "subscribe", channel });

    return () => {
      set!.delete(handler);
      if (set!.size === 0) {
        this.handlers.delete(channel);
      }
    };
  }

  /**
   * Send a message to the Hub.
   */
  send(msg: WsMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  private dispatch(msg: WsMessage) {
    // Route by channel if present
    if (msg.channel) {
      const channelHandlers = this.handlers.get(msg.channel);
      channelHandlers?.forEach((fn) => fn(msg.data));
    }

    // Also dispatch by type (for preload, system events)
    const typeHandlers = this.handlers.get(msg.type);
    typeHandlers?.forEach((fn) => fn(msg.data));

    // Global "*" handlers get everything
    const globalHandlers = this.handlers.get("*");
    globalHandlers?.forEach((fn) => fn(msg));
  }

  private scheduleReconnect() {
    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempt),
      this.maxReconnectDelay,
    );
    this.reconnectAttempt++;

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  private startHeartbeat() {
    this.heartbeatTimer = setInterval(() => {
      this.send({ type: "ping" });
    }, 30_000);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}

// Singleton
export const wsManager = new WebSocketManager();
