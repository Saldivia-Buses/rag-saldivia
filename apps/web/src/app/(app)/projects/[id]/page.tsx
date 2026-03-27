import { notFound } from "next/navigation"
import { requireUser } from "@/lib/auth/current-user"
import { getProject } from "@rag-saldivia/db"
import Link from "next/link"
import { ArrowLeft, FolderKanban } from "lucide-react"
import { formatDate } from "@/lib/utils"

export default async function ProjectDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const user = await requireUser()
  const { id } = await params
  const project = await getProject(id)

  if (!project || project.userId !== user.id) notFound()

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <Link href="/projects" className="flex items-center gap-1.5 text-sm mb-4 hover:opacity-70 transition-opacity" style={{ color: "var(--muted-foreground)" }}>
        <ArrowLeft size={14} /> Proyectos
      </Link>
      <div className="flex items-center gap-3 mb-6">
        <FolderKanban size={24} style={{ color: "var(--accent)" }} />
        <div>
          <h1 className="text-xl font-semibold">{project.name}</h1>
          {project.description && (
            <p className="text-sm mt-0.5" style={{ color: "var(--muted-foreground)" }}>{project.description}</p>
          )}
        </div>
      </div>

      {project.instructions && (
        <div className="mb-6 p-4 rounded-xl border text-sm" style={{ borderColor: "var(--border)", background: "var(--muted)" }}>
          <p className="text-xs font-medium mb-1" style={{ color: "var(--accent)" }}>📌 Instrucciones de contexto</p>
          <p style={{ color: "var(--foreground)" }}>{project.instructions}</p>
        </div>
      )}

      <div className="text-sm" style={{ color: "var(--muted-foreground)" }}>
        <p>Las sesiones y colecciones del proyecto se configuran via la CLI o la API.</p>
        <p className="mt-1">Creado: {formatDate(project.createdAt)}</p>
      </div>
    </div>
  )
}
