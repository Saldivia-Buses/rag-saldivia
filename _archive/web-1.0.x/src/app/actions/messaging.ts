"use server"

import { z } from "zod"
import { revalidatePath } from "next/cache"
import { authAction, clean } from "@/lib/safe-action"
import {
  sendMessage,
  editMessage,
  deleteMessage,
  getMessage,
  getChannel,
  createChannel,
  addChannelMember,
  removeChannelMember,
  pinMessage,
  unpinMessage,
  addReaction,
  removeReaction,
  updateLastRead,
} from "@rag-saldivia/db"
import { SendMessageSchema, CreateChannelSchema } from "@rag-saldivia/shared"
import { publishToChannel } from "@/lib/ws/publish"

// ── Messages ─────────────────────────────────────────────────────────────

export const actionSendMessage = authAction
  .schema(SendMessageSchema)
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    const message = await sendMessage(clean({
      channelId: data.channelId,
      userId: user.id,
      content: data.content,
      parentId: data.parentId,
      type: data.type,
    }))

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    publishToChannel(data.channelId, { type: "message_new", message: message as any })

    revalidatePath(`/messaging/${data.channelId}`)
    return message
  })

export const actionEditMessage = authAction
  .schema(z.object({ id: z.string().uuid(), content: z.string().min(1).max(10000) }))
  .action(async ({ parsedInput: { id, content }, ctx: { user } }) => {
    const msg = await getMessage(id)
    if (!msg) throw new Error("Mensaje no encontrado")
    if (msg.userId !== user.id) throw new Error("Solo podés editar tus propios mensajes")

    const updated = await editMessage(id, content)
    if (updated) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      publishToChannel(msg.channelId, { type: "message_updated", message: updated as any })
    }
    return updated
  })

export const actionDeleteMessage = authAction
  .schema(z.object({ id: z.string().uuid() }))
  .action(async ({ parsedInput: { id }, ctx: { user } }) => {
    const msg = await getMessage(id)
    if (!msg) throw new Error("Mensaje no encontrado")

    // Owner of message, channel admin/owner, or system admin can delete
    if (msg.userId !== user.id && user.role !== "admin") {
      const channel = await getChannel(msg.channelId)
      const member = channel?.channelMembers.find((m) => m.userId === user.id)
      if (member?.role !== "owner" && member?.role !== "admin") {
        throw new Error("Sin permisos para borrar este mensaje")
      }
    }

    await deleteMessage(id)
    publishToChannel(msg.channelId, {
      type: "message_deleted",
      messageId: id,
      channelId: msg.channelId,
    })
  })

// ── Channels ─────────────────────────────────────────────────────────────

export const actionCreateChannel = authAction
  .schema(CreateChannelSchema)
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    const channel = await createChannel(clean({
      type: data.type,
      name: data.name,
      description: data.description,
      createdBy: user.id,
      memberIds: data.memberIds,
    }))
    revalidatePath("/messaging")
    return channel
  })

export const actionJoinChannel = authAction
  .schema(z.object({ channelId: z.string().uuid() }))
  .action(async ({ parsedInput: { channelId }, ctx: { user } }) => {
    const channel = await getChannel(channelId)
    if (!channel) throw new Error("Canal no encontrado")
    if (channel.type === "private" || channel.type === "dm" || channel.type === "group_dm") {
      throw new Error("No se puede unir a un canal privado sin invitación")
    }

    await addChannelMember(channelId, user.id)
    publishToChannel(channelId, { type: "member_joined", channelId, userId: user.id })
    revalidatePath("/messaging")
  })

export const actionLeaveChannel = authAction
  .schema(z.object({ channelId: z.string().uuid() }))
  .action(async ({ parsedInput: { channelId }, ctx: { user } }) => {
    await removeChannelMember(channelId, user.id)
    publishToChannel(channelId, { type: "member_left", channelId, userId: user.id })
    revalidatePath("/messaging")
  })

// ── Pins ─────────────────────────────────────────────────────────────────

export const actionPinMessage = authAction
  .schema(z.object({ channelId: z.string().uuid(), messageId: z.string().uuid() }))
  .action(async ({ parsedInput: { channelId, messageId }, ctx: { user } }) => {
    await pinMessage(channelId, messageId, user.id)
  })

export const actionUnpinMessage = authAction
  .schema(z.object({ channelId: z.string().uuid(), messageId: z.string().uuid() }))
  .action(async ({ parsedInput: { channelId, messageId } }) => {
    await unpinMessage(channelId, messageId)
  })

// ── Reactions ────────────────────────────────────────────────────────────

export const actionReactToMessage = authAction
  .schema(z.object({ messageId: z.string().uuid(), emoji: z.string().min(1).max(8) }))
  .action(async ({ parsedInput: { messageId, emoji }, ctx: { user } }) => {
    const msg = await getMessage(messageId)
    if (!msg) throw new Error("Mensaje no encontrado")

    await addReaction(messageId, user.id, emoji)
    publishToChannel(msg.channelId, {
      type: "reaction_added",
      messageId,
      userId: user.id,
      emoji,
    })
  })

export const actionRemoveReaction = authAction
  .schema(z.object({ messageId: z.string().uuid(), emoji: z.string().min(1).max(8) }))
  .action(async ({ parsedInput: { messageId, emoji }, ctx: { user } }) => {
    const msg = await getMessage(messageId)
    if (!msg) throw new Error("Mensaje no encontrado")

    await removeReaction(messageId, user.id, emoji)
    publishToChannel(msg.channelId, {
      type: "reaction_removed",
      messageId,
      userId: user.id,
      emoji,
    })
  })

// ── Read status ──────────────────────────────────────────────────────────

export const actionMarkAsRead = authAction
  .schema(z.object({ channelId: z.string().uuid() }))
  .action(async ({ parsedInput: { channelId }, ctx: { user } }) => {
    await updateLastRead(channelId, user.id)
  })
