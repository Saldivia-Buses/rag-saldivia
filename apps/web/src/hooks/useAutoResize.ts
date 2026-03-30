"use client"

/* eslint-disable react-hooks/immutability, react-compiler/react-compiler */

import { useEffect, type RefObject } from "react"

/**
 * Auto-resize a textarea based on its content, up to maxHeight pixels.
 * Intentionally mutates DOM element style — this is the hook's purpose.
 */
export function useAutoResize(
  ref: RefObject<HTMLTextAreaElement | null> | undefined,
  value: string,
  maxHeight = 200
) {
  useEffect(() => {
    const ta = ref?.current
    if (!ta) return
    ta.style.height = "auto"
    ta.style.height = `${Math.min(ta.scrollHeight, maxHeight)}px`
  }, [ref, value, maxHeight])
}
