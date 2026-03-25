"use client"

import { useEffect, useRef, useState } from "react"
import { toast } from "sonner"

const STORAGE_KEY = "seen_notification_ids"
const POLL_INTERVAL_MS = 30_000

type Notification = {
  id: string
  type: string
  ts: number
  payload: Record<string, unknown>
}

function getSeenIds(): Set<string> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    return new Set(raw ? JSON.parse(raw) : [])
  } catch {
    return new Set()
  }
}

function markSeen(ids: string[]) {
  try {
    const existing = getSeenIds()
    for (const id of ids) existing.add(id)
    // Mantener solo los últimos 200 IDs vistos para no crecer indefinidamente
    const trimmed = Array.from(existing).slice(-200)
    localStorage.setItem(STORAGE_KEY, JSON.stringify(trimmed))
  } catch {
    // ignorar errores de localStorage
  }
}

function toastForNotification(n: Notification) {
  const labels: Record<string, string> = {
    "ingestion.completed": "✅ Ingesta completada",
    "ingestion.error": "❌ Error en ingesta",
    "user.created": "👤 Nuevo usuario registrado",
  }
  const label = labels[n.type] ?? n.type
  const detail = (n.payload as { filename?: string; name?: string }).filename
    ?? (n.payload as { name?: string }).name
    ?? ""

  if (n.type === "ingestion.error") {
    toast.error(label, { description: detail })
  } else {
    toast.success(label, { description: detail })
  }
}

export function useNotifications() {
  const [unreadCount, setUnreadCount] = useState(0)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  async function poll() {
    try {
      const res = await fetch("/api/notifications")
      if (!res.ok) return
      const data = await res.json() as { notifications: Notification[] }
      const seen = getSeenIds()
      const unseen = data.notifications.filter((n) => !seen.has(n.id))

      setUnreadCount(unseen.length)

      for (const n of unseen) {
        toastForNotification(n)
      }

      if (unseen.length > 0) {
        markSeen(unseen.map((n) => n.id))
        setUnreadCount(0)
      }
    } catch {
      // silencioso — no interrumpir la UI por errores de polling
    }
  }

  useEffect(() => {
    poll()
    timerRef.current = setInterval(poll, POLL_INTERVAL_MS)
    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [])

  return { unreadCount }
}
