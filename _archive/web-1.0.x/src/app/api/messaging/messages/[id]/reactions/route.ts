/**
 * /api/messaging/messages/[id]/reactions — Add / Remove reaction.
 */

import { NextResponse } from "next/server"
import { z } from "zod"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { getMessage, addReaction, removeReaction, getReactions } from "@rag-saldivia/db"
import { publishToChannel } from "@/lib/ws/publish"

const ReactionInputSchema = z.object({
  emoji: z.string().min(1).max(8),
})

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const { id } = await params
    const reactions = await getReactions(id)
    return apiOk(reactions)
  } catch (error) {
    return apiServerError(error, "GET /api/messaging/messages/[id]/reactions", Number(claims.sub))
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = ReactionInputSchema.safeParse(body)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Emoji inválido")
  }

  try {
    const { id } = await params
    const message = await getMessage(id)
    if (!message) return apiError("Mensaje no encontrado", 404)

    const userId = Number(claims.sub)
    await addReaction(id, userId, parsed.data.emoji)

    publishToChannel(message.channelId, {
      type: "reaction_added",
      messageId: id,
      userId,
      emoji: parsed.data.emoji,
    })

    return apiOk(null, 201)
  } catch (error) {
    return apiServerError(error, "POST /api/messaging/messages/[id]/reactions", Number(claims.sub))
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = ReactionInputSchema.safeParse(body)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Emoji inválido")
  }

  try {
    const { id } = await params
    const message = await getMessage(id)
    if (!message) return apiError("Mensaje no encontrado", 404)

    const userId = Number(claims.sub)
    await removeReaction(id, userId, parsed.data.emoji)

    publishToChannel(message.channelId, {
      type: "reaction_removed",
      messageId: id,
      userId,
      emoji: parsed.data.emoji,
    })

    return apiOk()
  } catch (error) {
    return apiServerError(error, "DELETE /api/messaging/messages/[id]/reactions", Number(claims.sub))
  }
}
