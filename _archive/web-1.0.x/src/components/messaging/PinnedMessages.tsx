/**
 * PinnedMessages — side drawer showing pinned messages in a channel.
 */
"use client"

import { Pin, X } from "lucide-react"

type PinnedMsg = {
  channelId: string
  messageId: string
  pinnedBy: number
  pinnedAt: number
  message: {
    id: string
    content: string
    userId: number
    createdAt: number
  } | null
}

type MemberInfo = { id: number; name: string }

export function PinnedMessages({
  pinnedMessages,
  members,
  onClose,
  onUnpin,
}: {
  pinnedMessages: PinnedMsg[]
  members: MemberInfo[]
  onClose: () => void
  onUnpin?: (messageId: string) => void
}) {
  function getMemberName(userId: number) {
    return members.find((m) => m.id === userId)?.name ?? "Usuario"
  }

  function formatDate(ts: number) {
    return new Date(ts).toLocaleDateString("es-AR", {
      day: "numeric",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
    })
  }

  return (
    <div className="w-80 shrink-0 border-l border-border bg-surface flex flex-col h-full">
      <div className="px-4 py-3 border-b border-border flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Pin className="h-4 w-4 text-fg-muted" />
          <h3 className="text-sm font-semibold text-fg">Mensajes fijados</h3>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="text-fg-subtle hover:text-fg transition-colors"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {pinnedMessages.length === 0 ? (
          <div className="flex items-center justify-center h-full px-4">
            <p className="text-sm text-fg-subtle text-center">
              No hay mensajes fijados en este canal
            </p>
          </div>
        ) : (
          <div className="flex flex-col gap-1 p-2">
            {pinnedMessages.map((pin) => (
              <div
                key={pin.messageId}
                className="rounded-lg border border-border bg-bg p-3 group"
              >
                {pin.message ? (
                  <>
                    <div className="flex items-baseline gap-2 mb-1">
                      <span className="text-xs font-semibold text-fg">
                        {getMemberName(pin.message.userId)}
                      </span>
                      <span className="text-[10px] text-fg-subtle">
                        {formatDate(pin.message.createdAt)}
                      </span>
                    </div>
                    <p className="text-sm text-fg whitespace-pre-wrap break-words line-clamp-4">
                      {pin.message.content}
                    </p>
                  </>
                ) : (
                  <p className="text-xs text-fg-subtle italic">Mensaje no disponible</p>
                )}
                {onUnpin && (
                  <button
                    type="button"
                    onClick={() => onUnpin(pin.messageId)}
                    className="mt-2 text-xs text-fg-subtle hover:text-destructive transition-colors opacity-0 group-hover:opacity-100"
                  >
                    Desfijar
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
