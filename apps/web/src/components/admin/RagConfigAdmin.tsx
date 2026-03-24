"use client"

import { useState, useTransition } from "react"
import { actionUpdateRagParams, actionResetRagParams } from "@/app/actions/config"
import type { RagParams } from "@rag-saldivia/shared"

type SliderParam = {
  key: keyof RagParams
  label: string
  description: string
  min: number
  max: number
  step: number
}

const SLIDERS: SliderParam[] = [
  { key: "temperature", label: "Temperature", description: "Creatividad del LLM (0 = determinístico, 2 = muy creativo)", min: 0, max: 2, step: 0.05 },
  { key: "top_p", label: "Top P", description: "Nucleus sampling — controla diversidad de tokens", min: 0, max: 1, step: 0.05 },
  { key: "max_tokens", label: "Max tokens", description: "Longitud máxima de la respuesta", min: 128, max: 8192, step: 128 },
  { key: "vdb_top_k", label: "VDB Top-K", description: "Documentos recuperados de Milvus por query", min: 1, max: 100, step: 1 },
  { key: "reranker_top_k", label: "Reranker Top-K", description: "Documentos que pasan al reranker", min: 1, max: 50, step: 1 },
]

export function RagConfigAdmin({ params }: { params: RagParams }) {
  const [values, setValues] = useState(params)
  const [saved, setSaved] = useState(false)
  const [isPending, startTransition] = useTransition()

  async function handleSave() {
    startTransition(async () => {
      await actionUpdateRagParams(values)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    })
  }

  async function handleReset() {
    if (!confirm("¿Resetear la configuración a los valores por defecto?")) return
    startTransition(async () => {
      await actionResetRagParams()
    })
  }

  return (
    <div className="space-y-6">
      {/* Sliders */}
      {SLIDERS.map((slider) => (
        <div key={slider.key} className="space-y-2">
          <div className="flex justify-between items-start">
            <div>
              <label className="text-sm font-medium">{slider.label}</label>
              <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{slider.description}</p>
            </div>
            <span className="text-sm font-mono font-medium">{values[slider.key]}</span>
          </div>
          <input
            type="range"
            min={slider.min}
            max={slider.max}
            step={slider.step}
            value={Number(values[slider.key])}
            onChange={(e) => {
              setSaved(false)
              setValues((prev) => ({ ...prev, [slider.key]: Number(e.target.value) }))
            }}
            className="w-full accent-current"
          />
          <div className="flex justify-between text-xs" style={{ color: "var(--muted-foreground)" }}>
            <span>{slider.min}</span>
            <span>{slider.max}</span>
          </div>
        </div>
      ))}

      {/* Toggles */}
      <div className="space-y-3 pt-2 border-t" style={{ borderColor: "var(--border)" }}>
        <ToggleRow
          label="Reranker"
          description="Mejora relevancia de documentos (aumenta latencia)"
          checked={values.use_reranker}
          onChange={(v) => { setSaved(false); setValues((prev) => ({ ...prev, use_reranker: v })) }}
        />
        <ToggleRow
          label="Guardrails"
          description="Filtrar respuestas inapropiadas"
          checked={values.use_guardrails}
          onChange={(v) => { setSaved(false); setValues((prev) => ({ ...prev, use_guardrails: v })) }}
        />
      </div>

      {/* Actions */}
      <div className="flex gap-3 pt-2">
        <button
          onClick={handleSave}
          disabled={isPending}
          className="px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
          style={{ background: saved ? "#dcfce7" : "var(--primary)", color: saved ? "#166534" : "var(--primary-foreground)" }}
        >
          {isPending ? "Guardando..." : saved ? "✓ Guardado" : "Guardar cambios"}
        </button>
        <button
          onClick={handleReset}
          disabled={isPending}
          className="px-4 py-2 rounded-md text-sm border disabled:opacity-50"
          style={{ borderColor: "var(--border)" }}
        >
          Resetear a defaults
        </button>
      </div>
    </div>
  )
}

function ToggleRow({ label, description, checked, onChange }: {
  label: string
  description: string
  checked: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{description}</p>
      </div>
      <button
        onClick={() => onChange(!checked)}
        className="relative w-10 h-6 rounded-full transition-colors"
        style={{ background: checked ? "#16a34a" : "var(--border)" }}
      >
        <span
          className="absolute top-1 w-4 h-4 rounded-full bg-white transition-transform"
          style={{ left: checked ? "calc(100% - 1.25rem)" : "0.25rem" }}
        />
      </button>
    </div>
  )
}
