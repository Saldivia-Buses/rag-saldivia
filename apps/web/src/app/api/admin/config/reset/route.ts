/**
 * POST /api/admin/config/reset — resetear parámetros RAG a defaults (CLI)
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { saveRagParams } from "@rag-saldivia/config"
import { RagParamsSchema } from "@rag-saldivia/shared"
import { log } from "@rag-saldivia/logger/backend"

export async function POST(request: Request) {
  const claims = await extractClaims(request)
  if (!claims) return NextResponse.json({ ok: false, error: "No autenticado" }, { status: 401 })
  if (claims.role !== "admin") return NextResponse.json({ ok: false, error: "Acceso denegado" }, { status: 403 })

  const defaults = RagParamsSchema.parse({})
  await saveRagParams(defaults)
  log.info("admin.config_changed", { reset: true }, { userId: Number(claims.sub) })
  return NextResponse.json({ ok: true, data: defaults })
}
