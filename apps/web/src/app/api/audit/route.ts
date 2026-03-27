/**
 * GET /api/audit
 * Retorna eventos del black box con filtros.
 * Requiere rol area_manager o admin.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { queryEvents } from "@rag-saldivia/db"
import type { LogLevel, EventSource, EventType } from "@rag-saldivia/shared"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  if (claims.role === "user") {
    return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })
  }

  const url = new URL(request.url)
  const limit = Math.min(parseInt(url.searchParams.get("limit") ?? "100"), 500)
  const offset = parseInt(url.searchParams.get("offset") ?? "0")
  const level = url.searchParams.get("level") as LogLevel | null
  const type = url.searchParams.get("type") as EventType | null
  const source = url.searchParams.get("source") as EventSource | null
  const fromTs = url.searchParams.get("from") ? parseInt(url.searchParams.get("from")!) : undefined
  const toTs = url.searchParams.get("to") ? parseInt(url.searchParams.get("to")!) : undefined
  const order = (url.searchParams.get("order") ?? "desc") as "asc" | "desc"

  // Usuarios no-admin solo ven sus propios eventos
  const userId = claims.role === "admin" ? undefined : Number(claims.sub)

  const events = await queryEvents({
    ...(fromTs !== undefined ? { fromTs } : {}),
    ...(toTs !== undefined ? { toTs } : {}),
    ...(level ? { level } : {}),
    ...(type ? { type } : {}),
    ...(source ? { source } : {}),
    ...(userId !== undefined ? { userId } : {}),
    limit,
    offset,
    order,
  })

  return NextResponse.json({ ok: true, data: events })
}
