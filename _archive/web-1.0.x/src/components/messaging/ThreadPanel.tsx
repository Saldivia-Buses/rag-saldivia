/**
 * Thread panel — side panel showing replies to a parent message.
 * Has its own composer for posting thread replies.
 */
"use client"

import { useState, useEffect, useTransition, useRef } from "react"
import { X } from "lucide-react"
import { MessageItem } from "./MessageItem"
import { useAutoResize } from "@/hooks/useAutoResize"
import { actionSendMessage } from "@/app/actions/messaging"
import { Send } from "lucide-react"
import { cn } from "@/lib/utils"

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

export function ThreadPanel({
  channelId,
  parentMessage,
  replies,
  members,
  currentUserId,
  onClose,
}: {
  channelId: string
  parentMessage: MessageData
  replies: MessageData[]
  members: MemberInfo[]
  currentUserId: number
  onClose: () => void
}) {
  const [value, setValue] = useState("")
  const [isPending, startTransition] = useTransition()
  const [localReplies, setLocalReplies] = useState<MessageData[]>(replies)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  useAutoResize(textareaRef, value)

  useEffect(() => {
    setLocalReplies(replies)
  }, [replies])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [localReplies.length])

  function handleSend() {
    const content = value.trim()
    if (!content) return

    // Optimistic
    const optimistic: MessageData = {
      id: `optimistic-thread-${Date.now()}`,
      userId: currentUserId,
      content,
      type: "text",
      createdAt: Date.now(),
      editedAt: null,
      deletedAt: null,
      replyCount: 0,
      parentId: parentMessage.id,
    }
    setLocalReplies((prev) => [...prev, optimistic])
    setValue("")
    textareaRef.current?.focus()

    startTransition(async () => {
      await actionSendMessage({
        channelId,
        content,
        parentId: parentMessage.id,
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
    <div className="shrink-0 border-l border-border bg-surface flex flex-col h-full" style={{ width: "340px" }}>
      {/* Header */}
      <div className="border-b border-border flex items-center justify-between" style={{ padding: "14px 16px", minHeight: "52px" }}>
        <h3 className="text-sm font-semibold text-fg">Hilo</h3>
        <button
          type="button"
          onClick={onClose}
          className="text-fg-subtle hover:text-fg transition-colors"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto">
        {/* Parent message — no thread link (we're already in the thread) */}
        <div className="border-b border-border" style={{ paddingBottom: "8px" }}>
          <MessageItem
            message={{ ...parentMessage, replyCount: 0 }}
            members={members}
            showHeader
            currentUserId={currentUserId}
          />
        </div>

        {/* Reply count */}
        <div className="px-4 py-2 flex items-center gap-2">
          <div className="flex-1 h-px bg-border" />
          <span className="text-xs text-fg-subtle">
            {localReplies.length} {localReplies.length === 1 ? "respuesta" : "respuestas"}
          </span>
          <div className="flex-1 h-px bg-border" />
        </div>

        {/* Replies */}
        {localReplies.map((reply, i) => {
          const prev = i > 0 ? localReplies[i - 1] : undefined
          const showHeader = !prev || prev.userId !== reply.userId || reply.createdAt - prev.createdAt > 300_000

          return (
            <MessageItem
              key={reply.id}
              message={reply}
              members={members}
              showHeader={showHeader}
              currentUserId={currentUserId}
            />
          )
        })}

        <div ref={bottomRef} />
      </div>

      {/* Thread composer */}
      <div className="shrink-0 border-t border-border" style={{ padding: "12px 16px" }}>
        <div
          className={cn(
            "border border-border bg-bg transition-colors focus-within:border-accent",
            isPending && "opacity-70",
          )}
          style={{ padding: "10px 12px", borderRadius: "12px" }}
        >
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Responder en hilo..."
            disabled={isPending}
            rows={1}
            className="w-full resize-none bg-transparent text-fg text-sm placeholder:text-fg-subtle outline-none disabled:opacity-50"
            style={{ minHeight: "22px", maxHeight: "120px", lineHeight: "1.5" }}
          />
          <div className="flex items-center justify-end" style={{ marginTop: "6px" }}>
            <button
              type="button"
              onClick={handleSend}
              disabled={!value.trim() || isPending}
              className={cn(
                "flex items-center justify-center transition-colors",
                value.trim()
                  ? "bg-accent text-accent-fg hover:bg-accent/90"
                  : "text-fg-subtle",
              )}
              style={{ width: "30px", height: "30px", borderRadius: "8px" }}
              title="Enviar (Enter)"
            >
              <Send size={15} />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
