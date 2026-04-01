/**
 * POST /api/rag/generate — Main chat streaming endpoint.
 *
 * Pipeline (executed in order):
 *   1. Auth check (JWT)
 *   2. Rate limiting (queries/hour per user)
 *   3. Collection access verification (multi-collection support)
 *   4. System prompt injection: focus mode → project context → memory → language hint
 *   5. Proxy to RAG server (or OpenRouter in mock mode)
 *   6. Transform SSE stream to AI SDK Data Stream protocol
 *
 * Critical: verifies HTTP status BEFORE streaming (prevents the gateway.py
 * bug where StreamingResponse always returned 200 even on upstream errors).
 *
 * Data flow: Client (useChat) → this route → ragGenerateStream → createRagStreamResponse → SSE
 * Depends on: lib/rag/client.ts, lib/rag/ai-stream.ts, lib/api-utils.ts
 */

import { NextResponse } from "next/server"
import { z } from "zod"
import { ragGenerateStream } from "@/lib/rag/client"
import { createRagStreamResponse } from "@/lib/rag/ai-stream"
import { canAccessCollection, getUserCollections } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"
import { FOCUS_MODES, type FocusModeId, CollectionNameSchema } from "@rag-saldivia/shared"
import { detectLanguageHint } from "@/lib/rag/client"
import { getRateLimit, countQueriesLastHour, getProjectBySession, getMemoryAsContext } from "@rag-saldivia/db"
import { dispatchEvent } from "@/lib/webhook"
import { requireAuth, apiError, apiServerError } from "@/lib/api-utils"

export const runtime = "nodejs" // SSE requiere Node runtime, no Edge

const MessageSchema = z.object({
  role: z.string(),
  content: z.string().optional(),
  parts: z.array(z.object({ type: z.string(), text: z.string().optional() })).optional(),
})

const GenerateBodySchema = z.object({
  messages: z.array(MessageSchema).min(1, "El campo 'messages' es requerido y no puede estar vacío"),
  collection_name: CollectionNameSchema.optional(),
  collection_names: z.array(CollectionNameSchema).optional(),
  session_id: z.string().optional(),
  use_knowledge_base: z.boolean().optional(),
  focus_mode: z.string().optional(),
  crossdoc: z.boolean().optional(),
})

type RawMessage = z.infer<typeof MessageSchema>

/** Normaliza mensajes del AI SDK (parts) o legacy (content) a formato OpenAI. */
function normalizeMessages(raw: RawMessage[]): Array<{ role: string; content: string }> {
  return raw.map((m) => {
    if (m.parts) {
      const text = m.parts
        .filter((p) => p.type === "text" && p.text)
        .map((p) => p.text!)
        .join("")
      return { role: m.role, content: text }
    }
    return { role: m.role, content: m.content ?? "" }
  })
}

export async function POST(request: Request) {
  const start = Date.now()

  // Auth check
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const userId = Number(claims.sub)

  try {
    const raw = await request.json().catch(() => null)
    const parsed = GenerateBodySchema.safeParse(raw)
    if (!parsed.success) {
      return apiError(parsed.error.issues[0]?.message ?? "Body inválido")
    }
    // Mutable copy — Zod validated, messages normalized for system prompt injection
    const body = {
      ...parsed.data,
      messages: normalizeMessages(parsed.data.messages),
    }

    // Rate limiting — reject if user exceeded queries/hour quota
    const maxQph = await getRateLimit(userId)
    if (maxQph !== null) {
      const count = await countQueriesLastHour(userId)
      if (count >= maxQph) {
        return apiError(`Límite de ${maxQph} queries/hora alcanzado.`, 429, {
          code: "RATE_LIMITED",
          retryAfterMs: 3600_000, // conservative: up to 60 min
          maxCount: maxQph,
        })
      }
    }

    const collectionName = body.collection_name as string | undefined
    const collectionNames = body.collection_names as string[] | undefined

    // Multi-colección: verificar acceso con una sola query + Set local (evita N queries)
    if (collectionNames && collectionNames.length > 0) {
      const userCollections = await getUserCollections(userId)
      const accessSet = new Set(
        userCollections
          .filter((c) => ["read", "write", "admin"].includes(c.permission))
          .map((c) => c.name)
      )
      for (const col of collectionNames) {
        if (!accessSet.has(col)) {
          return apiError("Sin acceso a la colección solicitada", 403)
        }
      }
      // El Blueprint acepta colecciones múltiples como array
      body.collection_names = collectionNames
    }

    // Prepend system message para el modo de foco seleccionado
    const focusModeId = body.focus_mode as FocusModeId | undefined
    const focusMode = FOCUS_MODES.find((m) => m.id === focusModeId)
    if (focusMode) {
      body.messages = [
        { role: "system", content: focusMode.systemPrompt },
        ...body.messages,
      ]
    }

    // Verificar acceso a la colección si se especificó
    if (collectionName) {
      const hasAccess = await canAccessCollection(userId, collectionName, "read")
      if (!hasAccess) {
        log.warn("rag.error", {
          reason: "forbidden",
          collection: collectionName,
        }, { userId })
        return apiError("Sin acceso a la colección solicitada", 403)
      }
    }

    log.info("rag.stream_started", {
      collection: collectionName,
      crossdoc: body.crossdoc ?? false,
    }, { userId, sessionId: body.session_id ?? null })

    // Inyectar memoria del usuario si existe — F3.44
    try {
      const memoryContext = await getMemoryAsContext(userId)
      if (memoryContext) {
        body.messages = [{ role: "system", content: memoryContext }, ...body.messages]
      }
    } catch { /* no bloquear */ }

    // Inyectar instrucciones del proyecto si la sesión pertenece a uno — F3.41
    const sessionId = body.session_id as string | undefined
    if (sessionId) {
      try {
        const project = await getProjectBySession(sessionId)
        if (project?.instructions) {
          body.messages = [
            { role: "system", content: `Project context: ${project.instructions}` },
            ...body.messages,
          ]
        }
      } catch { /* no bloquear si falla */ }
    }

    // Inyectar instrucción de idioma si el query no está en español
    const lastUserMessage = [...body.messages].reverse().find((m: { role: string }) => m.role === "user")
    const langHint = lastUserMessage ? detectLanguageHint(lastUserMessage.content as string) : ""
    if (langHint) {
      body.messages = [{ role: "system", content: langHint }, ...body.messages]
    }

    const result = await ragGenerateStream(body as import("@/lib/rag/client").RagGenerateRequest, request.signal)

    if ("error" in result) {
      log.error("rag.error", {
        code: result.error.code,
        message: result.error.message,
        collection: collectionName,
        duration: Date.now() - start,
      }, { userId })

      const status = result.error.code === "TIMEOUT" ? 504
        : result.error.code === "UNAVAILABLE" ? 503 : 502
      return apiError(result.error.message, status, {
        code: result.error.code,
        suggestion: result.error.suggestion,
      })
    }

    log.info("rag.stream_completed", {
      collection: collectionName,
      duration: Date.now() - start,
    }, { userId })

    // Dispatch webhook de baja confianza si no hay fuentes — F2.38
    // (la detección real ocurre en el cliente; aquí delegamos al hook de post-stream)
    dispatchEvent("query.completed", { userId, collection: collectionName }).catch(() => {})

    return createRagStreamResponse(result.stream)
  } catch (error) {
    return apiServerError(error, "POST /api/rag/generate", userId)
  }
}
