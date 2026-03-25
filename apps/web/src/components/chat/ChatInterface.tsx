"use client"

import { useState, useRef, useEffect, useTransition } from "react"
import { Send, ThumbsUp, ThumbsDown, Loader2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback } from "@/app/actions/chat"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { useRagStream } from "@/hooks/useRagStream"
import { ThinkingSteps } from "@/components/chat/ThinkingSteps"
import { FocusModeSelector, useFocusMode } from "@/components/chat/FocusModeSelector"
import { VoiceInput } from "@/components/chat/VoiceInput"

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

function updateLastAssistantMessage(messages: Message[], content: string): Message[] {
  const updated = [...messages]
  const last = updated[updated.length - 1]
  if (last?.role === "assistant") {
    updated[updated.length - 1] = { ...last, content }
  }
  return updated
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
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()
  const bottomRef = useRef<HTMLDivElement>(null)
  const pendingSourcesRef = useRef<unknown[]>([])
  const { focusMode, setFocusMode } = useFocusMode()

  const { phase, stream, abort } = useRagStream({
    sessionId: session.id,
    collection: session.collection,
    focusMode,
    onDelta: (fullContent) => setMessages((prev) => updateLastAssistantMessage(prev, fullContent)),
    onSources: (sources) => { pendingSourcesRef.current = sources },
    onError: (message) => {
      setError(message)
      setMessages((prev) => updateLastAssistantMessage(prev, `Error: ${message}`))
      clientLog.error(new Error(message))
    },
  })

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  async function handleSend() {
    const query = input.trim()
    if (!query || phase === "streaming") return

    setInput("")
    setError(null)
    pendingSourcesRef.current = []

    const userMsg: Message = { role: "user", content: query }
    const assistantMsg: Message = { role: "assistant", content: "" }
    setMessages((prev) => [...prev, userMsg, assistantMsg])

    startTransition(async () => {
      await actionAddMessage({ sessionId: session.id, role: "user", content: query })
    })

    const result = await stream([...messages, userMsg])
    if (!result) return

    clientLog.action("rag.query", { collection: session.collection, sessionId: session.id })

    startTransition(async () => {
      const saved = await actionAddMessage({
        sessionId: session.id,
        role: "assistant",
        content: result.fullContent,
        sources: result.sources,
      })
      if (saved) {
        setMessages((prev) => {
          const updated = [...prev]
          const last = updated[updated.length - 1]
          if (last?.role === "assistant") {
            updated[updated.length - 1] = { ...last, id: saved.id, sources: result.sources }
          }
          return updated
        })
      }
    })
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

              {msg.role === "assistant" && msg.id && msg.content && phase !== "streaming" && (
                <div className="flex gap-1 pt-1 opacity-40 hover:opacity-100 transition-opacity">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => handleFeedback(msg.id!, "up")}
                    style={msg.feedback === "up" ? { color: "var(--accent)", opacity: 1 } : {}}
                    title="Útil"
                  >
                    <ThumbsUp size={13} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => handleFeedback(msg.id!, "down")}
                    style={msg.feedback === "down" ? { color: "var(--destructive)", opacity: 1 } : {}}
                    title="No útil"
                  >
                    <ThumbsDown size={13} />
                  </Button>
                </div>
              )}
            </div>
          </div>
        ))}

        <ThinkingSteps phase={phase} />

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
        <FocusModeSelector
          value={focusMode}
          onChange={setFocusMode}
          disabled={phase === "streaming"}
        />
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
          <VoiceInput
            onTranscript={(text) => setInput(text)}
            disabled={phase === "streaming"}
          />
          <button
            type="submit"
            disabled={!input.trim() || phase === "streaming"}
            className="px-4 py-2.5 rounded-xl text-sm font-medium transition-opacity disabled:opacity-40"
            style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
          >
            {phase === "streaming"
              ? <Loader2 size={16} className="animate-spin" />
              : <Send size={16} />
            }
          </button>
        </form>
      </div>
    </div>
  )
}
