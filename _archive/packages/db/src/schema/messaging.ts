/**
 * Messaging domain tables: channels, members, messages, reactions,
 * mentions, pinned messages.
 *
 * Plan 25 — Internal Messaging System.
 */

import {
  sqliteTable,
  text,
  integer,
  primaryKey,
  index,
} from "drizzle-orm/sqlite-core"
import { users } from "./core"

// ── Channels ───────────────────────────────────────────────────────────────

export const channels = sqliteTable("channels", {
  id: text("id").primaryKey(),
  type: text("type", { enum: ["public", "private", "dm", "group_dm"] }).notNull(),
  name: text("name"),
  description: text("description"),
  topic: text("topic"),
  createdBy: integer("created_by").references(() => users.id),
  createdAt: integer("created_at").notNull(),
  updatedAt: integer("updated_at").notNull(),
  archivedAt: integer("archived_at"),
}, (t) => ({
  typeIdx: index("idx_channels_type").on(t.type),
}))

// ── Channel Members ────────────────────────────────────────────────────────

export const channelMembers = sqliteTable("channel_members", {
  channelId: text("channel_id").references(() => channels.id, { onDelete: "cascade" }).notNull(),
  userId: integer("user_id").references(() => users.id, { onDelete: "cascade" }).notNull(),
  role: text("role", { enum: ["owner", "admin", "member"] }).notNull().default("member"),
  lastReadAt: integer("last_read_at").notNull(),
  muted: integer("muted", { mode: "boolean" }).notNull().default(false),
  joinedAt: integer("joined_at").notNull(),
}, (t) => ({
  pk: primaryKey({ columns: [t.channelId, t.userId] }),
}))

// ── Messages ───────────────────────────────────────────────────────────────

export const msgMessages = sqliteTable("msg_messages", {
  id: text("id").primaryKey(),
  channelId: text("channel_id").references(() => channels.id, { onDelete: "cascade" }).notNull(),
  userId: integer("user_id").references(() => users.id, { onDelete: "cascade" }).notNull(),
  parentId: text("parent_id"),
  content: text("content").notNull(),
  type: text("type", { enum: ["text", "system", "file"] }).notNull().default("text"),
  replyCount: integer("reply_count").notNull().default(0),
  lastReplyAt: integer("last_reply_at"),
  editedAt: integer("edited_at"),
  deletedAt: integer("deleted_at"),
  metadata: text("metadata", { mode: "json" }).$type<Record<string, unknown>>(),
  createdAt: integer("created_at").notNull(),
}, (t) => ({
  channelCreatedIdx: index("idx_msg_channel_created").on(t.channelId, t.createdAt),
  parentIdx: index("idx_msg_parent").on(t.parentId),
}))

// ── Reactions ──────────────────────────────────────────────────────────────

export const msgReactions = sqliteTable("msg_reactions", {
  messageId: text("message_id").references(() => msgMessages.id, { onDelete: "cascade" }).notNull(),
  userId: integer("user_id").references(() => users.id, { onDelete: "cascade" }).notNull(),
  emoji: text("emoji").notNull(),
  createdAt: integer("created_at").notNull(),
}, (t) => ({
  pk: primaryKey({ columns: [t.messageId, t.userId, t.emoji] }),
}))

// ── Mentions ───────────────────────────────────────────────────────────────

export const msgMentions = sqliteTable("msg_mentions", {
  id: text("id").primaryKey(),
  messageId: text("message_id").references(() => msgMessages.id, { onDelete: "cascade" }).notNull(),
  userId: integer("user_id"),
  type: text("type", { enum: ["user", "channel", "everyone"] }).notNull(),
}, (t) => ({
  userIdx: index("idx_msg_mentions_user").on(t.userId),
}))

// ── Pinned Messages ────────────────────────────────────────────────────────

export const pinnedMessages = sqliteTable("pinned_messages", {
  channelId: text("channel_id").references(() => channels.id, { onDelete: "cascade" }).notNull(),
  messageId: text("message_id").references(() => msgMessages.id, { onDelete: "cascade" }).notNull(),
  pinnedBy: integer("pinned_by").references(() => users.id).notNull(),
  pinnedAt: integer("pinned_at").notNull(),
}, (t) => ({
  pk: primaryKey({ columns: [t.channelId, t.messageId] }),
}))

// ── Type exports (Drizzle inferred) ───────────────────────────────────────

export type DbChannel = typeof channels.$inferSelect
export type NewChannel = typeof channels.$inferInsert
export type DbChannelMember = typeof channelMembers.$inferSelect
export type NewChannelMember = typeof channelMembers.$inferInsert
export type DbMsgMessage = typeof msgMessages.$inferSelect
export type NewMsgMessage = typeof msgMessages.$inferInsert
export type DbMsgReaction = typeof msgReactions.$inferSelect
export type DbMsgMention = typeof msgMentions.$inferSelect
export type NewMsgMention = typeof msgMentions.$inferInsert
export type DbPinnedMessage = typeof pinnedMessages.$inferSelect
