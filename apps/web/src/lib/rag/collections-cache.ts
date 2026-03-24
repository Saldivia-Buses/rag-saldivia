import { unstable_cache } from "next/cache"
import { ragFetch } from "./client"

export const getCachedRagCollections = unstable_cache(
  async (): Promise<string[]> => {
    const res = await ragFetch("/v1/collections")
    if ("error" in res) return []
    if (!res.ok) return []
    try {
      const data = await res.json()
      return (data.collections ?? []) as string[]
    } catch {
      return []
    }
  },
  ["rag-collections-admin"],
  { revalidate: 60, tags: ["collections"] }
)
