"use client"

import { useState, useRef, useEffect, useTransition } from "react"
import { Send, ThumbsUp, ThumbsDown, Loader2 } from "lucide-react"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback } from "@/app/actions/chat"
import { clientLog } from "@rag-saldivia/logger/frontend"

type ChatPhase = "idle" | "streaming" | "done" | "error"

type Message = {
  id?: number
  role: "user" | "assistant"
  content: string
  sources?: unknown[]
  feedback?: "up" | "down" | null
}

function parseSessionMessages(session: DbChatSession & { messages?: DbChatMessage[] }): Message[] {
  return (session.messages ?? []).map((m) => ({
    id: m.id ?? undefined,
    role: m.role as "user" | "assistant",
    content: m.content,
    sources: (m.sources as unknown[]) ?? [],
    feedback: null,
  }))
}

export function ChatInterface({
  session,
  userId,
}: {
  session: DbChatSession & { messages?: DbChatMessage[] }
  userId: number
}) {
  const [messages, setMessages] = useState<Message[]>(() => parseSessionMessages(session))
  const [input, setInput] = useState("")
  const [phase, setPhase] = useState<ChatPhase>("idle")
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()
  const bottomRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  async function handleSend() {
    const query = input.trim()
    if (!query || phase === "streaming") return

    setInput("")
    setError(null)
    setPhase("streaming")

    // Guardar mensaje del usuario
    startTransition(async () => {
      await actionAddMessage({
        sessionId: session.id,
        role: "user",
        content: query,
      })
    })

    const userMsg: Message = { role: "user", content: query }
    setMessages((prev) => [...prev, userMsg])

    // Placeholder del asistente
    const assistantMsg: Message = { role: "assistant", content: "" }
    setMessages((prev) => [...prev, assistantMsg])

    // Stream
    abortRef.current = new AbortController()
    let fullContent = ""
    let sources: unknown[] = []

    try {
      const res = await fetch("/api/rag/generate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          messages: [...messages, userMsg].map((m) => ({
            role: m.role,
            content: m.content,
          })),
          collection_name: session.collection,
          session_id: session.id,
          use_knowledge_base: true,
        }),
        signal: abortRef.current.signal,
      })

      if (!res.ok) {
        const data = await res.json().catch(() => ({ error: "Error desconocido" }))
        throw new Error(data.error ?? `Error ${res.status}`)
      }

      if (!res.body) throw new Error("No stream")

      const reader = res.body.getReader()
      const decoder = new TextDecoder()

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split("\n")

        for (const line of lines) {
          if (!line.startsWith("data: ")) continue
          const data = line.slice(6).trim()
          if (data === "[DONE]") continue

          try {
            const parsed = JSON.parse(data)
            const delta = parsed.choices?.[0]?.delta?.content ?? ""
            if (delta) {
              fullContent += delta
              setMessages((prev) => {
                const updated = [...prev]
                const last = updated[updated.length - 1]
                if (last?.role === "assistant") {
                  updated[updated.length - 1] = { ...last, content: fullContent }
                }
                return updated
              })
            }

            // Capturar sources si vienen
            const srcData = parsed.choices?.[0]?.delta?.sources
            if (srcData) sources = srcData
          } catch {
            // Ignorar líneas malformadas
          }
        }
      }

      // Guardar respuesta del asistente
      startTransition(async () => {
        const saved = await actionAddMessage({
          sessionId: session.id,
          role: "assistant",
          content: fullContent,
          sources,
        })
        if (saved) {
          setMessages((prev) => {
            const updated = [...prev]
            const last = updated[updated.length - 1]
            if (last?.role === "assistant") {
              updated[updated.length - 1] = { ...last, id: saved.id, sources }
            }
            return updated
          })
        }
      })

      setPhase("done")
      clientLog.action("rag.query", { collection: session.collection, sessionId: session.id })
    } catch (err) {
      if (err instanceof Error && err.name === "AbortError") {
        setPhase("idle")
        return
      }

      const message = err instanceof Error ? err.message : String(err)
      setError(message)
      setPhase("error")
      clientLog.error(err instanceof Error ? err : new Error(message))

      // Actualizar mensaje del asistente con el error
      setMessages((prev) => {
        const updated = [...prev]
        const last = updated[updated.length - 1]
        if (last?.role === "assistant") {
          updated[updated.length - 1] = { ...last, content: `Error: ${message}` }
        }
        return updated
      })
    }
  }

  async function handleFeedback(messageId: number, rating: "up" | "down") {
    await actionAddFeedback(messageId, rating)
    setMessages((prev) =>
      prev.map((m) => (m.id === messageId ? { ...m, feedback: rating } : m))
    )
  }

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {messages.length === 0 && (
          <div className="h-full flex items-center justify-center" style={{ color: "var(--muted-foreground)" }}>
            <div className="text-center space-y-1">
              <p className="font-medium">Colección: {session.collection}</p>
              <p className="text-sm">Hacé tu primera pregunta</p>
            </div>
          </div>
        )}

        {messages.map((msg, i) => (
          <div
            key={i}
            className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`max-w-2xl rounded-xl px-4 py-3 text-sm space-y-1 ${
                msg.role === "user" ? "rounded-br-sm" : "rounded-bl-sm"
              }`}
              style={{
                background: msg.role === "user" ? "var(--primary)" : "var(--muted)",
                color: msg.role === "user" ? "var(--primary-foreground)" : "var(--foreground)",
              }}
            >
              <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>

              {/* Feedback para mensajes del asistente */}
              {msg.role === "assistant" && msg.id && msg.content && phase !== "streaming" && (
                <div className="flex gap-2 pt-1 opacity-50 hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => handleFeedback(msg.id!, "up")}
                    className={msg.feedback === "up" ? "opacity-100" : ""}
                  >
                    <ThumbsUp size={13} />
                  </button>
                  <button
                    onClick={() => handleFeedback(msg.id!, "down")}
                    className={msg.feedback === "down" ? "opacity-100" : ""}
                  >
                    <ThumbsDown size={13} />
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}

        {phase === "streaming" && messages[messages.length - 1]?.content === "" && (
          <div className="flex justify-start">
            <div className="px-4 py-3 rounded-xl rounded-bl-sm" style={{ background: "var(--muted)" }}>
              <Loader2 size={16} className="animate-spin" />
            </div>
          </div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t" style={{ borderColor: "var(--border)" }}>
        {error && (
          <div
            className="mb-3 px-3 py-2 rounded-md text-xs"
            style={{ background: "#fef2f2", color: "var(--destructive)" }}
          >
            {error}
          </div>
        )}
        <form
          onSubmit={(e) => { e.preventDefault(); handleSend() }}
          className="flex gap-2"
        >
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder={`Preguntá sobre ${session.collection}...`}
            disabled={phase === "streaming"}
            className="flex-1 px-4 py-2.5 rounded-xl border text-sm outline-none transition-all focus:ring-2 disabled:opacity-50"
            style={{
              borderColor: "var(--border)",
              background: "var(--background)",
              "--tw-ring-color": "var(--ring)",
            } as React.CSSProperties}
          />
          <button
            type="submit"
            disabled={!input.trim() || phase === "streaming"}
            className="px-4 py-2.5 rounded-xl text-sm font-medium transition-opacity disabled:opacity-40"
            style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
          >
            {phase === "streaming" ? <Loader2 size={16} className="animate-spin" /> : <Send size={16} />}
          </button>
        </form>
      </div>
    </div>
  )
}
