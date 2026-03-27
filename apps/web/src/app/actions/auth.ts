"use server"

import { cookies } from "next/headers"
import { redirect } from "next/navigation"
import { getRedisClient } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"
import { verifyJwt } from "@/lib/auth/jwt"

export async function actionLogout() {
  const cookieStore = await cookies()
  const raw = cookieStore.get("auth_token")?.value
  if (raw) {
    const claims = await verifyJwt(raw)
    if (claims?.sub) {
      log.info("auth.logout", { email: claims.email }, { userId: Number(claims.sub) })
    }
    if (claims?.jti && claims.exp) {
      const ttl = claims.exp - Math.floor(Date.now() / 1000)
      if (ttl > 0) {
        getRedisClient()
          .set(`revoked:${claims.jti}`, "1", "EX", ttl)
          .catch(() => {})
      }
    }
  }
  cookieStore.delete("auth_token")
  redirect("/login")
}
