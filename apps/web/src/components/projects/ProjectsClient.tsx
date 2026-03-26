"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Plus, Trash2, FolderKanban, MessageSquare } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { DbProject } from "@rag-saldivia/db"

type Props = {
  initialProjects: DbProject[]
}

export function ProjectsClient({ initialProjects }: Props) {
  const router = useRouter()
  const [projects, setProjects] = useState(initialProjects)
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ name: "", description: "", instructions: "" })
  const [creating, setCreating] = useState(false)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    try {
      const res = await fetch("/api/projects", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      })
      const d = await res.json() as { ok: boolean; data?: DbProject }
      if (d.ok && d.data) {
        setProjects((p) => [...p, d.data!])
        setShowCreate(false)
        setForm({ name: "", description: "", instructions: "" })
      }
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("¿Eliminar este proyecto?")) return
    await fetch(`/api/projects?id=${id}`, { method: "DELETE" })
    setProjects((p) => p.filter((pr) => pr.id !== id))
  }

  const inputClass = "w-full px-3 py-1.5 rounded-md border text-sm outline-none"
  const inputStyle = { borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }

  return (
    <div className="space-y-4">
      <Button size="sm" onClick={() => setShowCreate(!showCreate)} className="gap-1.5">
        <Plus size={13} /> Nuevo proyecto
      </Button>

      {showCreate && (
        <form onSubmit={handleCreate} className="p-4 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
          <div>
            <label className="text-xs font-medium">Nombre</label>
            <input value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle} required />
          </div>
          <div>
            <label className="text-xs font-medium">Descripción</label>
            <input value={form.description} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle} />
          </div>
          <div>
            <label className="text-xs font-medium">Instrucciones de contexto</label>
            <textarea value={form.instructions} onChange={(e) => setForm((p) => ({ ...p, instructions: e.target.value }))} className={`${inputClass} mt-1 min-h-[60px]`} style={inputStyle} placeholder="Ej: Responde siempre con el contexto legal de Argentina..." />
          </div>
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={creating}>{creating ? "Creando..." : "Crear"}</Button>
            <Button size="sm" variant="ghost" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {projects.length === 0 ? (
        <div className="rounded-xl border p-12 text-center" style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}>
          <FolderKanban size={32} className="mx-auto mb-3 opacity-40" />
          <p className="text-sm">Sin proyectos. Creá uno para agrupar sesiones con contexto compartido.</p>
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {projects.map((p) => (
            <div key={p.id} className="p-4 rounded-xl border space-y-2 hover:opacity-90 transition-opacity" style={{ borderColor: "var(--border)" }}>
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2 min-w-0">
                  <FolderKanban size={16} style={{ color: "var(--accent)", flexShrink: 0 }} />
                  <span className="font-medium truncate">{p.name}</span>
                </div>
                <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => handleDelete(p.id)} style={{ color: "var(--destructive)" }}>
                  <Trash2 size={13} />
                </Button>
              </div>
              {p.description && <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{p.description}</p>}
              {p.instructions && (
                <p className="text-xs line-clamp-2 px-2 py-1 rounded" style={{ background: "var(--muted)", color: "var(--muted-foreground)" }}>
                  📌 {p.instructions}
                </p>
              )}
              <Button size="sm" variant="outline" className="w-full gap-1.5 mt-2" onClick={() => router.push(`/projects/${p.id}`)}>
                <MessageSquare size={12} /> Ver sesiones
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
