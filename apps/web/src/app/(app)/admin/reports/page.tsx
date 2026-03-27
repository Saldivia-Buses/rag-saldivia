import { requireAdmin } from "@/lib/auth/current-user"
import { listReportsByUser } from "@rag-saldivia/db"
import { ReportsAdmin } from "@/components/admin/ReportsAdmin"

export default async function ReportsPage() {
  const admin = await requireAdmin()
  const reports = await listReportsByUser(admin.id)

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Informes programados</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          El sistema ejecuta el query automáticamente y guarda el resultado.
          {!process.env["SMTP_HOST"] && " (Email no configurado — se usará destino Guardados)"}
        </p>
      </div>
      <ReportsAdmin initialReports={reports} />
    </div>
  )
}
