"use client"

/**
 * Hook de notificaciones via SSE (Server-Sent Events).
 * F8.28 — Elimina localStorage y polling cada 30s.
 *
 * El servidor publica notificaciones via Redis Pub/Sub.
 * Este hook se suscribe al stream SSE y muestra toasts en tiempo real.
 * Los IDs vistos se persisten en Redis Sorted Set (server-side).
 */

import { useEffect, useState } from "react"
import { toast } from "sonner"

type Notification = {
  id: string
  type: string
  ts: number
  payload: Record<string, unknown>
}

function toastForNotification(n: Notification) {
  const labels: Record<string, string> = {
    "ingestion.completed": "✅ Ingesta completada",
    "ingestion.error": "❌ Error en ingesta",
    "user.created": "👤 Nuevo usuario registrado",
    "proactive.docs_available": "📄 Documentos relevantes disponibles",
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

  useEffect(() => {
    const es = new EventSource("/api/notifications/stream")

    es.onmessage = (e) => {
      try {
        const notif = JSON.parse(e.data as string) as Notification
        toastForNotification(notif)
        setUnreadCount((c) => c + 1)
        // Resetear contador después de mostrar el toast
        setTimeout(() => setUnreadCount((c) => Math.max(0, c - 1)), 5000)
      } catch {
        // Ignorar mensajes malformados
      }
    }

    es.onerror = () => {
      // EventSource reconecta automáticamente — no hacer nada
    }

    return () => {
      es.close()
    }
  }, [])

  return { unreadCount }
}
