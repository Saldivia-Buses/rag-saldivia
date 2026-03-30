/**
 * /api/messaging/channels/[id]/members — List / Add / Remove members.
 */

import { NextResponse } from "next/server"
import { z } from "zod"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import {
  getChannel,
  getChannelMembers,
  addChannelMember,
  removeChannelMember,
} from "@rag-saldivia/db"
import { publishToChannel } from "@/lib/ws/publish"

const AddMemberSchema = z.object({
  userId: z.number().int(),
  role: z.enum(["admin", "member"]).default("member"),
})

const RemoveMemberSchema = z.object({
  userId: z.number().int(),
})

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const { id } = await params
    const members = await getChannelMembers(id)
    return apiOk(members)
  } catch (error) {
    return apiServerError(error, "GET /api/messaging/channels/[id]/members", Number(claims.sub))
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = AddMemberSchema.safeParse(body)
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
    const canInvite = member?.role === "owner" || member?.role === "admin" || claims.role === "admin"
    if (!canInvite) return apiError("Sin permisos para invitar miembros", 403)

    await addChannelMember(id, parsed.data.userId, parsed.data.role)
    publishToChannel(id, { type: "member_joined", channelId: id, userId: parsed.data.userId })
    return apiOk(null, 201)
  } catch (error) {
    return apiServerError(error, "POST /api/messaging/channels/[id]/members", Number(claims.sub))
  }
}

export async function DELETE(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = RemoveMemberSchema.safeParse(body)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Datos inválidos")
  }

  try {
    const { id } = await params
    const channel = await getChannel(id)
    if (!channel) return apiError("Canal no encontrado", 404)

    const userId = Number(claims.sub)
    const member = channel.channelMembers.find((m) => m.userId === userId)
    const targetIsSelf = parsed.data.userId === userId

    // Can leave yourself, or kick if owner/admin/system admin
    const canRemove = targetIsSelf || member?.role === "owner" || member?.role === "admin" || claims.role === "admin"
    if (!canRemove) return apiError("Sin permisos para remover miembros", 403)

    await removeChannelMember(id, parsed.data.userId)
    publishToChannel(id, { type: "member_left", channelId: id, userId: parsed.data.userId })
    return apiOk()
  } catch (error) {
    return apiServerError(error, "DELETE /api/messaging/channels/[id]/members", Number(claims.sub))
  }
}
