/**
 * POST /api/admin/users/[id]/areas — asignar área a usuario
 * DELETE /api/admin/users/[id]/areas — quitar área de usuario
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { addUserArea, removeUserArea } from "@rag-saldivia/db"

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const { id } = await params
  const userId = parseInt(id)
  if (isNaN(userId)) return NextResponse.json({ ok: false, error: "ID inválido" }, { status: 400 })

  const body = await request.json().catch(() => null)
  if (!body?.areaId) return NextResponse.json({ ok: false, error: "areaId requerido" }, { status: 400 })

  await addUserArea(userId, body.areaId)
  return NextResponse.json({ ok: true })
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const { id } = await params
  const userId = parseInt(id)
  const body = await request.json().catch(() => null)
  if (!body?.areaId) return NextResponse.json({ ok: false, error: "areaId requerido" }, { status: 400 })

  await removeUserArea(userId, body.areaId)
  return NextResponse.json({ ok: true })
}
