import { redirect } from "next/navigation"
import { requireUser } from "@/lib/auth/current-user"
import { getUserCollections, listHistoryByCollection, listAreas } from "@rag-saldivia/db"
import { CollectionDetail } from "@/components/collections/CollectionDetail"

export default async function CollectionDetailPage({
  params,
}: {
  params: Promise<{ name: string }>
}) {
  const user = await requireUser()
  const { name } = await params
  const decodedName = decodeURIComponent(name)

  const isAdmin = user.role === "admin"

  // Check user permission for this collection
  let userPermission: string | null = null
  if (!isAdmin) {
    const userCols = await getUserCollections(user.id)
    const col = userCols.find((c) => c.name === decodedName)
    if (!col) redirect("/collections")
    userPermission = col.permission
  }

  // Load history and areas info in parallel
  const [history, areas] = await Promise.all([
    listHistoryByCollection(decodedName),
    isAdmin ? listAreas() : Promise.resolve([]),
  ])

  // Extract areas that reference this collection
  const collectionAreas = areas
    .filter((a) => a.areaCollections.some((ac) => ac.collectionName === decodedName))
    .map((a) => ({
      name: a.name,
      permission: a.areaCollections.find((ac) => ac.collectionName === decodedName)!.permission,
    }))

  return (
    <div className="max-w-3xl mx-auto" style={{ padding: "32px 24px" }}>
      <CollectionDetail
        name={decodedName}
        userPermission={userPermission}
        isAdmin={isAdmin}
        areas={collectionAreas}
        history={history}
      />
    </div>
  )
}
