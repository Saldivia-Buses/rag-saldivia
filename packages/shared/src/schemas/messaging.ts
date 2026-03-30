/**
 * Messaging Zod schemas — channels, messages, reactions, mentions
 */

import { z } from "zod"

// ── Channel types ──────────────────────────────────────────────────────────

export const ChannelTypeSchema = z.enum(["public", "private", "dm", "group_dm"])
export type ChannelType = z.infer<typeof ChannelTypeSchema>

export const ChannelMemberRoleSchema = z.enum(["owner", "admin", "member"])
export type ChannelMemberRole = z.infer<typeof ChannelMemberRoleSchema>

export const MessageTypeSchema = z.enum(["text", "system", "file"])
export type MessageType = z.infer<typeof MessageTypeSchema>

export const MentionTypeSchema = z.enum(["user", "channel", "everyone"])
export type MentionType = z.infer<typeof MentionTypeSchema>

// ── Channel schemas ────────────────────────────────────────────────────────

export const ChannelSchema = z.object({
  id: z.string().uuid(),
  type: ChannelTypeSchema,
  name: z.string().nullable(),
  description: z.string().nullable(),
  topic: z.string().nullable(),
  createdBy: z.number().int().nullable(),
  createdAt: z.number().int(),
  updatedAt: z.number().int(),
  archivedAt: z.number().int().nullable(),
})
export type Channel = z.infer<typeof ChannelSchema>

export const CreateChannelSchema = z.object({
  type: ChannelTypeSchema,
  name: z.string().min(1).max(80).optional(),
  description: z.string().max(500).optional(),
  memberIds: z.array(z.number().int()).optional(),
})
export type CreateChannel = z.infer<typeof CreateChannelSchema>

// ── Message schemas ────────────────────────────────────────────────────────

export const MsgMessageSchema = z.object({
  id: z.string().uuid(),
  channelId: z.string().uuid(),
  userId: z.number().int(),
  parentId: z.string().uuid().nullable(),
  content: z.string(),
  type: MessageTypeSchema,
  replyCount: z.number().int().default(0),
  lastReplyAt: z.number().int().nullable(),
  editedAt: z.number().int().nullable(),
  deletedAt: z.number().int().nullable(),
  metadata: z.record(z.string(), z.unknown()).nullable(),
  createdAt: z.number().int(),
})
export type MsgMessage = z.infer<typeof MsgMessageSchema>

export const SendMessageSchema = z.object({
  channelId: z.string().uuid(),
  content: z.string().min(1).max(10000),
  parentId: z.string().uuid().optional(),
  type: MessageTypeSchema.default("text"),
})
export type SendMessage = z.infer<typeof SendMessageSchema>

// ── Reaction schemas ───────────────────────────────────────────────────────

export const ReactionSchema = z.object({
  messageId: z.string().uuid(),
  userId: z.number().int(),
  emoji: z.string().min(1).max(8),
  createdAt: z.number().int(),
})
export type Reaction = z.infer<typeof ReactionSchema>

// ── WebSocket protocol ─────────────────────────────────────────────────────

export const ClientMessageSchema = z.discriminatedUnion("type", [
  z.object({ type: z.literal("auth"), token: z.string() }),
  z.object({ type: z.literal("subscribe"), channelId: z.string() }),
  z.object({ type: z.literal("unsubscribe"), channelId: z.string() }),
  z.object({ type: z.literal("typing_start"), channelId: z.string() }),
  z.object({ type: z.literal("typing_stop"), channelId: z.string() }),
  z.object({ type: z.literal("presence"), status: z.enum(["online", "away"]) }),
  z.object({ type: z.literal("sync"), channels: z.array(z.object({ id: z.string(), lastMessageAt: z.number() })) }),
  z.object({ type: z.literal("ping") }),
])
export type ClientMessage = z.infer<typeof ClientMessageSchema>

export type ServerMessage =
  | { type: "auth_ok"; userId: number }
  | { type: "auth_error"; reason: string }
  | { type: "message_new"; message: MsgMessage }
  | { type: "message_updated"; message: MsgMessage }
  | { type: "message_deleted"; messageId: string; channelId: string }
  | { type: "reaction_added"; messageId: string; userId: number; emoji: string }
  | { type: "reaction_removed"; messageId: string; userId: number; emoji: string }
  | { type: "typing"; channelId: string; userId: number; displayName: string }
  | { type: "presence_update"; userId: number; status: "online" | "away" | "offline" }
  | { type: "channel_updated"; channel: Channel }
  | { type: "member_joined"; channelId: string; userId: number }
  | { type: "member_left"; channelId: string; userId: number }
  | { type: "unread_update"; channelId: string; count: number }
  | { type: "sync_response"; channelId: string; messages: MsgMessage[] }
  | { type: "pong" }
