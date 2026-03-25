import { requireUser } from "@/lib/auth/current-user"
import { CollectionsList } from "@/components/collections/CollectionsList"

async function getCollections(userId: number, role: string): Promise<string[]> {
  try {
    // Llamamos al propio endpoint (server-side fetch relativo no funciona en Next.js — usamos la lógica directa)
    const { ragFetch } = await import("@/lib/rag/client")
    const { getUserCollections } = await import("@rag-saldivia/db")

    const res = await ragFetch("/v1/collections")
    let allCollections: string[] = []
    if (!("error" in res) && res.ok) {
      try {
        const data = await res.json() as { collections?: string[] }
        allCollections = data.collections ?? []
      } catch {
        allCollections = []
      }
    }

    if (role === "admin") return allCollections

    const userCols = await getUserCollections(userId)
    const allowed = new Set(userCols.map((c) => c.name))
    return allCollections.filter((name) => allowed.has(name))
  } catch {
    return []
  }
}

export default async function CollectionsPage() {
  const user = await requireUser()
  const collections = await getCollections(user.id, user.role)

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Colecciones</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          {collections.length} colección{collections.length !== 1 ? "es" : ""} disponible{collections.length !== 1 ? "s" : ""}
        </p>
      </div>
      <CollectionsList collections={collections} user={user} />
    </div>
  )
}
