import { requireAdmin } from "@/lib/auth/current-user"
import { KnowledgeGapsClient } from "@/components/admin/KnowledgeGapsClient"

export default async function KnowledgeGapsPage() {
  await requireAdmin()
  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Brechas de conocimiento</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Queries donde el RAG respondió con baja confianza. Guía para qué ingestar.
        </p>
      </div>
      <KnowledgeGapsClient />
    </div>
  )
}
