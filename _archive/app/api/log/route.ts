/**
 * POST /api/log
 * Recibe eventos del cliente (browser) y los persiste en la tabla events.
 * Accesible sin autenticación (el logger de frontend necesita enviar errores incluso al inicio).
 */

import { NextResponse } from "next/server"
import { writeEvent } from "@rag-saldivia/db"
import { extractClaims } from "@/lib/auth/jwt"

export async function POST(request: Request) {
  try {
    const body = await request.json()
    const eventsPayload = body.events as Array<{
      type: string
      payload?: Record<string, unknown>
      ts: number
    }>

    if (!Array.isArray(eventsPayload) || eventsPayload.length === 0) {
      return NextResponse.json({ ok: true, received: 0 })
    }

    // Intentar obtener usuario si hay cookie (opcional)
    const claims = await extractClaims(request).catch(() => null)
    const userId = claims ? Number(claims.sub) : null

    // Limitar batch a 50 eventos por request
    const toProcess = eventsPayload.slice(0, 50)

    await Promise.allSettled(
      toProcess.map((e) =>
        writeEvent({
          source: "frontend",
          level: "INFO",
          type: (e.type ?? "client.action") as Parameters<typeof writeEvent>[0]["type"],
          userId,
          payload: { ...e.payload, clientTs: e.ts },
        })
      )
    )

    return NextResponse.json({ ok: true, received: toProcess.length })
  } catch {
    // Silencioso — no romper el cliente por un error de logging
    return NextResponse.json({ ok: true, received: 0 })
  }
}
