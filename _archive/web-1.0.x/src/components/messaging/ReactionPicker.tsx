/**
 * Reaction picker — popover with native emoji grid, no external deps.
 */
"use client"

import { useState } from "react"
import { SmilePlus } from "lucide-react"
import { cn } from "@/lib/utils"

const FREQUENT_EMOJIS = [
  "👍", "👎", "❤️", "😂", "😮", "😢", "🎉", "🔥",
  "👀", "🙌", "💯", "✅", "❌", "🚀", "💡", "🤔",
  "👏", "😍", "🙏", "💪", "⭐", "📌", "🎯", "💬",
  "✨", "🤝", "📎", "🔗", "⚡", "🏆",
]

export function ReactionPicker({
  onSelect,
  className,
}: {
  onSelect: (emoji: string) => void
  className?: string
}) {
  const [open, setOpen] = useState(false)

  function handleSelect(emoji: string) {
    onSelect(emoji)
    setOpen(false)
  }

  return (
    <div className={cn("relative", className)}>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex items-center justify-center rounded-md text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
        style={{ width: "28px", height: "28px" }}
        title="Reaccionar"
      >
        <SmilePlus size={16} />
      </button>

      {open && (
        <>
          {/* Backdrop */}
          <div
            className="fixed inset-0 z-40"
            onClick={() => setOpen(false)}
          />
          {/* Picker */}
          <div
            className="absolute bottom-full right-0 mb-1 z-50 bg-surface border border-border rounded-lg shadow-lg p-2"
            style={{ width: "232px" }}
          >
            <div className="grid grid-cols-6 gap-0.5">
              {FREQUENT_EMOJIS.map((emoji) => (
                <button
                  key={emoji}
                  type="button"
                  onClick={() => handleSelect(emoji)}
                  className="flex items-center justify-center rounded-md hover:bg-surface-2 transition-colors"
                  style={{ width: "34px", height: "34px", fontSize: "18px" }}
                >
                  {emoji}
                </button>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  )
}

/** Compact reaction chips displayed under a message. */
export function ReactionChips({
  reactions,
  currentUserId,
  onToggle,
}: {
  reactions: Array<{ emoji: string; userId: number; count: number }>
  currentUserId: number
  onToggle: (emoji: string) => void
}) {
  if (reactions.length === 0) return null

  return (
    <div className="flex flex-wrap gap-1 mt-1">
      {reactions.map((r) => {
        const isOwn = r.userId === currentUserId
        return (
          <button
            key={r.emoji}
            type="button"
            onClick={() => onToggle(r.emoji)}
            className={cn(
              "inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-xs border transition-colors",
              isOwn
                ? "border-accent bg-accent-subtle text-accent"
                : "border-border bg-surface hover:bg-surface-2 text-fg-muted",
            )}
          >
            <span>{r.emoji}</span>
            <span>{r.count}</span>
          </button>
        )
      })}
    </div>
  )
}
