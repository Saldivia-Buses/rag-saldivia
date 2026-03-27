import { requireAdmin } from "@/lib/auth/current-user"
import { getDb, events, messageFeedback } from "@rag-saldivia/db"
import { desc, eq, sql } from "drizzle-orm"
import { AnalyticsDashboard, type AnalyticsData } from "@/components/admin/AnalyticsDashboard"

async function getAnalyticsData(): Promise<AnalyticsData> {
  const db = getDb()
  const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000

  const [queriesByDayRaw, topCollectionsRaw, feedbackRaw, topUsersRaw] = await Promise.all([
    db
      .select({
        day: sql<string>`date(${events.ts}/1000, 'unixepoch')`.as("day"),
        count: sql<number>`count(*)`.as("count"),
      })
      .from(events)
      .where(sql`${events.type} = 'rag.stream_started' AND ${events.ts} >= ${thirtyDaysAgo}`)
      .groupBy(sql`date(${events.ts}/1000, 'unixepoch')`)
      .orderBy(sql`day DESC`)
      .limit(30),

    db
      .select({
        collection: sql<string>`json_extract(${events.payload}, '$.collection')`.as("collection"),
        count: sql<number>`count(*)`.as("count"),
      })
      .from(events)
      .where(eq(events.type, "rag.stream_started"))
      .groupBy(sql`json_extract(${events.payload}, '$.collection')`)
      .orderBy(desc(sql`count(*)`))
      .limit(10),

    db
      .select({
        rating: messageFeedback.rating,
        count: sql<number>`count(*)`.as("count"),
      })
      .from(messageFeedback)
      .groupBy(messageFeedback.rating),

    db
      .select({
        userId: events.userId,
        queries: sql<number>`count(*)`.as("queries"),
      })
      .from(events)
      .where(eq(events.type, "rag.stream_started"))
      .groupBy(events.userId)
      .orderBy(desc(sql`count(*)`))
      .limit(10),
  ])

  return {
    queriesByDay: queriesByDayRaw.map((r) => ({ day: r.day, queries: r.count })),
    topCollections: topCollectionsRaw
      .filter((r) => r.collection)
      .map((r) => ({ name: r.collection, queries: r.count })),
    feedbackDistribution: [
      { name: "👍 Útil", value: feedbackRaw.find((r) => r.rating === "up")?.count ?? 0 },
      { name: "👎 No útil", value: feedbackRaw.find((r) => r.rating === "down")?.count ?? 0 },
    ],
    topUsers: topUsersRaw.map((r) => ({ userId: r.userId, queries: r.queries })),
  }
}

export default async function AnalyticsPage() {
  await requireAdmin()
  const data = await getAnalyticsData()

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Analytics</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Uso del sistema en los últimos 30 días
        </p>
      </div>
      <AnalyticsDashboard data={data} />
    </div>
  )
}
