"use client"

import { useState, useTransition } from "react"
import { Plus, Pencil, Trash2 } from "lucide-react"
import type { DbArea } from "@rag-saldivia/db"
import { actionCreateArea, actionUpdateArea, actionDeleteArea } from "@/app/actions/areas"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { FolderOpen } from "lucide-react"

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
        setShowCreate(false); setNewName(""); setNewDesc("")
      } catch (err) { setError(String(err)) }
    })
  }

  async function handleUpdate(id: number) {
    setError(null)
    startTransition(async () => {
      try {
        await actionUpdateArea(id, { name: editName, description: editDesc })
        setEditingId(null)
      } catch (err) { setError(String(err)) }
    })
  }

  async function handleDelete(id: number, name: string) {
    if (!confirm(`¿Eliminar el área "${name}"?\nTodos los usuarios del área perderán el acceso.`)) return
    startTransition(async () => {
      try { await actionDeleteArea(id) }
      catch (err) { setError(String(err)) }
    })
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Áreas</h1>
          <p className="text-sm text-fg-muted mt-0.5">{areas.length} área{areas.length !== 1 ? "s" : ""}</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <Plus className="h-3.5 w-3.5" /> Nueva área
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {showCreate && (
        <form onSubmit={handleCreate} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Nueva área</h3>
          <Input placeholder="Nombre del área" value={newName} onChange={(e) => setNewName(e.target.value)} required />
          <Input placeholder="Descripción (opcional)" value={newDesc} onChange={(e) => setNewDesc(e.target.value)} />
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={isPending}>{isPending ? "Creando..." : "Crear"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {areas.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FolderOpen} />
          <EmptyPlaceholder.Title>Sin áreas</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Creá un área para agrupar usuarios y colecciones.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="rounded-xl border border-border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead>Colecciones</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {areas.map((area) => (
                <TableRow key={area.id}>
                  <TableCell>
                    {editingId === area.id
                      ? <Input value={editName} onChange={(e) => setEditName(e.target.value)} className="h-7 text-xs" />
                      : <span className="font-medium text-fg">{area.name}</span>
                    }
                  </TableCell>
                  <TableCell className="text-fg-muted">
                    {editingId === area.id
                      ? <Input value={editDesc} onChange={(e) => setEditDesc(e.target.value)} className="h-7 text-xs" />
                      : area.description || "—"
                    }
                  </TableCell>
                  <TableCell className="text-xs text-fg-subtle">
                    {area.areaCollections?.map((ac) => ac.collectionName).join(", ") || "—"}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1 justify-end">
                      {editingId === area.id ? (
                        <>
                          <Button size="sm" className="h-7 text-xs" onClick={() => handleUpdate(area.id)} disabled={isPending}>Guardar</Button>
                          <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => setEditingId(null)}>Cancelar</Button>
                        </>
                      ) : (
                        <>
                          <Button variant="ghost" size="icon" className="h-7 w-7" title="Editar"
                            onClick={() => { setEditingId(area.id); setEditName(area.name); setEditDesc(area.description) }}>
                            <Pencil className="h-3.5 w-3.5" />
                          </Button>
                          <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" title="Eliminar"
                            onClick={() => handleDelete(area.id, area.name)} disabled={isPending}>
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
