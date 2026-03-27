import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionJobs } from "@rag-saldivia/db"
import { ingestionQueue } from "@/lib/queue"
import { eq } from "drizzle-orm"
import { log } from "@rag-saldivia/logger/backend"

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const { id } = await params
  const userId = Number(claims.sub)

  // Cancelar job en BullMQ
  try {
    const job = await ingestionQueue.getJob(id)
    if (!job) {
      return NextResponse.json({ ok: false, error: "Job no encontrado" }, { status: 404 })
    }

    // Verificar que el usuario tiene acceso al job
    if (claims.role !== "admin" && job.data.userId !== userId) {
      return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })
    }

    await job.remove()
    log.info("ingestion.cancelled", { jobId: id }, { userId })
    return NextResponse.json({ ok: true })
  } catch {
    return NextResponse.json({ ok: false, error: "No se pudo cancelar el job" }, { status: 500 })
  }
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

  // El retry de NV-Ingest jobs sigue en SQLite (ingestionJobs del blueprint)
  const db = getDb()
  await db
    .update(ingestionJobs)
    .set({ state: "pending", retryCount: 0, lastChecked: null })
    .where(eq(ingestionJobs.id, id))

  log.info("ingestion.started", { jobId: id }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true })
}
