"use client"

import { useState, useRef, useEffect, useTransition } from "react"
import { Send, ThumbsUp, ThumbsDown, Loader2, Bookmark, RefreshCw, Copy, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback, actionToggleSaved } from "@/app/actions/chat"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { useRagStream } from "@/hooks/useRagStream"
import { ThinkingSteps } from "@/components/chat/ThinkingSteps"
import { FocusModeSelector, useFocusMode } from "@/components/chat/FocusModeSelector"
import { VoiceInput } from "@/components/chat/VoiceInput"
import { ExportSession } from "@/components/chat/ExportSession"
import { SourcesPanel } from "@/components/chat/SourcesPanel"
import { RelatedQuestions } from "@/components/chat/RelatedQuestions"

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
  const [savedIds, setSavedIds] = useState<Set<number>>(new Set())
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [queryStats, setQueryStats] = useState<{ ms: number; sources: number } | null>(null)
  const [relatedQuestions, setRelatedQuestions] = useState<string[]>([])
  const streamStartRef = useRef<number>(0)
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
    setQueryStats(null)
    setRelatedQuestions([])
    pendingSourcesRef.current = []
    streamStartRef.current = Date.now()

    const userMsg: Message = { role: "user", content: query }
    const assistantMsg: Message = { role: "assistant", content: "" }
    setMessages((prev) => [...prev, userMsg, assistantMsg])

    startTransition(async () => {
      await actionAddMessage({ sessionId: session.id, role: "user", content: query })
    })

    const result = await stream([...messages, userMsg])
    if (!result) return

    setQueryStats({ ms: Date.now() - streamStartRef.current, sources: result.sources.length })

    // Generar preguntas relacionadas en background — F2.20
    fetch("/api/rag/suggest", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query, lastResponse: result.fullContent }),
    })
      .then((r) => r.json())
      .then((data: { ok: boolean; questions?: string[] }) => {
        if (data.ok && data.questions) setRelatedQuestions(data.questions)
      })
      .catch(() => {})

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

  function handleRegenerate() {
    const lastUser = [...messages].reverse().find((m) => m.role === "user")
    if (lastUser) {
      setInput(lastUser.content)
    }
  }

  async function handleCopy(messageId: number, content: string) {
    await navigator.clipboard.writeText(content)
    setCopiedId(messageId)
    setTimeout(() => setCopiedId(null), 2000)
  }

  async function handleToggleSaved(messageId: number, content: string) {
    const isSaved = savedIds.has(messageId)
    setSavedIds((prev) => {
      const next = new Set(prev)
      if (isSaved) next.delete(messageId)
      else next.add(messageId)
      return next
    })
    await actionToggleSaved(messageId, content, session.title, isSaved)
  }

  async function handleFeedback(messageId: number, rating: "up" | "down") {
    await actionAddFeedback(messageId, rating)
    setMessages((prev) =>
      prev.map((m) => (m.id === messageId ? { ...m, feedback: rating } : m))
    )
  }

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Header */}
      <div
        className="flex items-center justify-between px-6 py-3 border-b shrink-0 no-print"
        style={{ borderColor: "var(--border)" }}
      >
        <span className="text-sm font-medium truncate" style={{ color: "var(--muted-foreground)" }}>
          {session.collection}
        </span>
        <ExportSession session={session} />
      </div>

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
            className={`flex group ${msg.role === "user" ? "justify-end" : "justify-start"}`}
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

              {/* Panel de fuentes — F2.19 */}
              {msg.role === "assistant" && msg.sources && msg.sources.length > 0 && (
                <SourcesPanel sources={msg.sources} />
              )}

              {msg.role === "assistant" && msg.content && phase !== "streaming" && (
                <div className="flex gap-1 pt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  {/* Regenerar — F1.15 */}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={handleRegenerate}
                    title="Regenerar respuesta"
                  >
                    <RefreshCw size={13} />
                  </Button>
                  {/* Copy — F1.16 */}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => msg.id ? handleCopy(msg.id, msg.content) : navigator.clipboard.writeText(msg.content)}
                    title="Copiar respuesta"
                    style={msg.id && copiedId === msg.id ? { color: "var(--accent)" } : {}}
                  >
                    {msg.id && copiedId === msg.id ? <Check size={13} /> : <Copy size={13} />}
                  </Button>
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
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => handleToggleSaved(msg.id!, msg.content)}
                    style={savedIds.has(msg.id!) ? { color: "var(--accent)", opacity: 1 } : {}}
                    title={savedIds.has(msg.id!) ? "Quitar de guardados" : "Guardar respuesta"}
                  >
                    <Bookmark size={13} />
                  </Button>
                </div>
              )}
            </div>
          </div>
        ))}

        {/* Stats del último query — F1.17 */}
        {queryStats && phase === "done" && (
          <div
            className="flex justify-start px-1"
          >
            <span
              className="text-xs opacity-0 group-hover:opacity-100 transition-opacity"
              style={{ color: "var(--muted-foreground)" }}
            >
              {queryStats.ms}ms · {queryStats.sources} doc{queryStats.sources !== 1 ? "s" : ""}
            </span>
          </div>
        )}

        {/* Preguntas relacionadas tras la última respuesta — F2.20 */}
        {phase === "done" && relatedQuestions.length > 0 && (
          <div className="flex justify-start px-1">
            <div className="max-w-2xl w-full">
              <RelatedQuestions
                questions={relatedQuestions}
                onSelect={(q) => setInput(q)}
              />
            </div>
          </div>
        )}

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
