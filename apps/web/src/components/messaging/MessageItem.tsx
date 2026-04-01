/**
 * Single message — avatar, name, timestamp, content.
 * Shows (editado) badge and soft-delete placeholder.
 */
"use client"

import { useMemo } from "react"
import { MessageSquare, SmilePlus } from "lucide-react"
import { cn } from "@/lib/utils"
import { ReactionChips } from "./ReactionPicker"
import { renderWithMentions } from "./MentionSuggestions"

type MemberInfo = { id: number; name: string; email: string }

type MessageData = {
  id: string
  userId: number
  content: string
  type: string
  createdAt: number
  editedAt: number | null
  deletedAt: number | null
  replyCount: number
  parentId: string | null
}

function formatTime(ts: number): string {
  const d = new Date(ts)
  return d.toLocaleTimeString("es-AR", { hour: "2-digit", minute: "2-digit" })
}

function formatDate(ts: number): string {
  const d = new Date(ts)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (d.toDateString() === today.toDateString()) return "Hoy"
  if (d.toDateString() === yesterday.toDateString()) return "Ayer"
  return d.toLocaleDateString("es-AR", { day: "numeric", month: "long" })
}

function UserAvatar({ name }: { name: string }) {
  const initials = name
    .split(" ")
    .map((w) => w[0])
    .slice(0, 2)
    .join("")
    .toUpperCase()

  return (
    <div
      className="rounded-full bg-accent text-accent-fg flex items-center justify-center text-xs font-semibold shrink-0"
      style={{ width: "36px", height: "36px" }}
    >
      {initials || "?"}
    </div>
  )
}

export function MessageItem({
  message,
  members,
  showHeader,
  currentUserId,
  reactions,
  onReply,
  onReact,
  onToggleReaction,
  onOpenThread,
}: {
  message: MessageData
  members: MemberInfo[]
  showHeader: boolean
  currentUserId: number
  reactions?: Array<{ emoji: string; userId: number; count: number }>
  onReply?: (msg: MessageData) => void
  onReact?: (msgId: string) => void
  onToggleReaction?: (msgId: string, emoji: string) => void
  onOpenThread?: (msg: MessageData) => void
}) {
  const author = useMemo(
    () => members.find((m) => m.id === message.userId),
    [members, message.userId],
  )

  if (message.deletedAt) {
    return (
      <div className="px-4 py-1">
        <p className="text-xs text-fg-subtle italic">Mensaje eliminado</p>
      </div>
    )
  }

  if (message.type === "system") {
    return (
      <div className="px-4 py-1 text-center">
        <p className="text-xs text-fg-subtle">{message.content}</p>
      </div>
    )
  }

  return (
    <div
      className={cn(
        "group relative flex gap-3 hover:bg-surface transition-colors",
      )}
      style={{
        padding: showHeader ? "8px 20px 2px" : "0 20px",
      }}
    >
      {showHeader ? (
        <UserAvatar name={author?.name ?? "?"} />
      ) : (
        <div className="shrink-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity" style={{ width: "36px" }}>
          <span className="text-[10px] text-fg-subtle" style={{ lineHeight: 1 }}>
            {formatTime(message.createdAt).replace(/\s/g, "")}
          </span>
        </div>
      )}

      <div className="flex-1 min-w-0">
        {showHeader && (
          <div className="flex items-baseline gap-2">
            <span className="text-sm font-semibold text-fg">
              {author?.name ?? "Usuario"}
            </span>
            <span className="text-xs text-fg-subtle">
              {formatTime(message.createdAt)}
            </span>
          </div>
        )}
        <p className="text-sm text-fg whitespace-pre-wrap break-words" style={{ lineHeight: "1.4" }}>
          {renderWithMentions(message.content, members)}
          {message.editedAt && (
            <span className="text-xs text-fg-subtle ml-1">(editado)</span>
          )}
        </p>

        {/* Reaction chips */}
        {reactions && reactions.length > 0 && onToggleReaction && (
          <ReactionChips
            reactions={reactions}
            currentUserId={currentUserId}
            onToggle={(emoji) => onToggleReaction(message.id, emoji)}
          />
        )}

        {/* Thread reply count */}
        {message.replyCount > 0 && (
          <button
            onClick={() => onOpenThread?.(message)}
            className="text-xs text-accent hover:underline flex items-center gap-1"
            style={{ marginTop: "4px" }}
          >
            <MessageSquare size={12} />
            {message.replyCount} {message.replyCount === 1 ? "respuesta" : "respuestas"}
          </button>
        )}
      </div>

      {/* Hover actions */}
      {(onReply || onReact) && !message.deletedAt && message.type !== "system" && (
        <div
          className="absolute right-5 opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5 bg-surface border border-border shadow-sm"
          style={{ top: "-14px", borderRadius: "8px", padding: "3px" }}
        >
          {onReact && (
            <button
              type="button"
              onClick={() => onReact(message.id)}
              className="flex items-center justify-center rounded text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
              style={{ width: "26px", height: "26px" }}
              title="Reaccionar"
            >
              <SmilePlus size={14} />
            </button>
          )}
          {onReply && !message.parentId && (
            <button
              type="button"
              onClick={() => onReply(message)}
              className="flex items-center justify-center rounded text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
              style={{ width: "26px", height: "26px" }}
              title="Responder en hilo"
            >
              <MessageSquare size={14} />
            </button>
          )}
        </div>
      )}
    </div>
  )
}

export function DateSeparator({ timestamp }: { timestamp: number }) {
  return (
    <div className="flex items-center gap-3" style={{ padding: "16px 20px 8px" }}>
      <div className="flex-1 h-px bg-border" />
      <span
        className="text-xs text-fg-subtle font-medium"
        style={{
          padding: "2px 10px",
          borderRadius: "999px",
          border: "1px solid var(--border)",
          background: "var(--surface)",
        }}
      >
        {formatDate(timestamp)}
      </span>
      <div className="flex-1 h-px bg-border" />
    </div>
  )
}
