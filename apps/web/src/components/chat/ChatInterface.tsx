"use client"

import { useState, useRef, useEffect, useCallback, useTransition } from "react"
import { Send, ThumbsUp, ThumbsDown, Loader2, Bookmark, Copy, Check, GitBranch } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useChat } from "@ai-sdk/react"
import { DefaultChatTransport, type UIMessage } from "ai"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback, actionToggleSaved, actionForkSession } from "@/app/actions/chat"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { SourcesPanel } from "@/components/chat/SourcesPanel"
import type { Citation } from "@rag-saldivia/shared"

// ── Helpers: convert between DB messages and AI SDK UIMessage ──

function dbToUIMessages(session: DbChatSession & { messages?: DbChatMessage[] }): UIMessage[] {
  return (session.messages ?? []).map((m) => ({
    id: String(m.id ?? Math.random()),
    role: m.role as "user" | "assistant",
    parts: [{ type: "text" as const, text: m.content }],
    createdAt: new Date(m.timestamp ?? Date.now()),
  }))
}

function getMessageText(msg: UIMessage): string {
  return msg.parts
    .filter((p): p is { type: "text"; text: string } => p.type === "text")
    .map((p) => p.text)
    .join("")
}

function getMessageSources(msg: UIMessage): Citation[] {
  return msg.parts
    .filter((p): p is { type: `data-${string}`; data: { citations: Citation[] } } =>
      p.type === "data-sources"
    )
    .flatMap((p) => p.data.citations)
}

// ── Component ──

export function ChatInterface({
  session,
  userId: _userId,
}: {
  session: DbChatSession & { messages?: DbChatMessage[] }
  userId: number
}) {
  const [input, setInput] = useState("")
  const [savedIds, setSavedIds] = useState<Set<number>>(new Set())
  const [copiedId, setCopiedId] = useState<number | null>(null)
  const [_isPending, startTransition] = useTransition()
  const bottomRef = useRef<HTMLDivElement>(null)

  const { messages, sendMessage, status, error, stop } = useChat({
    id: session.id,
    transport: new DefaultChatTransport({
      api: "/api/rag/generate",
      body: {
        collection_name: session.collection,
        collection_names: [session.collection],
        session_id: session.id,
        use_knowledge_base: true,
        focus_mode: "detallado",
      },
    }),
    messages: dbToUIMessages(session),
    onFinish: ({ messages: allMessages }) => {
      const lastAssistant = [...allMessages].reverse().find((m) => m.role === "assistant")
      const lastUser = [...allMessages].reverse().find((m) => m.role === "user")

      if (lastUser && lastAssistant) {
        const userText = getMessageText(lastUser)
        const assistantText = getMessageText(lastAssistant)
        const sources = getMessageSources(lastAssistant)

        startTransition(async () => {
          await actionAddMessage({ sessionId: session.id, role: "user", content: userText })
          await actionAddMessage({
            sessionId: session.id,
            role: "assistant",
            content: assistantText,
            sources,
          })
        })

        clientLog.action("rag.query", { collection: session.collection, sessionId: session.id })
      }
    },
    onError: (err) => {
      clientLog.error(err instanceof Error ? err : new Error(String(err)))
    },
  })

  const isStreaming = status === "streaming" || status === "submitted"

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  const handleSend = useCallback(async () => {
    const query = input.trim()
    if (!query || isStreaming) return
    setInput("")
    await sendMessage({ text: query })
  }, [input, isStreaming, sendMessage])

  const handleCopy = useCallback(async (content: string, msgId?: string) => {
    await navigator.clipboard.writeText(content)
    if (msgId) {
      setCopiedId(Number(msgId))
      setTimeout(() => setCopiedId(null), 2000)
    }
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
  }, [])

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Header */}
      <div className="flex items-center justify-between px-6 py-3 border-b border-border shrink-0 no-print">
        <span className="text-sm font-medium text-fg-muted truncate">
          {session.collection}
        </span>
        {isStreaming && (
          <Button variant="ghost" size="sm" onClick={stop} className="text-xs">
            Detener
          </Button>
        )}
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

        {messages.map((msg) => {
          const text = getMessageText(msg)
          const sources = getMessageSources(msg)
          const numId = Number(msg.id) || undefined

          return (
            <div
              key={msg.id}
              className={`flex group ${msg.role === "user" ? "justify-end" : "justify-start"}`}
            >
              <div
                className={`max-w-2xl rounded-2xl px-4 py-3 text-sm space-y-1 ${
                  msg.role === "user"
                    ? "rounded-br-sm bg-accent text-accent-fg"
                    : "rounded-bl-sm bg-surface text-fg border border-border"
                }`}
              >
                <p className="whitespace-pre-wrap leading-relaxed">{text}</p>

                {msg.role === "assistant" && sources.length > 0 && (
                  <SourcesPanel sources={sources} />
                )}

                {msg.role === "assistant" && text && !isStreaming && (
                  <div className="flex gap-1 pt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    {numId && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6"
                        title="Bifurcar desde aquí"
                        onClick={async () => {
                          const newId = await actionForkSession(session.id, numId)
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
                      onClick={() => handleCopy(text, msg.id)}
                      title="Copiar respuesta"
                    >
                      {numId && copiedId === numId ? <Check size={13} /> : <Copy size={13} />}
                    </Button>
                    {numId && (
                      <>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-6 w-6"
                          onClick={() => handleFeedback(numId, "up")}
                          title="Útil"
                        >
                          <ThumbsUp size={13} />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-6 w-6"
                          onClick={() => handleFeedback(numId, "down")}
                          title="No útil"
                        >
                          <ThumbsDown size={13} />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className={`h-6 w-6 ${savedIds.has(numId) ? "text-accent opacity-100" : ""}`}
                          onClick={() => handleToggleSaved(numId, text)}
                          title={savedIds.has(numId) ? "Quitar de guardados" : "Guardar respuesta"}
                        >
                          <Bookmark size={13} />
                        </Button>
                      </>
                    )}
                  </div>
                )}
              </div>
            </div>
          )
        })}

        {(() => { const last = messages[messages.length - 1]; return isStreaming && last && last.role === "assistant" && getMessageText(last) === "" })() && (
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
            {error.message}
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
            disabled={isStreaming}
            className="flex-1 px-4 py-2.5 rounded-xl border border-border bg-bg text-fg text-sm placeholder:text-fg-subtle outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:border-accent disabled:opacity-50 transition-colors"
          />
          <Button
            type="submit"
            size="icon"
            className="h-10 w-10 rounded-xl shrink-0"
            disabled={!input.trim() || isStreaming}
          >
            {isStreaming
              ? <Loader2 size={16} className="animate-spin" />
              : <Send size={16} />
            }
          </Button>
        </form>
      </div>
    </div>
  )
}
