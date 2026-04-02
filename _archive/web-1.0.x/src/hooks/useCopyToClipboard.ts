"use client"

import { useState, useCallback, useRef } from "react"

/**
 * Copy text to clipboard with a temporary "copied" feedback state.
 * Optionally tracks which item was copied via `copiedKey`.
 */
export function useCopyToClipboard(timeout = 2000) {
  const [copiedKey, setCopiedKey] = useState<string | null>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined)

  const copy = useCallback(
    async (text: string, key?: string) => {
      await navigator.clipboard.writeText(text)
      setCopiedKey(key ?? "copied")
      clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => setCopiedKey(null), timeout)
    },
    [timeout]
  )

  return { copy, copiedKey, copied: copiedKey !== null }
}
