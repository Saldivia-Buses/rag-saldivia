/**
 * /api/rag/collections — CRUD for RAG document collections.
 *
 * GET:  List collections (filtered by user permissions, admins see all)
 * POST: Create collection (admin only, proxied to RAG server or mock)
 *
 * Data flow: request -> auth check -> RAG server/cache -> filtered response
 * Depends on: lib/rag/client.ts, lib/rag/collections-cache.ts, lib/api-utils.ts
 */

import { NextResponse } from "next/server"
import { ragFetch } from "@/lib/rag/client"
import { getUserCollections } from "@rag-saldivia/db"
import { getCachedRagCollections, invalidateCollectionsCache } from "@/lib/rag/collections-cache"
import { requireAuth, requireAdmin, apiOk, apiError, apiServerError } from "@/lib/api-utils"
import { CollectionNameSchema } from "@rag-saldivia/shared"

export async function GET(request: Request) {
  const claims = await requireAuth(request)
  if (claims instanceof NextResponse) return claims

  const userId = Number(claims.sub)
  const ragCollections = await getCachedRagCollections()

  // Admins see all collections, others see only permitted ones
  let filtered: string[]
  if (claims.role === "admin") {
    filtered = ragCollections
  } else {
    const userCollections = await getUserCollections(userId)
    const allowed = new Set(userCollections.map((c) => c.name))
    filtered = ragCollections.filter((name) => allowed.has(name))
  }

  const response = apiOk(filtered)
  response.headers.set("Cache-Control", "private, max-age=60")
  return response
}

export async function POST(request: Request) {
  const claims = await requireAdmin(request)
  if (claims instanceof NextResponse) return claims

  const body = await request.json().catch(() => null)
  const parsed = CollectionNameSchema.safeParse(body?.name)
  if (!parsed.success) {
    return apiError(parsed.error.issues[0]?.message ?? "Nombre de colección inválido")
  }

  try {
    const res = await ragFetch(`/v1/collections`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ collection_name: parsed.data }),
    } as Parameters<typeof ragFetch>[1])
    if ("error" in res) {
      return apiError(`No se pudo crear la colección: ${res.error.message}`, 502)
    }
  } catch (error) {
    return apiServerError(error, "POST /api/rag/collections", Number(claims.sub))
  }

  await invalidateCollectionsCache().catch(() => {})
  return apiOk()
}
