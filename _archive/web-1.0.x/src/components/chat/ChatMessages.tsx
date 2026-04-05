"use client"

import { ThumbsUp, ThumbsDown, Copy, Check, RotateCcw } from "lucide-react"
import type { UIMessage } from "ai"
import type { ParsedArtifact } from "@/lib/rag/artifact-parser"
import { stripArtifactTags } from "@/lib/rag/artifact-parser"
import type { Citation } from "@rag-saldivia/shared"
import { SourcesPanel } from "./SourcesPanel"
import { MarkdownMessage } from "./MarkdownMessage"
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip"

function SparkIcon({ className, size = 16 }: { className?: string; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" className={className}>
      <path d="M8 1L9.5 6.5L15 8L9.5 9.5L8 15L6.5 9.5L1 8L6.5 6.5L8 1Z" fill="currentColor" />
    </svg>
  )
}

const ICON_BTN = "flex items-center justify-center rounded-lg transition-colors"
const ICON_BTN_SIZE = { width: "42px", height: "42px" } as const
const ICON_PX = 18

function TipBtn({ label, onClick, disabled, className, style, children }: {
  label: string; onClick?: () => void; disabled?: boolean
  className?: string; style?: React.CSSProperties; children: React.ReactNode
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button type="button" onClick={onClick} disabled={disabled} className={className} style={style}>
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom" sideOffset={8}>{label}</TooltipContent>
    </Tooltip>
  )
}

export function getMessageText(msg: UIMessage): string {
  return msg.parts
    .filter((p): p is { type: "text"; text: string } => p.type === "text")
    .map((p) => p.text)
    .join("")
}

export function getMessageSources(msg: UIMessage): Citation[] {
  return msg.parts
    .filter((p): p is { type: `data-${string}`; data: { citations: Citation[] } } =>
      p.type === "data-sources"
    )
    .flatMap((p) => p.data.citations)
}

export function ChatMessages({
  messages,
  isStreaming,
  copiedKey,
  bottomRef,
  onCopy,
  onFeedback,
  onRetry,
  onOpenArtifact,
}: {
  messages: UIMessage[]
  isStreaming: boolean
  copiedKey: string | null
  bottomRef: React.RefObject<HTMLDivElement | null>
  onCopy: (content: string, msgId: string) => void
  onFeedback: (messageId: number, rating: "up" | "down") => void
  onRetry: () => void
  onOpenArtifact: (a: ParsedArtifact) => void
}) {
  return (
    <div role="log" aria-live="polite" aria-label="Mensajes del chat" style={{ padding: "24px 24px 0", maxWidth: "768px", marginLeft: "auto", marginRight: "auto" }}>
      {messages.map((msg) => {
        const text = getMessageText(msg)
        const sources = getMessageSources(msg)
        const numId = Number(msg.id) || undefined
        const isUser = msg.role === "user"
        const isLastAssistant = !isUser && msg === messages[messages.length - 1] && isStreaming

        return (
          <div key={msg.id} className="group" style={{ marginBottom: "24px" }}>
            {isUser ? (
              <div className="flex justify-end">
                <div
                  className="rounded-2xl text-sm text-fg leading-relaxed whitespace-pre-wrap"
                  style={{ background: "var(--surface-2)", padding: "12px 16px", maxWidth: "85%" }}
                >
                  {text}
                </div>
              </div>
            ) : (
              <div className="min-w-0">
                {isLastAssistant && text === "" && (
                  <div className="flex items-center" style={{ gap: "8px", marginBottom: "8px" }}>
                    <SparkIcon className="text-accent animate-pulse" size={20} />
                    <span className="text-sm text-fg-muted italic">Pensando...</span>
                  </div>
                )}

                {text && (
                  <>
                    {isLastAssistant ? (
                      <div className="text-sm text-fg leading-relaxed whitespace-pre-wrap">
                        {stripArtifactTags(text)}
                      </div>
                    ) : (
                      <MarkdownMessage content={text} onOpenArtifact={onOpenArtifact} />
                    )}
                  </>
                )}

                {sources.length > 0 && (
                  <div style={{ marginTop: "12px" }}>
                    <SourcesPanel sources={sources} />
                  </div>
                )}

                {text && !isStreaming && (
                  <div className="flex" style={{ gap: "2px", marginTop: "8px" }}>
                    <TipBtn label="Copiar" onClick={() => onCopy(text, msg.id)} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                      {copiedKey === msg.id ? <Check size={ICON_PX} /> : <Copy size={ICON_PX} />}
                    </TipBtn>
                    {numId && (
                      <>
                        <TipBtn label="Útil" onClick={() => onFeedback(numId, "up")} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                          <ThumbsUp size={ICON_PX} />
                        </TipBtn>
                        <TipBtn label="No útil" onClick={() => onFeedback(numId, "down")} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                          <ThumbsDown size={ICON_PX} />
                        </TipBtn>
                      </>
                    )}
                    <TipBtn label="Reintentar" onClick={onRetry} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE}>
                      <RotateCcw size={ICON_PX} />
                    </TipBtn>
                  </div>
                )}
              </div>
            )}
          </div>
        )
      })}

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
  )
}
