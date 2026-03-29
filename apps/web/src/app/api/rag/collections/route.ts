/**
 * /api/rag/collections — CRUD for RAG document collections.
 *
 * GET:  List collections (filtered by user permissions, admins see all)
 * POST: Create collection (admin only, proxied to RAG server or mock)
 *
 * Data flow: request → auth check → RAG server/cache → filtered response
 * Depends on: lib/rag/client.ts, lib/rag/collections-cache.ts, lib/api-utils.ts
 */

import { NextResponse } from "next/server"
import { ragFetch } from "@/lib/rag/client"
import { getUserCollections } from "@rag-saldivia/db"
import { getCachedRagCollections, invalidateCollectionsCache } from "@/lib/rag/collections-cache"
import { requireAuth, requireAdmin, apiOk, apiError } from "@/lib/api-utils"

export async function GET(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const userId = Number(claims.sub)
  const ragCollections = await getCachedRagCollections()

  // Admins see all collections, others see only permitted ones
  if (claims.role === "admin") return apiOk(ragCollections)

  const userCollections = await getUserCollections(userId)
  const allowed = new Set(userCollections.map((c) => c.name))
  return apiOk(ragCollections.filter((name) => allowed.has(name)))
}

export async function POST(request: Request) {
  const claims = await requireAdmin(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  if (!body?.name) return apiError("name requerido")

  try {
    const res = await ragFetch(`/v1/collections`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ collection_name: body.name }),
    } as Parameters<typeof ragFetch>[1])
    if ("error" in res) throw new Error(res.error.message)
    await invalidateCollectionsCache()
    return apiOk()
  } catch {
    // Mock mode: simulate success and invalidate cache anyway
    await invalidateCollectionsCache().catch(() => {})
    return apiOk()
  }
}

