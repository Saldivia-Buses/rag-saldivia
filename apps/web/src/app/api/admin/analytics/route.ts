/**
 * GET /api/admin/analytics
 * Retorna datos de analytics para el dashboard.
 * Solo accesible por admins.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, events, messageFeedback, chatSessions } from "@rag-saldivia/db"
import { desc, eq, sql, gte } from "drizzle-orm"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })
  }

  const db = getDb()
  const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000

  // Queries por día (últimos 30 días)
  const queriesByDayRaw = await db
    .select({
      day: sql<string>`date(${events.ts}/1000, 'unixepoch')`.as("day"),
      count: sql<number>`count(*)`.as("count"),
    })
    .from(events)
    .where(
      sql`${events.type} = 'rag.stream_started' AND ${events.ts} >= ${thirtyDaysAgo}`
    )
    .groupBy(sql`date(${events.ts}/1000, 'unixepoch')`)
    .orderBy(sql`day DESC`)
    .limit(30)

  // Colecciones más consultadas
  const topCollectionsRaw = await db
    .select({
      collection: sql<string>`json_extract(${events.payload}, '$.collection')`.as("collection"),
      count: sql<number>`count(*)`.as("count"),
    })
    .from(events)
    .where(eq(events.type, "rag.stream_started"))
    .groupBy(sql`json_extract(${events.payload}, '$.collection')`)
    .orderBy(desc(sql`count(*)`))
    .limit(10)

  // Distribución de feedback
  const feedbackRaw = await db
    .select({
      rating: messageFeedback.rating,
      count: sql<number>`count(*)`.as("count"),
    })
    .from(messageFeedback)
    .groupBy(messageFeedback.rating)

  // Usuarios más activos (top 10)
  const topUsersRaw = await db
    .select({
      userId: events.userId,
      queries: sql<number>`count(*)`.as("queries"),
    })
    .from(events)
    .where(eq(events.type, "rag.stream_started"))
    .groupBy(events.userId)
    .orderBy(desc(sql`count(*)`))
    .limit(10)

  return NextResponse.json({
    ok: true,
    queriesByDay: queriesByDayRaw.map((r) => ({ day: r.day, queries: r.count })),
    topCollections: topCollectionsRaw
      .filter((r) => r.collection)
      .map((r) => ({ name: r.collection, queries: r.count })),
    feedbackDistribution: [
      { name: "👍 Útil", value: feedbackRaw.find((r) => r.rating === "up")?.count ?? 0 },
      { name: "👎 No útil", value: feedbackRaw.find((r) => r.rating === "down")?.count ?? 0 },
    ],
    topUsers: topUsersRaw.map((r) => ({ userId: r.userId, queries: r.queries })),
  })
}
