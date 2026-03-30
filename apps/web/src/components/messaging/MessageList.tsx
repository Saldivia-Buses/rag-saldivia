/**
 * Message list — renders messages with date separators and scroll-to-bottom.
 * Uses a simple scroll container (TanStack Virtual deferred to phase 6+).
 */
"use client"

import { useRef, useEffect, useState } from "react"
import { ArrowDown } from "lucide-react"
import { cn } from "@/lib/utils"
import { MessageItem, DateSeparator } from "./MessageItem"

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

function isSameDay(a: number, b: number): boolean {
  const da = new Date(a)
  const db = new Date(b)
  return da.toDateString() === db.toDateString()
}

function shouldShowHeader(msg: MessageData, prev: MessageData | undefined): boolean {
  if (!prev) return true
  if (msg.userId !== prev.userId) return true
  if (msg.type === "system") return true
  // Show header if more than 5 minutes apart
  if (msg.createdAt - prev.createdAt > 5 * 60 * 1000) return true
  return false
}

export function MessageList({
  channelId: _channelId,
  initialMessages,
  currentUserId,
  members,
  onOpenThread,
  onReply,
}: {
  channelId: string
  initialMessages: MessageData[]
  currentUserId: number
  members: MemberInfo[]
  onOpenThread?: (msg: MessageData) => void
  onReply?: (msg: MessageData) => void
}) {
  const scrollRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const [showScrollBtn, setShowScrollBtn] = useState(false)

  // Use initialMessages directly — they update on revalidation and include optimistic ones
  const messages = initialMessages

  // Scroll to bottom on mount and new messages
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "instant" })
  }, []) // Only on mount

  useEffect(() => {
    if (!showScrollBtn) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" })
    }
  }, [messages.length]) // eslint-disable-line react-hooks/exhaustive-deps

  // Detect scroll position
  function handleScroll() {
    const el = scrollRef.current
    if (!el) return
    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight
    setShowScrollBtn(distanceFromBottom > 200)
  }

  function scrollToBottom() {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  return (
    <div className="flex-1 relative min-h-0">
      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="absolute inset-0 overflow-y-auto"
      >
        {messages.length === 0 && (
          <div className="flex items-center justify-center h-full">
            <p className="text-sm text-fg-subtle">
              No hay mensajes todavía. ¡Sé el primero!
            </p>
          </div>
        )}

        <div className="pb-4">
          {messages.map((msg, i) => {
            const prev = i > 0 ? messages[i - 1] : undefined
            const showDate = !prev || !isSameDay(msg.createdAt, prev.createdAt)
            const showHeader = shouldShowHeader(msg, prev)

            return (
              <div key={msg.id}>
                {showDate && <DateSeparator timestamp={msg.createdAt} />}
                <MessageItem
                  message={msg}
                  members={members}
                  showHeader={showHeader}
                  currentUserId={currentUserId}
                  {...(onOpenThread ? { onOpenThread } : {})}
                  {...(onReply ? { onReply } : {})}
                />
              </div>
            )
          })}
        </div>

        <div ref={bottomRef} />
      </div>

      {/* Scroll to bottom button */}
      {showScrollBtn && (
        <button
          onClick={scrollToBottom}
          className={cn(
            "absolute bottom-4 left-1/2 -translate-x-1/2",
            "flex items-center gap-1.5 px-3 py-1.5 rounded-full",
            "bg-surface border border-border shadow-md",
            "text-xs text-fg-muted hover:text-fg transition-colors",
          )}
        >
          <ArrowDown className="h-3.5 w-3.5" />
          Nuevos mensajes
        </button>
      )}
    </div>
  )
}
