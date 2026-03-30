/**
 * ChannelView — client component that combines MessageList, Composer,
 * TypingIndicator, and ThreadPanel for a channel.
 */
"use client"

import { useState, useCallback } from "react"
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
  const [optimisticMessages, setOptimisticMessages] = useState<MessageData[]>([])

  // TODO: Wire up useTyping when WS token is available
  const typingUsers: Array<{ userId: number; displayName: string }> = []

  const handleOpenThread = useCallback((msg: MessageData) => {
    setThreadMessage(msg)
  }, [])

  const handleReply = useCallback((msg: MessageData) => {
    setReplyTo({ id: msg.id, userId: msg.userId, content: msg.content })
  }, [])

  const handleOptimisticMessage = useCallback((msg: MessageData) => {
    setOptimisticMessages((prev) => [...prev, msg])
  }, [])

  return (
    <div className="flex flex-1 min-h-0">
      {/* Main message area */}
      <div className="flex-1 flex flex-col min-w-0">
        <MessageList
          channelId={channelId}
          initialMessages={[...initialMessages, ...optimisticMessages]}
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
          replies={[]}
          members={members}
          currentUserId={currentUserId}
          onClose={() => setThreadMessage(null)}
        />
      )}
    </div>
  )
}
