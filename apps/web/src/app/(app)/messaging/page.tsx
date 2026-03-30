import { MessageSquare } from "lucide-react"

export default function MessagingPage() {
  return (
    <div className="flex-1 flex flex-col items-center justify-center bg-bg">
      <h1 className="sr-only">Mensajería</h1>
      <MessageSquare className="h-8 w-8 text-fg-subtle mb-4 opacity-40" />
      <p className="text-sm text-fg-subtle">
        Seleccioná un canal para empezar
      </p>
    </div>
  )
}
