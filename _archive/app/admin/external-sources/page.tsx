import { requireAdmin } from "@/lib/auth/current-user"
import { listExternalSources } from "@rag-saldivia/db"
import { ExternalSourcesAdmin } from "@/components/admin/ExternalSourcesAdmin"

export default async function ExternalSourcesPage() {
  const admin = await requireAdmin()
  const sources = await listExternalSources(admin.id)

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Fuentes externas</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Auto-ingesta desde Google Drive, SharePoint y Confluence.
          El worker sincroniza automáticamente según el schedule configurado.
        </p>
      </div>
      <ExternalSourcesAdmin initialSources={sources} />
    </div>
  )
}
