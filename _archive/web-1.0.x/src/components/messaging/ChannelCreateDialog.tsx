/**
 * ChannelCreateDialog — create public or private channel.
 */
"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { Hash, Lock, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { actionCreateChannel } from "@/app/actions/messaging"
import { cn } from "@/lib/utils"

export function ChannelCreateDialog({
  open,
  onClose,
}: {
  open: boolean
  onClose: () => void
}) {
  const router = useRouter()
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [type, setType] = useState<"public" | "private">("public")
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState("")

  if (!open) return null

  function handleCreate() {
    const trimmed = name.trim()
    if (!trimmed) {
      setError("El nombre es obligatorio")
      return
    }
    if (trimmed.length > 80) {
      setError("Máximo 80 caracteres")
      return
    }

    setError("")
    startTransition(async () => {
      const result = await actionCreateChannel({
        type,
        name: trimmed,
        ...(description.trim() ? { description: description.trim() } : {}),
      })
      if (result?.data) {
        onClose()
        setName("")
        setDescription("")
        setType("public")
        router.push(`/messaging/${result.data.id}`)
      }
    })
  }

  return (
    <>
      {/* Backdrop */}
      <div className="fixed inset-0 z-50 bg-black/50" onClick={onClose} />

      {/* Dialog */}
      <div className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-md bg-surface border border-border rounded-xl shadow-xl">
        <div className="flex items-center justify-between px-5 py-4 border-b border-border">
          <h2 className="text-base font-semibold text-fg">Crear canal</h2>
          <button type="button" onClick={onClose} className="text-fg-subtle hover:text-fg">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="px-5 py-4 flex flex-col gap-4">
          {/* Type selector */}
          <div>
            <label className="text-xs font-medium text-fg-muted mb-1.5 block">Tipo</label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setType("public")}
                className={cn(
                  "flex-1 flex items-center gap-2 px-3 py-2 rounded-lg border text-sm transition-colors",
                  type === "public"
                    ? "border-accent bg-accent-subtle text-accent"
                    : "border-border text-fg-muted hover:bg-surface-2",
                )}
              >
                <Hash className="h-4 w-4" />
                Público
              </button>
              <button
                type="button"
                onClick={() => setType("private")}
                className={cn(
                  "flex-1 flex items-center gap-2 px-3 py-2 rounded-lg border text-sm transition-colors",
                  type === "private"
                    ? "border-accent bg-accent-subtle text-accent"
                    : "border-border text-fg-muted hover:bg-surface-2",
                )}
              >
                <Lock className="h-4 w-4" />
                Privado
              </button>
            </div>
          </div>

          {/* Name */}
          <div>
            <label htmlFor="channel-name" className="text-xs font-medium text-fg-muted mb-1.5 block">
              Nombre
            </label>
            <input
              id="channel-name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="ej. proyecto-flota"
              className="w-full px-3 py-2 rounded-lg border border-border bg-bg text-sm text-fg placeholder:text-fg-subtle outline-none focus:border-accent"
              maxLength={80}
              autoFocus
            />
          </div>

          {/* Description */}
          <div>
            <label htmlFor="channel-desc" className="text-xs font-medium text-fg-muted mb-1.5 block">
              Descripción <span className="text-fg-subtle">(opcional)</span>
            </label>
            <textarea
              id="channel-desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="De qué se trata este canal..."
              rows={2}
              className="w-full px-3 py-2 rounded-lg border border-border bg-bg text-sm text-fg placeholder:text-fg-subtle outline-none focus:border-accent resize-none"
              maxLength={500}
            />
          </div>

          {error && <p className="text-xs text-destructive">{error}</p>}
        </div>

        <div className="flex items-center justify-end gap-2 px-5 py-3 border-t border-border">
          <Button variant="outline" onClick={onClose} disabled={isPending}>
            Cancelar
          </Button>
          <Button onClick={handleCreate} disabled={isPending || !name.trim()}>
            {isPending ? "Creando..." : "Crear canal"}
          </Button>
        </div>
      </div>
    </>
  )
}
