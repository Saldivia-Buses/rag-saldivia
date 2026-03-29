/**
 * GET /api/audit/export
 * Exporta todos los eventos como JSON o CSV. Requiere rol admin.
 *
 * Query params:
 *   format=json (default) | csv
 */

import { NextResponse } from "next/server"
import Papa from "papaparse"
import { extractClaims } from "@/lib/auth/jwt"
import { queryEvents } from "@rag-saldivia/db"

export async function GET(request: Request) {
  const claims = await extractClaims(request)
  if (!claims || claims.role !== "admin") {
    return NextResponse.json({ ok: false, error: "Se requiere rol admin" }, { status: 403 })
  }

  const url = new URL(request.url)
  const format = url.searchParams.get("format") === "csv" ? "csv" : "json"

  const events = await queryEvents({ limit: 10000, order: "asc" })
  const timestamp = Date.now()

  if (format === "csv") {
    const rows = events.map((e) => ({
      ts: new Date(e.ts).toISOString(),
      level: e.level,
      type: e.type,
      userId: e.userId ?? "",
      sessionId: e.sessionId ?? "",
      payload: JSON.stringify(e.payload),
    }))

    const csv = Papa.unparse({
      fields: ["ts", "level", "type", "userId", "sessionId", "payload"],
      data: rows,
    })

    return new Response(csv, {
      headers: {
        "Content-Type": "text/csv; charset=utf-8",
        "Content-Disposition": `attachment; filename="audit-export-${timestamp}.csv"`,
      },
    })
  }

  return new Response(JSON.stringify(events, null, 2), {
    headers: {
      "Content-Type": "application/json",
      "Content-Disposition": `attachment; filename="audit-export-${timestamp}.json"`,
    },
  })
}
