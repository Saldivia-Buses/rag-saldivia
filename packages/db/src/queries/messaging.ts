/**
 * Messaging queries — messages, reactions, pins, threads, FTS5 search.
 * Cursor-based pagination with composite cursor (createdAt, id).
 */

import { eq, and, desc, sql, lt, isNull } from "drizzle-orm"
import { getDb } from "../connection"
import { msgMessages, msgReactions, msgMentions, pinnedMessages, channels } from "../schema"
import { randomUUID } from "crypto"

function now() { return Date.now() }

// ── Messages CRUD ─────────────────────────────────────────────────────────

export async function sendMessage(data: {
  channelId: string
  userId: number
  content: string
  parentId?: string
  type?: "text" | "system" | "file"
  metadata?: Record<string, unknown>
}) {
  const db = getDb()
  const id = randomUUID()
  const ts = now()

  const [message] = await db.insert(msgMessages).values({
    id,
    channelId: data.channelId,
    userId: data.userId,
    parentId: data.parentId ?? null,
    content: data.content,
    type: data.type ?? "text",
    replyCount: 0,
    lastReplyAt: null,
    editedAt: null,
    deletedAt: null,
    metadata: data.metadata ?? null,
    createdAt: ts,
  }).returning()

  // Update parent reply count if this is a thread reply
  if (data.parentId) {
    await db
      .update(msgMessages)
      .set({
        replyCount: sql`${msgMessages.replyCount} + 1`,
        lastReplyAt: ts,
      })
      .where(eq(msgMessages.id, data.parentId))
  }

  // Update channel updatedAt
  await db
    .update(channels)
    .set({ updatedAt: ts })
    .where(eq(channels.id, data.channelId))

  return message!
}

/**
 * Get messages with cursor-based pagination.
 * Returns messages older than the cursor, newest first.
 */
export async function getMessages(
  channelId: string,
  opts: { before?: number; limit?: number } = {}
) {
  const { before, limit = 50 } = opts
  const db = getDb()

  const conditions = [
    eq(msgMessages.channelId, channelId),
    isNull(msgMessages.parentId), // Only top-level messages, not thread replies
  ]
  if (before) {
    conditions.push(lt(msgMessages.createdAt, before))
  }

  return db
    .select()
    .from(msgMessages)
    .where(and(...conditions))
    .orderBy(desc(msgMessages.createdAt))
    .limit(limit)
}

export async function editMessage(id: string, content: string) {
  const [updated] = await getDb()
    .update(msgMessages)
    .set({ content, editedAt: now() })
    .where(eq(msgMessages.id, id))
    .returning()
  return updated
}

export async function deleteMessage(id: string) {
  // Soft delete
  const [updated] = await getDb()
    .update(msgMessages)
    .set({ deletedAt: now(), content: "" })
    .where(eq(msgMessages.id, id))
    .returning()
  return updated
}

export async function getMessage(id: string) {
  return getDb().query.msgMessages.findFirst({
    where: (m, { eq }) => eq(m.id, id),
  })
}

// ── Thread replies ────────────────────────────────────────────────────────

export async function getThreadReplies(parentId: string, limit = 50) {
  return getDb()
    .select()
    .from(msgMessages)
    .where(eq(msgMessages.parentId, parentId))
    .orderBy(msgMessages.createdAt) // oldest first for threads
    .limit(limit)
}

export async function getThreadPreview(parentId: string) {
  const replies = await getDb()
    .select()
    .from(msgMessages)
    .where(eq(msgMessages.parentId, parentId))
    .orderBy(desc(msgMessages.createdAt))
    .limit(3)
  return replies.reverse() // Show in chronological order
}

// ── Reactions ─────────────────────────────────────────────────────────────

export async function addReaction(messageId: string, userId: number, emoji: string) {
  await getDb()
    .insert(msgReactions)
    .values({ messageId, userId, emoji, createdAt: now() })
    .onConflictDoNothing()
}

export async function removeReaction(messageId: string, userId: number, emoji: string) {
  await getDb()
    .delete(msgReactions)
    .where(
      and(
        eq(msgReactions.messageId, messageId),
        eq(msgReactions.userId, userId),
        eq(msgReactions.emoji, emoji),
      )
    )
}

export async function getReactions(messageId: string) {
  return getDb().query.msgReactions.findMany({
    where: (r, { eq }) => eq(r.messageId, messageId),
  })
}

// ── Pins ──────────────────────────────────────────────────────────────────

export async function pinMessage(channelId: string, messageId: string, pinnedBy: number) {
  await getDb()
    .insert(pinnedMessages)
    .values({ channelId, messageId, pinnedBy, pinnedAt: now() })
    .onConflictDoNothing()
}

export async function unpinMessage(channelId: string, messageId: string) {
  await getDb()
    .delete(pinnedMessages)
    .where(
      and(eq(pinnedMessages.channelId, channelId), eq(pinnedMessages.messageId, messageId))
    )
}

export async function getPinnedMessages(channelId: string) {
  return getDb().query.pinnedMessages.findMany({
    where: (p, { eq }) => eq(p.channelId, channelId),
    with: { message: true },
    orderBy: (p, { desc }) => [desc(p.pinnedAt)],
  })
}

// ── Search (FTS5) ─────────────────────────────────────────────────────────

/**
 * Full-text search across messages using FTS5.
 * Returns messages matching the query, filtered by channel access.
 */
export async function searchMessages(
  query: string,
  channelIds: string[],
  limit = 20
): Promise<Array<{ id: string; channelId: string; userId: number; content: string; createdAt: number }>> {
  if (!query.trim() || channelIds.length === 0) return []

  const db = getDb()
  // FTS5 search with channel filter
  const placeholders = channelIds.map(() => "?").join(",")
  const results = await db.all(sql.raw(
    `SELECT m.id, m.channel_id as channelId, m.user_id as userId, m.content, m.created_at as createdAt
     FROM msg_messages m
     JOIN msg_messages_fts fts ON m.rowid = fts.rowid
     WHERE msg_messages_fts MATCH '${query.replace(/'/g, "''")}'
     AND m.channel_id IN (${placeholders})
     AND m.deleted_at IS NULL
     ORDER BY rank
     LIMIT ${limit}`
  ))

  return results as Array<{ id: string; channelId: string; userId: number; content: string; createdAt: number }>
}

// ── Mentions ──────────────────────────────────────────────────────────────

export async function addMentions(
  messageId: string,
  mentions: Array<{ userId?: number; type: "user" | "channel" | "everyone" }>
) {
  if (mentions.length === 0) return
  const db = getDb()
  await db.insert(msgMentions).values(
    mentions.map((m) => ({
      id: randomUUID(),
      messageId,
      userId: m.userId ?? null,
      type: m.type,
    }))
  )
}

export async function getMentionsForUser(userId: number, limit = 50) {
  return getDb().query.msgMentions.findMany({
    where: (m, { eq }) => eq(m.userId, userId),
    with: { message: true },
    orderBy: (m, { desc }) => [desc(m.id)],
    limit,
  })
}
