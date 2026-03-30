import { listAreas } from "@rag-saldivia/db"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"
import { AdminCollections } from "@/components/admin/AdminCollections"

export default async function AdminCollectionsPage() {
  const [collections, areas] = await Promise.all([
    getCachedRagCollections(),
    listAreas(),
  ])

  // Build reverse map: collection → areas that reference it
  const collectionAreas = new Map<string, Array<{ areaName: string; permission: string }>>()
  for (const area of areas) {
    for (const ac of area.areaCollections) {
      const existing = collectionAreas.get(ac.collectionName) ?? []
      existing.push({ areaName: area.name, permission: ac.permission })
      collectionAreas.set(ac.collectionName, existing)
    }
  }

  const rows = collections.map((name) => ({
    name,
    areas: collectionAreas.get(name) ?? [],
  }))

  return <AdminCollections collections={rows} />
}
