import { requireAdmin } from "@/lib/auth/current-user"
import { getDb, ingestionQueue, users, areas, events } from "@rag-saldivia/db"
import { count, eq, or, gt } from "drizzle-orm"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"
import { SystemStatus } from "@/components/admin/SystemStatus"

export const dynamic = "force-dynamic" // siempre fresco

export default async function AdminSystemPage() {
  await requireAdmin()
  const db = getDb()

  const [[totalUsers], [totalAreas], activeJobs, collections] = await Promise.all([
    db.select({ count: count() }).from(users).where(eq(users.active, true)),
    db.select({ count: count() }).from(areas),
    db.select().from(ingestionQueue).where(
      or(eq(ingestionQueue.status, "pending"), eq(ingestionQueue.status, "locked"))
    ).limit(20),
    getCachedRagCollections(),
  ])

  // Contar eventos de error en las últimas 24hs
  const oneDayAgo = Date.now() - 86400_000
  const [recentErrors] = await db
    .select({ count: count() })
    .from(events)
    .where(
      gt(events.ts, oneDayAgo)
    )

  const stats = {
    activeUsers: totalUsers?.count ?? 0,
    areas: totalAreas?.count ?? 0,
    collections: collections.length,
    activeJobs: activeJobs.length,
    recentErrors: recentErrors?.count ?? 0,
  }

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Estado del sistema</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Métricas en tiempo real
        </p>
      </div>
      <SystemStatus stats={stats} activeJobs={activeJobs} />
    </div>
  )
}
