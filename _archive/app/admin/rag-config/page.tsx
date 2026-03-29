import { requireAdmin } from "@/lib/auth/current-user"
import { loadRagParams } from "@rag-saldivia/config"
import { RagConfigAdmin } from "@/components/admin/RagConfigAdmin"

export default async function AdminRagConfigPage() {
  await requireAdmin()
  const params = loadRagParams()

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Configuración RAG</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Parámetros del servidor de inferencia. Los cambios se aplican inmediatamente.
        </p>
      </div>
      <RagConfigAdmin params={params} />
    </div>
  )
}
