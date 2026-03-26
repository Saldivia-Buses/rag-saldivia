import { getDb } from "../connection"
import { sql, eq, and } from "drizzle-orm"
import { chatSessions, chatMessages, savedResponses, promptTemplates } from "../schema"

export type SearchResult = {
  type: "session" | "message" | "saved" | "template"
  id: string
  title: string
  snippet: string
  sessionId?: string
  href: string
}

/**
 * Búsqueda universal con FTS5 (F3.39).
 * Si FTS5 no está disponible, cae a LIKE sobre las tablas principales.
 */
export async function universalSearch(
  query: string,
  userId: number,
  limit = 20
): Promise<SearchResult[]> {
  if (!query || query.trim().length < 2) return []

  const db = getDb()
  const results: SearchResult[] = []
  const q = query.trim()

  try {
    // Intentar FTS5
    const sessionRows = await db.run(
      sql`SELECT session_id, title, snippet(sessions_fts, 2, '<b>', '</b>', '...', 10) as snip
          FROM sessions_fts
          WHERE sessions_fts MATCH ${q} AND user_id = ${userId}
          LIMIT ${Math.floor(limit / 2)}`
    )
    for (const row of sessionRows.rows) {
      results.push({
        type: "session",
        id: String(row[0]),
        title: String(row[1]),
        snippet: String(row[2] ?? ""),
        href: `/chat/${row[0]}`,
      })
    }

    const msgRows = await db.run(
      sql`SELECT m.message_id, m.session_id, s.title, snippet(messages_fts, 2, '<b>', '</b>', '...', 10) as snip
          FROM messages_fts m
          JOIN chat_sessions s ON s.id = m.session_id
          WHERE messages_fts MATCH ${q} AND s.user_id = ${userId}
          LIMIT ${Math.floor(limit / 3)}`
    )
    for (const row of msgRows.rows) {
      results.push({
        type: "message",
        id: String(row[0]),
        title: String(row[2]),
        snippet: String(row[3] ?? ""),
        sessionId: String(row[1]),
        href: `/chat/${row[1]}`,
      })
    }
  } catch {
    // FTS5 no disponible — usar LIKE como fallback
    const likeQ = `%${q}%`

    const sessions = await db
      .select({ id: chatSessions.id, title: chatSessions.title })
      .from(chatSessions)
      .where(and(eq(chatSessions.userId, userId), sql`${chatSessions.title} LIKE ${likeQ}`))
      .limit(Math.floor(limit / 2))

    for (const s of sessions) {
      results.push({
        type: "session",
        id: s.id,
        title: s.title,
        snippet: "",
        href: `/chat/${s.id}`,
      })
    }
  }

  // Buscar también en templates (no requieren FTS)
  try {
    const templates = await db
      .select({ id: promptTemplates.id, title: promptTemplates.title, prompt: promptTemplates.prompt })
      .from(promptTemplates)
      .where(and(eq(promptTemplates.active, true), sql`(${promptTemplates.title} LIKE ${"%" + q + "%"} OR ${promptTemplates.prompt} LIKE ${"%" + q + "%"})`))
      .limit(5)

    for (const t of templates) {
      results.push({
        type: "template",
        id: String(t.id),
        title: t.title,
        snippet: t.prompt.slice(0, 80),
        href: "/chat",
      })
    }
  } catch { /* ignorar */ }

  // Buscar en guardados del usuario
  try {
    const saved = await db
      .select({ id: savedResponses.id, content: savedResponses.content, sessionTitle: savedResponses.sessionTitle })
      .from(savedResponses)
      .where(and(eq(savedResponses.userId, userId), sql`${savedResponses.content} LIKE ${"%" + q + "%"}`))
      .limit(5)

    for (const s of saved) {
      results.push({
        type: "saved",
        id: String(s.id),
        title: s.sessionTitle ?? "Respuesta guardada",
        snippet: s.content.slice(0, 100),
        href: "/saved",
      })
    }
  } catch { /* ignorar */ }

  return results.slice(0, limit)
}
