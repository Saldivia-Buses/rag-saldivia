/**
 * GET  /api/admin/ingestion — lista jobs activos/recientes
 *
 * F8.30 — Lee job history desde BullMQ en lugar de SQLite ingestion_queue.
 * Los ingestion_jobs del blueprint (NV-Ingest) se siguen leyendo de SQLite.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionJobs } from "@rag-saldivia/db"
import { ingestionQueue, type IngestionJobData } from "@/lib/queue"
import { desc, or, eq, and } from "drizzle-orm"
import type { Job } from "bullmq"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const userId = Number(claims.sub)

  // Jobs de BullMQ — nuestra cola de ingesta
  const allBullJobs = await ingestionQueue
    .getJobs(["active", "waiting", "completed", "failed"])
    .catch(() => [] as Job<IngestionJobData>[])

  const queueItems = claims.role === "admin"
    ? allBullJobs
    : allBullJobs.filter((j) => j.data.userId === userId)

  const queueDto = queueItems.map((j) => ({
    id: j.id,
    collection: j.data.collection,
    filePath: j.data.filePath,
    userId: j.data.userId,
    filename: j.data.filename,
    status: j.returnvalue !== undefined ? "done" : j.failedReason ? "error" : "pending",
    failedReason: j.failedReason,
    timestamp: j.timestamp,
    finishedOn: j.finishedOn,
  }))

  // Jobs de NV-Ingest (blueprint) — se leen de SQLite como antes
  const db = getDb()
  const nvIngestJobs = await db
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

  return NextResponse.json({ ok: true, data: { queue: queueDto, jobs: nvIngestJobs } })
}
