/**
 * GET /api/audit/replay?from=YYYY-MM-DD
 * Reconstruye el estado del sistema desde una fecha.
 * Requiere rol admin.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getEventsForReplay } from "@rag-saldivia/db"
import { reconstructFromEvents } from "@rag-saldivia/logger/blackbox"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return NextResponse.json({ ok: false, error: "Se requiere rol admin" }, { status: 403 })
  }

  const url = new URL(request.url)
  const fromStr = url.searchParams.get("from")
  const toStr = url.searchParams.get("to")

  if (!fromStr) {
    return NextResponse.json(
      { ok: false, error: "Parámetro 'from' requerido (YYYY-MM-DD o epoch ms)" },
      { status: 400 }
    )
  }

  // Aceptar tanto YYYY-MM-DD como epoch ms
  const fromTs = isNaN(Number(fromStr))
    ? new Date(fromStr).getTime()
    : Number(fromStr)

  const toTs = toStr
    ? isNaN(Number(toStr)) ? new Date(toStr).getTime() : Number(toStr)
    : undefined

  if (isNaN(fromTs)) {
    return NextResponse.json({ ok: false, error: "Fecha inválida" }, { status: 400 })
  }

  const events = await getEventsForReplay(fromTs, toTs)
  const state = reconstructFromEvents(events)

  return NextResponse.json({
    ok: true,
    data: {
      timeline: events,
      stats: state.stats,
      errors: state.errors,
      ragQueryCount: state.ragQueries.length,
      uniqueUsers: state.stats.uniqueUsers,
    },
  })
}
