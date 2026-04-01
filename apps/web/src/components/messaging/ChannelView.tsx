/**
 * ChannelView — client component that combines MessageList, Composer,
 * TypingIndicator, and ThreadPanel for a channel.
 */
"use client"

import { useState, useCallback, useEffect, useRef, useTransition } from "react"
import { MessageList } from "./MessageList"
import { MessageComposer } from "./MessageComposer"
import { TypingIndicator } from "./TypingIndicator"
import { ThreadPanel } from "./ThreadPanel"

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

export function ChannelView({
  channelId,
  initialMessages,
  currentUserId,
  members,
}: {
  channelId: string
  initialMessages: MessageData[]
  currentUserId: number
  members: MemberInfo[]
}) {
  const [replyTo, setReplyTo] = useState<{ id: string; userId: number; content: string } | null>(null)
  const [threadMessage, setThreadMessage] = useState<MessageData | null>(null)
  const [threadReplies, setThreadReplies] = useState<MessageData[]>([])
  const [optimisticMessages, setOptimisticMessages] = useState<MessageData[]>([])
  const [_isLoadingThread, startThreadTransition] = useTransition()

  // Track the previous initialMessages length to detect server re-renders
  const prevInitialLengthRef = useRef(initialMessages.length)
  useEffect(() => {
    if (initialMessages.length !== prevInitialLengthRef.current) {
      // Server re-rendered with new data — clear optimistic messages
      // Filter out any optimistic message whose content matches a real message
      setOptimisticMessages((prev) =>
        prev.filter((opt) =>
          !initialMessages.some((real) =>
            real.content === opt.content && real.userId === opt.userId
          )
        )
      )
      prevInitialLengthRef.current = initialMessages.length
    }
  }, [initialMessages])

  // TODO: Wire up useTyping when WS token is available
  const typingUsers: Array<{ userId: number; displayName: string }> = []

  const handleOpenThread = useCallback((msg: MessageData) => {
    setThreadMessage(msg)
    // Load thread replies from server
    startThreadTransition(async () => {
      try {
        const res = await fetch(`/api/messaging/channels/${channelId}/threads/${msg.id}`)
        if (res.ok) {
          const data = await res.json()
          if (data.ok && Array.isArray(data.data)) {
            setThreadReplies(data.data)
          }
        }
      } catch {
        // Thread replies will show as empty — user can still post
      }
    })
  }, [channelId, startThreadTransition])

  const handleReply = useCallback((msg: MessageData) => {
    setReplyTo({ id: msg.id, userId: msg.userId, content: msg.content })
  }, [])

  const handleOptimisticMessage = useCallback((msg: MessageData) => {
    setOptimisticMessages((prev) => [...prev, msg])
  }, [])

  // Merge: server messages + optimistic (that haven't been confirmed yet)
  const allMessages = [...initialMessages, ...optimisticMessages]

  return (
    <div className="flex flex-1 min-h-0">
      {/* Main message area */}
      <div className="flex-1 flex flex-col min-w-0">
        <MessageList
          channelId={channelId}
          initialMessages={allMessages}
          currentUserId={currentUserId}
          members={members}
          onOpenThread={handleOpenThread}
          onReply={handleReply}
        />
        <TypingIndicator typingUsers={typingUsers} />
        <MessageComposer
          channelId={channelId}
          currentUserId={currentUserId}
          members={members}
          replyTo={replyTo}
          onClearReply={() => setReplyTo(null)}
          onOptimisticMessage={handleOptimisticMessage}
        />
      </div>

      {/* Thread panel (side) */}
      {threadMessage && (
        <ThreadPanel
          channelId={channelId}
          parentMessage={threadMessage}
          replies={threadReplies}
          members={members}
          currentUserId={currentUserId}
          onClose={() => { setThreadMessage(null); setThreadReplies([]) }}
        />
      )}
    </div>
  )
}
