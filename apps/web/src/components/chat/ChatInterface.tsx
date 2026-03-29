"use client"

import { useState, useRef, useEffect, useCallback, useTransition } from "react"
import { Send, ThumbsUp, ThumbsDown, Loader2, Copy, Check } from "lucide-react"
import { useChat } from "@ai-sdk/react"
import { DefaultChatTransport, type UIMessage } from "ai"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback } from "@/app/actions/chat"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { SourcesPanel } from "@/components/chat/SourcesPanel"
import { MarkdownMessage } from "@/components/chat/MarkdownMessage"
import { ArtifactPanel, type Artifact } from "@/components/chat/ArtifactPanel"
import type { Citation } from "@rag-saldivia/shared"

// ── Helpers ──

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
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const [artifacts, setArtifacts] = useState<Artifact[]>([])
  const [activeArtifactIndex, setActiveArtifactIndex] = useState(0)
  const [_isPending, startTransition] = useTransition()
  const bottomRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

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

  // Solo auto-scroll si el usuario está cerca del fondo (no si scrolleó arriba)
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const container = scrollContainerRef.current
    if (!container) return
    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 150
    if (isNearBottom) {
      requestAnimationFrame(() => {
        bottomRef.current?.scrollIntoView({ behavior: "instant" })
      })
    }
  }, [messages])

  useEffect(() => {
    const ta = textareaRef.current
    if (ta) {
      ta.style.height = "auto"
      ta.style.height = `${Math.min(ta.scrollHeight, 200)}px`
    }
  }, [input])

  const handleSend = useCallback(async () => {
    const query = input.trim()
    if (!query || isStreaming) return
    setInput("")
    await sendMessage({ text: query })
  }, [input, isStreaming, sendMessage])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }, [handleSend])

  const handleCopy = useCallback(async (content: string, msgId: string) => {
    await navigator.clipboard.writeText(content)
    setCopiedId(msgId)
    setTimeout(() => setCopiedId(null), 2000)
  }, [])

  const handleFeedback = useCallback(async (messageId: number, rating: "up" | "down") => {
    await actionAddFeedback(messageId, rating)
  }, [])

  return (
    <div className="flex-1 flex min-h-0">
    <div className="flex-1 flex flex-col min-h-0 bg-bg">
      <h1 className="sr-only">Chat — {session.collection}</h1>
      {/* Messages */}
      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto">
        {messages.length === 0 && (
          <div className="h-full flex items-center justify-center">
            <div className="flex flex-col items-center" style={{ gap: "16px" }}>
              <div
                className="flex items-center justify-center rounded-2xl bg-accent"
                style={{ width: "48px", height: "48px" }}
              >
                <span className="text-xl font-bold text-accent-fg select-none">S</span>
              </div>
              <div className="text-center" style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
                <h1 className="text-lg font-semibold text-fg">¿En qué puedo ayudarte?</h1>
                <p className="text-sm text-fg-subtle">
                  Preguntá sobre los documentos de {session.collection}
                </p>
              </div>
            </div>
          </div>
        )}

        {messages.length > 0 && (
          <div className="max-w-3xl mx-auto" style={{ padding: "24px 24px 0" }}>
            {messages.map((msg) => {
              const text = getMessageText(msg)
              const sources = getMessageSources(msg)
              const numId = Number(msg.id) || undefined
              const isUser = msg.role === "user"

              return (
                <div
                  key={msg.id}
                  className="group"
                  style={{ marginBottom: "32px" }}
                >
                  {isUser ? (
                    /* ── User message ── */
                    <div
                      className="rounded-2xl text-sm text-fg leading-relaxed whitespace-pre-wrap"
                      style={{
                        background: "var(--surface)",
                        padding: "16px 20px",
                      }}
                    >
                      {text}
                    </div>
                  ) : (
                    /* ── Assistant message ── */
                    <div className="flex" style={{ gap: "12px" }}>
                      {/* Avatar */}
                      <div
                        className="shrink-0 flex items-center justify-center rounded-full bg-accent"
                        style={{ width: "28px", height: "28px", marginTop: "2px" }}
                      >
                        <span className="text-xs font-bold text-accent-fg select-none">S</span>
                      </div>

                      {/* Content */}
                      <div className="flex-1 min-w-0">
                        <MarkdownMessage content={text} onOpenArtifact={(a) => {
                          setArtifacts(prev => {
                            const exists = prev.findIndex(p => p.content === a.content)
                            if (exists >= 0) {
                              setActiveArtifactIndex(exists)
                              return prev
                            }
                            setActiveArtifactIndex(prev.length)
                            return [...prev, a]
                          })
                        }} />

                        {sources.length > 0 && (
                          <div style={{ marginTop: "12px" }}>
                            <SourcesPanel sources={sources} />
                          </div>
                        )}

                        {/* Actions on hover */}
                        {text && !isStreaming && (
                          <div
                            className="flex opacity-0 group-hover:opacity-100 transition-opacity"
                            style={{ gap: "2px", marginTop: "8px" }}
                          >
                            <button
                              onClick={() => handleCopy(text, msg.id)}
                              className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
                              title="Copiar"
                            >
                              {copiedId === msg.id ? <Check size={14} /> : <Copy size={14} />}
                            </button>
                            {numId && (
                              <>
                                <button
                                  onClick={() => handleFeedback(numId, "up")}
                                  className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
                                  title="Útil"
                                >
                                  <ThumbsUp size={14} />
                                </button>
                                <button
                                  onClick={() => handleFeedback(numId, "down")}
                                  className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
                                  title="No útil"
                                >
                                  <ThumbsDown size={14} />
                                </button>
                              </>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              )
            })}

            {/* Streaming indicator */}
            {(() => {
              const last = messages[messages.length - 1]
              return isStreaming && last && last.role === "assistant" && getMessageText(last) === ""
            })() && (
              <div className="flex" style={{ gap: "12px", marginBottom: "32px" }}>
                <div
                  className="shrink-0 flex items-center justify-center rounded-full bg-accent"
                  style={{ width: "28px", height: "28px" }}
                >
                  <span className="text-xs font-bold text-accent-fg select-none">S</span>
                </div>
                <Loader2 size={16} className="animate-spin text-fg-muted" style={{ marginTop: "6px" }} />
              </div>
            )}

            <div ref={bottomRef} />
          </div>
        )}
      </div>

      {/* Input area */}
      <div style={{ padding: "12px 24px 24px" }}>
        <div className="max-w-3xl mx-auto">
          {error && (
            <div
              className="text-sm text-destructive"
              style={{
                marginBottom: "12px",
                padding: "12px 16px",
                borderRadius: "12px",
                background: "var(--destructive-subtle)",
                border: "1px solid color-mix(in srgb, var(--destructive) 20%, transparent)",
              }}
            >
              {error.message}
            </div>
          )}

          <div
            className="border border-border rounded-2xl bg-bg transition-colors focus-within:border-accent"
            style={{ padding: "12px 16px" }}
          >
            <textarea
              ref={textareaRef}
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Respondeme..."
              disabled={isStreaming}
              rows={1}
              className="w-full resize-none bg-transparent text-fg text-sm placeholder:text-fg-subtle outline-none disabled:opacity-50"
              style={{ minHeight: "24px", maxHeight: "200px", lineHeight: "1.5" }}
            />
            <div className="flex items-center justify-between" style={{ marginTop: "8px" }}>
              <div className="text-xs text-fg-subtle">
                {isStreaming ? (
                  <button onClick={stop} className="text-accent hover:underline">
                    Detener generación
                  </button>
                ) : (
                  <span>{session.collection}</span>
                )}
              </div>
              <button
                onClick={handleSend}
                disabled={!input.trim() || isStreaming}
                className="flex items-center justify-center rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity hover:opacity-90"
                style={{ width: "32px", height: "32px" }}
              >
                {isStreaming
                  ? <Loader2 size={16} className="animate-spin" />
                  : <Send size={16} />
                }
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    {/* Artifact panel */}
    {artifacts.length > 0 && (
      <ArtifactPanel
        artifacts={artifacts}
        activeIndex={activeArtifactIndex}
        onSelect={setActiveArtifactIndex}
        onClose={() => { setArtifacts([]); setActiveArtifactIndex(0) }}
      />
    )}
    </div>
  )
}
