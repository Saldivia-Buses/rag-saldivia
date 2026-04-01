"use client"

import { useState, useEffect, useTransition } from "react"
import { MessageSquareWarning, Send, X } from "lucide-react"
import { reportError, registerFeedbackSetter } from "@/lib/error-feedback"
import { toast } from "sonner"

/**
 * Error feedback dialog — renders at layout level.
 * Opens when showErrorFeedback() is called from anywhere.
 *
 * Mount once in a layout or page wrapper:
 *   <ErrorFeedbackMount />
 */
export function ErrorFeedbackMount() {
  const [state, setState] = useState<{ error: string; context: string } | null>(null)

  useEffect(() => {
    registerFeedbackSetter(setState)
    return () => registerFeedbackSetter(null)
  }, [])

  if (!state) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" style={{ background: "rgba(0,0,0,0.3)" }}>
      <ErrorFeedbackDialog error={state.error} context={state.context} onClose={() => setState(null)} />
    </div>
  )
}

function ErrorFeedbackDialog({
  error,
  context,
  onClose,
}: {
  error: string
  context: string
  onClose: () => void
}) {
  const [comment, setComment] = useState("")
  const [isPending, startTransition] = useTransition()

  function handleSend() {
    startTransition(async () => {
      const ok = await reportError({ error, context, comment: comment.trim() || undefined })
      if (ok) {
        toast.success("Reporte enviado. Gracias por tu feedback.")
      } else {
        toast.error("No se pudo enviar el reporte.")
      }
      onClose()
    })
  }

  return (
    <div
      className="rounded-xl border border-border bg-surface shadow-lg"
      style={{ padding: "16px", maxWidth: "400px", width: "90%" }}
    >
      <div className="flex items-start justify-between" style={{ marginBottom: "12px" }}>
        <div className="flex items-center" style={{ gap: "8px" }}>
          <MessageSquareWarning size={18} className="text-warning shrink-0" />
          <h3 className="text-sm font-semibold text-fg">Reportar error</h3>
        </div>
        <button onClick={onClose} className="text-fg-subtle hover:text-fg transition-colors">
          <X size={16} />
        </button>
      </div>

      <p className="text-xs text-fg-muted" style={{ marginBottom: "8px" }}>
        Error: <span className="text-fg">{error}</span>
      </p>
      <p className="text-xs text-fg-subtle" style={{ marginBottom: "12px" }}>
        Contexto: {context}
      </p>

      <textarea
        value={comment}
        onChange={(e) => setComment(e.target.value)}
        placeholder="¿Qué estabas haciendo? (opcional)"
        className="w-full rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent resize-none transition-colors"
        style={{ padding: "8px 12px", minHeight: "60px" }}
      />

      <div className="flex justify-end" style={{ gap: "8px", marginTop: "12px" }}>
        <button
          onClick={onClose}
          className="text-xs text-fg-muted hover:text-fg transition-colors"
          style={{ padding: "6px 12px" }}
        >
          Cancelar
        </button>
        <button
          onClick={handleSend}
          disabled={isPending}
          className="flex items-center text-xs font-medium rounded-lg bg-accent text-accent-fg disabled:opacity-50 transition-opacity"
          style={{ padding: "6px 12px", gap: "4px" }}
        >
          <Send size={12} />
          {isPending ? "Enviando..." : "Enviar reporte"}
        </button>
      </div>
    </div>
  )
}
