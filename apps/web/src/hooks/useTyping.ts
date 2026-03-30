"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { wsClient } from "@/lib/ws/client"
import type { ServerMessage } from "@rag-saldivia/shared"

type TypingUser = { userId: number; displayName: string }

const TYPING_TIMEOUT_MS = 3000 // Auto-expire typing indicator after 3s
const TYPING_DEBOUNCE_MS = 1000 // Don't send typing_start more than once per second

/**
 * Typing indicator hook — tracks who's typing in a channel.
 *
 * Usage:
 *   const { typingUsers, startTyping, stopTyping } = useTyping(channelId)
 */
export function useTyping(channelId: string | null) {
  const [typingUsers, setTypingUsers] = useState<TypingUser[]>([])
  const timers = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map())
  const lastSent = useRef(0)

  useEffect(() => {
    if (!channelId) return
    const timerMap = timers.current

    const unsub = wsClient.on((msg: ServerMessage) => {
      if (msg.type !== "typing" || msg.channelId !== channelId) return
      if (msg.userId === wsClient.userId) return

      const { userId, displayName } = msg

      setTypingUsers((prev) => {
        const existing = prev.find((t) => t.userId === userId)
        if (!existing) return [...prev, { userId, displayName }]
        return prev
      })

      const existing = timerMap.get(userId)
      if (existing) clearTimeout(existing)

      timerMap.set(
        userId,
        setTimeout(() => {
          setTypingUsers((prev) => prev.filter((t) => t.userId !== userId))
          timerMap.delete(userId)
        }, TYPING_TIMEOUT_MS)
      )
    })

    return () => {
      unsub()
      for (const timer of timerMap.values()) clearTimeout(timer)
      timerMap.clear()
      setTypingUsers([])
    }
  }, [channelId])

  const startTyping = useCallback(() => {
    if (!channelId) return
    const now = Date.now()
    if (now - lastSent.current < TYPING_DEBOUNCE_MS) return
    lastSent.current = now
    wsClient.send({ type: "typing_start", channelId })
  }, [channelId])

  const stopTyping = useCallback(() => {
    if (!channelId) return
    wsClient.send({ type: "typing_stop", channelId })
  }, [channelId])

  return { typingUsers, startTyping, stopTyping }
}
