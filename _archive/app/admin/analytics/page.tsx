import { requireAdmin } from "@/lib/auth/current-user"
import { getDb, events, messageFeedback } from "@rag-saldivia/db"
import { desc, eq, sql } from "drizzle-orm"
import { AnalyticsDashboard, type AnalyticsData } from "@/components/admin/AnalyticsDashboard"

async function getAnalyticsData(): Promise<AnalyticsData> {
  const db = getDb()
  const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000

  // Drizzle 0.45: SQL<T> es invariante — usar SQL sin genérico y castear al mapear
  const [queriesByDayRaw, topCollectionsRaw, feedbackRaw, topUsersRaw] = await Promise.all([
    db
      .select({
        day: sql`date(${events.ts}/1000, 'unixepoch')`,
        count: sql`count(*)`,
      })
      .from(events)
      .where(sql`${events.type} = 'rag.stream_started' AND ${events.ts} >= ${thirtyDaysAgo}`)
      .groupBy(sql`date(${events.ts}/1000, 'unixepoch')`)
      .orderBy(sql`date(${events.ts}/1000, 'unixepoch') DESC`)
      .limit(30),

    db
      .select({
        collection: sql`json_extract(${events.payload}, '$.collection')`,
        count: sql`count(*)`,
      })
      .from(events)
      .where(eq(events.type, "rag.stream_started"))
      .groupBy(sql`json_extract(${events.payload}, '$.collection')`)
      .orderBy(desc(sql`count(*)`))
      .limit(10),

    db
      .select({
        rating: messageFeedback.rating,
        count: sql`count(*)`,
      })
      .from(messageFeedback)
      .groupBy(messageFeedback.rating),

    db
      .select({
        userId: events.userId,
        queries: sql`count(*)`,
      })
      .from(events)
      .where(eq(events.type, "rag.stream_started"))
      .groupBy(events.userId)
      .orderBy(desc(sql`count(*)`))
      .limit(10),
  ])

  return {
    queriesByDay: queriesByDayRaw.map((r) => ({ day: r.day as string, queries: r.count as number })),
    topCollections: topCollectionsRaw
      .filter((r) => r.collection)
      .map((r) => ({ name: r.collection as string, queries: r.count as number })),
    feedbackDistribution: [
      { name: "👍 Útil", value: (feedbackRaw.find((r) => r.rating === "up")?.count as number) ?? 0 },
      { name: "👎 No útil", value: (feedbackRaw.find((r) => r.rating === "down")?.count as number) ?? 0 },
    ],
    topUsers: topUsersRaw.map((r) => ({ userId: r.userId, queries: r.queries as number })),
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
