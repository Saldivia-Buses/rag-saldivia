/**
 * WebSocket client — singleton connection to the sidecar server.
 *
 * Features:
 *   - Auto-reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s)
 *   - Auth via first message after connect
 *   - Event emitter for hooks to subscribe
 *   - Single shared connection across all hooks
 */

import type { ServerMessage, ClientMessage } from "@rag-saldivia/shared"
import { WS_RECONNECT_BASE_MS, WS_RECONNECT_MAX_MS } from "@rag-saldivia/config"

type Listener = (msg: ServerMessage) => void

const WS_URL = typeof window !== "undefined"
  ? (process.env["NEXT_PUBLIC_WS_URL"] ?? `ws://${window.location.hostname}:3001`)
  : ""

class WsClient {
  private ws: WebSocket | null = null
  private listeners = new Set<Listener>()
  private token: string | null = null
  private reconnectAttempt = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private _connected = false
  private _userId: number | null = null

  get connected() { return this._connected }
  get userId() { return this._userId }

  /** Connect with a JWT token. Safe to call multiple times. */
  connect(token: string) {
    if (this.ws && this.token === token) return // Already connected with same token
    this.token = token
    this.reconnectAttempt = 0
    this.doConnect()
  }

  /** Disconnect and stop reconnecting. */
  disconnect() {
    this.token = null
    this._connected = false
    this._userId = null
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    if (this.ws) {
      this.ws.close(1000, "Client disconnect")
      this.ws = null
    }
  }

  /** Subscribe to all server messages. Returns unsubscribe function. */
  on(listener: Listener): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  /** Send a typed client message. */
  send(msg: ClientMessage) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return
    this.ws.send(JSON.stringify(msg))
  }

  private doConnect() {
    if (!this.token || !WS_URL) return

    try {
      this.ws = new WebSocket(WS_URL)
    } catch {
      this.scheduleReconnect()
      return
    }

    this.ws.onopen = () => {
      this.reconnectAttempt = 0
      // Auth is the first message
      this.send({ type: "auth", token: this.token! })
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data as string) as ServerMessage
        // Track auth state
        if (msg.type === "auth_ok") {
          this._connected = true
          this._userId = msg.userId
        } else if (msg.type === "auth_error") {
          this._connected = false
          this.disconnect()
          return
        }
        // Notify all listeners
        for (const listener of this.listeners) {
          listener(msg)
        }
      } catch {
        // Ignore malformed messages
      }
    }

    this.ws.onclose = (event) => {
      this._connected = false
      this.ws = null
      // Don't reconnect if explicitly disconnected or auth failed
      if (event.code === 1000 || event.code === 4001 || !this.token) return
      this.scheduleReconnect()
    }

    this.ws.onerror = () => {
      // onclose will fire after onerror — reconnect handled there
    }
  }

  private scheduleReconnect() {
    if (!this.token) return
    const delay = Math.min(
      WS_RECONNECT_BASE_MS * Math.pow(2, this.reconnectAttempt),
      WS_RECONNECT_MAX_MS
    )
    this.reconnectAttempt++
    this.reconnectTimer = setTimeout(() => this.doConnect(), delay)
  }
}

/** Singleton WebSocket client — shared across all hooks. */
export const wsClient = new WsClient()
