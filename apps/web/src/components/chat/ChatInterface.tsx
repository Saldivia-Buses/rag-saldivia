"use client"

import { useState, useRef, useEffect, useCallback, useTransition, useMemo } from "react"
import { useLocalStorage } from "@/hooks/useLocalStorage"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"
import { useAutoResize } from "@/hooks/useAutoResize"
import { ChevronDown, ArrowDown, PanelRightClose } from "lucide-react"
import { useChat } from "@ai-sdk/react"
import { DefaultChatTransport, type UIMessage } from "ai"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"
import { actionAddMessage, actionAddFeedback, actionRenameSession } from "@/app/actions/chat"
import { useRouter } from "next/navigation"
import { clientLog } from "@rag-saldivia/logger/frontend"
import { ArtifactPanel } from "@/components/chat/ArtifactPanel"
import type { ParsedArtifact } from "@/lib/rag/artifact-parser"
import { extractArtifacts, extractCodeBlocks, extractStreamingArtifact } from "@/lib/rag/artifact-parser"
import { ChatInputBar } from "./ChatInputBar"
import { CollectionSelector } from "./CollectionSelector"
import { useSidebar } from "./ChatLayout"
import { ChatEmptyState } from "./ChatEmptyState"
import { ChatMessages, getMessageText, getMessageSources } from "./ChatMessages"
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

const ICON_BTN = "flex items-center justify-center rounded-lg transition-colors"
const ICON_BTN_SIZE = { width: "42px", height: "42px" } as const
const ICON_PX = 18

function TipBtn({ label, onClick, className, style, children }: {
  label: string; onClick?: () => void; className?: string
  style?: React.CSSProperties; children: React.ReactNode
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button type="button" onClick={onClick} className={className} style={style}>
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom" sideOffset={8}>{label}</TooltipContent>
    </Tooltip>
  )
}

type PromptTemplate = { id: number; title: string; prompt: string; focusMode: string }

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
            actionAddMessage({ sessionId: session.id, role: "assistant", content: assistantText, sources }),
          ])
        })

        clientLog.action("rag.query", { collection: session.collection, sessionId: session.id })

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

  // ── Artifacts ──
  const allArtifacts = useMemo(() => {
    const result: ParsedArtifact[] = []
    for (const msg of messages) {
      if (msg.role !== "assistant") continue
      const text = getMessageText(msg)
      const { artifacts: tagArtifacts } = extractArtifacts(text)
      result.push(...(tagArtifacts.length > 0 ? tagArtifacts : extractCodeBlocks(text)))
    }
    return result
  }, [messages])

  const streamingArtifact = useMemo(() => {
    if (!isStreaming) return null
    const lastMsg = messages[messages.length - 1]
    if (!lastMsg || lastMsg.role !== "assistant") return null
    return extractStreamingArtifact(getMessageText(lastMsg))
  }, [messages, isStreaming])

  const panelArtifacts = useMemo(() => {
    const all = [...allArtifacts]
    for (const a of adhocArtifacts) {
      if (!all.some((x) => x.id === a.id)) all.push(a)
    }
    if (streamingArtifact) all.push(streamingArtifact)
    return all
  }, [allArtifacts, adhocArtifacts, streamingArtifact])

  useEffect(() => { setAdhocArtifacts([]); setActiveArtifactId(null); setShowArtifactPanel(false) }, [session.id])

  useEffect(() => {
    if (streamingArtifact && !showArtifactPanel) {
      setActiveArtifactId(streamingArtifact.id)
      setShowArtifactPanel(true)
      if (sidebarOpen) toggleSidebar()
    }
  }, [streamingArtifact, showArtifactPanel, sidebarOpen, toggleSidebar])

  // ── Scroll ──
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const [showScrollButton, setShowScrollButton] = useState(false)

  useEffect(() => {
    const container = scrollContainerRef.current
    if (!container) return
    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 150
    if (isNearBottom) requestAnimationFrame(() => { bottomRef.current?.scrollIntoView({ behavior: "instant" }) })
  }, [messages])

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

  const scrollToBottom = useCallback(() => { bottomRef.current?.scrollIntoView({ behavior: "smooth" }) }, [])

  useAutoResize(textareaRef, input)

  // ── Handlers ──
  async function handleSend() {
    const query = input.trim()
    if (!query || isStreaming) return
    setInput("")
    await sendMessage({ text: query })
    requestAnimationFrame(() => textareaRef.current?.focus())
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend() }
  }

  const handleCopy = useCallback(async (content: string, msgId: string) => { await copy(content, msgId) }, [copy])
  const handleFeedback = useCallback(async (messageId: number, rating: "up" | "down") => { await actionAddFeedback({ messageId, rating }) }, [])
  const handleTitleSave = useCallback(async () => {
    const newTitle = titleDraft.trim()
    if (newTitle && newTitle !== session.title) {
      await actionRenameSession({ id: session.id, title: newTitle })
      router.refresh()
    }
    setEditingTitle(false)
  }, [titleDraft, session.id, session.title, router])

  function handleOpenArtifact(a: ParsedArtifact) {
    setAdhocArtifacts((prev) => prev.some((x) => x.id === a.id) ? prev : [...prev, a])
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

  // ── Render ──
  return (
    <TooltipProvider delayDuration={200}>
    <div className="flex-1 flex min-h-0">
    <div className="flex-1 flex flex-col min-h-0 bg-bg">
      {/* Header */}
      {messages.length > 0 && (
        <div className="shrink-0 flex items-center border-b border-border" style={{ height: "48px", padding: "0 8px" }}>
          <div style={{ width: "42px" }} />
          <div className="flex-1 flex items-center justify-center">
            {editingTitle ? (
              <input
                ref={titleInputRef} value={titleDraft}
                onChange={(e) => setTitleDraft(e.target.value)}
                onBlur={handleTitleSave}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleTitleSave()
                  if (e.key === "Escape") { setTitleDraft(session.title); setEditingTitle(false) }
                }}
                className="text-sm font-medium text-fg bg-transparent text-center outline-none border-b border-accent"
                style={{ maxWidth: "300px" }} autoFocus
              />
            ) : (
              <button
                onClick={() => { setEditingTitle(true); setTitleDraft(session.title) }}
                className="flex items-center text-sm font-medium text-fg hover:text-fg-muted transition-colors"
                style={{ gap: "4px" }} title="Click para renombrar"
              >
                {session.title}
                <ChevronDown size={14} className="text-fg-subtle" />
              </button>
            )}
          </div>
          <div style={{ width: "42px" }}>
            {panelArtifacts.length > 0 && (
              <TipBtn label={showArtifactPanel ? "Cerrar artifacts" : "Abrir artifacts"} onClick={() => {
                const next = !showArtifactPanel
                setShowArtifactPanel(next)
                if (next && !activeArtifactId && panelArtifacts.length > 0) setActiveArtifactId(panelArtifacts[0]!.id)
              }} className={`${ICON_BTN} hover:bg-surface-2 ${showArtifactPanel ? "text-accent" : "text-fg-muted hover:text-fg"}`} style={ICON_BTN_SIZE}>
                <PanelRightClose size={ICON_PX} />
              </TipBtn>
            )}
          </div>
        </div>
      )}

      <h1 className="sr-only">Chat — {session.collection}</h1>

      {/* Messages area */}
      <div ref={scrollContainerRef} className="flex-1 overflow-y-auto w-full">
        {messages.length === 0 && (
          <ChatEmptyState
            input={input} setInput={setInput} onKeyDown={handleKeyDown} onSend={handleSend}
            textareaRef={textareaRef} collection={session.collection}
            availableCollections={availableCollections} onCollectionsChange={setSelectedCollections}
            templates={templates}
          />
        )}

        {messages.length > 0 && (
          <ChatMessages
            messages={messages} isStreaming={isStreaming} copiedKey={copiedKey} bottomRef={bottomRef}
            onCopy={handleCopy} onFeedback={handleFeedback} onRetry={handleRetry} onOpenArtifact={handleOpenArtifact}
          />
        )}
      </div>

      {/* Scroll to bottom */}
      {showScrollButton && messages.length > 0 && (
        <div className="flex justify-center" style={{ position: "relative" }}>
          <button
            onClick={scrollToBottom}
            className="absolute flex items-center justify-center rounded-full border border-border bg-bg text-fg-muted hover:text-fg hover:bg-surface shadow-md transition-all"
            style={{ width: "36px", height: "36px", bottom: "8px" }} title="Ir al final"
          >
            <ArrowDown size={ICON_PX} />
          </button>
        </div>
      )}

      {/* Input area (when messages exist) */}
      {messages.length > 0 && (
        <div style={{ padding: "12px 24px 16px" }}>
          <div style={{ maxWidth: "768px", marginLeft: "auto", marginRight: "auto" }}>
            {error && (
              <div className="text-sm text-destructive" style={{
                marginBottom: "12px", padding: "12px 16px", borderRadius: "12px",
                background: "var(--destructive-subtle)",
                border: "1px solid color-mix(in srgb, var(--destructive) 20%, transparent)",
              }}>
                {error.message}
              </div>
            )}
            <ChatInputBar
              value={input} onChange={setInput} onKeyDown={handleKeyDown} onSend={handleSend} onStop={stop}
              textareaRef={textareaRef} placeholder="Responder..." collection={session.collection}
              collectionSlot={availableCollections.length > 1 ? (
                <CollectionSelector
                  defaultCollection={session.collection} availableCollections={availableCollections}
                  onCollectionsChange={setSelectedCollections} disabled={isStreaming}
                />
              ) : undefined}
              isStreaming={isStreaming} disabled={isStreaming}
            />
            <p className="text-xs text-fg-subtle text-center" style={{ marginTop: "8px" }}>
              Saldivia RAG es IA y puede cometer errores. Por favor, verificá las respuestas.
            </p>
          </div>
        </div>
      )}
    </div>

    {/* Artifact panel */}
    {panelArtifacts.length > 0 && (
      <ArtifactPanel
        artifacts={panelArtifacts} activeId={activeArtifactId}
        onSelect={(id) => { setActiveArtifactId(id); setShowArtifactPanel(true) }}
        onClose={() => setShowArtifactPanel(false)}
        panelWidth={showArtifactPanel && activeArtifactId ? artifactPanelWidth : 0}
        onWidthChange={setArtifactPanelWidth} isResizing={isResizingPanel}
        onResizeStart={() => setIsResizingPanel(true)} onResizeEnd={() => setIsResizingPanel(false)}
      />
    )}
    </div>
    </TooltipProvider>
  )
}
