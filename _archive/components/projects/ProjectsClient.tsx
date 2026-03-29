"use client"

import { useOptimistic, useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { Plus, Trash2, FolderKanban, MessageSquare } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { DbProject } from "@rag-saldivia/db"
import { actionCreateProject, actionDeleteProject } from "@/app/actions/projects"

type Props = { initialProjects: DbProject[] }

export function ProjectsClient({ initialProjects }: Props) {
  const router = useRouter()
  const [optimisticProjects, applyOptimistic] = useOptimistic(
    initialProjects,
    (state, action: { type: "delete"; id: string }) => {
      if (action.type === "delete") return state.filter((p) => p.id !== action.id)
      return state
    }
  )
  const [_isPending, startTransition] = useTransition()
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: "", description: "", instructions: "" })
  const [creating, setCreating] = useState(false)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    try {
      await actionCreateProject(form)
      setShowCreate(false)
      setForm({ name: "", description: "", instructions: "" })
    } finally {
      setCreating(false)
    }
  }

  function handleDelete(id: string) {
    if (!confirm("¿Eliminar este proyecto?")) return
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      await actionDeleteProject(id)
    })
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Proyectos</h1>
          <p className="text-sm text-fg-muted mt-0.5">{optimisticProjects.length} proyecto{optimisticProjects.length !== 1 ? "s" : ""}</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <Plus className="h-3.5 w-3.5" /> Nuevo proyecto
        </Button>
      </div>

      {showCreate && (
        <form onSubmit={handleCreate} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Crear proyecto</h3>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Nombre</label>
            <Input value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} required placeholder="Mi proyecto" />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Descripción</label>
            <Input value={form.description} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} placeholder="Descripción opcional" />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Instrucciones de contexto</label>
            <Textarea
              value={form.instructions}
              onChange={(e) => setForm((p) => ({ ...p, instructions: e.target.value }))}
              placeholder="Ej: Responde siempre con el contexto legal de Argentina..."
              className="min-h-[60px]"
            />
          </div>
          <div className="flex gap-2 pt-1">
            <Button size="sm" type="submit" disabled={creating}>{creating ? "Creando..." : "Crear proyecto"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {optimisticProjects.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FolderKanban} />
          <EmptyPlaceholder.Title>Sin proyectos</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            Creá un proyecto para agrupar sesiones con contexto compartido.
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {optimisticProjects.map((p) => (
            <div key={p.id} className="rounded-xl border border-border bg-surface p-4 space-y-3 hover:shadow-sm transition-shadow">
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2 min-w-0">
                  <FolderKanban size={16} className="text-accent shrink-0" />
                  <span className="font-medium text-fg truncate">{p.name}</span>
                </div>
                <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0 text-destructive hover:text-destructive"
                  onClick={() => handleDelete(p.id)}>
                  <Trash2 size={13} />
                </Button>
              </div>
              {p.description && <p className="text-xs text-fg-muted">{p.description}</p>}
              {p.instructions && (
                <p className="text-xs line-clamp-2 px-2 py-1.5 rounded-lg bg-surface-2 text-fg-muted">
                  📌 {p.instructions}
                </p>
              )}
              <Button size="sm" variant="outline" className="w-full gap-1.5"
                onClick={() => router.push(`/projects/${p.id}`)}>
                <MessageSquare size={12} /> Ver sesiones
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
