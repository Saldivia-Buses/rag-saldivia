#!/usr/bin/env bun
/**
 * Bun WebSocket sidecar server for real-time messaging.
 *
 * Runs as a separate process on WS_PORT (default 3001).
 * Auth via JWT (same JWT_SECRET as Next.js).
 * Uses Bun's built-in pub/sub for channel subscriptions.
 * Bridges Redis Pub/Sub for cross-instance fan-out.
 *
 * Usage:
 *   bun apps/web/src/lib/ws/sidecar.ts
 *   bun run ws
 */

import { jwtVerify } from "jose"
import Redis from "ioredis"
import { getChannelsByUser } from "@rag-saldivia/db"
import { parseClientMessage, serializeServerMessage, redisTopic } from "./protocol"
import { setPresence, clearPresence } from "./presence"
import type { ServerMessage } from "@rag-saldivia/shared"
import type { ServerWebSocket } from "bun"

// ── Config ────────────────────────────────────────────────────────────────

const WS_PORT = parseInt(process.env["WS_PORT"] ?? "3001")
const JWT_SECRET = new TextEncoder().encode(process.env["JWT_SECRET"] ?? "")
const REDIS_URL = process.env["REDIS_URL"] ?? "redis://localhost:6379"
const PRESENCE_HEARTBEAT_MS = 15_000

// ── Per-connection state ──────────────────────────────────────────────────

type WsData = {
  userId: number | null
  email: string | null
  name: string | null
  authenticated: boolean
  lastPing: number
}

// ── Redis Pub/Sub ─────────────────────────────────────────────────────────

const redisSub = new Redis(REDIS_URL)
const redisPub = new Redis(REDIS_URL)
const redisPresence = new Redis(REDIS_URL)

// Track active connections for presence heartbeat
const activeConnections = new Map<number, Set<ServerWebSocket<WsData>>>()

function trackConnection(userId: number, ws: ServerWebSocket<WsData>) {
  const set = activeConnections.get(userId) ?? new Set()
  set.add(ws)
  activeConnections.set(userId, set)
}

function untrackConnection(userId: number, ws: ServerWebSocket<WsData>) {
  const set = activeConnections.get(userId)
  if (set) {
    set.delete(ws)
    if (set.size === 0) activeConnections.delete(userId)
  }
}

// Subscribe to all messaging events from Redis
redisSub.psubscribe("messaging:*").catch(() => {
  console.error("[ws] Failed to subscribe to Redis messaging channel")
})

redisSub.on("pmessage", (_pattern, channel, message) => {
  // channel format: "messaging:{channelId}"
  const channelId = channel.split(":")[1]
  if (!channelId) return

  // Republish to Bun's built-in pub/sub
  // All WS connections subscribed to this channel will receive it
  try {
    server.publish(`channel:${channelId}`, message)
  } catch {
    // Server not ready yet
  }
})

// ── JWT verification ──────────────────────────────────────────────────────

async function verifyToken(token: string): Promise<{ sub: string; email: string; name: string } | null> {
  try {
    const { payload } = await jwtVerify(token, JWT_SECRET)
    return {
      sub: String(payload.sub),
      email: String(payload.email ?? ""),
      name: String(payload.name ?? ""),
    }
  } catch {
    return null
  }
}

// ── Helpers ───────────────────────────────────────────────────────────────

function send(ws: ServerWebSocket<WsData>, msg: ServerMessage) {
  ws.send(serializeServerMessage(msg))
}

/** Publish a server message to a channel via Redis (for cross-instance fan-out). */
export function publishToChannel(channelId: string, msg: ServerMessage) {
  redisPub.publish(redisTopic(channelId), serializeServerMessage(msg))
}

// ── Server ────────────────────────────────────────────────────────────────

const server = Bun.serve<WsData>({
  port: WS_PORT,

  fetch(req, server) {
    // Upgrade HTTP to WebSocket
    const success = server.upgrade(req, {
      data: {
        userId: null,
        email: null,
        name: null,
        authenticated: false,
        lastPing: Date.now(),
      },
    })
    if (success) return undefined

    // Non-WebSocket requests: health check
    return new Response(JSON.stringify({ ok: true, ws: true }), {
      headers: { "Content-Type": "application/json" },
    })
  },

  websocket: {
    async message(ws, raw) {
      const text = typeof raw === "string" ? raw : new TextDecoder().decode(raw as unknown as ArrayBuffer)
      const msg = parseClientMessage(text)
      if (!msg) return

      // ── Auth (must be first message) ──
      if (msg.type === "auth") {
        if (ws.data.authenticated) return // Already authenticated

        const claims = await verifyToken(msg.token)
        if (!claims) {
          send(ws, { type: "auth_error", reason: "Token inválido o expirado" })
          ws.close(4001, "Authentication failed")
          return
        }

        const userId = Number(claims.sub)
        ws.data.userId = userId
        ws.data.email = claims.email
        ws.data.name = claims.name
        ws.data.authenticated = true

        // Subscribe to user's channels
        try {
          const userChannels = await getChannelsByUser(userId)
          for (const ch of userChannels) {
            ws.subscribe(`channel:${ch.id}`)
          }
        } catch {
          // DB might be unavailable — subscribe to nothing, they can subscribe manually
        }

        // Set presence
        await setPresence(redisPresence, userId, "online").catch(() => {})
        trackConnection(userId, ws)

        send(ws, { type: "auth_ok", userId })
        return
      }

      // ── All other messages require auth ──
      if (!ws.data.authenticated || !ws.data.userId) {
        send(ws, { type: "auth_error", reason: "No autenticado. Enviar auth primero." })
        return
      }

      const userId = ws.data.userId

      switch (msg.type) {
        case "subscribe":
          ws.subscribe(`channel:${msg.channelId}`)
          break

        case "unsubscribe":
          ws.unsubscribe(`channel:${msg.channelId}`)
          break

        case "typing_start":
          // Broadcast typing to the channel (except sender)
          server.publish(
            `channel:${msg.channelId}`,
            serializeServerMessage({
              type: "typing",
              channelId: msg.channelId,
              userId,
              displayName: ws.data.name ?? "Usuario",
            })
          )
          break

        case "typing_stop":
          // No-op: typing indicators auto-expire on the client
          break

        case "presence":
          await setPresence(redisPresence, userId, msg.status).catch(() => {})
          break

        case "sync": {
          // Reconnection: client sends timestamps per channel, server responds with missed messages
          // This is handled via HTTP API (getMessages with cursor), not WS
          // Just acknowledge the sync request
          break
        }

        case "ping":
          ws.data.lastPing = Date.now()
          send(ws, { type: "pong" })
          break
      }
    },

    open(ws) {
      ws.data.lastPing = Date.now()
    },

    close(ws) {
      if (ws.data.userId) {
        untrackConnection(ws.data.userId, ws)
        // If no more connections for this user, clear presence
        if (!activeConnections.has(ws.data.userId)) {
          clearPresence(redisPresence, ws.data.userId).catch(() => {})
        }
      }
    },
  },
})

// ── Heartbeat: close dead connections ─────────────────────────────────────

setInterval(() => {
  // Presence heartbeat for all active users
  for (const [userId] of activeConnections) {
    setPresence(redisPresence, userId, "online").catch(() => {})
  }
}, PRESENCE_HEARTBEAT_MS)

console.warn(`[ws] WebSocket sidecar listening on ws://localhost:${WS_PORT}`)
