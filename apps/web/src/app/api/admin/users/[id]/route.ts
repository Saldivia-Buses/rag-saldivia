/**
 * DELETE /api/admin/users/[id] — eliminar usuario (CLI)
 * PATCH  /api/admin/users/[id] — actualizar usuario (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { deleteUser, updateUser, getUserById } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const { id } = await params
  const userId = parseInt(id)
  if (isNaN(userId)) return NextResponse.json({ ok: false, error: "ID inválido" }, { status: 400 })

  const existing = await getUserById(userId)
  if (!existing) return NextResponse.json({ ok: false, error: "Usuario no encontrado" }, { status: 404 })

  await deleteUser(userId)
  log.info("user.deleted", { targetUserId: userId }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true })
}

export async function PATCH(
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
  if (!body) return NextResponse.json({ ok: false, error: "Body requerido" }, { status: 400 })

  await updateUser(userId, body)
  log.info("user.updated", { targetUserId: userId, changes: body }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true })
}
