/**
 * WebSocket protocol helpers — parse client messages, serialize server messages.
 * Uses Zod schemas from @rag-saldivia/shared for validation.
 */

import { ClientMessageSchema, type ClientMessage, type ServerMessage } from "@rag-saldivia/shared"

/** Parse a raw WebSocket text message into a typed ClientMessage, or null if invalid. */
export function parseClientMessage(raw: string): ClientMessage | null {
  try {
    const json = JSON.parse(raw)
    const result = ClientMessageSchema.safeParse(json)
    return result.success ? result.data : null
  } catch {
    return null
  }
}

/** Serialize a ServerMessage to JSON string for sending over WebSocket. */
export function serializeServerMessage(msg: ServerMessage): string {
  return JSON.stringify(msg)
}

/** Redis Pub/Sub channel prefix for messaging events. */
export const REDIS_MESSAGING_CHANNEL = "messaging"

/** Build a Redis Pub/Sub channel name for a specific messaging event. */
export function redisTopic(channelId: string): string {
  return `${REDIS_MESSAGING_CHANNEL}:${channelId}`
}
