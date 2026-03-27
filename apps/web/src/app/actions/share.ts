"use server"

import { requireUser } from "@/lib/auth/current-user"
import { createShare, revokeShare } from "@rag-saldivia/db"
import { headers } from "next/headers"

export async function actionCreateShare(sessionId: string, ttlDays = 7) {
  const user = await requireUser()
  const share = await createShare(sessionId, user.id, ttlDays)
  const h = await headers()
  const host = h.get("host") ?? "localhost:3000"
  const proto = h.get("x-forwarded-proto") ?? "http"
  const baseUrl = `${proto}://${host}`
  return { share, url: `${baseUrl}/share/${share.token}` }
}

export async function actionRevokeShare(id: string) {
  const user = await requireUser()
  await revokeShare(id, user.id)
}
