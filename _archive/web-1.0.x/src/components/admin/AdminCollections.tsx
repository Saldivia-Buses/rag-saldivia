"use client"

import { useState, useTransition, useOptimistic, useCallback } from "react"
import { Plus, Trash2, FolderOpen, ExternalLink } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { actionCreateCollection, actionDeleteCollection } from "@/app/actions/collections"

type AreaInfo = { areaName: string; permission: string }
type CollectionRow = { name: string; areas: AreaInfo[] }

type Action =
  | { type: "create"; name: string }
  | { type: "delete"; name: string }

const PERM_COLORS: Record<string, string> = {
  read: "var(--accent)",
  write: "var(--success)",
  admin: "var(--warning)",
}

export function AdminCollections({
  collections: initial,
}: {
  collections: CollectionRow[]
}) {
  const [optimistic, applyOptimistic] = useOptimistic(
    initial,
    (state: CollectionRow[], action: Action) => {
      if (action.type === "delete") return state.filter((c) => c.name !== action.name)
      if (action.type === "create") return [...state, { name: action.name, areas: [] }]
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [search, setSearch] = useState("")

  const filtered = search.trim()
    ? optimistic.filter((c) => c.name.toLowerCase().includes(search.toLowerCase()))
    : optimistic

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!newName.trim()) return
    const name = newName.trim()
    setError(null)
    setNewName("")
    setShowCreate(false)
    startTransition(async () => {
      applyOptimistic({ type: "create", name })
      try {
        await actionCreateCollection({ name })
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }

  const confirmDelete = useCallback(() => {
    if (!deleteTarget) return
    const name = deleteTarget
    setDeleteTarget(null)
    setError(null)
    startTransition(async () => {
      applyOptimistic({ type: "delete", name })
      try {
        await actionDeleteCollection({ name })
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }, [deleteTarget, applyOptimistic, startTransition])

  return (
    <div className="flex flex-col" style={{ gap: "20px" }}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <p className="text-sm text-fg-muted">
          {optimistic.length} colección{optimistic.length !== 1 ? "es" : ""}
        </p>
        <Button onClick={() => setShowCreate(!showCreate)} style={{ gap: "6px" }}>
          <Plus size={16} /> Nueva colección
        </Button>
      </div>

      {error && (
        <p className="text-sm text-destructive" style={{ padding: "8px 12px", borderRadius: "8px", background: "color-mix(in srgb, var(--destructive) 8%, transparent)" }}>
          {error}
        </p>
      )}

      {/* Create form */}
      {showCreate && (
        <form onSubmit={handleCreate} className="flex items-center" style={{ gap: "8px" }}>
          <Input
            autoFocus
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="nombre-de-coleccion (minúsculas, guiones)"
            style={{ maxWidth: "320px" }}
          />
          <Button type="submit" disabled={isPending || !newName.trim()}>
            {isPending ? "Creando..." : "Crear"}
          </Button>
          <Button variant="ghost" type="button" onClick={() => setShowCreate(false)}>
            Cancelar
          </Button>
        </form>
      )}

      {/* Search */}
      {optimistic.length > 5 && (
        <Input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Buscar colección..."
          style={{ maxWidth: "300px" }}
        />
      )}

      {/* List */}
      {filtered.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FolderOpen} />
          <EmptyPlaceholder.Title>
            {search ? "Sin resultados" : "Sin colecciones"}
          </EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            {search ? "Probá con otro término." : "Creá una colección para empezar."}
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="flex flex-col" style={{ gap: "8px" }}>
          {filtered.map((col) => (
            <div
              key={col.name}
              className="group flex items-center justify-between rounded-xl border border-border bg-surface transition-colors hover:bg-surface-2"
              style={{ padding: "16px 20px" }}
            >
              <div className="flex items-center min-w-0" style={{ gap: "12px" }}>
                <FolderOpen size={18} className="text-accent shrink-0" />
                <Link
                  href={`/collections/${encodeURIComponent(col.name)}`}
                  className="font-medium text-fg hover:text-accent transition-colors"
                >
                  {col.name}
                </Link>
                {/* Area badges */}
                <div className="flex items-center" style={{ gap: "4px" }}>
                  {col.areas.map((a) => (
                    <Badge
                      key={a.areaName}
                      variant="outline"
                      className="text-xs"
                      style={{ color: PERM_COLORS[a.permission] ?? "var(--fg-subtle)" }}
                    >
                      {a.areaName}
                    </Badge>
                  ))}
                </div>
              </div>

              <div className="flex items-center shrink-0" style={{ gap: "4px" }}>
                <Link
                  href={`/collections/${encodeURIComponent(col.name)}`}
                  className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                  style={{ width: "32px", height: "32px" }}
                  title="Ver detalle"
                >
                  <ExternalLink size={14} />
                </Link>
                <button
                  onClick={() => setDeleteTarget(col.name)}
                  className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-destructive hover:bg-surface-2 transition-colors opacity-0 group-hover:opacity-100"
                  style={{ width: "32px", height: "32px" }}
                  title="Eliminar"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null) }}
        title={`¿Eliminar la colección "${deleteTarget}"?`}
        description="Se perderán todos los documentos indexados. Las áreas que la referencian perderán el acceso."
        onConfirm={confirmDelete}
      />
    </div>
  )
}
