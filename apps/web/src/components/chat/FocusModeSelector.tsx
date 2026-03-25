"use client"

import { useEffect, useState } from "react"
import { FOCUS_MODES, type FocusModeId } from "@rag-saldivia/shared"

const STORAGE_KEY = "rag-focus-mode"

type Props = {
  value: FocusModeId
  onChange: (mode: FocusModeId) => void
  disabled?: boolean
}

export function FocusModeSelector({ value, onChange, disabled }: Props) {
  return (
    <div className="flex items-center gap-1 mb-2 flex-wrap">
      {FOCUS_MODES.map((mode) => {
        const active = value === mode.id
        return (
          <button
            key={mode.id}
            onClick={() => onChange(mode.id)}
            disabled={disabled}
            className="px-2.5 py-1 rounded-full text-xs font-medium transition-colors disabled:opacity-40"
            style={{
              background: active ? "var(--accent)" : "var(--muted)",
              color: active ? "var(--accent-foreground)" : "var(--muted-foreground)",
            }}
            title={mode.systemPrompt}
          >
            {mode.label}
          </button>
        )
      })}
    </div>
  )
}

/** Hook para persistir el modo seleccionado en localStorage */
export function useFocusMode() {
  const [focusMode, setFocusModeState] = useState<FocusModeId>("detallado")

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY) as FocusModeId | null
    if (saved && FOCUS_MODES.some((m) => m.id === saved)) {
      setFocusModeState(saved)
    }
  }, [])

  function setFocusMode(mode: FocusModeId) {
    setFocusModeState(mode)
    localStorage.setItem(STORAGE_KEY, mode)
  }

  return { focusMode, setFocusMode }
}
