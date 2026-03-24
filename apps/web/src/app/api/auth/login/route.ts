/**
 * POST /api/auth/login
 * Verifica credenciales contra la DB y emite JWT en cookie HttpOnly.
 */

import { NextResponse } from "next/server"
import { LoginRequestSchema } from "@rag-saldivia/shared"
import { verifyPassword } from "@rag-saldivia/db"
import { createJwt, makeAuthCookie } from "@/lib/auth/jwt"
import { log } from "@rag-saldivia/logger/backend"

export async function POST(request: Request) {
  const start = Date.now()

  try {
    const body = await request.json()
    const parsed = LoginRequestSchema.safeParse(body)

    if (!parsed.success) {
      return NextResponse.json(
        { ok: false, error: "Datos inválidos", details: parsed.error.flatten() },
        { status: 400 }
      )
    }

    const { email, password } = parsed.data
    const user = await verifyPassword(email, password)

    if (!user) {
      log.warn("auth.failed", { email, reason: "invalid_credentials" })
      // Delay deliberado para dificultar brute force
      await new Promise((r) => setTimeout(r, 300))
      return NextResponse.json(
        { ok: false, error: "Email o contraseña incorrectos" },
        { status: 401 }
      )
    }

    if (!user.active) {
      log.warn("auth.failed", { email, reason: "account_inactive", userId: user.id })
      return NextResponse.json(
        { ok: false, error: "Cuenta desactivada. Contactá al administrador." },
        { status: 403 }
      )
    }

    const token = await createJwt({
      sub: String(user.id),
      email: user.email,
      name: user.name,
      role: user.role as "admin" | "area_manager" | "user",
    })

    log.info("auth.login", { email, role: user.role, duration: Date.now() - start }, { userId: user.id })

    const response = NextResponse.json({
      ok: true,
      data: {
        user: {
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role,
          active: user.active,
          preferences: user.preferences ?? {},
          createdAt: user.createdAt,
          lastLogin: user.lastLogin ?? null,
        },
      },
    })

    response.headers.set("Set-Cookie", makeAuthCookie(token))
    return response
  } catch (error) {
    log.error("system.error", { error: String(error), endpoint: "POST /api/auth/login" })
    return NextResponse.json(
      { ok: false, error: "Error interno del servidor" },
      { status: 500 }
    )
  }
}
