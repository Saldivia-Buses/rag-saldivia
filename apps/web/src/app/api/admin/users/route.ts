/**
 * GET  /api/admin/users — listar usuarios (CLI)
 * POST /api/admin/users — crear usuario (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, listUsers, createUser, users } from "@rag-saldivia/db"
import { eq } from "drizzle-orm"
import { log } from "@rag-saldivia/logger/backend"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const list = await listUsers()
  return NextResponse.json({ ok: true, data: list })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body?.email || !body?.password || !body?.name) {
    return NextResponse.json({ ok: false, error: "email, name y password son requeridos" }, { status: 400 })
  }

  try {
    const user = await createUser({
      email: body.email,
      name: body.name,
      password: body.password,
      role: body.role ?? "user",
    })
    log.info("user.created", { email: body.email, role: body.role ?? "user" }, { userId: Number(claims.sub) })
    return NextResponse.json({ ok: true, data: user })
  } catch (err) {
    const msg = String(err)
    if (msg.includes("UNIQUE") || msg.includes("unique")) {
      return NextResponse.json({ ok: false, error: "El email ya existe" }, { status: 409 })
    }
    return NextResponse.json({ ok: false, error: msg }, { status: 500 })
  }
}
