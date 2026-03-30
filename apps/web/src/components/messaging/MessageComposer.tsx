/**
 * Message composer — textarea with auto-resize, Enter to send, Shift+Enter newline.
 * Supports optimistic send and reply-in-thread indicator.
 */
"use client"

import { useRef, useState, useTransition } from "react"
import { Send, X } from "lucide-react"
import { useAutoResize } from "@/hooks/useAutoResize"
import { actionSendMessage } from "@/app/actions/messaging"
import { cn } from "@/lib/utils"

type MemberInfo = { id: number; name: string; email: string }

export function MessageComposer({
  channelId,
  currentUserId,
  members,
  replyTo,
  onClearReply,
  onOptimisticMessage,
}: {
  channelId: string
  currentUserId: number
  members: MemberInfo[]
  replyTo?: { id: string; userId: number; content: string } | null
  onClearReply?: () => void
  onOptimisticMessage?: (msg: {
    id: string
    channelId: string
    userId: number
    content: string
    type: string
    createdAt: number
    editedAt: null
    deletedAt: null
    replyCount: number
    parentId: string | null
  }) => void
}) {
  const [value, setValue] = useState("")
  const [isPending, startTransition] = useTransition()
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  useAutoResize(textareaRef, value)

  const replyAuthor = replyTo
    ? members.find((m) => m.id === replyTo.userId)?.name ?? "Usuario"
    : null

  function handleSend() {
    const content = value.trim()
    if (!content) return

    // Optimistic: show message instantly
    const optimisticId = `optimistic-${Date.now()}`
    onOptimisticMessage?.({
      id: optimisticId,
      channelId,
      userId: currentUserId,
      content,
      type: "text",
      createdAt: Date.now(),
      editedAt: null,
      deletedAt: null,
      replyCount: 0,
      parentId: replyTo?.id ?? null,
    })

    setValue("")
    onClearReply?.()
    textareaRef.current?.focus()

    // Persist
    startTransition(async () => {
      await actionSendMessage({
        channelId,
        content,
        ...(replyTo ? { parentId: replyTo.id } : {}),
      })
    })
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="shrink-0 border-t border-border bg-bg px-4 py-3">
      {/* Reply indicator */}
      {replyTo && (
        <div className="flex items-center gap-2 mb-2 text-xs text-fg-subtle">
          <span>
            Respondiendo a <span className="font-medium text-fg">{replyAuthor}</span>
          </span>
          <button
            type="button"
            onClick={onClearReply}
            className="hover:text-fg transition-colors"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </div>
      )}

      <div
        className={cn(
          "border border-border rounded-xl bg-bg transition-colors focus-within:border-accent",
          isPending && "opacity-70",
        )}
        style={{ padding: "10px 14px" }}
      >
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Escribí un mensaje..."
          disabled={isPending}
          rows={1}
          className="w-full resize-none bg-transparent text-fg text-sm placeholder:text-fg-subtle outline-none disabled:opacity-50"
          style={{ minHeight: "24px", maxHeight: "200px", lineHeight: "1.5" }}
        />
        <div className="flex items-center justify-end" style={{ marginTop: "6px" }}>
          <button
            type="button"
            onClick={handleSend}
            disabled={!value.trim() || isPending}
            className={cn(
              "flex items-center justify-center rounded-lg transition-colors",
              value.trim()
                ? "bg-accent text-white hover:bg-accent/90"
                : "bg-surface-2 text-fg-subtle",
            )}
            style={{ width: "34px", height: "34px" }}
            title="Enviar (Enter)"
          >
            <Send size={16} />
          </button>
        </div>
      </div>
    </div>
  )
}
