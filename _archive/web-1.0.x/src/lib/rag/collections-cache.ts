/**
 * Cache de colecciones RAG via Redis.
 * F8.27 — Elimina unstable_cache (Next.js per-process) por Redis (compartido entre instancias).
 *
 * TTL: 60 segundos (mismo que el unstable_cache anterior).
 * Invalidation: call invalidateCollectionsCache() after POST/DELETE on /api/rag/collections.
 * Graceful degradation: bypasses cache if Redis is down, fetches directly from RAG server.
 */

import { ragFetch } from "./client"
import { getRedisClient } from "@rag-saldivia/db"
import { COLLECTIONS_CACHE_TTL_S } from "@rag-saldivia/config"

const CACHE_KEY = "rag:collections"

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
 * Get collection list from Redis cache (TTL: 60s). Falls back to direct RAG fetch if Redis is down.
 * Call `invalidateCollectionsCache()` after any POST or DELETE on `/api/rag/collections`.
 */
export async function getCachedRagCollections(): Promise<string[]> {
  try {
    const redis = getRedisClient()
    const cached = await redis.get(CACHE_KEY)
    if (cached) return JSON.parse(cached) as string[]
    const fresh = await fetchCollectionsFromRAG()
    await redis.set(CACHE_KEY, JSON.stringify(fresh), "EX", COLLECTIONS_CACHE_TTL_S).catch(() => {})
    return fresh
  } catch {
    // Redis unavailable — bypass cache, fetch directly
    return fetchCollectionsFromRAG()
  }
}

export async function invalidateCollectionsCache(): Promise<void> {
  try {
    await getRedisClient().del(CACHE_KEY)
  } catch {
    // Redis unavailable — cache will expire naturally via TTL
  }
}
