/**
 * /api/messaging/channels/[id] — Get / Update / Archive channel.
 */

import { NextResponse } from "next/server"
import { z } from "zod"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { getChannel, updateChannel, archiveChannel } from "@rag-saldivia/db"
import { publishToChannel } from "@/lib/ws/publish"
import { clean } from "@/lib/safe-action"

const UpdateChannelSchema = z.object({
  name: z.string().min(1).max(80).optional(),
  description: z.string().max(500).optional(),
  topic: z.string().max(200).optional(),
})

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const { id } = await params
    const channel = await getChannel(id)
    if (!channel) return apiError("Canal no encontrado", 404)

    // Check membership
    const userId = Number(claims.sub)
    const isMember = channel.channelMembers.some((m) => m.userId === userId)
    if (!isMember && claims.role !== "admin") {
      return apiError("No sos miembro de este canal", 403)
    }

    return apiOk(channel)
  } catch (error) {
    return apiServerError(error, "GET /api/messaging/channels/[id]", Number(claims.sub))
  }
}

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = UpdateChannelSchema.safeParse(body)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Datos inválidos")
  }

  try {
    const { id } = await params
    const channel = await getChannel(id)
    if (!channel) return apiError("Canal no encontrado", 404)

    // Check permission: owner, admin, or system admin
    const userId = Number(claims.sub)
    const member = channel.channelMembers.find((m) => m.userId === userId)
    const canEdit = member?.role === "owner" || member?.role === "admin" || claims.role === "admin"
    if (!canEdit) return apiError("Sin permisos para editar este canal", 403)

    const updated = await updateChannel(id, clean(parsed.data))
    if (updated) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      publishToChannel(id, { type: "channel_updated", channel: updated as any })
    }
    return apiOk(updated)
  } catch (error) {
    return apiServerError(error, "PATCH /api/messaging/channels/[id]", Number(claims.sub))
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
    const channel = await getChannel(id)
    if (!channel) return apiError("Canal no encontrado", 404)

    // Only owner or system admin can archive
    const userId = Number(claims.sub)
    const member = channel.channelMembers.find((m) => m.userId === userId)
    const canArchive = member?.role === "owner" || claims.role === "admin"
    if (!canArchive) return apiError("Solo el propietario puede archivar el canal", 403)

    const archived = await archiveChannel(id)
    return apiOk(archived)
  } catch (error) {
    return apiServerError(error, "DELETE /api/messaging/channels/[id]", Number(claims.sub))
  }
}
