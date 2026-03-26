/**
 * POST /api/slack
 * Handler de slash commands y eventos de Slack.
 * F3.49 — Bot Slack/Teams.
 *
 * Configurar en Slack App:
 * - Slash command: /rag [query]
 * - Request URL: https://tu-dominio.com/api/slack
 * - Signing secret: SLACK_SIGNING_SECRET env var
 */

import { NextResponse } from "next/server"
import { getDb, botUserMappings, users } from "@rag-saldivia/db"
import { eq, and } from "drizzle-orm"
import { createHmac, timingSafeEqual } from "crypto"

const SLACK_BOT_TOKEN = process.env["SLACK_BOT_TOKEN"] ?? ""
const SLACK_SIGNING_SECRET = process.env["SLACK_SIGNING_SECRET"] ?? ""
const SYSTEM_API_KEY = process.env["SYSTEM_API_KEY"] ?? ""
const BASE_URL = process.env["NEXTAUTH_URL"] ?? process.env["NEXT_PUBLIC_BASE_URL"] ?? "http://localhost:3000"

async function verifySlackSignature(request: Request, rawBody: string): Promise<boolean> {
  if (!SLACK_SIGNING_SECRET) return false
  const timestamp = request.headers.get("x-slack-request-timestamp") ?? ""
  const signature = request.headers.get("x-slack-signature") ?? ""
  const sigBaseString = `v0:${timestamp}:${rawBody}`
  const hmac = createHmac("sha256", SLACK_SIGNING_SECRET).update(sigBaseString).digest("hex")
  const expected = `v0=${hmac}`
  try {
    return timingSafeEqual(Buffer.from(signature), Buffer.from(expected))
  } catch {
    return false
  }
}

export async function POST(request: Request) {
  const rawBody = await request.text()

  if (SLACK_SIGNING_SECRET && !(await verifySlackSignature(request, rawBody))) {
    return NextResponse.json({ error: "Invalid signature" }, { status: 401 })
  }

  let body: Record<string, string>
  try {
    body = Object.fromEntries(new URLSearchParams(rawBody))
  } catch {
    return NextResponse.json({ error: "Invalid body" }, { status: 400 })
  }

  const slackUserId = body["user_id"] ?? ""
  const text = (body["text"] ?? "").trim()

  if (!text) {
    return NextResponse.json({ response_type: "ephemeral", text: "Uso: /rag [tu pregunta]" })
  }

  // Respuesta inmediata a Slack (evitar timeout de 3s)
  const responseUrl = body["response_url"]

  // Resolver el userId del sistema
  const db = getDb()
  const mapping = await db
    .select({ systemUserId: botUserMappings.systemUserId })
    .from(botUserMappings)
    .where(and(eq(botUserMappings.platform, "slack"), eq(botUserMappings.externalUserId, slackUserId)))
    .limit(1)
    .then((r) => r[0] ?? null)

  if (!mapping) {
    return NextResponse.json({
      response_type: "ephemeral",
      text: `Tu usuario de Slack (${slackUserId}) no está vinculado a ninguna cuenta del sistema. Pedile a un admin que te configure en /admin/integrations.`,
    })
  }

  // Hacer la consulta al RAG en background
  if (responseUrl) {
    ; (async () => {
      try {
        const ragRes = await fetch(`${BASE_URL}/api/rag/generate`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Api-Key": SYSTEM_API_KEY,
            "x-user-id": String(mapping.systemUserId),
            "x-user-role": "user",
          },
          body: JSON.stringify({
            messages: [{ role: "user", content: text }],
            use_knowledge_base: true,
          }),
        })

        let answer = "Sin respuesta del sistema RAG."
        if (ragRes.ok) {
          const reader = ragRes.body?.getReader()
          const decoder = new TextDecoder()
          let fullContent = ""
          if (reader) {
            while (true) {
              const { done, value } = await reader.read()
              if (done) break
              const chunk = decoder.decode(value, { stream: true })
              for (const line of chunk.split("\n")) {
                if (!line.startsWith("data: ")) continue
                const data = line.slice(6).trim()
                if (data === "[DONE]") continue
                try {
                  const parsed = JSON.parse(data) as { choices?: Array<{ delta?: { content?: string } }> }
                  fullContent += parsed.choices?.[0]?.delta?.content ?? ""
                } catch { /* ignorar */ }
              }
            }
            answer = fullContent || answer
          }
        }

        await fetch(responseUrl, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            response_type: "in_channel",
            text: `*Pregunta:* ${text}\n\n*Respuesta:* ${answer.slice(0, 3000)}`,
          }),
        })
      } catch {
        await fetch(responseUrl, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ response_type: "ephemeral", text: "Error al consultar el sistema RAG." }),
        }).catch(() => {})
      }
    })()
  }

  return NextResponse.json({
    response_type: "ephemeral",
    text: "⏳ Procesando tu consulta...",
  })
}
