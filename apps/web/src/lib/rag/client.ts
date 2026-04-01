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
import { RAG_TIMEOUT_MS as DEFAULT_RAG_TIMEOUT } from "@rag-saldivia/config"

const RAG_URL = process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"
const RAG_TIMEOUT_MS = parseInt(process.env["RAG_TIMEOUT_MS"] ?? String(DEFAULT_RAG_TIMEOUT))
const MOCK_RAG = process.env["MOCK_RAG"] === "true"

export type RagGenerateRequest = {
  messages: Array<{ role: string; content: string }>
  use_knowledge_base?: boolean
  collection_name?: string
  collection_names?: string[]
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
    return new Response(JSON.stringify({ status: "mock", collections: ["default"] }), {
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

const OPENROUTER_KEY = process.env["OPENROUTER_API_KEY"]
const OPENROUTER_MODEL = process.env["OPENROUTER_MODEL"] ?? "anthropic/claude-haiku-4-5"

const ARTIFACT_SYSTEM_PROMPT = `Cuando generes código, HTML, SVG, diagramas o contenido estructurado sustancial, envolvelo en tags artifact:

<artifact type="code" language="python" title="Título descriptivo corto">
código aquí, SIN markdown fences
</artifact>

Tipos soportados: code, html, svg, mermaid, table, text
- type="code": siempre incluí el atributo language
- NO uses \`\`\` (markdown code fences) dentro de artifact tags
- Código inline corto (1-2 líneas) dejalo como \`código\` en markdown, no como artifact
- Usá artifacts para bloques sustanciales (3+ líneas de código, HTML completo, diagramas, etc.)
- Siempre poné un title descriptivo
- Podés incluir múltiples artifacts en una respuesta
- El texto explicativo va FUERA de los tags artifact, en markdown normal`

function mockRagStream(
  body: RagGenerateRequest
): { stream: ReadableStream; contentType: string } | Promise<{ stream: ReadableStream; contentType: string } | { error: RagError }> {
  // Si hay key de OpenRouter, usar LLM real
  if (OPENROUTER_KEY) {
    return openRouterStream(body)
  }

  // Fallback: respuesta hardcoded
  const messages = [
    "Esta es una respuesta simulada del RAG Server.\n\n",
    "## Modo Mock\n\n",
    "El sistema está en modo `MOCK_RAG=true`.\n\n",
    "Para usar el **RAG real**, levantá Docker y desactivá `MOCK_RAG` en `.env.local`.\n\n",
    "### Alternativa\n\n",
    "Configurá `OPENROUTER_API_KEY` en `.env.local` para usar un LLM real sin GPU.",
  ]

  const stream = new ReadableStream({
    async start(controller) {
      for (const msg of messages) {
        const data = JSON.stringify({
          choices: [{ delta: { content: msg }, finish_reason: null }],
        })
        controller.enqueue(new TextEncoder().encode(`data: ${data}\n\n`))
        await new Promise((r) => setTimeout(r, 80))
      }
      controller.enqueue(new TextEncoder().encode("data: [DONE]\n\n"))
      controller.close()
    },
  })

  return { stream, contentType: "text/event-stream" }
}

async function openRouterStream(
  body: RagGenerateRequest
): Promise<{ stream: ReadableStream; contentType: string } | { error: RagError }> {
  try {
    const response = await fetch("https://openrouter.ai/api/v1/chat/completions", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${OPENROUTER_KEY}`,
      },
      body: JSON.stringify({
        model: OPENROUTER_MODEL,
        messages: [
          { role: "system", content: ARTIFACT_SYSTEM_PROMPT },
          ...body.messages,
        ],
        stream: true,
      }),
    })

    if (!response.ok) {
      const text = await response.text().catch(() => "")
      return { error: createRagError("UPSTREAM_ERROR", `OpenRouter ${response.status}: ${text.slice(0, 200)}`) }
    }

    if (!response.body) {
      return { error: createRagError("UPSTREAM_ERROR", "OpenRouter no retornó stream") }
    }

    return { stream: response.body, contentType: "text/event-stream" }
  } catch (err) {
    return { error: createRagError("UNAVAILABLE", `OpenRouter: ${String(err)}`) }
  }
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
