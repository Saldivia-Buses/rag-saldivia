/**
 * DELETE /api/auth/logout
 * Limpia la cookie de autenticación.
 */

import { NextResponse } from "next/server"
import { makeClearAuthCookie } from "@/lib/auth/jwt"
import { extractClaims } from "@/lib/auth/jwt"
import { log } from "@rag-saldivia/logger/backend"

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)

  if (claims) {
    log.info("auth.logout", { email: claims.email }, { userId: Number(claims.sub) })
  }

  const response = NextResponse.json({ ok: true })
  response.headers.set("Set-Cookie", makeClearAuthCookie())
  return response
}

// Alias para POST (algunos clientes usan POST para logout)
export const POST = DELETE
