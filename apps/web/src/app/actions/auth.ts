"use server"

import { cookies } from "next/headers"
import { redirect } from "next/navigation"
import { log } from "@rag-saldivia/logger/backend"
import { verifyJwt, revokeToken } from "@/lib/auth/jwt"
import { authAction } from "@/lib/safe-action"

export const actionLogout = authAction.action(async ({ ctx: { user } }) => {
  const cookieStore = await cookies()

  // Revoke access token
  const accessRaw = cookieStore.get("auth_token")?.value
  if (accessRaw) {
    const claims = await verifyJwt(accessRaw)
    if (claims?.jti && claims.exp) await revokeToken(claims.jti, claims.exp)
  }

  // Revoke refresh token
  const refreshRaw = cookieStore.get("refresh_token")?.value
  if (refreshRaw) {
    const claims = await verifyJwt(refreshRaw)
    if (claims?.jti && claims.exp) await revokeToken(claims.jti, claims.exp)
  }

  log.info("auth.logout", { email: user.email }, { userId: user.id })
  cookieStore.delete("auth_token")
  cookieStore.delete("refresh_token")
  redirect("/login")
})
