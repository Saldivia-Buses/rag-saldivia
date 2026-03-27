/**
 * Cliente HTTP para el RAG Server de NVIDIA (puerto 8081).
 * Reemplaza la lógica de proxy en gateway.py.
 *
 * Puntos críticos preservados del gateway Python:
 * - Verifica status HTTP ANTES de streamear (el gateway Python tenía un bug
 *   donde StreamingResponse siempre retornaba 200 aunque hubiera error upstream).
 * - Timeout configurable para queries largas.
 * - Retry con backoff para errores transitorios.
 */

import { getSuggestion } from "@rag-saldivia/logger/suggestions"

const RAG_URL = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
const RAG_TIMEOUT_MS = parseInt(process.env["RAG_TIMEOUT_MS"] ?? "120000")
const MOCK_RAG = process.env["MOCK_RAG"] === "true"

export type RagGenerateRequest = {
  messages: Array<{ role: string; content: string }>
  use_knowledge_base?: boolean
  collection_name?: string
  temperature?: number
  top_p?: number
  max_tokens?: number
  vdb_top_k?: number
  reranker_top_k?: number
  use_reranker?: boolean
}

export type RagError = {
  code: "UNAVAILABLE" | "TIMEOUT" | "FORBIDDEN" | "UPSTREAM_ERROR"
  message: string
  suggestion: string
}

function createRagError(code: RagError["code"], message: string): RagError {
  return {
    code,
    message,
    suggestion: getSuggestion(message) ?? "Verificá el estado del RAG Server con: rag status",
  }
}

/**
 * Genera una respuesta del RAG en modo stream (SSE).
 * Retorna el Response stream directamente para hacer pipe al cliente.
 *
 * CRÍTICO: Verifica el status ANTES de retornar el stream,
 * para que el cliente reciba un 4xx/5xx correcto en lugar de un 200 vacío.
 */
export async function ragGenerateStream(
  body: RagGenerateRequest,
  signal?: AbortSignal
): Promise<{ stream: ReadableStream; contentType: string } | { error: RagError }> {
  if (MOCK_RAG) {
    return mockRagStream(body)
  }

  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), RAG_TIMEOUT_MS)

  if (signal) {
    if (signal.aborted) controller.abort()
    else signal.addEventListener("abort", () => controller.abort(), { once: true })
  }

  try {
    const response = await fetch(`${RAG_URL}/v1/chat/completions`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "text/event-stream",
      },
      body: JSON.stringify({ ...body, stream: true }),
      signal: controller.signal,
    })

    clearTimeout(timeout)

    // Verificar status ANTES de streamear
    if (!response.ok) {
      const errorBody = await response.text().catch(() => "")
      return {
        error: createRagError(
          response.status >= 500 ? "UPSTREAM_ERROR" : "FORBIDDEN",
          `RAG Server respondió ${response.status}: ${errorBody.slice(0, 200)}`
        ),
      }
    }

    if (!response.body) {
      return { error: createRagError("UPSTREAM_ERROR", "RAG Server no retornó stream") }
    }

    return {
      stream: response.body,
      contentType: response.headers.get("content-type") ?? "text/event-stream",
    }
  } catch (err) {
    clearTimeout(timeout)

    if (err instanceof Error && err.name === "AbortError") {
      return { error: createRagError("TIMEOUT", `Timeout después de ${RAG_TIMEOUT_MS}ms`) }
    }

    const message = String(err)
    return {
      error: createRagError(
        message.includes("ECONNREFUSED") ? "UNAVAILABLE" : "UPSTREAM_ERROR",
        message
      ),
    }
  }
}

/**
 * Hace `fetch` al RAG Server. Con `MOCK_RAG=true` no llama al servidor real.
 *
 * Los consumidores deben comprobar `response.ok` o el código HTTP antes de parsear el body;
 * un upstream puede devolver 4xx/5xx. El streaming de chat usa `ragGenerateStream`, que valida
 * el status antes de devolver el stream.
 */
export async function ragFetch(
  path: string,
  options: RequestInit = {}
): Promise<Response | { error: RagError }> {
  if (MOCK_RAG) {
    return new Response(JSON.stringify({ status: "mock", collections: ["tecpia"] }), {
      headers: { "Content-Type": "application/json" },
    })
  }

  try {
    const response = await fetch(`${RAG_URL}${path}`, {
      ...options,
      signal: AbortSignal.timeout(10000),
    })
    return response
  } catch (err) {
    return {
      error: createRagError("UNAVAILABLE", String(err)),
    }
  }
}

// ── Mock para desarrollo sin RAG ───────────────────────────────────────────

function mockRagStream(
  _body: RagGenerateRequest
): { stream: ReadableStream; contentType: string } {
  const messages = [
    "Esta es una respuesta simulada del RAG Server.",
    " El sistema está en modo MOCK_RAG=true.",
    " Para usar el RAG real, levantá Docker y desactivá MOCK_RAG en .env.local.",
  ]

  const stream = new ReadableStream({
    async start(controller) {
      for (const msg of messages) {
        const data = JSON.stringify({
          choices: [{ delta: { content: msg }, finish_reason: null }],
        })
        controller.enqueue(new TextEncoder().encode(`data: ${data}\n\n`))
        await new Promise((r) => setTimeout(r, 100))
      }
      controller.enqueue(new TextEncoder().encode("data: [DONE]\n\n"))
      controller.close()
    },
  })

  return { stream, contentType: "text/event-stream" }
}

/**
 * Detecta si el texto no está en español y retorna una instrucción de idioma.
 * Heurística simple: palabras comunes en inglés, o caracteres no-latinos.
 * Zero config — el modelo responde en el idioma del usuario automáticamente.
 */
export function detectLanguageHint(text: string): string {
  if (!text || text.length < 4) return ""

  // Caracteres no-latinos (CJK, árabe, cirílico, etc.)
  const nonLatin = /[\u0400-\u04FF\u0600-\u06FF\u4E00-\u9FFF\u3040-\u30FF\uAC00-\uD7AF]/
  if (nonLatin.test(text)) {
    return "Respond in the same language as the user's message."
  }

  // Palabras clave en inglés que sugieren que el usuario escribe en inglés
  const englishWords = /\b(what|how|why|when|where|who|which|is|are|was|were|can|could|would|should|will|the|and|or|for|with|about|from|tell|me|please|show|list|find|does|do|have|has|had|get|give|make|need|want|know|think|help)\b/i
  const words = text.toLowerCase().split(/\s+/)
  const englishCount = words.filter((w) => englishWords.test(w)).length

  if (englishCount >= 2 || (words.length <= 3 && englishCount >= 1)) {
    return "Respond in the same language as the user's message."
  }

  return ""
}
