/**
 * /api/messaging/channels — List user's channels / Create channel.
 */

import { NextResponse } from "next/server"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { getChannelsByUser, createChannel, getUnreadCounts } from "@rag-saldivia/db"
import { CreateChannelSchema } from "@rag-saldivia/shared"
import { clean } from "@/lib/safe-action"

export async function GET(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const userId = Number(claims.sub)
    const [channels, unreadCounts] = await Promise.all([
      getChannelsByUser(userId),
      getUnreadCounts(userId),
    ])
    return apiOk({ channels, unreadCounts })
  } catch (error) {
    return apiServerError(error, "GET /api/messaging/channels", Number(claims.sub))
  }
}

export async function POST(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = CreateChannelSchema.safeParse(body)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Datos inválidos")
  }

  try {
    const userId = Number(claims.sub)
    const channel = await createChannel(clean({
      type: parsed.data.type,
      name: parsed.data.name,
      description: parsed.data.description,
      createdBy: userId,
      memberIds: parsed.data.memberIds,
    }))
    return apiOk(channel, 201)
  } catch (error) {
    return apiServerError(error, "POST /api/messaging/channels", Number(claims.sub))
  }
}
