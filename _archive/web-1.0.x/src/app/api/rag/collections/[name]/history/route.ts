/**
 * GET /api/rag/collections/[name]/history — ingestion history for a collection
 */

import { NextResponse } from "next/server"
import { listHistoryByCollection, canAccessCollection } from "@rag-saldivia/db"
import { requireAuth, apiError } from "@/lib/api-utils"

export async function GET(
  request: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const { name } = await params
  const userId = Number(claims.sub)

  // Admin sees all, users need collection access
  if (claims.role !== "admin") {
    const hasAccess = await canAccessCollection(userId, name, "read")
    if (!hasAccess) return apiError("Sin acceso a esta colección", 403)
  }

  const history = await listHistoryByCollection(name)
  return NextResponse.json({ ok: true, data: history })
}
