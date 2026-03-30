"use client"

import { useState, useTransition } from "react"
import { actionUpdateRagParams, actionResetRagParams } from "@/app/actions/config"
import type { RagParams } from "@rag-saldivia/shared"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip"
import { Info } from "lucide-react"

type SliderParam = {
  key: keyof RagParams
  label: string
  tip: string
  min: number
  max: number
  step: number
}

const SLIDERS: SliderParam[] = [
  { key: "temperature",    label: "Temperature",    tip: "Creatividad del LLM. 0 = determinístico, 2 = muy creativo.",            min: 0,   max: 2,    step: 0.05 },
  { key: "top_p",          label: "Top P",          tip: "Nucleus sampling. Controla diversidad de tokens seleccionados.",        min: 0,   max: 1,    step: 0.05 },
  { key: "max_tokens",     label: "Max Tokens",     tip: "Longitud máxima de la respuesta en tokens.",                           min: 128, max: 8192, step: 128  },
  { key: "vdb_top_k",      label: "VDB Top-K",      tip: "Documentos recuperados de Milvus por query.",                          min: 1,   max: 100,  step: 1    },
  { key: "reranker_top_k", label: "Reranker Top-K", tip: "Documentos que pasan al reranker después del retrieval.",             min: 1,   max: 50,   step: 1    },
  { key: "chunk_size",     label: "Chunk Size",     tip: "Tamaño de cada chunk de texto al indexar documentos.",                 min: 128, max: 2048, step: 64   },
  { key: "chunk_overlap",  label: "Chunk Overlap",  tip: "Solapamiento entre chunks para preservar contexto en bordes.",         min: 0,   max: 512,  step: 16   },
]

const TOGGLES: Array<{ key: keyof RagParams; label: string; tip: string }> = [
  { key: "use_reranker",   label: "Reranker",   tip: "Mejora la relevancia de documentos. Aumenta latencia." },
  { key: "use_guardrails", label: "Guardrails", tip: "Filtra respuestas inapropiadas o fuera de tema." },
]

export function AdminRagConfig({
  params: initial,
  defaults,
}: {
  params: RagParams
  defaults: RagParams
}) {
  const [values, setValues] = useState<RagParams>(initial)
  const [saved, setSaved] = useState(false)
  const [showReset, setShowReset] = useState(false)
  const [isPending, startTransition] = useTransition()

  const isDirty = JSON.stringify(values) !== JSON.stringify(initial)

  function handleSave() {
    startTransition(async () => {
      await actionUpdateRagParams(values)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    })
  }

  function handleReset() {
    setShowReset(false)
    startTransition(async () => {
      await actionResetRagParams()
      setValues(defaults)
      setSaved(false)
    })
  }

  return (
    <TooltipProvider delayDuration={200}>
      <div className="flex flex-col" style={{ maxWidth: "600px", gap: "24px" }}>
        {/* Sliders */}
        {SLIDERS.map((s) => {
          const val = Number(values[s.key])
          const def = Number(defaults[s.key])
          const isDefault = val === def
          return (
            <div key={s.key}>
              <div className="flex items-center justify-between" style={{ marginBottom: "6px" }}>
                <div className="flex items-center" style={{ gap: "6px" }}>
                  <label className="text-sm font-medium text-fg">{s.label}</label>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info size={13} className="text-fg-subtle" />
                    </TooltipTrigger>
                    <TooltipContent side="right" sideOffset={8}>{s.tip}</TooltipContent>
                  </Tooltip>
                </div>
                <div className="flex items-center" style={{ gap: "6px" }}>
                  <span className="text-sm font-mono font-semibold text-accent">{val}</span>
                  {!isDefault && (
                    <span className="text-xs text-fg-subtle">(def: {def})</span>
                  )}
                </div>
              </div>
              <input
                type="range"
                min={s.min}
                max={s.max}
                step={s.step}
                value={val}
                onChange={(e) => {
                  setSaved(false)
                  setValues((prev) => ({ ...prev, [s.key]: Number(e.target.value) }))
                }}
                className="w-full accent-accent"
                style={{ height: "6px" }}
              />
              <div className="flex justify-between text-xs text-fg-subtle" style={{ marginTop: "2px" }}>
                <span>{s.min}</span>
                <span>{s.max}</span>
              </div>
            </div>
          )
        })}

        {/* Toggles */}
        <div className="border-t border-border" style={{ paddingTop: "16px" }}>
          {TOGGLES.map((t) => (
            <div key={t.key} className="flex items-center justify-between" style={{ padding: "12px 0" }}>
              <div className="flex items-center" style={{ gap: "6px" }}>
                <p className="text-sm font-medium text-fg">{t.label}</p>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info size={13} className="text-fg-subtle" />
                  </TooltipTrigger>
                  <TooltipContent side="right" sideOffset={8}>{t.tip}</TooltipContent>
                </Tooltip>
              </div>
              <button
                onClick={() => {
                  setSaved(false)
                  setValues((prev) => ({ ...prev, [t.key]: !prev[t.key] }))
                }}
                className={`relative rounded-full transition-colors ${values[t.key] ? "bg-success" : "bg-border"}`}
                style={{ width: "40px", height: "24px" }}
              >
                <span
                  className="absolute top-1 rounded-full bg-white transition-transform"
                  style={{ width: "16px", height: "16px", transform: values[t.key] ? "translateX(18px)" : "translateX(4px)" }}
                />
              </button>
            </div>
          ))}
        </div>

        {/* Actions */}
        <div className="flex items-center" style={{ gap: "8px" }}>
          <Button onClick={handleSave} disabled={isPending || (!isDirty && !saved)}>
            {isPending ? "Guardando..." : saved ? "Guardado" : "Guardar cambios"}
          </Button>
          {saved && (
            <Badge variant="outline" className="text-success border-success">
              Guardado
            </Badge>
          )}
          <Button variant="ghost" onClick={() => setShowReset(true)} disabled={isPending}>
            Restaurar defaults
          </Button>
        </div>

        <ConfirmDialog
          open={showReset}
          onOpenChange={setShowReset}
          title="¿Restaurar configuración por defecto?"
          description="Todos los parámetros volverán a sus valores originales."
          confirmLabel="Restaurar"
          variant="default"
          onConfirm={handleReset}
        />
      </div>
    </TooltipProvider>
  )
}
