/**
 * GET /api/search?q=...
 * Búsqueda universal sobre sesiones, mensajes, templates y guardados del usuario.
 * F3.39 — usa FTS5 con fallback a LIKE.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { universalSearch } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })

  const { searchParams } = new URL(request.url)
  const q = searchParams.get("q") ?? ""

  if (q.trim().length < 2) {
    return NextResponse.json({ ok: true, results: [] })
  }

  const results = await universalSearch(q, Number(claims.sub), 20)
  return NextResponse.json({ ok: true, results })
}
