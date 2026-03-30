"use client"

import { AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"

export default function MessagingError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  const message =
    process.env.NODE_ENV === "production"
      ? "Ha ocurrido un error inesperado."
      : error.message

  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
      <EmptyPlaceholder>
        <EmptyPlaceholder.Icon icon={AlertTriangle} />
        <EmptyPlaceholder.Title>Algo salió mal</EmptyPlaceholder.Title>
        <EmptyPlaceholder.Description>{message}</EmptyPlaceholder.Description>
      </EmptyPlaceholder>
      <Button onClick={reset} variant="outline">
        Reintentar
      </Button>
    </div>
  )
}
