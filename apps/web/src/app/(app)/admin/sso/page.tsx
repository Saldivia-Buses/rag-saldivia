import { listAllSsoProviders } from "@rag-saldivia/db"
import { AdminSso } from "@/components/admin/AdminSso"

export default async function AdminSsoPage() {
  const providers = await listAllSsoProviders()

  return (
    <AdminSso
      initialProviders={providers.map((p) => ({
        id: p.id,
        name: p.name,
        type: p.type,
        clientId: p.clientId,
        active: p.active,
        autoProvision: p.autoProvision,
        defaultRole: p.defaultRole,
        createdAt: p.createdAt,
      }))}
    />
  )
}
