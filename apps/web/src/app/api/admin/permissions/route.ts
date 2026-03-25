/**
 * POST /api/admin/permissions — asignar colección a área (CLI / E2E)
 * DELETE /api/admin/permissions — quitar colección de área
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, areaCollections } from "@rag-saldivia/db"
import { and, eq } from "drizzle-orm"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body?.areaId || !body?.collectionName || !body?.permission) {
    return NextResponse.json({ ok: false, error: "areaId, collectionName y permission son requeridos" }, { status: 400 })
  }

  const db = getDb()
  await db.insert(areaCollections).values({
    areaId: body.areaId,
    collectionName: body.collectionName,
    permission: body.permission,
  }).onConflictDoNothing()

  return NextResponse.json({ ok: true })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body?.areaId || !body?.collectionName) {
    return NextResponse.json({ ok: false, error: "areaId y collectionName son requeridos" }, { status: 400 })
  }

  const db = getDb()
  await db.delete(areaCollections).where(
    and(eq(areaCollections.areaId, body.areaId), eq(areaCollections.collectionName, body.collectionName))
  )

  return NextResponse.json({ ok: true })
}
