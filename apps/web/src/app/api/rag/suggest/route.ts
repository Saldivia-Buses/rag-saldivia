/**
 * POST /api/rag/suggest
 *
 * Genera 3-4 preguntas de follow-up basadas en el último intercambio Q&A.
 * En modo MOCK_RAG retorna sugerencias hardcodeadas.
 * En modo real: envía un prompt al RAG server para generar sugerencias.
 */

import { NextResponse } from "next/server"
import { requireAuth, apiOk, apiError, apiServerError } from "@/lib/api-utils"

const MOCK_RAG = process.env["MOCK_RAG"] === "true"

const MOCK_SUGGESTIONS = [
  "¿Podés ampliar más sobre este tema?",
  "¿Cuáles son los casos de uso más comunes?",
  "¿Hay alguna excepción o caso borde importante?",
  "¿Cómo se compara esto con alternativas?",
]

export async function POST(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  try {
    const body = await request.json().catch(() => null)
    if (!body?.query) return apiError("El campo 'query' es requerido")

    if (MOCK_RAG) return apiOk({ questions: MOCK_SUGGESTIONS })

    // En producción: usar el RAG server para generar sugerencias contextuales
    const ragUrl = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
    const lang = (body.language as string | undefined) ?? "es"
    const prompt = `Based on this question and answer, suggest 3 concise follow-up questions in ${lang === "es" ? "Spanish" : "the same language"}. Return ONLY a JSON array of strings, no other text.
Question: ${body.query}
Answer: ${String(body.lastResponse ?? "").slice(0, 500)}`

    try {
      const res = await fetch(`${ragUrl}/v1/chat/completions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: [{ role: "user", content: prompt }],
          use_knowledge_base: false,
          max_tokens: 200,
        }),
        signal: AbortSignal.timeout(10000),
      })

      if (res.ok) {
        const data = await res.json() as { choices?: Array<{ message?: { content?: string } }> }
        const content = data.choices?.[0]?.message?.content ?? ""
        const match = content.match(/\[[\s\S]*\]/)
        if (match) {
          const questions = JSON.parse(match[0]) as string[]
          return apiOk({ questions: questions.slice(0, 4) })
        }
      }
    } catch {
      // Si falla el RAG, caer al mock
    }

    return apiOk({ questions: MOCK_SUGGESTIONS })
  } catch (err) {
    return apiServerError(err, "POST /api/rag/suggest")
  }
}
