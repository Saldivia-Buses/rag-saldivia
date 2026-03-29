import { notFound } from "next/navigation"
import { getShareWithSession } from "@rag-saldivia/db"
import { MessageSquare } from "lucide-react"

export default async function SharePage({
  params,
}: {
  params: Promise<{ token: string }>
}) {
  const { token } = await params
  const data = await getShareWithSession(token)

  if (!data) notFound()

  const { session, messages } = data

  return (
    <div className="min-h-screen bg-bg text-fg">
      {/* Header */}
      <div className="border-b border-border px-6 py-4 bg-surface">
        <div className="max-w-3xl mx-auto">
          <div className="flex items-center gap-2 mb-1">
            <MessageSquare size={16} className="text-accent" />
            <h1 className="text-lg font-semibold text-fg">{session.title}</h1>
          </div>
          <p className="text-sm text-fg-muted">
            Colección: {session.collection} · Compartido (solo lectura)
          </p>
          <div className="mt-2 px-3 py-2 rounded-lg text-xs bg-warning-subtle text-warning border border-warning/20">
            ⚠️ Esta sesión fue compartida públicamente. No contiene mecanismo de autenticación.
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="max-w-3xl mx-auto px-6 py-6 space-y-6">
        {messages.length === 0 && (
          <p className="text-center text-fg-muted text-sm">Esta sesión no tiene mensajes.</p>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
            <div
              className={`max-w-2xl rounded-2xl px-4 py-3 text-sm ${
                msg.role === "user"
                  ? "rounded-br-sm bg-accent text-accent-fg"
                  : "rounded-bl-sm bg-surface border border-border text-fg"
              }`}
            >
              <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
