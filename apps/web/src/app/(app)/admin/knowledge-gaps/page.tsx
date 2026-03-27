import { requireAdmin } from "@/lib/auth/current-user"
import { getDb, chatMessages, chatSessions } from "@rag-saldivia/db"
import { eq, sql } from "drizzle-orm"
import { KnowledgeGapsClient, type KnowledgeGap } from "@/components/admin/KnowledgeGapsClient"

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

async function getKnowledgeGaps(): Promise<KnowledgeGap[]> {
  const db = getDb()
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

  return messages
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
}

export default async function KnowledgeGapsPage() {
  await requireAdmin()
  const gaps = await getKnowledgeGaps()

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Brechas de conocimiento</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Queries donde el RAG respondió con baja confianza. Guía para qué ingestar.
        </p>
      </div>
      <KnowledgeGapsClient gaps={gaps} />
    </div>
  )
}
