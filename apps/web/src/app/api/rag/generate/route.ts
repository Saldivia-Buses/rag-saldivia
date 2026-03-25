/**
 * POST /api/rag/generate
 *
 * Proxy SSE al RAG Server en :8081.
 * Verifica permisos de colección antes de streamear.
 *
 * Preserva el fix crítico del gateway Python: verifica el status HTTP
 * ANTES de retornar el stream (evita el bug donde siempre se retornaba 200).
 */

import { NextResponse } from "next/server"
import { ragGenerateStream } from "@/lib/rag/client"
import { extractClaims } from "@/lib/auth/jwt"
import { canAccessCollection } from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"
import { FOCUS_MODES, type FocusModeId } from "@rag-saldivia/shared"

export const runtime = "nodejs" // SSE requiere Node runtime, no Edge

export async function POST(request: Request) {
  const start = Date.now()

  // Autenticación
  const claims = await extractClaims(request)
  if (!claims) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  const userId = Number(claims.sub)

  try {
    const body = await request.json().catch(() => null)
    if (!body || !Array.isArray(body.messages) || body.messages.length === 0) {
      return NextResponse.json(
        { ok: false, error: "El campo 'messages' es requerido y no puede estar vacío" },
        { status: 400 }
      )
    }

    const collectionName = body.collection_name as string | undefined

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
        return NextResponse.json(
          { ok: false, error: `Sin acceso a la colección '${collectionName}'` },
          { status: 403 }
        )
      }
    }

    log.info("rag.stream_started", {
      collection: collectionName,
      crossdoc: body.crossdoc ?? false,
    }, { userId, sessionId: body.session_id })

    const result = await ragGenerateStream(body, request.signal)

    if ("error" in result) {
      log.error("rag.error", {
        code: result.error.code,
        message: result.error.message,
        collection: collectionName,
        duration: Date.now() - start,
      }, { userId })

      return NextResponse.json(
        {
          ok: false,
          error: result.error.message,
          suggestion: result.error.suggestion,
        },
        {
          status: result.error.code === "TIMEOUT" ? 504
            : result.error.code === "UNAVAILABLE" ? 503
            : 502,
        }
      )
    }

    log.info("rag.stream_completed", {
      collection: collectionName,
      duration: Date.now() - start,
    }, { userId })

    return new Response(result.stream, {
      headers: {
        "Content-Type": result.contentType,
        "Cache-Control": "no-cache",
        Connection: "keep-alive",
        "X-Accel-Buffering": "no", // Deshabilitar buffering en nginx
      },
    })
  } catch (error) {
    log.error("system.error", {
      error: String(error),
      endpoint: "POST /api/rag/generate",
      duration: Date.now() - start,
    }, { userId })

    return NextResponse.json(
      { ok: false, error: "Error interno del servidor" },
      { status: 500 }
    )
  }
}
