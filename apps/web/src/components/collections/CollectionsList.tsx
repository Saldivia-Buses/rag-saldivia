/**
 * Milvus collection list with optimistic UI, search, permission badges, and detail navigation.
 */
"use client"

import { useOptimistic, useState, useTransition, useCallback } from "react"
import { useRouter } from "next/navigation"
import { FolderOpen, Trash2, MessageSquare, Plus, Search } from "lucide-react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionCreateCollection, actionDeleteCollection } from "@/app/actions/collections"

type CollectionRow = { name: string; permission: string | null }

type Props = {
  collections: CollectionRow[]
  user: CurrentUser
}

type Action =
  | { type: "create"; name: string }
  | { type: "delete"; name: string }

const PERM_COLORS: Record<string, string> = {
  read: "var(--accent)",
  write: "var(--success)",
  admin: "var(--warning)",
}

export function CollectionsList({ collections: initial, user }: Props) {
  const router = useRouter()
  const [optimistic, applyOptimistic] = useOptimistic(
    initial,
    (state: CollectionRow[], action: Action) => {
      if (action.type === "delete") return state.filter((c) => c.name !== action.name)
      if (action.type === "create") return [...state, { name: action.name, permission: null }]
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [newName, setNewName] = useState("")
  const [showCreate, setShowCreate] = useState(false)
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
      try { await actionCreateCollection({ name }) }
      catch (err) { setError(err instanceof Error ? err.message : "Error al crear") }
    })
  }

  const confirmDelete = useCallback(() => {
    if (!deleteTarget) return
    const name = deleteTarget
    setDeleteTarget(null)
    startTransition(async () => {
      applyOptimistic({ type: "delete", name })
      try { await actionDeleteCollection({ name }) }
      catch { /* silencioso */ }
    })
  }, [deleteTarget, applyOptimistic, startTransition])

  function handleChat(name: string) {
    router.push(`/chat?collection=${encodeURIComponent(name)}`)
  }

  return (
    <div className="flex flex-col" style={{ gap: "24px" }}>
      {user.role === "admin" && (
        <div>
          {!showCreate ? (
            <Button onClick={() => setShowCreate(true)} style={{ gap: "6px" }}>
              <Plus size={16} /> Nueva colección
            </Button>
          ) : (
            <form onSubmit={handleCreate} className="flex items-center" style={{ gap: "8px" }}>
              <Input
                autoFocus
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="nombre-de-coleccion"
                className="h-11 rounded-[10px]"
                style={{ width: "240px" }}
              />
              <Button type="submit" disabled={isPending}>
                {isPending ? "Creando..." : "Crear"}
              </Button>
              <Button variant="ghost" type="button" onClick={() => setShowCreate(false)}>
                Cancelar
              </Button>
            </form>
          )}
          {error && <p className="text-sm text-destructive" style={{ marginTop: "8px" }}>{error}</p>}
        </div>
      )}

      {/* Search */}
      {optimistic.length > 5 && (
        <div className="flex items-center rounded-lg border border-border bg-bg" style={{ padding: "4px 10px", gap: "6px", maxWidth: "300px" }}>
          <Search size={14} className="text-fg-subtle shrink-0" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar colección..."
            className="flex-1 bg-transparent text-sm text-fg placeholder:text-fg-subtle outline-none"
          />
        </div>
      )}

      {filtered.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FolderOpen} />
          <EmptyPlaceholder.Title>
            {search ? "Sin resultados" : "Sin colecciones disponibles"}
          </EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            {search
              ? "Probá con otro término."
              : user.role === "admin"
                ? "Creá una colección para empezar a ingestar documentos."
                : "No tenés acceso a ninguna colección todavía."}
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
              <div className="flex items-center" style={{ gap: "12px" }}>
                <FolderOpen size={18} className="text-accent" />
                <Link
                  href={`/collections/${encodeURIComponent(col.name)}`}
                  className="font-medium text-fg hover:text-accent transition-colors"
                >
                  {col.name}
                </Link>
                {col.permission && (
                  <Badge
                    variant="outline"
                    className="text-xs"
                    style={{ color: PERM_COLORS[col.permission] ?? "var(--fg-subtle)" }}
                  >
                    {col.permission}
                  </Badge>
                )}
              </div>
              <div className="flex items-center" style={{ gap: "4px" }}>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleChat(col.name)}
                  style={{ gap: "6px" }}
                >
                  <MessageSquare size={14} /> Chat
                </Button>
                {user.role === "admin" && (
                  <button
                    onClick={() => setDeleteTarget(col.name)}
                    className="p-2 rounded-md text-fg-subtle opacity-0 group-hover:opacity-100 hover:text-destructive hover:bg-surface transition-all"
                    title="Eliminar colección"
                  >
                    <Trash2 size={14} />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null) }}
        title={`¿Eliminar la colección "${deleteTarget}"?`}
        description="Esta acción no se puede deshacer. Se perderán todos los documentos indexados."
        onConfirm={confirmDelete}
      />
    </div>
  )
}
