"use client"

import { useState, useEffect, useCallback } from "react"
import { wsClient } from "@/lib/ws/client"
import type { ServerMessage } from "@rag-saldivia/shared"

type PresenceStatus = "online" | "away" | "offline"

/**
 * Presence hook — tracks online status of users via WebSocket.
 *
 * Usage:
 *   const { presenceMap, setStatus } = usePresence()
 *   presenceMap[userId] // "online" | "away" | "offline"
 */
export function usePresence() {
  const [presenceMap, setPresenceMap] = useState<Record<number, PresenceStatus>>({})

  useEffect(() => {
    const unsub = wsClient.on((msg: ServerMessage) => {
      if (msg.type === "presence_update") {
        setPresenceMap((prev) => ({
          ...prev,
          [msg.userId]: msg.status,
        }))
      }
    })
    return unsub
  }, [])

  const setStatus = useCallback((status: "online" | "away") => {
    wsClient.send({ type: "presence", status })
  }, [])

  return { presenceMap, setStatus }
}
