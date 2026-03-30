/**
 * MessageActions — context menu for messages: edit, delete, copy, reply.
 */
"use client"

import { useState, useTransition } from "react"
import { Copy, Pencil, Trash2, MessageSquare, Pin, Check } from "lucide-react"
import { actionEditMessage, actionDeleteMessage } from "@/app/actions/messaging"
import { cn } from "@/lib/utils"

type MessageData = {
  id: string
  userId: number
  content: string
  channelId: string
}

export function MessageActions({
  message,
  currentUserId,
  isAdmin,
  onReply,
  onPin,
  onClose,
}: {
  message: MessageData
  currentUserId: number
  isAdmin: boolean
  onReply?: () => void
  onPin?: () => void
  onClose: () => void
}) {
  const [mode, setMode] = useState<"menu" | "edit" | "delete">("menu")
  const [editContent, setEditContent] = useState(message.content)
  const [isPending, startTransition] = useTransition()
  const [copied, setCopied] = useState(false)

  const isOwn = message.userId === currentUserId
  const canDelete = isOwn || isAdmin

  function handleCopy() {
    navigator.clipboard.writeText(message.content)
    setCopied(true)
    setTimeout(() => {
      setCopied(false)
      onClose()
    }, 800)
  }

  function handleEdit() {
    const trimmed = editContent.trim()
    if (!trimmed || trimmed === message.content) {
      onClose()
      return
    }
    startTransition(async () => {
      await actionEditMessage({ id: message.id, content: trimmed })
      onClose()
    })
  }

  function handleDelete() {
    startTransition(async () => {
      await actionDeleteMessage({ id: message.id })
      onClose()
    })
  }

  if (mode === "edit") {
    return (
      <div className="p-2 w-72">
        <textarea
          value={editContent}
          onChange={(e) => setEditContent(e.target.value)}
          rows={3}
          className="w-full px-2 py-1.5 rounded border border-border bg-bg text-sm text-fg outline-none focus:border-accent resize-none"
          autoFocus
        />
        <div className="flex justify-end gap-1.5 mt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={isPending}
            className="px-2.5 py-1 text-xs rounded border border-border text-fg-muted hover:bg-surface-2"
          >
            Cancelar
          </button>
          <button
            type="button"
            onClick={handleEdit}
            disabled={isPending || !editContent.trim()}
            className="px-2.5 py-1 text-xs rounded bg-accent text-white hover:bg-accent/90 disabled:opacity-50"
          >
            {isPending ? "Guardando..." : "Guardar"}
          </button>
        </div>
      </div>
    )
  }

  if (mode === "delete") {
    return (
      <div className="p-3 w-64">
        <p className="text-sm text-fg mb-3">¿Eliminar este mensaje?</p>
        <div className="flex justify-end gap-1.5">
          <button
            type="button"
            onClick={onClose}
            disabled={isPending}
            className="px-2.5 py-1 text-xs rounded border border-border text-fg-muted hover:bg-surface-2"
          >
            Cancelar
          </button>
          <button
            type="button"
            onClick={handleDelete}
            disabled={isPending}
            className="px-2.5 py-1 text-xs rounded bg-destructive text-white hover:bg-destructive/90"
          >
            {isPending ? "Eliminando..." : "Eliminar"}
          </button>
        </div>
      </div>
    )
  }

  // Menu mode
  return (
    <div className="py-1 w-48">
      {onReply && (
        <ActionButton icon={MessageSquare} label="Responder en hilo" onClick={() => { onReply(); onClose() }} />
      )}
      <ActionButton
        icon={copied ? Check : Copy}
        label={copied ? "¡Copiado!" : "Copiar texto"}
        onClick={handleCopy}
      />
      {onPin && (
        <ActionButton icon={Pin} label="Fijar mensaje" onClick={() => { onPin(); onClose() }} />
      )}
      {isOwn && (
        <ActionButton icon={Pencil} label="Editar" onClick={() => setMode("edit")} />
      )}
      {canDelete && (
        <ActionButton icon={Trash2} label="Eliminar" onClick={() => setMode("delete")} destructive />
      )}
    </div>
  )
}

function ActionButton({
  icon: Icon,
  label,
  onClick,
  destructive,
}: {
  icon: React.ComponentType<{ className?: string; size?: number }>
  label: string
  onClick: () => void
  destructive?: boolean
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full flex items-center gap-2 px-3 py-1.5 text-sm transition-colors",
        destructive
          ? "text-destructive hover:bg-destructive/10"
          : "text-fg hover:bg-surface-2",
      )}
    >
      <Icon size={14} className="shrink-0" />
      {label}
    </button>
  )
}
