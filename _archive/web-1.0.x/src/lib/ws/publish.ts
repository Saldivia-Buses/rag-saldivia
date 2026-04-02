/**
 * Publish messaging events to Redis Pub/Sub from Next.js server-side code.
 * The WS sidecar subscribes to these channels and broadcasts to connected clients.
 */

import { getRedisClient } from "@rag-saldivia/db"
import { redisTopic } from "./protocol"
import type { ServerMessage } from "@rag-saldivia/shared"

/** Publish a server message to a channel via Redis (for WS sidecar fan-out). */
export function publishToChannel(channelId: string, msg: ServerMessage) {
  const redis = getRedisClient()
  redis.publish(redisTopic(channelId), JSON.stringify(msg))
}
