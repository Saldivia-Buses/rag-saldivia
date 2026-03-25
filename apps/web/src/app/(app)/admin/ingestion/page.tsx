import { requireAdmin } from "@/lib/auth/current-user"
import { IngestionKanban } from "@/components/admin/IngestionKanban"

export default async function IngestionMonitorPage() {
  await requireAdmin()

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Monitoring de ingesta</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Estado de los jobs en tiempo real
        </p>
      </div>
      <IngestionKanban />
    </div>
  )
}
