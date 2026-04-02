/**
 * error-recovery.ts — Contextual error recovery for end users.
 *
 * Maps error codes/statuses to user-facing recovery suggestions in Spanish.
 * Each recovery includes a title, description, suggestion, and actionable
 * buttons the UI can render (retry, navigate, report).
 *
 * Used by: ErrorRecovery component (ui/error-recovery.tsx)
 * Depends on: nothing (pure functions)
 */

export type ErrorAction = {
  label: string
  type: "retry" | "navigate" | "dismiss" | "report"
  href?: string
}

export type UserErrorRecovery = {
  title: string
  description: string
  suggestion: string
  actions: ErrorAction[]
  icon: "unavailable" | "timeout" | "forbidden" | "rate-limit" | "upstream" | "auth" | "not-found" | "generic"
}

export type ErrorInput = {
  message?: string | undefined
  status?: number | undefined
  code?: string | undefined
  details?: Record<string, unknown> | undefined
}

/**
 * Maps an error to a user-facing recovery with contextual actions.
 *
 * Priority: code > status > message pattern > generic fallback.
 */
export function getErrorRecovery(error: ErrorInput): UserErrorRecovery {
  const { message = "", status, code, details } = error

  // --- By explicit code (from RagError) ---

  if (code === "UNAVAILABLE" || message.includes("ECONNREFUSED")) {
    return {
      title: "Servidor no disponible",
      description: "El sistema de búsqueda no está respondiendo.",
      suggestion: "Podés reintentar en unos minutos. Si el problema persiste, contactá al administrador.",
      actions: [{ label: "Reintentar", type: "retry" }],
      icon: "unavailable",
    }
  }

  if (code === "TIMEOUT") {
    return {
      title: "La consulta tardó demasiado",
      description: "El servidor no respondió a tiempo.",
      suggestion: "Intentá con una pregunta más corta o específica.",
      actions: [{ label: "Reintentar", type: "retry" }],
      icon: "timeout",
    }
  }

  if (code === "RATE_LIMITED" || status === 429) {
    const retryAfterMs = (details?.retryAfterMs as number) ?? 3600_000
    const minutes = Math.ceil(retryAfterMs / 60_000)
    const maxCount = (details?.maxCount as number) ?? null
    return {
      title: "Límite de consultas alcanzado",
      description: maxCount
        ? `Alcanzaste el límite de ${maxCount} consultas por hora.`
        : "Alcanzaste el límite de consultas por hora.",
      suggestion: `Próximo slot disponible en ~${minutes} minutos.`,
      actions: [
        { label: "Respuestas guardadas", type: "navigate", href: "/chat" },
      ],
      icon: "rate-limit",
    }
  }

  if (code === "FORBIDDEN" || status === 403) {
    if (message.includes("colección") || message.includes("collection")) {
      return {
        title: "Sin acceso a la colección",
        description: "No tenés permisos para esta colección.",
        suggestion: "Contactá al administrador para solicitar acceso.",
        actions: [
          { label: "Ver colecciones", type: "navigate", href: "/collections" },
        ],
        icon: "forbidden",
      }
    }
    return {
      title: "Acceso denegado",
      description: "No tenés permisos para esta acción.",
      suggestion: "Si creés que es un error, contactá al administrador.",
      actions: [{ label: "Volver al inicio", type: "navigate", href: "/chat" }],
      icon: "forbidden",
    }
  }

  if (code === "UPSTREAM_ERROR" || status === 502) {
    return {
      title: "Error en el servidor de búsqueda",
      description: "El servidor de búsqueda tuvo un problema interno.",
      suggestion: "Reintentá en unos segundos. Si persiste, reportá el error.",
      actions: [
        { label: "Reintentar", type: "retry" },
        { label: "Reportar error", type: "report" },
      ],
      icon: "upstream",
    }
  }

  // --- By HTTP status ---

  if (status === 401) {
    return {
      title: "Sesión expirada",
      description: "Tu sesión expiró.",
      suggestion: "Iniciá sesión de nuevo para continuar.",
      actions: [{ label: "Iniciar sesión", type: "navigate", href: "/login" }],
      icon: "auth",
    }
  }

  if (status === 404) {
    return {
      title: "No encontrado",
      description: "El recurso que buscás no existe o fue eliminado.",
      suggestion: "Verificá la URL o volvé al inicio.",
      actions: [{ label: "Ir al inicio", type: "navigate", href: "/chat" }],
      icon: "not-found",
    }
  }

  if (status === 503) {
    return {
      title: "Servicio no disponible",
      description: "El sistema está temporalmente fuera de servicio.",
      suggestion: "Intentá de nuevo en unos minutos.",
      actions: [{ label: "Reintentar", type: "retry" }],
      icon: "unavailable",
    }
  }

  // --- By message pattern ---

  const lowerMsg = message.toLowerCase()

  if (lowerMsg.includes("sqlite_busy") || lowerMsg.includes("database is locked")) {
    return {
      title: "Base de datos ocupada",
      description: "El sistema está procesando otra operación.",
      suggestion: "Reintentá en unos segundos.",
      actions: [{ label: "Reintentar", type: "retry" }],
      icon: "timeout",
    }
  }

  if (lowerMsg.includes("colección no encontrada") || lowerMsg.includes("collection not found")) {
    return {
      title: "Colección no encontrada",
      description: "La colección ya no existe o fue eliminada.",
      suggestion: "Seleccioná otra colección para continuar.",
      actions: [
        { label: "Ver colecciones", type: "navigate", href: "/collections" },
      ],
      icon: "not-found",
    }
  }

  // --- Generic fallback ---

  return {
    title: "Error inesperado",
    description: message || "Ocurrió un error.",
    suggestion: "Si el problema persiste, reportá el error.",
    actions: [
      { label: "Reintentar", type: "retry" },
      { label: "Reportar error", type: "report" },
    ],
    icon: "generic",
  }
}

/**
 * Parses a useChat error (generic Error) into structured ErrorInput.
 *
 * The useChat hook throws generic Error objects. Our API returns JSON with
 * { ok, error, details? } — the error message may contain the status code
 * and the details may contain the error code and suggestions.
 */
export function parseUseChatError(error: Error & { status?: number }): ErrorInput {
  const message = error.message ?? ""

  // Use error.status directly — don't guess from message text (false positives with numbers like "45000ms")
  const status = error.status

  // Try to extract code from message
  let code: string | undefined
  if (message.includes("ECONNREFUSED") || message.includes("no disponible")) code = "UNAVAILABLE"
  else if (message.includes("timeout") || message.includes("tardó demasiado")) code = "TIMEOUT"
  else if (status === 429 || message.includes("límite") || message.includes("Límite")) code = "RATE_LIMITED"
  else if (status === 403) code = "FORBIDDEN"
  else if (status === 502) code = "UPSTREAM_ERROR"

  const result: ErrorInput = { message }
  if (status !== undefined) result.status = status
  if (code !== undefined) result.code = code
  return result
}
