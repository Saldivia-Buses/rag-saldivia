/**
 * Cache de colecciones RAG via Redis.
 * F8.27 — Elimina unstable_cache (Next.js per-process) por Redis (compartido entre instancias).
 *
 * TTL: 60 segundos (mismo que el unstable_cache anterior).
 * Invalidación explícita: llamar a invalidateCollectionsCache() después de POST/DELETE en /api/rag/collections.
 */

import { ragFetch } from "./client"
import { getRedisClient } from "@rag-saldivia/db"

const CACHE_KEY = "rag:collections"
const CACHE_TTL_S = 60

async function fetchCollectionsFromRAG(): Promise<string[]> {
  const res = await ragFetch("/v1/collections")
  if ("error" in res) return []
  if (!res.ok) return []
  try {
    const data = await res.json()
    return (data.collections ?? []) as string[]
  } catch {
    return []
  }
}

/**
 * Obtiene la lista de colecciones desde Redis (TTL: 60s). Llamar a `invalidateCollectionsCache()`
 * después de cualquier POST o DELETE en `/api/rag/collections` — si no, la UI puede mostrar
 * datos obsoletos hasta ~60 segundos.
 */
export async function getCachedRagCollections(): Promise<string[]> {
  const redis = getRedisClient()
  const cached = await redis.get(CACHE_KEY)
  if (cached) return JSON.parse(cached) as string[]
  const fresh = await fetchCollectionsFromRAG()
  await redis.set(CACHE_KEY, JSON.stringify(fresh), "EX", CACHE_TTL_S)
  return fresh
}

export async function invalidateCollectionsCache(): Promise<void> {
  await getRedisClient().del(CACHE_KEY)
}
