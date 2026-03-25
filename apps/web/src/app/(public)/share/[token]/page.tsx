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
    <div className="min-h-screen" style={{ background: "var(--bg)", color: "var(--fg)" }}>
      {/* Header */}
      <div
        className="border-b px-6 py-4"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="max-w-3xl mx-auto">
          <div className="flex items-center gap-2 mb-1">
            <MessageSquare size={16} style={{ color: "var(--accent)" }} />
            <h1 className="text-lg font-semibold">{session.title}</h1>
          </div>
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
            Colección: {session.collection} · Compartido (solo lectura)
          </p>
          <div
            className="mt-2 px-3 py-2 rounded-md text-xs"
            style={{ background: "var(--muted-bg)", color: "var(--muted-foreground)" }}
          >
            ⚠️ Esta sesión fue compartida públicamente. No contiene mecanismo de autenticación.
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="max-w-3xl mx-auto px-6 py-6 space-y-6">
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`max-w-2xl rounded-xl px-4 py-3 text-sm ${
                msg.role === "user" ? "rounded-br-sm" : "rounded-bl-sm"
              }`}
              style={{
                background: msg.role === "user" ? "var(--accent)" : "var(--muted)",
                color: msg.role === "user" ? "var(--accent-foreground)" : "var(--fg)",
              }}
            >
              <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>
            </div>
          </div>
        ))}

        {messages.length === 0 && (
          <p className="text-center" style={{ color: "var(--muted-foreground)" }}>
            Esta sesión no tiene mensajes.
          </p>
        )}
      </div>
    </div>
  )
}
