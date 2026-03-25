"use client"

import { useHotkeys } from "react-hotkeys-hook"
import { useRouter } from "next/navigation"

/**
 * Atajos de teclado globales del sistema.
 * Registrados en AppShellChrome para que estén disponibles en toda la app.
 *
 * Atajos:
 * - Cmd+N / Ctrl+N: navegar a /chat (nueva sesión)
 * - Esc: cierra modales/drawers (manejado por los componentes individuales via useZenMode)
 *
 * Nota: j/k para navegar sesiones requiere estado de la lista de sesiones —
 * se implementa en Fase 2 cuando el panel de sesiones tenga estado centralizado.
 */
export function useGlobalHotkeys() {
  const router = useRouter()

  useHotkeys(
    "mod+n",
    (e) => {
      e.preventDefault()
      router.push("/chat")
    },
    { enableOnFormTags: false }
  )
}
