"use client"

import { useEffect } from "react"
import { actionCompleteOnboarding } from "@/app/actions/settings"

type Props = {
  completed: boolean
}

const TOUR_STEPS = [
  {
    element: "nav",
    popover: {
      title: "Barra de navegación",
      description: "Accedé rápidamente a chat, colecciones, uploads y configuración desde aquí.",
      side: "right" as const,
    },
  },
  {
    element: "main",
    popover: {
      title: "Área de chat",
      description: "Hacé tus preguntas sobre los documentos de tu organización aquí.",
      side: "top" as const,
    },
  },
  {
    element: "[data-tour='focus-mode']",
    popover: {
      title: "Modos de foco",
      description: "Elegí cómo querés que el sistema responda: detallado, ejecutivo, técnico o comparativo.",
      side: "top" as const,
    },
  },
  {
    element: "a[href='/collections']",
    popover: {
      title: "Colecciones",
      description: "Tus documentos están organizados en colecciones. Podés crear y gestionar colecciones desde aquí.",
      side: "right" as const,
    },
  },
  {
    element: "a[href='/settings']",
    popover: {
      title: "Configuración",
      description: "Cambiá tu perfil, contraseña y preferencias de respuesta.",
      side: "right" as const,
    },
  },
]

export function OnboardingTour({ completed }: Props) {
  useEffect(() => {
    if (completed) return

    async function startTour() {
      // Import dinámico de driver.js para no afectar el bundle inicial
      const { driver } = await import("driver.js")
      await import("driver.js/dist/driver.css")

      const driverObj = driver({
        showProgress: true,
        animate: true,
        allowClose: true,
        steps: TOUR_STEPS,
        onDestroyStarted: () => {
          driverObj.destroy()
          actionCompleteOnboarding().catch(() => {})
        },
        onDestroyed: () => {
          actionCompleteOnboarding().catch(() => {})
        },
      })

      // Pequeño delay para que el DOM esté listo
      setTimeout(() => driverObj.drive(), 800)
    }

    startTour().catch(() => {})
  }, [completed])

  return null
}
