import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionQueue, ingestionJobs } from "@rag-saldivia/db"
import { eq, and } from "drizzle-orm"
import { log } from "@rag-saldivia/logger/backend"

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const { id } = await params
  const userId = Number(claims.sub)
  const db = getDb()

  // Verificar que el job existe y el usuario tiene acceso
  const condition = claims.role === "admin"
    ? eq(ingestionQueue.id, id)
    : and(eq(ingestionQueue.id, id), eq(ingestionQueue.userId, userId))

  const existing = await db
    .select({ id: ingestionQueue.id })
    .from(ingestionQueue)
    .where(condition)
    .limit(1)

  if (existing.length === 0) {
    return NextResponse.json({ ok: false, error: "Job no encontrado" }, { status: 404 })
  }

  await db
    .update(ingestionQueue)
    .set({ status: "error", error: "Cancelado manualmente", completedAt: Date.now() })
    .where(condition)

  log.info("ingestion.cancelled", { jobId: id }, { userId })

  return NextResponse.json({ ok: true })
}

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })
  }

  const { id } = await params
  const body = await request.json().catch(() => ({})) as { action?: string }
  if (body.action !== "retry") {
    return NextResponse.json({ ok: false, error: "action inválida" }, { status: 400 })
  }

  const db = getDb()
  await db
    .update(ingestionJobs)
    .set({ state: "pending", retryCount: 0, lastChecked: null })
    .where(eq(ingestionJobs.id, id))

  log.info("ingestion.retry", { jobId: id }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true })
}
