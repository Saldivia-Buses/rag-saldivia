/**
 * Presence tracking via Redis TTL keys.
 *
 * Pattern: `presence:user:{id}` with 30s TTL.
 * Heartbeat every 15s to keep alive.
 * Status: "online" | "away" | "offline" (offline = key expired).
 */

import type { Redis } from "ioredis"

const PRESENCE_TTL = 30 // seconds
const PRESENCE_PREFIX = "presence:user:"

export type PresenceStatus = "online" | "away" | "offline"

/** Update user presence in Redis with TTL. */
export async function setPresence(
  redis: Redis,
  userId: number,
  status: "online" | "away"
): Promise<void> {
  await redis.set(`${PRESENCE_PREFIX}${userId}`, status, "EX", PRESENCE_TTL)
}

/** Get a single user's presence status. */
export async function getPresence(
  redis: Redis,
  userId: number
): Promise<PresenceStatus> {
  const val = await redis.get(`${PRESENCE_PREFIX}${userId}`)
  return (val as PresenceStatus) ?? "offline"
}

/** Get presence for multiple users in a single pipeline call. */
export async function getPresenceBulk(
  redis: Redis,
  userIds: number[]
): Promise<Record<number, PresenceStatus>> {
  if (userIds.length === 0) return {}
  const pipeline = redis.pipeline()
  for (const id of userIds) {
    pipeline.get(`${PRESENCE_PREFIX}${id}`)
  }
  const results = await pipeline.exec()
  const map: Record<number, PresenceStatus> = {}
  for (let i = 0; i < userIds.length; i++) {
    const val = results?.[i]?.[1] as string | null
    map[userIds[i]!] = (val as PresenceStatus) ?? "offline"
  }
  return map
}

/** Remove user presence (explicit disconnect). */
export async function clearPresence(
  redis: Redis,
  userId: number
): Promise<void> {
  await redis.del(`${PRESENCE_PREFIX}${userId}`)
}
