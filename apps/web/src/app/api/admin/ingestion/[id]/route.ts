import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionQueue } from "@rag-saldivia/db"
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

  const condition = claims.role === "admin"
    ? eq(ingestionQueue.id, id)
    : and(eq(ingestionQueue.id, id), eq(ingestionQueue.userId, userId))

  await db
    .update(ingestionQueue)
    .set({ status: "error", error: "Cancelado manualmente", completedAt: Date.now() })
    .where(condition)

  log.info("ingestion.cancelled", { jobId: id }, { userId })

  return NextResponse.json({ ok: true })
}
