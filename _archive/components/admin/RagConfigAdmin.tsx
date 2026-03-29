"use client"

import { useState, useTransition } from "react"
import { actionUpdateRagParams, actionResetRagParams } from "@/app/actions/config"
import type { RagParams } from "@rag-saldivia/shared"
import { Button } from "@/components/ui/button"

type SliderParam = {
  key: keyof RagParams
  label: string
  description: string
  min: number
  max: number
  step: number
}

const SLIDERS: SliderParam[] = [
  { key: "temperature",   label: "Temperature",     description: "Creatividad del LLM (0 = determinístico, 2 = muy creativo)", min: 0,   max: 2,    step: 0.05 },
  { key: "top_p",         label: "Top P",           description: "Nucleus sampling — controla diversidad de tokens",           min: 0,   max: 1,    step: 0.05 },
  { key: "max_tokens",    label: "Max tokens",      description: "Longitud máxima de la respuesta",                           min: 128, max: 8192, step: 128  },
  { key: "vdb_top_k",     label: "VDB Top-K",       description: "Documentos recuperados de Milvus por query",                min: 1,   max: 100,  step: 1    },
  { key: "reranker_top_k",label: "Reranker Top-K",  description: "Documentos que pasan al reranker",                         min: 1,   max: 50,   step: 1    },
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
    startTransition(async () => { await actionResetRagParams() })
  }

  return (
    <div className="p-6 max-w-xl space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-fg">Configuración RAG</h1>
        <p className="text-sm text-fg-muted mt-0.5">Parámetros del modelo y recuperación de documentos</p>
      </div>

      {/* Sliders */}
      <div className="space-y-6">
        {SLIDERS.map((slider) => (
          <div key={slider.key} className="space-y-2">
            <div className="flex justify-between items-start">
              <div>
                <label className="text-sm font-medium text-fg">{slider.label}</label>
                <p className="text-xs text-fg-muted">{slider.description}</p>
              </div>
              <span className="text-sm font-mono font-semibold text-accent">{values[slider.key]}</span>
            </div>
            <input
              type="range"
              min={slider.min} max={slider.max} step={slider.step}
              value={Number(values[slider.key])}
              onChange={(e) => { setSaved(false); setValues((prev) => ({ ...prev, [slider.key]: Number(e.target.value) })) }}
              className="w-full accent-accent"
            />
            <div className="flex justify-between text-xs text-fg-subtle">
              <span>{slider.min}</span><span>{slider.max}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Toggles */}
      <div className="space-y-3 pt-4 border-t border-border">
        <ToggleRow label="Reranker" description="Mejora relevancia de documentos (aumenta latencia)"
          checked={values.use_reranker}
          onChange={(v) => { setSaved(false); setValues((p) => ({ ...p, use_reranker: v })) }} />
        <ToggleRow label="Guardrails" description="Filtrar respuestas inapropiadas"
          checked={values.use_guardrails}
          onChange={(v) => { setSaved(false); setValues((p) => ({ ...p, use_guardrails: v })) }} />
      </div>

      <div className="flex gap-3">
        <Button onClick={handleSave} disabled={isPending} variant={saved ? "outline" : "default"}
          className={saved ? "text-success border-success" : ""}>
          {isPending ? "Guardando..." : saved ? "✓ Guardado" : "Guardar cambios"}
        </Button>
        <Button onClick={handleReset} disabled={isPending} variant="outline">Resetear a defaults</Button>
      </div>
    </div>
  )
}

function ToggleRow({ label, description, checked, onChange }: {
  label: string; description: string; checked: boolean; onChange: (v: boolean) => void
}) {
  return (
    <div className="flex items-center justify-between py-2">
      <div>
        <p className="text-sm font-medium text-fg">{label}</p>
        <p className="text-xs text-fg-muted">{description}</p>
      </div>
      <button
        onClick={() => onChange(!checked)}
        className={`relative w-10 h-6 rounded-full transition-colors ${checked ? "bg-success" : "bg-border"}`}
      >
        <span className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${checked ? "translate-x-4" : "translate-x-1"}`} />
      </button>
    </div>
  )
}
