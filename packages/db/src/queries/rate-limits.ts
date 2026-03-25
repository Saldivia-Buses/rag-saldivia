import { eq, and, gte, sql } from "drizzle-orm"
import { getDb } from "../connection"
import { rateLimits, events, userAreas } from "../schema"
import type { NewRateLimit } from "../schema"

export async function getRateLimit(userId: number): Promise<number | null> {
  const db = getDb()

  // Primero verificar límite a nivel usuario
  const userLimit = await db
    .select({ max: rateLimits.maxQueriesPerHour })
    .from(rateLimits)
    .where(and(eq(rateLimits.targetType, "user"), eq(rateLimits.targetId, userId), eq(rateLimits.active, true)))
    .limit(1)
  if (userLimit[0]) return userLimit[0].max

  // Luego verificar límite de área (toma el mínimo entre todas las áreas del usuario)
  const areas = await db
    .select({ areaId: userAreas.areaId })
    .from(userAreas)
    .where(eq(userAreas.userId, userId))
  if (areas.length === 0) return null

  const areaIds = areas.map((a) => a.areaId)
  for (const areaId of areaIds) {
    const areaLimit = await db
      .select({ max: rateLimits.maxQueriesPerHour })
      .from(rateLimits)
      .where(and(eq(rateLimits.targetType, "area"), eq(rateLimits.targetId, areaId), eq(rateLimits.active, true)))
      .limit(1)
    if (areaLimit[0]) return areaLimit[0].max
  }

  return null
}

export async function countQueriesLastHour(userId: number): Promise<number> {
  const db = getDb()
  const oneHourAgo = Date.now() - 60 * 60 * 1000
  const rows = await db
    .select({ count: sql<number>`count(*)` })
    .from(events)
    .where(
      and(
        eq(events.type, "rag.stream_started"),
        eq(events.userId, userId),
        gte(events.ts, oneHourAgo)
      )
    )
  return rows[0]?.count ?? 0
}

export async function createRateLimit(data: Omit<NewRateLimit, "createdAt">) {
  const db = getDb()
  const [row] = await db
    .insert(rateLimits)
    .values({ ...data, createdAt: Date.now() })
    .returning()
  return row!
}

export async function listRateLimits() {
  const db = getDb()
  return db.select().from(rateLimits)
}

export async function deleteRateLimit(id: number) {
  const db = getDb()
  await db.delete(rateLimits).where(eq(rateLimits.id, id))
}
