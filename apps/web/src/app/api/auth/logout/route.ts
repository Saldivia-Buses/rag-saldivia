/**
 * DELETE /api/auth/logout
 * Limpia la cookie de autenticación y agrega el jti a la blacklist Redis.
 */

import { NextResponse } from "next/server"
import { makeClearAuthCookie, extractClaims } from "@/lib/auth/jwt"
import { getRedisClient } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)

  if (claims) {
    log.info("auth.logout", { email: claims.email }, { userId: Number(claims.sub) })

    // Agregar jti a la blacklist con TTL = tiempo restante del token
    if (claims.jti && claims.exp > 0) {
      const ttl = claims.exp - Math.floor(Date.now() / 1000)
      if (ttl > 0) {
        getRedisClient()
          .set(`revoked:${claims.jti}`, "1", "EX", ttl)
          .catch(() => {})
      }
    }
  }

  const response = NextResponse.json({ ok: true })
  response.headers.set("Set-Cookie", makeClearAuthCookie())
  return response
}

// Alias para POST (algunos clientes usan POST para logout)
export const POST = DELETE
