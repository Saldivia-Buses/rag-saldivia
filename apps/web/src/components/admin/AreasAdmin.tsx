"use client"

import { useOptimistic, useTransition, useState } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Plus, Pencil, Trash2 } from "lucide-react"
import type { DbArea } from "@rag-saldivia/db"
import { actionCreateArea, actionUpdateArea, actionDeleteArea } from "@/app/actions/areas"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

const AreaSchema = z.object({
  name: z.string().min(2, "El nombre debe tener al menos 2 caracteres"),
  description: z.string().default(""),
})
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { FolderOpen } from "lucide-react"

type AreaWithCollections = DbArea & {
  areaCollections?: Array<{ collectionName: string; permission: string }>
}

type AreaAction =
  | { type: "delete"; id: number }
  | { type: "create"; area: AreaWithCollections }

export function AreasAdmin({ areas: initialAreas }: { areas: AreaWithCollections[] }) {
  const [optimisticAreas, applyOptimistic] = useOptimistic(
    initialAreas,
    (state, action: AreaAction) => {
      if (action.type === "delete") return state.filter((a) => a.id !== action.id)
      if (action.type === "create") return [...state, action.area]
      return state
    }
  )
  const [showCreate, setShowCreate] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  const createForm = useForm<{ name: string; description: string }>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(AreaSchema as any),
    defaultValues: { name: "", description: "" },
  })

  const editForm = useForm<{ name: string; description: string }>({
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    resolver: zodResolver(AreaSchema as any),
    defaultValues: { name: "", description: "" },
  })

  function handleCreate(data: { name: string; description: string }) {
    setError(null)
    startTransition(async () => {
      applyOptimistic({ type: "create", area: { id: Date.now(), name: data.name, description: data.description, createdAt: Date.now(), areaCollections: [] } })
      try {
        await actionCreateArea(data.name, data.description || undefined)
        setShowCreate(false)
        createForm.reset()
      } catch (err) { setError(String(err)) }
    })
  }

  function handleUpdate(id: number, data: { name: string; description: string }) {
    setError(null)
    startTransition(async () => {
      try {
        await actionUpdateArea(id, { name: data.name, description: data.description })
        setEditingId(null)
      } catch (err) { setError(String(err)) }
    })
  }

  function handleDelete(id: number, name: string) {
    if (!confirm(`¿Eliminar el área "${name}"?\nTodos los usuarios del área perderán el acceso.`)) return
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      try { await actionDeleteArea(id) }
      catch (err) { setError(String(err)) }
    })
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Áreas</h1>
          <p className="text-sm text-fg-muted mt-0.5">{optimisticAreas.length} área{optimisticAreas.length !== 1 ? "s" : ""}</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <Plus className="h-3.5 w-3.5" /> Nueva área
        </Button>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      {showCreate && (
        <form onSubmit={createForm.handleSubmit(handleCreate)} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Nueva área</h3>
          <Input placeholder="Nombre del área" {...createForm.register("name")} />
          {createForm.formState.errors.name && <p className="text-xs text-destructive">{createForm.formState.errors.name.message}</p>}
          <Input placeholder="Descripción (opcional)" {...createForm.register("description")} />
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={isPending}>{isPending ? "Creando..." : "Crear"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {optimisticAreas.length === 0 ? (
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
              {optimisticAreas.map((area) => (
                <TableRow key={area.id}>
                  <TableCell>
                    {editingId === area.id
                      ? <Input {...editForm.register("name")} className="h-7 text-xs" />
                      : <span className="font-medium text-fg">{area.name}</span>
                    }
                  </TableCell>
                  <TableCell className="text-fg-muted">
                    {editingId === area.id
                      ? <Input {...editForm.register("description")} className="h-7 text-xs" />
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
                          <Button size="sm" className="h-7 text-xs" onClick={() => handleUpdate(area.id, editForm.getValues())} disabled={isPending}>Guardar</Button>
                          <Button size="sm" variant="outline" className="h-7 text-xs" onClick={() => setEditingId(null)}>Cancelar</Button>
                        </>
                      ) : (
                        <>
                          <Button variant="ghost" size="icon" className="h-7 w-7" title="Editar"
                            onClick={() => { setEditingId(area.id); editForm.reset({ name: area.name, description: area.description ?? "" }) }}>
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
