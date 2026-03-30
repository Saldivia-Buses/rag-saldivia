/**
 * Milvus collection management UI — list, create, delete, and jump to chat.
 *
 * Displays all available vector collections with optimistic UI updates.
 * Admin users see create/delete controls; regular users see read-only list.
 *
 * Key patterns:
 *   - useOptimistic: create/delete actions update the UI instantly, then the
 *     server action runs in a transition. On error, React re-renders with the
 *     actual server state (automatic rollback).
 *   - "Chat" button navigates to /chat?collection=<name> to start a session
 *     scoped to that collection.
 *
 * Data flow: /collections page (RSC, fetches from RAG API) -> collections prop
 * Depends on: server actions (create/delete collection), EmptyPlaceholder
 */
"use client"

import { useOptimistic, useState, useTransition, useCallback } from "react"
import { useRouter } from "next/navigation"
import { FolderOpen, Trash2, MessageSquare, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionCreateCollection, actionDeleteCollection } from "@/app/actions/collections"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"

type Props = {
  collections: string[]
  user: CurrentUser
}

export function CollectionsList({ collections: initial, user }: Props) {
  const router = useRouter()
  const [optimisticCollections, applyOptimistic] = useOptimistic(
    initial,
    (state, action: { type: "delete"; name: string } | { type: "create"; name: string }) => {
      if (action.type === "delete") return state.filter((c) => c !== action.name)
      if (action.type === "create") return [...state, action.name]
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [newName, setNewName] = useState("")
  const [showCreate, setShowCreate] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)

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

      {optimisticCollections.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FolderOpen} />
          <EmptyPlaceholder.Title>Sin colecciones disponibles</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            {user.role === "admin"
              ? "Creá una colección para empezar a ingestar documentos."
              : "No tenés acceso a ninguna colección todavía."}
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="flex flex-col" style={{ gap: "8px" }}>
          {optimisticCollections.map((name) => (
            <div
              key={name}
              className="group flex items-center justify-between rounded-xl border border-border bg-surface transition-colors hover:bg-surface-2"
              style={{ padding: "16px 20px" }}
            >
              <div className="flex items-center" style={{ gap: "12px" }}>
                <FolderOpen size={18} className="text-accent" />
                <span className="font-medium text-fg">{name}</span>
              </div>
              <div className="flex items-center" style={{ gap: "4px" }}>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleChat(name)}
                  style={{ gap: "6px" }}
                >
                  <MessageSquare size={14} /> Chat
                </Button>
                {user.role === "admin" && (
                  <button
                    onClick={() => setDeleteTarget(name)}
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
