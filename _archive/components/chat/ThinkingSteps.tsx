"use client"

import { useState, useEffect } from "react"
import { ChevronDown, ChevronRight, Sparkles } from "lucide-react"
import type { ChatPhase } from "@/hooks/useRagStream"

/**
 * Muestra steps del proceso de razonamiento durante el streaming.
 *
 * Contrato del stream (futuro): cuando el RAG server emita eventos SSE del tipo
 *   event: thinking
 *   data: {"step": "Buscando en colección..."}
 * conectar onThinkingStep en useRagStream y reemplazar STEPS por los datos reales.
 *
 * Por ahora: simulación UI-side con timing — el mock y el blueprint actual no emiten thinking events.
 */

const STEPS = [
  "Buscando en la colección...",
  "Encontré fragmentos relevantes...",
  "Sintetizando respuesta...",
]

export function ThinkingSteps({ phase }: { phase: ChatPhase }) {
  const [visibleCount, setVisibleCount] = useState(0)
  const [collapsed, setCollapsed] = useState(false)

  useEffect(() => {
    if (phase === "streaming") {
      setCollapsed(false)
      setVisibleCount(1)
      const t1 = setTimeout(() => setVisibleCount(2), 700)
      const t2 = setTimeout(() => setVisibleCount(3), 1500)
      return () => {
        clearTimeout(t1)
        clearTimeout(t2)
      }
    }
    if (phase === "done" || phase === "error") {
      setVisibleCount(STEPS.length)
      const t = setTimeout(() => setCollapsed(true), 1800)
      return () => clearTimeout(t)
    }
    if (phase === "idle") {
      setVisibleCount(0)
      setCollapsed(false)
    }
  }, [phase])

  if (visibleCount === 0) return null

  return (
    <div className="flex justify-start mb-1">
      <div
        className="rounded-lg px-3 py-2 text-xs max-w-xs"
        style={{
          background: "var(--muted)",
          color: "var(--muted-foreground)",
          border: "1px solid var(--border)",
        }}
      >
        <button
          onClick={() => setCollapsed((c) => !c)}
          className="flex items-center gap-1.5 w-full text-left"
        >
          <Sparkles size={11} className="shrink-0" style={{ color: "var(--accent)" }} />
          <span className="font-medium flex-1">Proceso de razonamiento</span>
          {collapsed
            ? <ChevronRight size={11} />
            : <ChevronDown size={11} />}
        </button>

        {!collapsed && (
          <div className="mt-1.5 space-y-1 pl-4">
            {STEPS.slice(0, visibleCount).map((step, i) => (
              <div
                key={i}
                className="flex items-center gap-1.5"
                style={{
                  opacity: i === visibleCount - 1 && phase === "streaming" ? 1 : 0.6,
                  animation: i === visibleCount - 1 ? "fadeIn 0.3s ease" : undefined,
                }}
              >
                <span
                  className="w-1 h-1 rounded-full shrink-0"
                  style={{ background: "var(--accent)" }}
                />
                {step}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
