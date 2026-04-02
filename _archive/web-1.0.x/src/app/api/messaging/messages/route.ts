/**
 * /api/messaging/messages — List messages (cursor-based) / Send message.
 */

import { NextResponse } from "next/server"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { getMessages, sendMessage, getChannel, updateLastRead } from "@rag-saldivia/db"
import { SendMessageSchema } from "@rag-saldivia/shared"
import { publishToChannel } from "@/lib/ws/publish"
import { clean } from "@/lib/safe-action"

export async function GET(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const { searchParams } = new URL(request.url)
  const channelId = searchParams.get("channelId")
  if (!channelId) return apiError("channelId requerido")

  const before = searchParams.get("before") ? Number(searchParams.get("before")) : undefined
  const limit = searchParams.get("limit") ? Number(searchParams.get("limit")) : 50

  try {
    const userId = Number(claims.sub)

    // Verify membership
    const channel = await getChannel(channelId)
    if (!channel) return apiError("Canal no encontrado", 404)
    const isMember = channel.channelMembers.some((m) => m.userId === userId)
    if (!isMember && claims.role !== "admin") {
      return apiError("No sos miembro de este canal", 403)
    }

    const messages = await getMessages(channelId, clean({ before, limit }))

    // Mark as read
    await updateLastRead(channelId, userId)

    return apiOk(messages)
  } catch (error) {
    return apiServerError(error, "GET /api/messaging/messages", Number(claims.sub))
  }
}

export async function POST(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = SendMessageSchema.safeParse(body)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Datos inválidos")
  }

  try {
    const userId = Number(claims.sub)

    // Verify membership
    const channel = await getChannel(parsed.data.channelId)
    if (!channel) return apiError("Canal no encontrado", 404)
    const isMember = channel.channelMembers.some((m) => m.userId === userId)
    if (!isMember && claims.role !== "admin") {
      return apiError("No sos miembro de este canal", 403)
    }

    const message = await sendMessage(clean({
      channelId: parsed.data.channelId,
      userId,
      content: parsed.data.content,
      parentId: parsed.data.parentId,
      type: parsed.data.type,
    }))

    // Broadcast via WS
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    publishToChannel(parsed.data.channelId, { type: "message_new", message: message as any })

    return apiOk(message, 201)
  } catch (error) {
    return apiServerError(error, "POST /api/messaging/messages", Number(claims.sub))
  }
}
