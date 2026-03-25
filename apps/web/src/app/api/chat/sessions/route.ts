/**
 * GET /api/chat/sessions
 * Lista las sesiones del usuario actual para la command palette.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { listSessionsByUser } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const userId = Number(claims.sub)
  const sessions = await listSessionsByUser(userId)

  return NextResponse.json({
    ok: true,
    data: sessions.map((s) => ({ id: s.id, title: s.title, collection: s.collection })),
  })
}
