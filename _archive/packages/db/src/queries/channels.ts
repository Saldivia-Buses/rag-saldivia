/**
 * Channel queries — CRUD, members, unread counts.
 */

import { eq, and, desc, sql, inArray } from "drizzle-orm"
import { getDb } from "../connection"
import { channels, channelMembers, msgMessages } from "../schema"
import { randomUUID } from "crypto"

function now() { return Date.now() }

// ── Channel CRUD ──────────────────────────────────────────────────────────

export async function createChannel(data: {
  type: "public" | "private" | "dm" | "group_dm"
  name?: string
  description?: string
  createdBy: number
  memberIds?: number[]
}) {
  const db = getDb()
  const id = randomUUID()
  const ts = now()

  const [channel] = await db.insert(channels).values({
    id,
    type: data.type,
    name: data.name ?? null,
    description: data.description ?? null,
    topic: null,
    createdBy: data.createdBy,
    createdAt: ts,
    updatedAt: ts,
    archivedAt: null,
  }).returning()

  // Add creator as owner
  await db.insert(channelMembers).values({
    channelId: id,
    userId: data.createdBy,
    role: "owner",
    lastReadAt: ts,
    muted: false,
    joinedAt: ts,
  })

  // Add additional members
  if (data.memberIds && data.memberIds.length > 0) {
    const otherMembers = data.memberIds.filter((uid) => uid !== data.createdBy)
    if (otherMembers.length > 0) {
      await db.insert(channelMembers).values(
        otherMembers.map((userId) => ({
          channelId: id,
          userId,
          role: "member" as const,
          lastReadAt: ts,
          muted: false,
          joinedAt: ts,
        }))
      )
    }
  }

  return channel!
}

export async function getChannel(id: string) {
  return getDb().query.channels.findFirst({
    where: (c, { eq }) => eq(c.id, id),
    with: { channelMembers: true },
  })
}

export async function getChannelsByUser(userId: number) {
  const db = getDb()
  const memberships = await db.query.channelMembers.findMany({
    where: (cm, { eq }) => eq(cm.userId, userId),
  })
  if (memberships.length === 0) return []

  const channelIds = memberships.map((m) => m.channelId)
  return db.query.channels.findMany({
    where: (c, { inArray }) => inArray(c.id, channelIds),
    with: { channelMembers: true },
    orderBy: (c, { desc }) => [desc(c.updatedAt)],
  })
}

export async function updateChannel(
  id: string,
  data: { name?: string; description?: string; topic?: string }
) {
  const [updated] = await getDb()
    .update(channels)
    .set({ ...data, updatedAt: now() })
    .where(eq(channels.id, id))
    .returning()
  return updated
}

export async function archiveChannel(id: string) {
  const [updated] = await getDb()
    .update(channels)
    .set({ archivedAt: now(), updatedAt: now() })
    .where(eq(channels.id, id))
    .returning()
  return updated
}

// ── Members ───────────────────────────────────────────────────────────────

export async function addChannelMember(channelId: string, userId: number, role: "admin" | "member" = "member") {
  const ts = now()
  await getDb().insert(channelMembers).values({
    channelId,
    userId,
    role,
    lastReadAt: ts,
    muted: false,
    joinedAt: ts,
  }).onConflictDoNothing()
}

export async function removeChannelMember(channelId: string, userId: number) {
  await getDb()
    .delete(channelMembers)
    .where(and(eq(channelMembers.channelId, channelId), eq(channelMembers.userId, userId)))
}

export async function getChannelMembers(channelId: string) {
  return getDb().query.channelMembers.findMany({
    where: (cm, { eq }) => eq(cm.channelId, channelId),
    with: { user: true },
  })
}

export async function updateLastRead(channelId: string, userId: number) {
  await getDb()
    .update(channelMembers)
    .set({ lastReadAt: now() })
    .where(and(eq(channelMembers.channelId, channelId), eq(channelMembers.userId, userId)))
}

// ── Unread counts ─────────────────────────────────────────────────────────

export async function getUnreadCounts(userId: number): Promise<Record<string, number>> {
  const db = getDb()
  const memberships = await db.query.channelMembers.findMany({
    where: (cm, { eq }) => eq(cm.userId, userId),
  })

  const counts: Record<string, number> = {}
  for (const m of memberships) {
    const result = await db
      .select({ count: sql<number>`count(*)` })
      .from(msgMessages)
      .where(
        and(
          eq(msgMessages.channelId, m.channelId),
          sql`${msgMessages.createdAt} > ${m.lastReadAt}`,
          sql`${msgMessages.deletedAt} IS NULL`,
          sql`${msgMessages.userId} != ${userId}`,
        )
      )
    counts[m.channelId] = result[0]?.count ?? 0
  }

  return counts
}
