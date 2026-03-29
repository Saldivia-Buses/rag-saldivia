/**
 * DELETE /api/admin/areas/[id] — eliminar área (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, areas } from "@rag-saldivia/db"
import { eq } from "drizzle-orm"

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const { id } = await params
  const areaId = parseInt(id)
  if (isNaN(areaId)) return NextResponse.json({ ok: false, error: "ID inválido" }, { status: 400 })

  const db = getDb()
  const existing = await db.select({ id: areas.id }).from(areas).where(eq(areas.id, areaId)).limit(1)
  if (existing.length === 0) return NextResponse.json({ ok: false, error: "Área no encontrada" }, { status: 404 })

  await db.delete(areas).where(eq(areas.id, areaId))
  return NextResponse.json({ ok: true })
}
