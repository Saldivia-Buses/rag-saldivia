import { eq, and, gt } from "drizzle-orm"
import { getDb } from "../connection"
import { sessionShares, chatSessions, chatMessages } from "../schema"
import { randomBytes } from "crypto"

const DEFAULT_TTL_DAYS = 7

export async function createShare(sessionId: string, userId: number, ttlDays = DEFAULT_TTL_DAYS) {
  const db = getDb()
  const token = randomBytes(32).toString("hex") // 64-char hex
  const now = Date.now()
  const expiresAt = now + ttlDays * 24 * 60 * 60 * 1000
  const id = randomBytes(16).toString("hex")

  const [row] = await db
    .insert(sessionShares)
    .values({ id, sessionId, userId, token, expiresAt, createdAt: now })
    .returning()
  return row!
}

export async function getShareByToken(token: string) {
  const db = getDb()
  const now = Date.now()
  const rows = await db
    .select()
    .from(sessionShares)
    .where(and(eq(sessionShares.token, token), gt(sessionShares.expiresAt, now)))
    .limit(1)
  return rows[0] ?? null
}

export async function getShareWithSession(token: string) {
  const share = await getShareByToken(token)
  if (!share) return null

  const db = getDb()
  const sessions = await db
    .select()
    .from(chatSessions)
    .where(eq(chatSessions.id, share.sessionId))
    .limit(1)
  const session = sessions[0]
  if (!session) return null

  const messages = await db
    .select()
    .from(chatMessages)
    .where(eq(chatMessages.sessionId, share.sessionId))

  return { share, session, messages }
}

export async function revokeShare(id: string, userId: number) {
  const db = getDb()
  await db
    .delete(sessionShares)
    .where(and(eq(sessionShares.id, id), eq(sessionShares.userId, userId)))
}

export async function listSharesByUser(userId: number) {
  const db = getDb()
  return db
    .select()
    .from(sessionShares)
    .where(eq(sessionShares.userId, userId))
}
