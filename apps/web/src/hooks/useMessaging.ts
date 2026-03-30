"use client"

import { useState, useEffect, useCallback, useRef } from "react"
import { wsClient } from "@/lib/ws/client"
import type { ServerMessage, MsgMessage } from "@rag-saldivia/shared"

/**
 * Core messaging hook — connects to WS, subscribes to channel, receives messages.
 *
 * Usage:
 *   const { connected, messages, sendMessage, subscribe } = useMessaging(authToken)
 */
export function useMessaging(token: string | null) {
  const [connected, setConnected] = useState(false)
  const [messages, setMessages] = useState<MsgMessage[]>([])
  const subscribedChannels = useRef(new Set<string>())

  // Connect/disconnect on token change
  useEffect(() => {
    if (!token) {
      wsClient.disconnect()
      setConnected(false)
      return
    }
    wsClient.connect(token)

    const unsub = wsClient.on((msg: ServerMessage) => {
      switch (msg.type) {
        case "auth_ok":
          setConnected(true)
          break
        case "auth_error":
          setConnected(false)
          break
        case "message_new":
          setMessages((prev) => [...prev, msg.message])
          break
        case "message_updated":
          setMessages((prev) =>
            prev.map((m) => (m.id === msg.message.id ? msg.message : m))
          )
          break
        case "message_deleted":
          setMessages((prev) =>
            prev.map((m) =>
              m.id === msg.messageId ? { ...m, deletedAt: Date.now(), content: "" } : m
            )
          )
          break
      }
    })

    return () => {
      unsub()
    }
  }, [token])

  const subscribe = useCallback((channelId: string) => {
    if (subscribedChannels.current.has(channelId)) return
    subscribedChannels.current.add(channelId)
    wsClient.send({ type: "subscribe", channelId })
  }, [])

  const unsubscribe = useCallback((channelId: string) => {
    subscribedChannels.current.delete(channelId)
    wsClient.send({ type: "unsubscribe", channelId })
  }, [])

  /** Clear local messages (e.g. when switching channels). */
  const clearMessages = useCallback(() => setMessages([]), [])

  /** Set initial messages loaded from API. */
  const setInitialMessages = useCallback((msgs: MsgMessage[]) => setMessages(msgs), [])

  return {
    connected,
    messages,
    subscribe,
    unsubscribe,
    clearMessages,
    setInitialMessages,
  }
}
