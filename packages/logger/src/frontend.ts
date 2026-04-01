/**
 * Logger del cliente (browser).
 *
 * Captura acciones del usuario y errores en el browser.
 * Usa batching para no saturar el servidor: acumula eventos y los envía
 * cada 5 segundos o cuando el batch llega a 20 eventos.
 *
 * Implementación completa en Fase 5.
 */

import type { EventType } from "@rag-saldivia/shared"
import { LOGGER_BATCH_SIZE, LOGGER_FLUSH_INTERVAL_MS } from "@rag-saldivia/config"

type ClientEvent = {
  type: EventType
  payload?: Record<string, unknown>
  ts: number
}

class ClientLogger {
  private batch: ClientEvent[] = []
  private timer: ReturnType<typeof setTimeout> | null = null
  private readonly BATCH_SIZE = LOGGER_BATCH_SIZE
  private readonly FLUSH_INTERVAL_MS = LOGGER_FLUSH_INTERVAL_MS
  private endpoint = "/api/log"

  private schedule() {
    if (this.timer) return
    this.timer = setTimeout(() => {
      this.flush()
    }, this.FLUSH_INTERVAL_MS)
  }

  private async flush() {
    this.timer = null
    if (this.batch.length === 0) return

    const toSend = this.batch.splice(0)

    try {
      await fetch(this.endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ events: toSend }),
        keepalive: true, // Permite enviar incluso al cerrar la pestaña
      })
    } catch {
      // Silencioso — el logger de frontend no debe romper la app
    }
  }

  private push(type: EventType, payload?: Record<string, unknown>) {
    this.batch.push({ type, payload: payload ?? {}, ts: Date.now() })
    if (this.batch.length >= this.BATCH_SIZE) {
      this.flush()
    } else {
      this.schedule()
    }
  }

  // ── API pública ──────────────────────────────────────────────────────────

  action(type: EventType, payload?: Record<string, unknown>) {
    this.push(type, payload)
  }

  navigation(path: string, from?: string) {
    this.push("client.navigation", { path, from })
  }

  error(error: Error, context?: Record<string, unknown>) {
    this.push("client.error", {
      message: error.message,
      stack: error.stack?.slice(0, 500), // Limitar tamaño
      ...context,
    })
    // Los errores se envían inmediatamente
    this.flush()
  }

  // Llamar al cerrar sesión o cuando se sepa que el usuario se va
  forceFlush() {
    return this.flush()
  }
}

export const clientLog = new ClientLogger()

// Captura global de errores no manejados
if (typeof window !== "undefined") {
  window.addEventListener("error", (e) => {
    clientLog.error(new Error(e.message), {
      filename: e.filename,
      lineno: e.lineno,
      colno: e.colno,
    })
  })

  window.addEventListener("unhandledrejection", (e) => {
    clientLog.error(new Error(String(e.reason)), { type: "unhandledRejection" })
  })
}
