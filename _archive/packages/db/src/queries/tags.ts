import { eq, and, inArray } from "drizzle-orm"
import { getDb } from "../connection"
import { sessionTags, chatSessions } from "../schema"

export async function addTag(sessionId: string, tag: string) {
  const db = getDb()
  await db
    .insert(sessionTags)
    .values({ sessionId, tag: tag.toLowerCase().trim() })
    .onConflictDoNothing()
}

export async function removeTag(sessionId: string, tag: string) {
  const db = getDb()
  await db
    .delete(sessionTags)
    .where(and(eq(sessionTags.sessionId, sessionId), eq(sessionTags.tag, tag.toLowerCase().trim())))
}

export async function listTagsBySession(sessionId: string) {
  const db = getDb()
  const rows = await db
    .select({ tag: sessionTags.tag })
    .from(sessionTags)
    .where(eq(sessionTags.sessionId, sessionId))
  return rows.map((r) => r.tag)
}

export async function listTagsByUser(userId: number): Promise<string[]> {
  const db = getDb()
  // Obtener tags de sesiones del usuario
  const sessions = await db
    .select({ id: chatSessions.id })
    .from(chatSessions)
    .where(eq(chatSessions.userId, userId))
  if (sessions.length === 0) return []
  const sessionIds = sessions.map((s) => s.id)
  const rows = await db
    .selectDistinct({ tag: sessionTags.tag })
    .from(sessionTags)
    .where(inArray(sessionTags.sessionId, sessionIds))
  return rows.map((r) => r.tag)
}
