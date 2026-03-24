/**
 * GET  /api/admin/ingestion — lista jobs activos/recientes
 * DELETE /api/admin/ingestion/[id] — cancelar un job
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionQueue, ingestionJobs } from "@rag-saldivia/db"
import { desc, or, eq, and } from "drizzle-orm"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const db = getDb()
  const userId = Number(claims.sub)

  // Admins ven todos, otros solo los suyos
  const queueItems = await db
    .select()
    .from(ingestionQueue)
    .where(
      claims.role === "admin"
        ? undefined
        : eq(ingestionQueue.userId, userId)
    )
    .orderBy(desc(ingestionQueue.createdAt))
    .limit(50)

  const jobs = await db
    .select()
    .from(ingestionJobs)
    .where(
      claims.role === "admin"
        ? or(eq(ingestionJobs.state, "pending"), eq(ingestionJobs.state, "running"), eq(ingestionJobs.state, "stalled"))
        : and(
            eq(ingestionJobs.userId, userId),
            or(eq(ingestionJobs.state, "pending"), eq(ingestionJobs.state, "running"), eq(ingestionJobs.state, "stalled"))
          )
    )
    .orderBy(desc(ingestionJobs.createdAt))
    .limit(50)

  return NextResponse.json({ ok: true, data: { queue: queueItems, jobs } })
}
