/**
 * GET /api/admin/ingestion/stream
 *
 * SSE endpoint que emite el estado de los jobs de ingesta cada 3 segundos.
 * Solo accesible por admins.
 */

import { extractClaims } from "@/lib/auth/jwt"
import { getDb, ingestionJobs } from "@rag-saldivia/db"
import { desc } from "drizzle-orm"

export const runtime = "nodejs"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return new Response("No autorizado", { status: 401 })
  }

  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder()

      function emit(data: unknown) {
        try {
          controller.enqueue(encoder.encode(`data: ${JSON.stringify(data)}\n\n`))
        } catch {
          // cliente desconectado
        }
      }

      async function fetchJobs() {
        const db = getDb()
        return db
          .select()
          .from(ingestionJobs)
          .orderBy(desc(ingestionJobs.createdAt))
          .limit(50)
      }

      // Emitir estado inicial
      const initialJobs = await fetchJobs()
      emit({ jobs: initialJobs })

      // Polling cada 3s
      const interval = setInterval(async () => {
        try {
          const jobs = await fetchJobs()
          emit({ jobs })
        } catch {
          clearInterval(interval)
          controller.close()
        }
      }, 3000)

      // Limpiar cuando el cliente desconecta
      request.signal.addEventListener("abort", () => {
        clearInterval(interval)
        try { controller.close() } catch { /* ya cerrado */ }
      })
    },
  })

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
      "X-Accel-Buffering": "no",
    },
  })
}
