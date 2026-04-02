/**
 * /api/messaging/messages/[id]/pin — Pin / Unpin message.
 */

import { NextResponse } from "next/server"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { getMessage, getChannel, pinMessage, unpinMessage } from "@rag-saldivia/db"

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const { id } = await params
    const message = await getMessage(id)
    if (!message) return apiError("Mensaje no encontrado", 404)

    const userId = Number(claims.sub)
    const channel = await getChannel(message.channelId)
    if (!channel) return apiError("Canal no encontrado", 404)

    // Only owner, admin, or system admin can pin
    const member = channel.channelMembers.find((m) => m.userId === userId)
    const canPin = member?.role === "owner" || member?.role === "admin" || claims.role === "admin"
    if (!canPin) return apiError("Sin permisos para fijar mensajes", 403)

    await pinMessage(message.channelId, id, userId)
    return apiOk()
  } catch (error) {
    return apiServerError(error, "POST /api/messaging/messages/[id]/pin", Number(claims.sub))
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const { id } = await params
    const message = await getMessage(id)
    if (!message) return apiError("Mensaje no encontrado", 404)

    const userId = Number(claims.sub)
    const channel = await getChannel(message.channelId)
    if (!channel) return apiError("Canal no encontrado", 404)

    const member = channel.channelMembers.find((m) => m.userId === userId)
    const canUnpin = member?.role === "owner" || member?.role === "admin" || claims.role === "admin"
    if (!canUnpin) return apiError("Sin permisos para desfijar mensajes", 403)

    await unpinMessage(message.channelId, id)
    return apiOk()
  } catch (error) {
    return apiServerError(error, "DELETE /api/messaging/messages/[id]/pin", Number(claims.sub))
  }
}
