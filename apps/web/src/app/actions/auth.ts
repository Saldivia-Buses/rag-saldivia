"use server"

import { cookies } from "next/headers"
import { redirect } from "next/navigation"
import { getRedisClient } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"
import { verifyJwt } from "@/lib/auth/jwt"
import { authAction } from "@/lib/safe-action"

export const actionLogout = authAction.action(async ({ ctx: { user } }) => {
  const cookieStore = await cookies()
  const raw = cookieStore.get("auth_token")?.value
  if (raw) {
    const claims = await verifyJwt(raw)
    if (claims?.jti && claims.exp) {
      const ttl = claims.exp - Math.floor(Date.now() / 1000)
      if (ttl > 0) {
        getRedisClient()
          .set(`revoked:${claims.jti}`, "1", "EX", ttl)
          .catch(() => {})
      }
    }
  }
  log.info("auth.logout", { email: user.email }, { userId: user.id })
  cookieStore.delete("auth_token")
  redirect("/login")
})
