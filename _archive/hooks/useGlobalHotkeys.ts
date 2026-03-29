"use client"

import { useHotkeys } from "react-hotkeys-hook"
import { useRouter } from "next/navigation"

type Options = {
  onOpenPalette?: () => void
}

/**
 * Atajos de teclado globales del sistema.
 * - Cmd+N: nueva sesión
 * - Cmd+K: abrir command palette (F2.23)
 */
export function useGlobalHotkeys({ onOpenPalette }: Options = {}) {
  const router = useRouter()

  useHotkeys(
    "mod+n",
    (e) => {
      e.preventDefault()
      router.push("/chat")
    },
    { enableOnFormTags: false }
  )

  useHotkeys(
    "mod+k",
    (e) => {
      e.preventDefault()
      onOpenPalette?.()
    },
    { enableOnFormTags: false }
  )
}
