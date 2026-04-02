"use client"

import { useState, useEffect, useCallback } from "react"

/**
 * SSR-safe localStorage hook. Returns defaultValue during SSR and initial
 * hydration, then syncs from localStorage after mount to avoid hydration
 * mismatches. Values are JSON-serialized.
 */
export function useLocalStorage<T>(key: string, defaultValue: T): [T, (value: T | ((prev: T) => T)) => void] {
  const [value, setValue] = useState<T>(defaultValue)

  // Sync from localStorage after mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(key)
      if (stored !== null) setValue(JSON.parse(stored) as T)
    } catch { /* ignore parse errors */ }
  }, [key])

  const set = useCallback(
    (newValue: T | ((prev: T) => T)) => {
      setValue((prev) => {
        const resolved =
          typeof newValue === "function"
            ? (newValue as (prev: T) => T)(prev)
            : newValue
        try {
          localStorage.setItem(key, JSON.stringify(resolved))
        } catch { /* quota exceeded or private browsing */ }
        return resolved
      })
    },
    [key]
  )

  return [value, set]
}
