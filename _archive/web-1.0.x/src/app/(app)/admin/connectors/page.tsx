import { listExternalSources, countSyncDocuments } from "@rag-saldivia/db"
import { requireAdmin } from "@/lib/auth/current-user"
import { AdminConnectors } from "@/components/admin/AdminConnectors"

export default async function AdminConnectorsPage() {
  const user = await requireAdmin()
  const sources = await listExternalSources(user.id)

  const connectors = await Promise.all(
    sources.map(async (s) => ({
      id: s.id,
      provider: s.provider,
      name: s.name,
      collectionDest: s.collectionDest,
      schedule: s.schedule,
      active: s.active,
      lastSync: s.lastSync,
      docCount: await countSyncDocuments(s.id),
    }))
  )

  return <AdminConnectors initialConnectors={connectors} />
}
