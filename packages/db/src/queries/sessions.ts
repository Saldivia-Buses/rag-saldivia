/**
 * Queries de sesiones de chat y mensajes.
 */

import { eq, desc, and } from "drizzle-orm"
import { getDb } from "../connection.js"
import { chatSessions, chatMessages, messageFeedback } from "../schema.js"

const db = getDb()

function now() {
  return Date.now()
}

// ── Sessions ───────────────────────────────────────────────────────────────

export async function listSessionsByUser(userId: number) {
  return db.query.chatSessions.findMany({
    where: (s, { eq }) => eq(s.userId, userId),
    orderBy: (s, { desc }) => [desc(s.updatedAt)],
  })
}

export async function getSessionById(id: string, userId?: number) {
  return db.query.chatSessions.findFirst({
    where: (s, { and, eq }) =>
      userId ? and(eq(s.id, id), eq(s.userId, userId)) : eq(s.id, id),
    with: {
      messages: {
        orderBy: (m, { asc }) => [asc(m.timestamp)],
      },
    },
  })
}

export async function createSession(data: {
  userId: number
  collection: string
  crossdoc?: boolean
  title?: string
}) {
  const id = crypto.randomUUID()
  const ts = now()
  const [session] = await db
    .insert(chatSessions)
    .values({
      id,
      userId: data.userId,
      title: data.title ?? "Nueva sesión",
      collection: data.collection,
      crossdoc: data.crossdoc ?? false,
      createdAt: ts,
      updatedAt: ts,
    })
    .returning()
  return session
}

export async function updateSessionTitle(id: string, userId: number, title: string) {
  const [updated] = await db
    .update(chatSessions)
    .set({ title, updatedAt: now() })
    .where(and(eq(chatSessions.id, id), eq(chatSessions.userId, userId)))
    .returning()
  return updated
}

export async function deleteSession(id: string, userId: number) {
  await db
    .delete(chatSessions)
    .where(and(eq(chatSessions.id, id), eq(chatSessions.userId, userId)))
}

// ── Messages ───────────────────────────────────────────────────────────────

export async function addMessage(data: {
  sessionId: string
  role: "user" | "assistant" | "system"
  content: string
  sources?: unknown[]
}) {
  const [message] = await db
    .insert(chatMessages)
    .values({
      sessionId: data.sessionId,
      role: data.role,
      content: data.content,
      sources: data.sources ?? null,
      timestamp: now(),
    })
    .returning()

  // Actualizar updated_at de la sesión
  await db
    .update(chatSessions)
    .set({ updatedAt: now() })
    .where(eq(chatSessions.id, data.sessionId))

  return message
}

export async function addFeedback(
  messageId: number,
  userId: number,
  rating: "up" | "down"
) {
  await db
    .insert(messageFeedback)
    .values({ messageId, userId, rating, createdAt: now() })
    .onConflictDoUpdate({
      target: [messageFeedback.messageId, messageFeedback.userId],
      set: { rating },
    })
}
