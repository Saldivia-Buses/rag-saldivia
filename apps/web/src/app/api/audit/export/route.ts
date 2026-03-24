/**
 * GET /api/audit/export
 * Exporta todos los eventos como JSON. Requiere rol admin.
 */

import { NextResponse } from "next/server"
import { extractClaims } from "@/lib/auth/jwt"
import { queryEvents } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return NextResponse.json({ ok: false, error: "Se requiere rol admin" }, { status: 403 })
  }

  const events = await queryEvents({ limit: 10000, order: "asc" })

  const response = new Response(JSON.stringify(events, null, 2), {
    headers: {
      "Content-Type": "application/json",
      "Content-Disposition": `attachment; filename="audit-export-${Date.now()}.json"`,
    },
  })

  return response
}
