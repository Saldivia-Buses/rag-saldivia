import { MessageSquare } from "lucide-react"

export default function MessagingPage() {
  return (
    <div className="flex-1 flex flex-col items-center justify-center bg-bg">
      <h1 className="sr-only">Mensajería</h1>
      <div
        className="flex flex-col items-center text-center"
        style={{ marginBottom: "80px" }}
      >
        <div
          className="text-fg-subtle"
          style={{
            width: "56px",
            height: "56px",
            borderRadius: "16px",
            background: "var(--surface)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            marginBottom: "16px",
          }}
        >
          <MessageSquare className="h-6 w-6 opacity-50" />
        </div>
        <p className="text-sm text-fg-muted" style={{ marginBottom: "4px" }}>
          Seleccioná un canal para empezar
        </p>
        <p className="text-xs text-fg-subtle">
          O creá uno nuevo con el botón + en la sidebar
        </p>
      </div>
    </div>
  )
}
