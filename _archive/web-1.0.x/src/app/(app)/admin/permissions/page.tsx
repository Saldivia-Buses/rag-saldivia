import { listAreas } from "@rag-saldivia/db"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"
import { AdminPermissions } from "@/components/admin/AdminPermissions"

export default async function PermissionsPage({
  searchParams,
}: {
  searchParams: Promise<{ area?: string }>
}) {
  const [areas, collections, params] = await Promise.all([
    listAreas(),
    getCachedRagCollections(),
    searchParams,
  ])

  return (
    <AdminPermissions
      areas={areas}
      collections={collections}
      {...(params.area ? { preselectedAreaId: Number(params.area) } : {})}
    />
  )
}
