/**
 * DELETE /api/auth/logout
 * Revoca ambos tokens (access + refresh) y limpia las cookies.
 *
 * Plan 26: migrado a access+refresh con revocación de ambos.
 */

import { NextResponse } from "next/server"
import {
  extractClaims,
  verifyJwt,
  makeClearAuthCookie,
  makeClearRefreshCookie,
  revokeToken,
} from "@/lib/auth/jwt"
import { log } from "@rag-saldivia/logger/backend"

export async function DELETE(request: Request) {
  const claims = await extractClaims(request)

  if (claims) {
    log.info("auth.logout", { email: claims.email }, { userId: Number(claims.sub) })

    // Revoke access token
    if (claims.jti && claims.exp > 0) {
      await revokeToken(claims.jti, claims.exp)
    }
  }

  // Revoke refresh token (from cookie)
  const cookieHeader = request.headers.get("cookie")
  if (cookieHeader) {
    const match = cookieHeader.match(/(?:^|;\s*)refresh_token=([^;]+)/)
    if (match?.[1]) {
      const refreshClaims = await verifyJwt(decodeURIComponent(match[1]))
      if (refreshClaims?.jti && refreshClaims.exp > 0) {
        await revokeToken(refreshClaims.jti, refreshClaims.exp)
      }
    }
  }

  const response = NextResponse.json({ ok: true })
  response.headers.append("Set-Cookie", makeClearAuthCookie())
  response.headers.append("Set-Cookie", makeClearRefreshCookie())
  return response
}

// Alias for POST (some clients use POST for logout)
export const POST = DELETE
