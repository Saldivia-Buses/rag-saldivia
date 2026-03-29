import { requireAdmin } from "@/lib/auth/current-user"
import { listAreas } from "@rag-saldivia/db"
import { AreasAdmin } from "@/components/admin/AreasAdmin"

export default async function AdminAreasPage() {
  await requireAdmin()
  const areas = await listAreas()

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Áreas</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Grupos de usuarios con acceso a colecciones específicas
        </p>
      </div>
      <AreasAdmin areas={areas} />
    </div>
  )
}
