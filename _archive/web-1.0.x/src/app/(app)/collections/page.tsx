import { requireUser } from "@/lib/auth/current-user"
import { getUserCollections } from "@rag-saldivia/db"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"
import { CollectionsList } from "@/components/collections/CollectionsList"

export default async function CollectionsPage() {
  const user = await requireUser()

  const ragCollections = await getCachedRagCollections()

  let collections: Array<{ name: string; permission: string | null }>
  if (user.role === "admin") {
    collections = ragCollections.map((name) => ({ name, permission: null }))
  } else {
    const userCols = await getUserCollections(user.id)
    const allowed = new Map(userCols.map((c) => [c.name, c.permission]))
    collections = ragCollections
      .filter((name) => allowed.has(name))
      .map((name) => ({ name, permission: allowed.get(name) ?? null }))
  }

  return (
    <div className="max-w-3xl mx-auto" style={{ padding: "32px 24px" }}>
      <div style={{ marginBottom: "32px" }}>
        <h1 className="text-2xl font-semibold text-fg">Colecciones</h1>
        <p className="text-sm text-fg-muted" style={{ marginTop: "4px" }}>
          {collections.length} colección{collections.length !== 1 ? "es" : ""} disponible{collections.length !== 1 ? "s" : ""}
        </p>
      </div>
      <CollectionsList collections={collections} user={user} />
    </div>
  )
}
