/**
 * GET  /api/admin/areas — listar áreas (CLI)
 * POST /api/admin/areas — crear área (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, areas } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const db = getDb()
  const list = await db.select().from(areas)
  return NextResponse.json({ ok: true, data: list })
}

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body?.name) return NextResponse.json({ ok: false, error: "name es requerido" }, { status: 400 })

  const db = getDb()
  const [area] = await db.insert(areas).values({
    name: body.name,
    description: body.description ?? "",
    createdAt: Date.now(),
  }).returning()

  log.info("user.area_assigned", { areaName: body.name }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true, data: area })
}
