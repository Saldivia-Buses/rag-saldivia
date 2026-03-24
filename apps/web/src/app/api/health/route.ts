/**
 * GET /api/health
 * Health check público del servidor Next.js.
 */

import { NextResponse } from "next/server"

export async function GET() {
  return NextResponse.json({
    ok: true,
    status: "healthy",
    service: "rag-saldivia-web",
    ts: Date.now(),
  })
}
