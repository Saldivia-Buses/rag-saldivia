/**
 * POST /api/share — crear token de compartir sesión (requiere auth)
 * DELETE /api/share?id=X — revocar share (requiere auth)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { createShare, revokeShare } from "@rag-saldivia/db"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const body = await request.json().catch(() => null)
  if (!body?.sessionId) {
    return NextResponse.json({ ok: false, error: "sessionId requerido" }, { status: 400 })
  }

  const share = await createShare(body.sessionId as string, Number(claims.sub), body.ttlDays as number | undefined)
  const baseUrl = process.env["NEXTAUTH_URL"] ?? process.env["NEXT_PUBLIC_BASE_URL"] ?? "http://localhost:3000"

  return NextResponse.json({
    ok: true,
    share,
    url: `${baseUrl}/share/${share.token}`,
  })
}

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const { searchParams } = new URL(request.url)
  const id = searchParams.get("id")
  if (!id) return NextResponse.json({ ok: false, error: "id requerido" }, { status: 400 })

  await revokeShare(id, Number(claims.sub))
  return NextResponse.json({ ok: true })
}
