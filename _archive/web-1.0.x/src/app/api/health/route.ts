/**
 * GET /api/health
 * Health check público del servidor Next.js.
 * Verifica conectividad con Redis — retorna 503 si está caído.
 */

import { NextResponse } from "next/server"
import { getRedisClient } from "@rag-saldivia/db"

export async function GET() {
  const redisOk = await getRedisClient()
    .ping()
    .then(() => true)
    .catch(() => false)

  if (!redisOk) {
    return NextResponse.json(
      { ok: false, service: "redis", status: "down", ts: Date.now() },
      { status: 503 }
    )
  }

  return NextResponse.json({
    ok: true,
    status: "healthy",
    service: "rag-saldivia-web",
    ts: Date.now(),
  })
}
