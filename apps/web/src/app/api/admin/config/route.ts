/**
 * GET   /api/admin/config — obtener parámetros RAG actuales (CLI)
 * PATCH /api/admin/config — actualizar un parámetro (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { loadRagParams, saveRagParams } from "@rag-saldivia/config"
import { log } from "@rag-saldivia/logger/backend"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const params = loadRagParams()
  return NextResponse.json({ ok: true, data: params })
}

export async function PATCH(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const body = await request.json().catch(() => null)
  if (!body) return NextResponse.json({ ok: false, error: "Body requerido" }, { status: 400 })

  await saveRagParams(body)
  log.info("admin.config_changed", { changes: body }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true })
}
