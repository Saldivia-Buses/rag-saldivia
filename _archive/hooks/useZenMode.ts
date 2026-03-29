"use client"

import { useState, useEffect } from "react"

export function useZenMode() {
  const [isZen, setIsZen] = useState(false)

  useEffect(() => {
    function handler(e: KeyboardEvent) {
      const mod = e.metaKey || e.ctrlKey
      if (mod && e.shiftKey && e.key === "Z") {
        e.preventDefault()
        setIsZen((z) => !z)
      }
      if (e.key === "Escape") {
        setIsZen(false)
      }
    }
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [])

  return { isZen, toggleZen: () => setIsZen((z) => !z) }
}
