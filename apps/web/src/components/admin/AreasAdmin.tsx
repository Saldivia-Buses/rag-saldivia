"use client"

import { useState, useTransition } from "react"
import { Plus, Pencil, Trash2, Users } from "lucide-react"
import type { DbArea } from "@rag-saldivia/db"
import { actionCreateArea, actionUpdateArea, actionDeleteArea } from "@/app/actions/areas"

type AreaWithCollections = DbArea & {
  areaCollections?: Array<{ collectionName: string; permission: string }>
}

export function AreasAdmin({ areas: initialAreas }: { areas: AreaWithCollections[] }) {
  const [areas, setAreas] = useState(initialAreas)
  const [showCreate, setShowCreate] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [newName, setNewName] = useState("")
  const [newDesc, setNewDesc] = useState("")
  const [editName, setEditName] = useState("")
  const [editDesc, setEditDesc] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    startTransition(async () => {
      try {
        await actionCreateArea(newName, newDesc)
        setShowCreate(false)
        setNewName(""); setNewDesc("")
      } catch (err) {
        setError(String(err))
      }
    })
  }

  async function handleUpdate(id: number) {
    setError(null)
    startTransition(async () => {
      try {
        await actionUpdateArea(id, { name: editName, description: editDesc })
        setEditingId(null)
      } catch (err) {
        setError(String(err))
      }
    })
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`¿Eliminar el área "${name}"?\nTodos los usuarios del área perderán el acceso.`)) return
    setError(null)
    startTransition(async () => {
      try {
        await actionDeleteArea(id)
      } catch (err) {
        setError(String(err))
      }
    })
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>
          {areas.length} área(s)
        </span>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium"
          style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
        >
          <Plus size={15} /> Nueva área
        </button>
      </div>

      {error && (
        <div className="px-3 py-2 rounded-md text-sm" style={{ background: "#fef2f2", color: "var(--destructive)" }}>
          {error}
        </div>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="p-4 rounded-lg border space-y-3" style={{ borderColor: "var(--border)" }}>
          <h3 className="font-medium text-sm">Nueva área</h3>
          <input
            placeholder="Nombre del área"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            required
            className="w-full px-3 py-2 rounded-md border text-sm"
            style={{ borderColor: "var(--border)" }}
          />
          <input
            placeholder="Descripción (opcional)"
            value={newDesc}
            onChange={(e) => setNewDesc(e.target.value)}
            className="w-full px-3 py-2 rounded-md border text-sm"
            style={{ borderColor: "var(--border)" }}
          />
          <div className="flex gap-2">
            <button type="submit" disabled={isPending} className="px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50" style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}>
              {isPending ? "Creando..." : "Crear"}
            </button>
            <button type="button" onClick={() => setShowCreate(false)} className="px-4 py-2 rounded-md text-sm border" style={{ borderColor: "var(--border)" }}>
              Cancelar
            </button>
          </div>
        </form>
      )}

      <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
        <table className="w-full text-sm">
          <thead>
            <tr style={{ background: "var(--muted)" }}>
              <th className="text-left px-4 py-3 font-medium">Nombre</th>
              <th className="text-left px-4 py-3 font-medium">Descripción</th>
              <th className="text-left px-4 py-3 font-medium">Colecciones</th>
              <th className="px-4 py-3" />
            </tr>
          </thead>
          <tbody>
            {areas.map((area, i) => (
              <tr key={area.id} style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}>
                <td className="px-4 py-3">
                  {editingId === area.id ? (
                    <input
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      className="px-2 py-1 rounded border text-sm w-full"
                      style={{ borderColor: "var(--border)" }}
                    />
                  ) : (
                    <span className="font-medium">{area.name}</span>
                  )}
                </td>
                <td className="px-4 py-3" style={{ color: "var(--muted-foreground)" }}>
                  {editingId === area.id ? (
                    <input
                      value={editDesc}
                      onChange={(e) => setEditDesc(e.target.value)}
                      className="px-2 py-1 rounded border text-sm w-full"
                      style={{ borderColor: "var(--border)" }}
                    />
                  ) : (
                    area.description || "—"
                  )}
                </td>
                <td className="px-4 py-3 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {area.areaCollections?.map((ac) => ac.collectionName).join(", ") || "—"}
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-1 justify-end">
                    {editingId === area.id ? (
                      <>
                        <button onClick={() => handleUpdate(area.id)} disabled={isPending} className="px-2 py-1 rounded text-xs font-medium" style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}>
                          Guardar
                        </button>
                        <button onClick={() => setEditingId(null)} className="px-2 py-1 rounded text-xs border" style={{ borderColor: "var(--border)" }}>
                          Cancelar
                        </button>
                      </>
                    ) : (
                      <>
                        <button onClick={() => { setEditingId(area.id); setEditName(area.name); setEditDesc(area.description) }} className="p-1.5 rounded hover:opacity-80" title="Editar">
                          <Pencil size={14} />
                        </button>
                        <button onClick={() => handleDelete(area.id, area.name)} disabled={isPending} className="p-1.5 rounded hover:opacity-80" style={{ color: "var(--destructive)" }} title="Eliminar">
                          <Trash2 size={14} />
                        </button>
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
