import { requireAdmin } from "@/lib/auth/current-user"
import { AnalyticsDashboard } from "@/components/admin/AnalyticsDashboard"

export default async function AnalyticsPage() {
  await requireAdmin()

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Analytics</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Uso del sistema en los últimos 30 días
        </p>
      </div>
      <AnalyticsDashboard />
    </div>
  )
}
