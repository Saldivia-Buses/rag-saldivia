"use client"

import { useState, useRef, useEffect, useTransition, useCallback } from "react"
import { Send, ThumbsUp, ThumbsDown, Loader2, Bookmark, RefreshCw, Copy, Check, GitBranch } from "lucide-react"
import { Button } from "@/components/ui/button"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback, actionToggleSaved, actionForkSession } from "@/app/actions/chat"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { useRagStream } from "@/hooks/useRagStream"
import { SourcesPanel } from "@/components/chat/SourcesPanel"

type Message = {
  id?: number
  role: "user" | "assistant"
  content: string
  sources?: import("@rag-saldivia/shared").Citation[]
  feedback?: "up" | "down" | null
}

function parseSessionMessages(session: DbChatSession & { messages?: DbChatMessage[] }): Message[] {
  return (session.messages ?? []).map((m) => ({
    id: m.id ?? undefined,
    role: m.role as "user" | "assistant",
    content: m.content,
    sources: (m.sources as import("@rag-saldivia/shared").Citation[]) ?? [],
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
  userId: _userId,
}: {
  session: DbChatSession & { messages?: DbChatMessage[] }
  userId: number
}) {
  const [messages, setMessages] = useState<Message[]>(() => parseSessionMessages(session))
  const [input, setInput] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [_isPending, startTransition] = useTransition()
  const [savedIds, setSavedIds] = useState<Set<number>>(new Set())
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [queryStats, setQueryStats] = useState<{ ms: number; sources: number } | null>(null)
  const streamStartRef = useRef<number>(0)
  const bottomRef = useRef<HTMLDivElement>(null)
  const pendingSourcesRef = useRef<import("@rag-saldivia/shared").Citation[]>([])

  const { phase, stream } = useRagStream({
    sessionId: session.id,
    collection: session.collection,
    collections: [session.collection],
    focusMode: "detailed",
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

  const handleSend = useCallback(async () => {
    const query = input.trim()
    if (!query || phase === "streaming") return

    setInput("")
    setError(null)
    setQueryStats(null)
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
  }, [input, phase, messages, stream, session.id, session.collection, startTransition])

  const handleRegenerate = useCallback(() => {
    const lastUser = [...messages].reverse().find((m) => m.role === "user")
    if (lastUser) {
      setInput(lastUser.content)
    }
  }, [messages])

  const handleCopy = useCallback(async (messageId: number, content: string) => {
    await navigator.clipboard.writeText(content)
    setCopiedId(messageId)
    setTimeout(() => setCopiedId(null), 2000)
  }, [])

  const handleToggleSaved = useCallback(async (messageId: number, content: string) => {
    const isSaved = savedIds.has(messageId)
    setSavedIds((prev) => {
      const next = new Set(prev)
      if (isSaved) next.delete(messageId)
      else next.add(messageId)
      return next
    })
    await actionToggleSaved(messageId, content, session.title, isSaved)
  }, [savedIds, session.title])

  const handleFeedback = useCallback(async (messageId: number, rating: "up" | "down") => {
    await actionAddFeedback(messageId, rating)
    setMessages((prev) =>
      prev.map((m) => (m.id === messageId ? { ...m, feedback: rating } : m))
    )
  }, [])

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-border shrink-0 no-print">
        <span className="text-sm font-medium text-fg-muted truncate">
          {session.collection}
        </span>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-6">
        {messages.length === 0 && (
          <div className="h-full flex items-center justify-center">
            <div className="text-center space-y-1 text-fg-muted">
              <p className="font-medium text-fg">Colección: {session.collection}</p>
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
              className={`max-w-2xl rounded-2xl px-4 py-3 text-sm space-y-1 ${
                msg.role === "user"
                  ? "rounded-br-sm bg-accent text-accent-fg"
                  : "rounded-bl-sm bg-surface text-fg border border-border"
              }`}
            >
              <p className="whitespace-pre-wrap leading-relaxed">{msg.content}</p>

              {msg.role === "assistant" && msg.sources && msg.sources.length > 0 && (
                <SourcesPanel sources={msg.sources} />
              )}

              {msg.role === "assistant" && msg.content && phase !== "streaming" && (
                <div className="flex gap-1 pt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  {msg.id && (
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-6 w-6"
                      title="Bifurcar desde aquí"
                      onClick={async () => {
                        const newId = await actionForkSession(session.id, msg.id!)
                        if (newId) window.location.href = `/chat/${newId}`
                      }}
                    >
                      <GitBranch size={13} />
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={handleRegenerate}
                    title="Regenerar respuesta"
                  >
                    <RefreshCw size={13} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => msg.id ? handleCopy(msg.id, msg.content) : navigator.clipboard.writeText(msg.content)}
                    title="Copiar respuesta"
                  >
                    {msg.id && copiedId === msg.id ? <Check size={13} /> : <Copy size={13} />}
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={`h-6 w-6 ${msg.feedback === "up" ? "text-accent opacity-100" : ""}`}
                    onClick={() => handleFeedback(msg.id!, "up")}
                    title="Útil"
                  >
                    <ThumbsUp size={13} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={`h-6 w-6 ${msg.feedback === "down" ? "text-destructive opacity-100" : ""}`}
                    onClick={() => handleFeedback(msg.id!, "down")}
                    title="No útil"
                  >
                    <ThumbsDown size={13} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className={`h-6 w-6 ${savedIds.has(msg.id!) ? "text-accent opacity-100" : ""}`}
                    onClick={() => handleToggleSaved(msg.id!, msg.content)}
                    title={savedIds.has(msg.id!) ? "Quitar de guardados" : "Guardar respuesta"}
                  >
                    <Bookmark size={13} />
                  </Button>
                </div>
              )}
            </div>
          </div>
        ))}

        {queryStats && phase === "done" && (
          <div className="flex justify-start px-1">
            <span className="text-xs text-fg-subtle">
              {queryStats.ms}ms · {queryStats.sources} doc{queryStats.sources !== 1 ? "s" : ""}
            </span>
          </div>
        )}

        {phase === "streaming" && messages[messages.length - 1]?.content === "" && (
          <div className="flex justify-start">
            <div className="px-4 py-3 rounded-2xl rounded-bl-sm bg-surface border border-border">
              <Loader2 size={16} className="animate-spin text-fg-muted" />
            </div>
          </div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t border-border bg-bg">
        {error && (
          <div className="mb-3 px-3 py-2 rounded-lg bg-destructive-subtle text-destructive text-xs border border-destructive/20">
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
            className="flex-1 px-4 py-2.5 rounded-xl border border-border bg-bg text-fg text-sm placeholder:text-fg-subtle outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:border-accent disabled:opacity-50 transition-colors"
          />
          <Button
            type="submit"
            size="icon"
            className="h-10 w-10 rounded-xl shrink-0"
            disabled={!input.trim() || phase === "streaming"}
          >
            {phase === "streaming"
              ? <Loader2 size={16} className="animate-spin" />
              : <Send size={16} />
            }
          </Button>
        </form>
      </div>
    </div>
  )
}
