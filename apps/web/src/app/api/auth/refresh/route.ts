/**
 * POST /api/auth/refresh
 * Renueva el JWT si aún es válido (aunque esté por expirar).
 */

import { NextResponse } from "next/server"
import { extractClaims, createJwt, makeAuthCookie } from "@/lib/auth/jwt"
import { getUserById } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function POST(request: Request) {
  const claims = await extractClaims(request)

  if (!claims) {
    return NextResponse.json(
      { ok: false, error: "Token inválido o expirado. Re-autenticarse." },
      { status: 401 }
    )
  }

  // Verificar que el usuario siga activo en la DB
  const user = await getUserById(Number(claims.sub))

  if (!user || !user.active) {
    return NextResponse.json(
      { ok: false, error: "Cuenta no encontrada o desactivada" },
      { status: 403 }
    )
  }

  const newToken = await createJwt({
    sub: String(user.id),
    email: user.email,
    name: user.name,
    role: user.role as "admin" | "area_manager" | "user",
  })

  log.debug("auth.refresh", { email: user.email }, { userId: user.id })

  const response = NextResponse.json({ ok: true })
  response.headers.set("Set-Cookie", makeAuthCookie(newToken))
  return response
}
