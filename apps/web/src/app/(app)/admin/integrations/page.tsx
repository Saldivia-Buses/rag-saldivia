import { requireAdmin } from "@/lib/auth/current-user"
import { IntegrationsAdmin } from "@/components/admin/IntegrationsAdmin"

export default async function IntegrationsPage() {
  await requireAdmin()
  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Integraciones</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Configurá el bot de Slack y Teams para que los usuarios puedan consultar el RAG directamente desde sus plataformas.
        </p>
      </div>
      <IntegrationsAdmin />
    </div>
  )
}
