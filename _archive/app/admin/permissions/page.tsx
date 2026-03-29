import { requireAdmin } from "@/lib/auth/current-user"
import { listAreas } from "@rag-saldivia/db"
import { getCachedRagCollections } from "@/lib/rag/collections-cache"
import { PermissionsAdmin } from "@/components/admin/PermissionsAdmin"

export default async function AdminPermissionsPage() {
  await requireAdmin()
  const [areas, collections] = await Promise.all([
    listAreas(),
    getCachedRagCollections(),
  ])

  return (
    <div className="p-6 max-w-5xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Permisos</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Asignación de colecciones a áreas con nivel de acceso
        </p>
      </div>
      <PermissionsAdmin areas={areas} collections={collections} />
    </div>
  )
}
