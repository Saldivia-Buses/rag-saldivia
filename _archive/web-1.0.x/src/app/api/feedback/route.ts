/**
 * POST /api/feedback
 * Receives error reports from users and stores them in audit_log.
 * Protected: requires authenticated user (proxy.ts verifies JWT).
 */

import { NextResponse, type NextRequest } from "next/server"
import { getDb, auditLog } from "@rag-saldivia/db"

export async function POST(request: NextRequest) {
  const userId = Number(request.headers.get("x-user-id"))
  if (!userId) {
    return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  }

  let body: { error?: string; context?: string; comment?: string }
  try {
    body = await request.json()
  } catch {
    return NextResponse.json({ ok: false, error: "Body inválido" }, { status: 400 })
  }

  const { error, context, comment } = body
  if (!error || typeof error !== "string") {
    return NextResponse.json({ ok: false, error: "Campo 'error' requerido" }, { status: 400 })
  }

  const db = getDb()
  await db.insert(auditLog).values({
    userId,
    action: "error_feedback",
    queryPreview: JSON.stringify({
      error: error.slice(0, 500),
      context: (context ?? "").slice(0, 200),
      comment: (comment ?? "").slice(0, 1000),
    }),
    ipAddress: request.headers.get("x-forwarded-for") ?? request.headers.get("x-real-ip") ?? "",
    timestamp: Date.now(),
  })

  return NextResponse.json({ ok: true })
}
