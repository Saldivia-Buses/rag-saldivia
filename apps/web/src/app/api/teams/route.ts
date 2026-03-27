/**
 * POST /api/teams
 * Handler de mensajes de Microsoft Teams via Bot Framework.
 * F3.49 — Bot Slack/Teams.
 */

import { NextResponse } from "next/server"
import { getDb, botUserMappings } from "@rag-saldivia/db"
import { eq, and } from "drizzle-orm"
import { collectSseText } from "@/lib/rag/stream"

const SYSTEM_API_KEY = process.env["SYSTEM_API_KEY"] ?? ""
const BASE_URL = process.env["NEXTAUTH_URL"] ?? "http://localhost:3000"

export async function POST(request: Request) {
  try {
    const body = await request.json() as {
      type?: string
      text?: string
      from?: { id?: string }
      serviceUrl?: string
      conversation?: { id?: string }
      id?: string
    }

    if (body.type !== "message" || !body.text) {
      return NextResponse.json({ type: "message", text: "Usa el bot para hacer consultas." })
    }

    const teamsUserId = body.from?.id ?? ""
    const text = body.text.trim()

    const db = getDb()
    const mapping = await db
      .select({ systemUserId: botUserMappings.systemUserId })
      .from(botUserMappings)
      .where(and(eq(botUserMappings.platform, "teams"), eq(botUserMappings.externalUserId, teamsUserId)))
      .limit(1)
      .then((r) => r[0] ?? null)

    if (!mapping) {
      return NextResponse.json({ type: "message", text: `Usuario ${teamsUserId} no vinculado. Contactá al administrador.` })
    }

    // Consultar RAG
    let answer = "Sin respuesta del sistema."
    try {
      const ragRes = await fetch(`${BASE_URL}/api/rag/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Api-Key": SYSTEM_API_KEY, "x-user-id": String(mapping.systemUserId), "x-user-role": "user" },
        body: JSON.stringify({ messages: [{ role: "user", content: text }], use_knowledge_base: true }),
        signal: AbortSignal.timeout(30000),
      })
      if (ragRes.ok) {
        const full = await collectSseText(ragRes)
        answer = full || answer
      }
    } catch { /* ignorar */ }

    return NextResponse.json({ type: "message", text: answer.slice(0, 4096) })
  } catch {
    return NextResponse.json({ type: "message", text: "Error interno." })
  }
}
