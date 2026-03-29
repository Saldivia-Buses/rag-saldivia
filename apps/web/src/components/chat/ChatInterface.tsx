"use client"

import { useState, useRef, useEffect, useCallback, useTransition } from "react"
import { ThumbsUp, ThumbsDown, Copy, Check, RotateCcw, Square, Plus, ChevronDown } from "lucide-react"
import { useChat } from "@ai-sdk/react"
import { DefaultChatTransport, type UIMessage } from "ai"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback, actionRenameSession } from "@/app/actions/chat"
import { useRouter } from "next/navigation"
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

// ── Spark thinking indicator ──

function SparkIcon({ className, size = 16 }: { className?: string; size?: number }) {
  return (
    <svg
      width={size} height={size} viewBox="0 0 16 16" fill="none"
      className={className}
    >
      <path
        d="M8 1L9.5 6.5L15 8L9.5 9.5L8 15L6.5 9.5L1 8L6.5 6.5L8 1Z"
        fill="currentColor"
      />
    </svg>
  )
}

// ── Suggestion chips for empty state ──

const SUGGESTIONS = [
  { icon: "📄", label: "Buscar documentos" },
  { icon: "❓", label: "Hacer preguntas" },
  { icon: "📊", label: "Analizar datos" },
  { icon: "📝", label: "Resumir contenido" },
]

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
  const [editingTitle, setEditingTitle] = useState(false)
  const [titleDraft, setTitleDraft] = useState(session.title)
  const titleInputRef = useRef<HTMLInputElement>(null)
  const router = useRouter()
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

        // Auto-rename session from first user message
        if (allMessages.filter(m => m.role === "user").length === 1) {
          const title = userText.slice(0, 60) + (userText.length > 60 ? "..." : "")
          actionRenameSession(session.id, title)
        }
      }
    },
    onError: (err) => {
      clientLog.error(err instanceof Error ? err : new Error(String(err)))
    },
  })

  const isStreaming = status === "streaming" || status === "submitted"

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

  const handleTitleSave = useCallback(async () => {
    const newTitle = titleDraft.trim()
    if (newTitle && newTitle !== session.title) {
      await actionRenameSession(session.id, newTitle)
      router.refresh()
    }
    setEditingTitle(false)
  }, [titleDraft, session.id, session.title, router])

  const handleRetry = useCallback(async () => {
    const lastUserMsg = [...messages].reverse().find((m) => m.role === "user")
    if (!lastUserMsg || isStreaming) return
    const text = getMessageText(lastUserMsg)
    if (text) await sendMessage({ text })
  }, [messages, isStreaming, sendMessage])

  return (
    <div className="flex-1 flex min-h-0">
    <div className="flex-1 flex flex-col min-h-0 bg-bg">
      {/* ── Conversation header ── */}
      {messages.length > 0 && (
        <div
          className="shrink-0 flex items-center justify-center border-b border-border"
          style={{ height: "48px" }}
        >
          {editingTitle ? (
            <input
              ref={titleInputRef}
              value={titleDraft}
              onChange={(e) => setTitleDraft(e.target.value)}
              onBlur={handleTitleSave}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleTitleSave()
                if (e.key === "Escape") { setTitleDraft(session.title); setEditingTitle(false) }
              }}
              className="text-sm font-medium text-fg bg-transparent text-center outline-none border-b border-accent"
              style={{ maxWidth: "300px" }}
              autoFocus
            />
          ) : (
            <button
              onClick={() => { setEditingTitle(true); setTitleDraft(session.title) }}
              className="flex items-center text-sm font-medium text-fg hover:text-fg-muted transition-colors"
              style={{ gap: "4px" }}
              title="Click para renombrar"
            >
              {session.title}
              <ChevronDown size={14} className="text-fg-subtle" />
            </button>
          )}
        </div>
      )}

      <h1 className="sr-only">Chat — {session.collection}</h1>

      {/* ── Messages ── */}
      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto">
        {/* Empty state */}
        {messages.length === 0 && (
          <div className="h-full flex flex-col items-center justify-center" style={{ padding: "0 24px" }}>
            <div className="flex items-center justify-center" style={{ gap: "12px", marginBottom: "32px" }}>
              <SparkIcon className="text-accent" size={28} />
              <h1
                className="font-semibold text-fg text-center"
                style={{ fontSize: "40px", lineHeight: "1.1", letterSpacing: "-0.03em" }}
              >
                ¿En qué pensamos?
              </h1>
            </div>

            {/* Input in empty state — centered */}
            <div className="w-full" style={{ maxWidth: "640px" }}>
              <div
                className="border border-border rounded-2xl bg-bg transition-colors focus-within:border-accent"
                style={{ padding: "12px 16px" }}
              >
                <textarea
                  ref={textareaRef}
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="¿Cómo puedo ayudarte hoy?"
                  rows={1}
                  className="w-full resize-none bg-transparent text-fg text-sm placeholder:text-fg-subtle outline-none"
                  style={{ minHeight: "24px", maxHeight: "200px", lineHeight: "1.5" }}
                />
                <div className="flex items-center justify-between" style={{ marginTop: "8px" }}>
                  <button
                    type="button"
                    className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
                    style={{ width: "28px", height: "28px" }}
                    title="Adjuntar"
                  >
                    <Plus size={16} />
                  </button>
                  <div className="flex items-center" style={{ gap: "8px" }}>
                    <span className="text-xs text-fg-subtle">{session.collection}</span>
                    <button
                      onClick={handleSend}
                      disabled={!input.trim()}
                      className="flex items-center justify-center rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity hover:opacity-90"
                      style={{ width: "32px", height: "32px" }}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <line x1="22" y1="2" x2="11" y2="13" />
                        <polygon points="22 2 15 22 11 13 2 9 22 2" />
                      </svg>
                    </button>
                  </div>
                </div>
              </div>

              {/* Suggestion chips */}
              <div className="flex items-center justify-center flex-wrap" style={{ gap: "8px", marginTop: "16px" }}>
                {SUGGESTIONS.map((s) => (
                  <button
                    key={s.label}
                    onClick={() => {
                      setInput(s.label)
                      textareaRef.current?.focus()
                    }}
                    className="flex items-center border border-border rounded-full text-sm text-fg-muted hover:text-fg hover:bg-surface transition-colors"
                    style={{ padding: "6px 14px", gap: "6px" }}
                  >
                    <span>{s.icon}</span>
                    <span>{s.label}</span>
                  </button>
                ))}
              </div>
            </div>

            {/* Disclaimer */}
            <p className="text-xs text-fg-subtle text-center" style={{ marginTop: "24px" }}>
              Saldivia RAG es IA y puede cometer errores. Por favor, verificá las respuestas.
            </p>
          </div>
        )}

        {/* Messages list */}
        {messages.length > 0 && (
          <div className="max-w-3xl mx-auto" style={{ padding: "24px 24px 0" }}>
            {messages.map((msg) => {
              const text = getMessageText(msg)
              const sources = getMessageSources(msg)
              const numId = Number(msg.id) || undefined
              const isUser = msg.role === "user"
              const isLastAssistant = !isUser && msg === messages[messages.length - 1] && isStreaming

              return (
                <div
                  key={msg.id}
                  className="group"
                  style={{ marginBottom: "24px" }}
                >
                  {isUser ? (
                    /* ── User message: right-aligned bubble ── */
                    <div className="flex justify-end">
                      <div
                        className="rounded-2xl text-sm text-fg leading-relaxed whitespace-pre-wrap"
                        style={{
                          background: "var(--surface-2)",
                          padding: "12px 16px",
                          maxWidth: "85%",
                        }}
                      >
                        {text}
                      </div>
                    </div>
                  ) : (
                    /* ── Assistant message: no avatar, no bubble ── */
                    <div className="min-w-0">
                      {/* Thinking indicator */}
                      {isLastAssistant && text === "" && (
                        <div className="flex items-center" style={{ gap: "8px", marginBottom: "8px" }}>
                          <SparkIcon className="text-accent animate-pulse" size={20} />
                          <span className="text-sm text-fg-muted italic">Pensando...</span>
                        </div>
                      )}

                      {/* Content */}
                      {text && (
                        <>
                          {isLastAssistant ? (
                            <div className="text-sm text-fg leading-relaxed whitespace-pre-wrap">{text}</div>
                          ) : (
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
                          )}
                        </>
                      )}

                      {sources.length > 0 && (
                        <div style={{ marginTop: "12px" }}>
                          <SourcesPanel sources={sources} />
                        </div>
                      )}

                      {/* Actions — always visible like Claude */}
                      {text && !isStreaming && (
                        <div
                          className="flex"
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
                          <button
                            onClick={handleRetry}
                            className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
                            title="Reintentar"
                          >
                            <RotateCcw size={14} />
                          </button>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}

            {/* Streaming indicator when assistant message hasn't started yet */}
            {(() => {
              const last = messages[messages.length - 1]
              return isStreaming && last && last.role !== "assistant"
            })() && (
              <div className="flex items-center" style={{ gap: "8px", marginBottom: "24px" }}>
                <SparkIcon className="text-accent animate-pulse" />
                <span className="text-sm text-fg-muted italic">Pensando...</span>
              </div>
            )}

            <div ref={bottomRef} />
          </div>
        )}
      </div>

      {/* ── Input area (only when there are messages) ── */}
      {messages.length > 0 && (
        <div style={{ padding: "12px 24px 16px" }}>
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
                ref={messages.length > 0 ? textareaRef : undefined}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Responder..."
                disabled={isStreaming}
                rows={1}
                className="w-full resize-none bg-transparent text-fg text-sm placeholder:text-fg-subtle outline-none disabled:opacity-50"
                style={{ minHeight: "24px", maxHeight: "200px", lineHeight: "1.5" }}
              />
              <div className="flex items-center justify-between" style={{ marginTop: "8px" }}>
                <button
                  type="button"
                  className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
                  style={{ width: "28px", height: "28px" }}
                  title="Adjuntar"
                >
                  <Plus size={16} />
                </button>
                <div className="flex items-center" style={{ gap: "8px" }}>
                  <span className="text-xs text-fg-subtle">{session.collection}</span>
                  {isStreaming ? (
                    <button
                      onClick={stop}
                      className="flex items-center justify-center rounded-full border border-border text-fg-muted hover:text-fg hover:border-fg-subtle transition-colors"
                      style={{ width: "32px", height: "32px" }}
                      title="Detener"
                    >
                      <Square size={12} fill="currentColor" />
                    </button>
                  ) : (
                    <button
                      onClick={handleSend}
                      disabled={!input.trim()}
                      className="flex items-center justify-center rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity hover:opacity-90"
                      style={{ width: "32px", height: "32px" }}
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <line x1="22" y1="2" x2="11" y2="13" />
                        <polygon points="22 2 15 22 11 13 2 9 22 2" />
                      </svg>
                    </button>
                  )}
                </div>
              </div>
            </div>

            {/* Disclaimer */}
            <p className="text-xs text-fg-subtle text-center" style={{ marginTop: "8px" }}>
              Saldivia RAG es IA y puede cometer errores. Por favor, verificá las respuestas.
            </p>
          </div>
        </div>
      )}
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
