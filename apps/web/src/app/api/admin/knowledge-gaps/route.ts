/**
 * GET /api/admin/knowledge-gaps
 *
 * Detecta queries donde el RAG respondió con baja confianza.
 * Heurística: mensajes del asistente cortos (< 80 palabras) con keywords de incertidumbre.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { getDb, chatMessages, chatSessions } from "@rag-saldivia/db"
import { eq, sql } from "drizzle-orm"

const UNCERTAINTY_PATTERNS = [
  "no encuentro",
  "no tengo información",
  "no sé",
  "no encontré",
  "no puedo encontrar",
  "no hay información",
  "no tengo datos",
  "i don't know",
  "i couldn't find",
  "no information",
  "not found",
  "unable to find",
]

function isLowConfidence(content: string): boolean {
  const wordCount = content.trim().split(/\s+/).length
  if (wordCount >= 80) return false
  const lower = content.toLowerCase()
  return UNCERTAINTY_PATTERNS.some((p) => lower.includes(p))
}

type Gap = {
  messageId: number
  content: string
  sessionId: string
  sessionTitle: string
  collection: string
  timestamp: number
}

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return NextResponse.json({ ok: false, error: "Solo admins" }, { status: 403 })
  }

  const db = getDb()

  // Obtener los últimos 500 mensajes de asistentes con sus sesiones
  const messages = await db
    .select({
      id: chatMessages.id,
      content: chatMessages.content,
      sessionId: chatMessages.sessionId,
      timestamp: chatMessages.timestamp,
      title: chatSessions.title,
      collection: chatSessions.collection,
    })
    .from(chatMessages)
    .innerJoin(chatSessions, eq(chatMessages.sessionId, chatSessions.id))
    .where(eq(chatMessages.role, "assistant"))
    .orderBy(sql`${chatMessages.timestamp} DESC`)
    .limit(500)

  const gaps: Gap[] = messages
    .filter((m) => isLowConfidence(m.content))
    .slice(0, 100)
    .map((m) => ({
      messageId: m.id,
      content: m.content,
      sessionId: m.sessionId,
      sessionTitle: m.title,
      collection: m.collection,
      timestamp: m.timestamp,
    }))

  return NextResponse.json({ ok: true, data: gaps, total: gaps.length })
}
