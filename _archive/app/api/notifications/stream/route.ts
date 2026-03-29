/**
 * GET /api/notifications/stream
 *
 * SSE endpoint para notificaciones en tiempo real via Redis Pub/Sub.
 * F8.28 — Reemplaza el polling cada 30s por push server-sent events.
 *
 * El worker de ingesta publica en `notifications:{userId}` cuando completa o falla.
 * IDs de notificaciones vistas se persisten en Redis Sorted Set (no localStorage).
 */

import { extractClaims } from "@/lib/auth/jwt"
import { getRedisClient } from "@rag-saldivia/db"
import { getDb } from "@rag-saldivia/db"
import { events } from "@rag-saldivia/db"
import { inArray, desc } from "drizzle-orm"

export const runtime = "nodejs"

const NOTIFICATION_TYPES = [
  "ingestion.completed",
  "ingestion.error",
  "user.created",
  "proactive.docs_available",
]

const SEEN_SET_KEY = (userId: number) => `notifications:seen:${userId}`
const SEEN_TTL_MS = 30 * 24 * 60 * 60 * 1000 // 30 días

async function isNotificationSeen(userId: number, notifId: string): Promise<boolean> {
  const score = await getRedisClient().zscore(SEEN_SET_KEY(userId), notifId)
  return score !== null
}

async function markNotificationSeen(userId: number, notifId: string): Promise<void> {
  const redis = getRedisClient()
  const now = Date.now()
  await redis.zadd(SEEN_SET_KEY(userId), now, notifId)
  // Limpiar IDs más viejos de 30 días
  const cutoff = now - SEEN_TTL_MS
  redis.zremrangebyscore(SEEN_SET_KEY(userId), 0, cutoff).catch(() => {})
}

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) {
    return new Response("No autenticado", { status: 401 })
  }

  const userId = Number(claims.sub)
  const role = claims.role as string
  const allowedTypes = role === "admin"
    ? NOTIFICATION_TYPES
    : NOTIFICATION_TYPES.filter((t) => t !== "user.created")

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

      // Enviar notificaciones recientes no vistas como estado inicial
      try {
        const db = getDb()
        const recent = await db
          .select({ id: events.id, type: events.type, ts: events.ts, payload: events.payload })
          .from(events)
          .where(inArray(events.type, allowedTypes))
          .orderBy(desc(events.ts))
          .limit(20)

        for (const n of recent) {
          if (!await isNotificationSeen(userId, n.id)) {
            emit({ id: n.id, type: n.type, ts: n.ts, payload: n.payload })
            await markNotificationSeen(userId, n.id)
          }
        }
      } catch {
        // Continuar aunque falle la carga inicial
      }

      // Suscribirse al canal Redis para notificaciones en tiempo real
      // Se necesita una conexión dedicada — subscribe bloquea la conexión
      const sub = getRedisClient().duplicate()
      const channel = `notifications:${userId}`

      await sub.subscribe(channel)

      sub.on("message", async (_ch: string, message: string) => {
        try {
          const notif = JSON.parse(message) as { id?: string; type: string; ts?: number; payload?: unknown }
          const notifId = notif.id ?? `${notif.type}-${Date.now()}`

          if (!allowedTypes.includes(notif.type)) return
          if (await isNotificationSeen(userId, notifId)) return

          emit({ id: notifId, type: notif.type, ts: notif.ts ?? Date.now(), payload: notif.payload ?? {} })
          await markNotificationSeen(userId, notifId)
        } catch {
          // Ignorar mensajes malformados
        }
      })

      request.signal.addEventListener("abort", () => {
        sub.unsubscribe(channel).catch(() => {})
        sub.quit().catch(() => {})
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
