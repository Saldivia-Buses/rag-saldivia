"use client"

import { Plus, Square } from "lucide-react"
import type { RefObject } from "react"

const ICON_BTN = "flex items-center justify-center rounded-lg transition-colors"
const BTN_SIZE = { width: "42px", height: "42px" } as const

function SendIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="22" y1="2" x2="11" y2="13" />
      <polygon points="22 2 15 22 11 13 2 9 22 2" />
    </svg>
  )
}

export function ChatInputBar({
  value,
  onChange,
  onKeyDown,
  onSend,
  onStop,
  textareaRef,
  placeholder,
  collection,
  collectionSlot,
  isStreaming,
  disabled,
}: {
  value: string
  onChange: (value: string) => void
  onKeyDown: (e: React.KeyboardEvent) => void
  onSend: () => void
  onStop?: () => void
  textareaRef?: RefObject<HTMLTextAreaElement | null>
  placeholder: string
  collection: string
  collectionSlot?: React.ReactNode
  isStreaming?: boolean
  disabled?: boolean
}) {
  return (
    <div
      className="border border-border rounded-2xl bg-bg transition-colors focus-within:border-accent"
      style={{ padding: "12px 16px" }}
    >
      <textarea
        ref={textareaRef}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        rows={1}
        className="w-full resize-none bg-transparent text-fg text-sm placeholder:text-fg-subtle outline-none disabled:opacity-50"
        style={{ minHeight: "24px", maxHeight: "200px", lineHeight: "1.5" }}
      />
      <div className="flex items-center justify-between" style={{ marginTop: "8px" }}>
        <button
          type="button"
          className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface`}
          style={BTN_SIZE}
          title="Adjuntar"
        >
          <Plus size={18} />
        </button>
        <div className="flex items-center" style={{ gap: "8px" }}>
          {collectionSlot ?? <span className="text-xs text-fg-subtle">{collection}</span>}
          {isStreaming && onStop ? (
            <button
              onClick={onStop}
              className="flex items-center justify-center rounded-full border border-border text-fg-muted hover:text-fg hover:border-fg-subtle transition-colors"
              style={BTN_SIZE}
              title="Detener"
            >
              <Square size={12} fill="currentColor" />
            </button>
          ) : (
            <button
              onClick={onSend}
              disabled={!value.trim()}
              className="flex items-center justify-center rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity hover:opacity-90"
              style={BTN_SIZE}
            >
              <SendIcon />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
