"use client"

import { useState, useRef, useEffect, useCallback, useTransition, useMemo } from "react"
import { useLocalStorage } from "@/hooks/useLocalStorage"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"
import { useAutoResize } from "@/hooks/useAutoResize"
import { ThumbsUp, ThumbsDown, Copy, Check, RotateCcw, ChevronDown, ArrowDown, PanelRightClose } from "lucide-react"
import { useChat } from "@ai-sdk/react"
import { DefaultChatTransport, type UIMessage } from "ai"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback, actionRenameSession } from "@/app/actions/chat"
import { useRouter } from "next/navigation"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { SourcesPanel } from "@/components/chat/SourcesPanel"
import { MarkdownMessage } from "@/components/chat/MarkdownMessage"
import { ArtifactPanel } from "@/components/chat/ArtifactPanel"
import type { ParsedArtifact } from "@/lib/rag/artifact-parser"
import { extractArtifacts, extractCodeBlocks, extractStreamingArtifact, stripArtifactTags } from "@/lib/rag/artifact-parser"
import type { Citation } from "@rag-saldivia/shared"
import { ChatInputBar } from "./ChatInputBar"
import { CollectionSelector } from "./CollectionSelector"
import { useSidebar } from "./ChatLayout"
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip"

// ── Helpers ──

function dbToUIMessages(session: DbChatSession & { messages?: DbChatMessage[] }): UIMessage[] {
  return (session.messages ?? []).map((m, i) => ({
    id: String(m.id ?? `temp-${i}`),
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

// ── Fallback suggestions when no templates are loaded from DB ──

const FALLBACK_SUGGESTIONS = [
  { title: "Buscar documentos", prompt: "Buscá información sobre " },
  { title: "Hacer preguntas", prompt: "¿Qué es " },
  { title: "Analizar datos", prompt: "Analizá los datos sobre " },
  { title: "Resumir contenido", prompt: "Hacé un resumen de " },
]

/** Icon map for template titles — matches keywords to emoji */
function getTemplateIcon(title: string): string {
  const lower = title.toLowerCase()
  if (lower.includes("buscar") || lower.includes("documento")) return "📄"
  if (lower.includes("resumir") || lower.includes("resumen")) return "📝"
  if (lower.includes("analizar") || lower.includes("dato")) return "📊"
  if (lower.includes("comparar") || lower.includes("alternativa")) return "⚖️"
  if (lower.includes("técnic") || lower.includes("explicar")) return "🔧"
  if (lower.includes("pregunta") || lower.includes("frecuente")) return "❓"
  return "💬"
}

/** Template type from DB */
type PromptTemplate = { id: number; title: string; prompt: string; focusMode: string }

// ── Shared icon button with tooltip ──
const ICON_BTN = "flex items-center justify-center rounded-lg transition-colors"
const ICON_BTN_SIZE = { width: "42px", height: "42px" } as const
const ICON_PX = 18

function TipBtn({ label, onClick, disabled, className, style, children }: {
  label: string
  onClick?: () => void
  disabled?: boolean
  className?: string
  style?: React.CSSProperties
  children: React.ReactNode
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          onClick={onClick}
          disabled={disabled}
          className={className}
          style={style}
        >
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom" sideOffset={8}>{label}</TooltipContent>
    </Tooltip>
  )
}

// ── Component ──

export function ChatInterface({
  session,
  userId: _userId,
  templates = [],
  availableCollections = [],
}: {
  session: DbChatSession & { messages?: DbChatMessage[] }
  userId: number
  templates?: PromptTemplate[]
  availableCollections?: string[]
}) {
  const { open: sidebarOpen, toggle: toggleSidebar } = useSidebar()
  const [input, setInput] = useState("")
  const { copy, copiedKey } = useCopyToClipboard()
  const [activeArtifactId, setActiveArtifactId] = useState<string | null>(null)
  const [showArtifactPanel, setShowArtifactPanel] = useState(false)
  // Store artifacts opened from MarkdownMessage (which creates its own ids)
  const [adhocArtifacts, setAdhocArtifacts] = useState<ParsedArtifact[]>([])
  const [selectedCollections, setSelectedCollections] = useState<string[]>([session.collection])
  const [artifactPanelWidth, setArtifactPanelWidth] = useLocalStorage("saldivia-artifact-width", 480)
  const [isResizingPanel, setIsResizingPanel] = useState(false)
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
        collection_name: selectedCollections[0] ?? session.collection,
        collection_names: selectedCollections,
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
          await Promise.all([
            actionAddMessage({ sessionId: session.id, role: "user", content: userText }),
            actionAddMessage({
              sessionId: session.id,
              role: "assistant",
              content: assistantText,
              sources,
            }),
          ])
        })

        clientLog.action("rag.query", { collection: session.collection, sessionId: session.id })

        // Auto-rename session from first user message
        if (allMessages.filter(m => m.role === "user").length === 1) {
          const title = userText.slice(0, 60) + (userText.length > 60 ? "..." : "")
          actionRenameSession({ id: session.id, title })
        }
      }
    },
    onError: (err) => {
      clientLog.error(err instanceof Error ? err : new Error(String(err)))
    },
  })

  const isStreaming = status === "streaming" || status === "submitted"

  // Derive artifacts from all messages
  const allArtifacts = useMemo(() => {
    const result: ParsedArtifact[] = []
    for (const msg of messages) {
      if (msg.role !== "assistant") continue
      const text = getMessageText(msg)
      const { artifacts: tagArtifacts } = extractArtifacts(text)
      if (tagArtifacts.length > 0) {
        result.push(...tagArtifacts)
      } else {
        result.push(...extractCodeBlocks(text))
      }
    }
    return result
  }, [messages])

  // Detect streaming artifact (partial, still being generated)
  const streamingArtifact = useMemo(() => {
    if (!isStreaming) return null
    const lastMsg = messages[messages.length - 1]
    if (!lastMsg || lastMsg.role !== "assistant") return null
    return extractStreamingArtifact(getMessageText(lastMsg))
  }, [messages, isStreaming])

  // Combined artifacts for the panel (includes adhoc from MarkdownMessage clicks)
  const panelArtifacts = useMemo(() => {
    const all = [...allArtifacts]
    // Add adhoc artifacts that aren't already in allArtifacts
    for (const a of adhocArtifacts) {
      if (!all.some((x) => x.id === a.id)) all.push(a)
    }
    if (streamingArtifact) all.push(streamingArtifact)
    return all
  }, [allArtifacts, adhocArtifacts, streamingArtifact])

  // Clear adhoc artifacts when session changes
  useEffect(() => {
    setAdhocArtifacts([])
    setActiveArtifactId(null)
    setShowArtifactPanel(false)
  }, [session.id])

  // Auto-open panel when streaming artifact starts
  useEffect(() => {
    if (streamingArtifact && !showArtifactPanel) {
      setActiveArtifactId(streamingArtifact.id)
      setShowArtifactPanel(true)
      if (sidebarOpen) toggleSidebar()
    }
  }, [streamingArtifact, showArtifactPanel, sidebarOpen, toggleSidebar])

  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const [showScrollButton, setShowScrollButton] = useState(false)

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

  // Track scroll position for scroll-to-bottom button
  useEffect(() => {
    const container = scrollContainerRef.current
    if (!container) return
    function onScroll() {
      const gap = container!.scrollHeight - container!.scrollTop - container!.clientHeight
      setShowScrollButton(gap > 300)
    }
    container.addEventListener("scroll", onScroll, { passive: true })
    return () => container.removeEventListener("scroll", onScroll)
  }, [])

  const scrollToBottom = useCallback(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [])

  useAutoResize(textareaRef, input)

  async function handleSend() {
    const query = input.trim()
    if (!query || isStreaming) return
    setInput("")
    await sendMessage({ text: query })
    // Return focus to textarea after sending
    requestAnimationFrame(() => textareaRef.current?.focus())
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleCopy = useCallback(async (content: string, msgId: string) => {
    await copy(content, msgId)
  }, [copy])

  const handleFeedback = useCallback(async (messageId: number, rating: "up" | "down") => {
    await actionAddFeedback({ messageId, rating })
  }, [])

  const handleTitleSave = useCallback(async () => {
    const newTitle = titleDraft.trim()
    if (newTitle && newTitle !== session.title) {
      await actionRenameSession({ id: session.id, title: newTitle })
      router.refresh()
    }
    setEditingTitle(false)
  }, [titleDraft, session.id, session.title, router])

  function handleOpenArtifact(a: ParsedArtifact) {
    // Register adhoc artifacts (e.g. from MarkdownMessage code block detection)
    // so the panel can find them by id
    setAdhocArtifacts((prev) => {
      if (prev.some((x) => x.id === a.id)) return prev
      return [...prev, a]
    })
    setActiveArtifactId(a.id)
    setShowArtifactPanel(true)
    if (sidebarOpen) toggleSidebar()
  }

  const handleRetry = useCallback(async () => {
    const lastUserMsg = [...messages].reverse().find((m) => m.role === "user")
    if (!lastUserMsg || isStreaming) return
    const text = getMessageText(lastUserMsg)
    if (text) await sendMessage({ text })
  }, [messages, isStreaming, sendMessage])

  return (
    <TooltipProvider delayDuration={200}>
    <div className="flex-1 flex min-h-0">
    <div className="flex-1 flex flex-col min-h-0 bg-bg">
      {/* ── Conversation header ── */}
      {messages.length > 0 && (
        <div
          className="shrink-0 flex items-center border-b border-border"
          style={{ height: "48px", padding: "0 8px" }}
        >
          {/* Left spacer (sidebar toggle is in NavRail) */}
          <div style={{ width: "42px" }} />

          {/* Center: title */}
          <div className="flex-1 flex items-center justify-center">
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

          {/* Right: artifact panel toggle */}
          <div style={{ width: "42px" }}>
            {panelArtifacts.length > 0 && (
              <TipBtn label={showArtifactPanel ? "Cerrar artifacts" : "Abrir artifacts"} onClick={() => {
                const next = !showArtifactPanel
                setShowArtifactPanel(next)
                // Auto-select first artifact if none selected
                if (next && !activeArtifactId && panelArtifacts.length > 0) {
                  setActiveArtifactId(panelArtifacts[0]!.id)
                }
              }} className={`${ICON_BTN} hover:bg-surface-2 ${showArtifactPanel ? "text-accent" : "text-fg-muted hover:text-fg"}`} style={ICON_BTN_SIZE}>
                <PanelRightClose size={ICON_PX} />
              </TipBtn>
            )}
          </div>
        </div>
      )}

      <h1 className="sr-only">Chat — {session.collection}</h1>

      {/* ── Messages ── */}
      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto w-full">
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
              <ChatInputBar
                value={input}
                onChange={setInput}
                onKeyDown={handleKeyDown}
                onSend={handleSend}
                textareaRef={textareaRef}
                placeholder="¿Cómo puedo ayudarte hoy?"
                collection={session.collection}
                collectionSlot={availableCollections.length > 1 ? (
                  <CollectionSelector
                    defaultCollection={session.collection}
                    availableCollections={availableCollections}
                    onCollectionsChange={setSelectedCollections}
                  />
                ) : undefined}
              />

              {/* Prompt template chips — loaded from DB, fallback to defaults */}
              <div className="flex items-center justify-center flex-wrap" style={{ gap: "8px", marginTop: "16px" }}>
                {(templates.length > 0 ? templates : FALLBACK_SUGGESTIONS).map((t) => (
                  <button
                    key={t.title}
                    onClick={() => {
                      setInput(t.prompt)
                      textareaRef.current?.focus()
                    }}
                    className="flex items-center border border-border rounded-full text-sm text-fg-muted hover:text-fg hover:bg-surface transition-colors"
                    style={{ padding: "6px 14px", gap: "6px" }}
                  >
                    <span>{getTemplateIcon(t.title)}</span>
                    <span>{t.title}</span>
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
          <div role="log" aria-live="polite" aria-label="Mensajes del chat" style={{ padding: "24px 24px 0", maxWidth: "768px", marginLeft: "auto", marginRight: "auto" }}>
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
                            <div className="text-sm text-fg leading-relaxed whitespace-pre-wrap">
                              {stripArtifactTags(text)}
                            </div>
                          ) : (
                            <MarkdownMessage content={text} onOpenArtifact={handleOpenArtifact} />
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
                        <div className="flex" style={{ gap: "2px", marginTop: "8px" }}>
                          <TipBtn label="Copiar" onClick={() => handleCopy(text, msg.id)} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                            {copiedKey === msg.id ? <Check size={ICON_PX} /> : <Copy size={ICON_PX} />}
                          </TipBtn>
                          {numId && (
                            <>
                              <TipBtn label="Útil" onClick={() => handleFeedback(numId, "up")} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                                <ThumbsUp size={ICON_PX} />
                              </TipBtn>
                              <TipBtn label="No útil" onClick={() => handleFeedback(numId, "down")} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                                <ThumbsDown size={ICON_PX} />
                              </TipBtn>
                            </>
                          )}
                          <TipBtn label="Reintentar" onClick={handleRetry} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                            <RotateCcw size={ICON_PX} />
                          </TipBtn>
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

      {/* ── Scroll to bottom ── */}
      {showScrollButton && messages.length > 0 && (
        <div className="flex justify-center" style={{ position: "relative" }}>
          <button
            onClick={scrollToBottom}
            className="absolute flex items-center justify-center rounded-full border border-border bg-bg text-fg-muted hover:text-fg hover:bg-surface shadow-md transition-all"
            style={{ width: "36px", height: "36px", bottom: "8px" }}
            title="Ir al final"
          >
            <ArrowDown size={ICON_PX} />
          </button>
        </div>
      )}

      {/* ── Input area (only when there are messages) ── */}
      {messages.length > 0 && (
        <div style={{ padding: "12px 24px 16px" }}>
          <div style={{ maxWidth: "768px", marginLeft: "auto", marginRight: "auto" }}>
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

            <ChatInputBar
              value={input}
              onChange={setInput}
              onKeyDown={handleKeyDown}
              onSend={handleSend}
              onStop={stop}
              textareaRef={textareaRef}
              placeholder="Responder..."
              collection={session.collection}
              collectionSlot={availableCollections.length > 1 ? (
                <CollectionSelector
                  defaultCollection={session.collection}
                  availableCollections={availableCollections}
                  onCollectionsChange={setSelectedCollections}
                  disabled={isStreaming}
                />
              ) : undefined}
              isStreaming={isStreaming}
              disabled={isStreaming}
            />

            {/* Disclaimer */}
            <p className="text-xs text-fg-subtle text-center" style={{ marginTop: "8px" }}>
              Saldivia RAG es IA y puede cometer errores. Por favor, verificá las respuestas.
            </p>
          </div>
        </div>
      )}
    </div>

    {/* Artifact panel — always rendered when artifacts exist for smooth transitions */}
    {panelArtifacts.length > 0 && (
      <ArtifactPanel
        artifacts={panelArtifacts}
        activeId={activeArtifactId}
        onSelect={(id) => { setActiveArtifactId(id); setShowArtifactPanel(true) }}
        onClose={() => setShowArtifactPanel(false)}
        panelWidth={showArtifactPanel && activeArtifactId ? artifactPanelWidth : 0}
        onWidthChange={setArtifactPanelWidth}
        isResizing={isResizingPanel}
        onResizeStart={() => setIsResizingPanel(true)}
        onResizeEnd={() => setIsResizingPanel(false)}
      />
    )}
    </div>
    </TooltipProvider>
  )
}
