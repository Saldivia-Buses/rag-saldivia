/**
 * /api/messaging/search — Full-text search across messages using FTS5.
 */

import { NextResponse } from "next/server"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { searchMessages, getChannelsByUser } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const { searchParams } = new URL(request.url)
  const query = searchParams.get("q")
  if (!query || !query.trim()) return apiError("Query requerida")

  const limit = searchParams.get("limit") ? Number(searchParams.get("limit")) : 20

  try {
    const userId = Number(claims.sub)

    // Search only in user's channels
    const userChannels = await getChannelsByUser(userId)
    const channelIds = userChannels.map((c) => c.id)

    if (channelIds.length === 0) return apiOk([])

    const results = await searchMessages(query, channelIds, limit)
    return apiOk(results)
  } catch (error) {
    return apiServerError(error, "GET /api/messaging/search", Number(claims.sub))
  }
}
