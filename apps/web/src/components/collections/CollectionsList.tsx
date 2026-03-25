"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { FolderOpen, Trash2, MessageSquare, Plus } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import type { CurrentUser } from "@/lib/auth/current-user"

type Props = {
  collections: string[]
  user: CurrentUser
}

export function CollectionsList({ collections: initial, user }: Props) {
  const router = useRouter()
  const [collections, setCollections] = useState(initial)
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState("")
  const [showCreate, setShowCreate] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!newName.trim()) return
    setCreating(true)
    setError(null)
    try {
      const res = await fetch("/api/rag/collections", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName.trim() }),
      })
      const data = await res.json() as { ok: boolean; error?: string }
      if (!data.ok) throw new Error(data.error ?? "Error")
      setCollections((prev) => [...prev, newName.trim()])
      setNewName("")
      setShowCreate(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Error al crear")
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(name: string) {
    if (!confirm(`¿Eliminar la colección "${name}"? Esta acción no se puede deshacer.`)) return
    try {
      await fetch(`/api/rag/collections/${encodeURIComponent(name)}`, { method: "DELETE" })
      setCollections((prev) => prev.filter((c) => c !== name))
    } catch {
      // silencioso
    }
  }

  function handleChat(name: string) {
    router.push(`/chat?collection=${encodeURIComponent(name)}`)
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
              <input
                autoFocus
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="nombre-de-coleccion"
                className="px-3 py-1.5 rounded-md border text-sm outline-none"
                style={{ borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }}
              />
              <Button size="sm" type="submit" disabled={creating}>
                {creating ? "Creando..." : "Crear"}
              </Button>
              <Button size="sm" variant="ghost" type="button" onClick={() => setShowCreate(false)}>
                Cancelar
              </Button>
            </form>
          )}
          {error && <p className="text-sm mt-1" style={{ color: "var(--destructive)" }}>{error}</p>}
        </div>
      )}

      <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
        <Table>
          <TableHeader>
            <TableRow style={{ background: "var(--muted)" }}>
              <TableHead>Colección</TableHead>
              <TableHead className="text-right">Acciones</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {collections.length === 0 ? (
              <TableRow>
                <TableCell colSpan={2} className="text-center py-8" style={{ color: "var(--muted-foreground)" }}>
                  No hay colecciones disponibles
                </TableCell>
              </TableRow>
            ) : (
              collections.map((name) => (
                <TableRow key={name}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <FolderOpen size={15} style={{ color: "var(--accent)" }} />
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
                      {user.role === "admin" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-7 w-7"
                          onClick={() => handleDelete(name)}
                          style={{ color: "var(--destructive)" }}
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
    </div>
  )
}
