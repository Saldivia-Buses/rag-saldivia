/**
 * POST /api/auth/refresh
 * Renueva el par access+refresh. Revoca el refresh token viejo (rotation).
 *
 * Plan 26: migrado de token único 24h a access (15m) + refresh (7d) con rotation.
 * El refresh token viaja en cookie `refresh_token` (Path=/api/auth/refresh, SameSite=Strict).
 */

import { NextResponse } from "next/server"
import {
  verifyJwt,
  createAccessToken,
  createRefreshToken,
  makeAuthCookie,
  makeRefreshCookie,
  revokeToken,
} from "@/lib/auth/jwt"
import { getUserById } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

function extractRefreshFromCookie(request: Request): string | null {
  const cookieHeader = request.headers.get("cookie")
  if (!cookieHeader) return null
  const match = cookieHeader.match(/(?:^|;\s*)refresh_token=([^;]+)/)
  return match?.[1] ? decodeURIComponent(match[1]) : null
}

export async function POST(request: Request) {
  const raw = extractRefreshFromCookie(request)

  if (!raw) {
    return NextResponse.json(
      { ok: false, error: "No refresh token" },
      { status: 401 }
    )
  }

  const claims = await verifyJwt(raw)

  if (!claims || (claims as Record<string, unknown>).type !== "refresh") {
    return NextResponse.json(
      { ok: false, error: "Token inválido o expirado. Re-autenticarse." },
      { status: 401 }
    )
  }

  // Revoke old refresh token (rotation — prevents reuse)
  if (claims.jti && claims.exp > 0) {
    await revokeToken(claims.jti, claims.exp)
  }

  // Verify user is still active
  const user = await getUserById(Number(claims.sub))

  if (!user || !user.active) {
    return NextResponse.json(
      { ok: false, error: "Cuenta no encontrada o desactivada" },
      { status: 403 }
    )
  }

  const newAccess = await createAccessToken({
    sub: String(user.id),
    email: user.email,
    name: user.name,
    role: user.role as "admin" | "area_manager" | "user",
  })
  const newRefresh = await createRefreshToken(String(user.id))

  log.debug("auth.refresh", { email: user.email }, { userId: user.id })

  const response = NextResponse.json({ ok: true })
  response.headers.append("Set-Cookie", makeAuthCookie(newAccess))
  response.headers.append("Set-Cookie", makeRefreshCookie(newRefresh))
  return response
}
