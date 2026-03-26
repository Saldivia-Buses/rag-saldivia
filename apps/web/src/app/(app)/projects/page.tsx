import { requireUser } from "@/lib/auth/current-user"
import { listProjects } from "@rag-saldivia/db"
import { ProjectsClient } from "@/components/projects/ProjectsClient"

export default async function ProjectsPage() {
  const user = await requireUser()
  const projects = await listProjects(user.id)
  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Proyectos</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Agrupá sesiones y colecciones con instrucciones de contexto compartidas.
        </p>
      </div>
      <ProjectsClient initialProjects={projects} />
    </div>
  )
}
