"use client"

import { useOptimistic, useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { FolderOpen, Trash2, MessageSquare, Plus, Network } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { CurrentUser } from "@/lib/auth/current-user"
import { actionCreateSessionForDoc } from "@/app/actions/chat"
import { actionCreateCollection, actionDeleteCollection } from "@/app/actions/collections"

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
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState("")
  const [showCreate, setShowCreate] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!newName.trim()) return
    const name = newName.trim()
    setError(null)
    setNewName("")
    setShowCreate(false)
    startTransition(async () => {
      applyOptimistic({ type: "create", name })
      try { await actionCreateCollection(name) }
      catch (err) { setError(err instanceof Error ? err.message : "Error al crear") }
    })
  }

  function handleDelete(name: string) {
    if (!confirm(`¿Eliminar la colección "${name}"? Esta acción no se puede deshacer.`)) return
    startTransition(async () => {
      applyOptimistic({ type: "delete", name })
      try { await actionDeleteCollection(name) }
      catch { /* silencioso */ }
    })
  }

  function handleChat(name: string) {
    router.push(`/chat?collection=${encodeURIComponent(name)}`)
  }

  async function handleChatWithDoc(collection: string, docName: string) {
    const session = await actionCreateSessionForDoc(collection, docName)
    if (session) router.push(`/chat/${session.id}`)
  }

  return (
    <div className="space-y-4">
      {user.role === "admin" && (
        <div>
          {!showCreate ? (
            <Button size="sm" onClick={() => setShowCreate(true)} className="gap-1.5">
              <Plus size={14} /> Nueva colección
            </Button>
          ) : (
            <form onSubmit={handleCreate} className="flex items-center gap-2">
              <Input
                autoFocus
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="nombre-de-coleccion"
                className="w-48"
              />
              <Button size="sm" type="submit" disabled={creating}>
                {creating ? "Creando..." : "Crear"}
              </Button>
              <Button size="sm" variant="ghost" type="button" onClick={() => setShowCreate(false)}>
                Cancelar
              </Button>
            </form>
          )}
          {error && <p className="text-sm mt-1 text-destructive">{error}</p>}
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
      <div className="rounded-lg border border-border overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Colección</TableHead>
              <TableHead className="text-right">Acciones</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(
              optimisticCollections.map((name) => (
                <TableRow key={name}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <FolderOpen size={15} className="text-accent" />
                      <span className="font-medium">{name}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 gap-1 text-xs"
                        onClick={() => handleChat(name)}
                      >
                        <MessageSquare size={12} /> Chat
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 gap-1 text-xs"
                        onClick={() => router.push(`/collections/${encodeURIComponent(name)}/graph`)}
                        title="Ver grafo de documentos"
                      >
                        <Network size={12} /> Grafo
                      </Button>
                      {user.role === "admin" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7 text-destructive hover:text-destructive"
                          onClick={() => handleDelete(name)}
                          title="Eliminar colección"
                        >
                          <Trash2 size={13} />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
      )}
    </div>
  )
}
