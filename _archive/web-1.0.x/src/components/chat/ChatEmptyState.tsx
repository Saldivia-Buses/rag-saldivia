"use client"

import { ChatInputBar } from "./ChatInputBar"
import { CollectionSelector } from "./CollectionSelector"

const FALLBACK_SUGGESTIONS = [
  { title: "Buscar documentos", prompt: "Buscá información sobre " },
  { title: "Hacer preguntas", prompt: "¿Qué es " },
  { title: "Analizar datos", prompt: "Analizá los datos sobre " },
  { title: "Resumir contenido", prompt: "Hacé un resumen de " },
]

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

function SparkIcon({ className, size = 16 }: { className?: string; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" className={className}>
      <path d="M8 1L9.5 6.5L15 8L9.5 9.5L8 15L6.5 9.5L1 8L6.5 6.5L8 1Z" fill="currentColor" />
    </svg>
  )
}

type PromptTemplate = { id: number; title: string; prompt: string; focusMode: string }

export function ChatEmptyState({
  input,
  setInput,
  onKeyDown,
  onSend,
  textareaRef,
  collection,
  availableCollections,
  onCollectionsChange,
  templates = [],
}: {
  input: string
  setInput: (v: string) => void
  onKeyDown: (e: React.KeyboardEvent) => void
  onSend: () => void
  textareaRef: React.RefObject<HTMLTextAreaElement | null>
  collection: string
  availableCollections: string[]
  onCollectionsChange: (cols: string[]) => void
  templates?: PromptTemplate[]
}) {
  return (
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

      <div className="w-full" style={{ maxWidth: "640px" }}>
        <ChatInputBar
          value={input}
          onChange={setInput}
          onKeyDown={onKeyDown}
          onSend={onSend}
          textareaRef={textareaRef}
          placeholder="¿Cómo puedo ayudarte hoy?"
          collection={collection}
          collectionSlot={availableCollections.length > 1 ? (
            <CollectionSelector
              defaultCollection={collection}
              availableCollections={availableCollections}
              onCollectionsChange={onCollectionsChange}
            />
          ) : undefined}
        />

        <div className="flex items-center justify-center flex-wrap" style={{ gap: "8px", marginTop: "16px" }}>
          {(templates.length > 0 ? templates : FALLBACK_SUGGESTIONS).map((t) => (
            <button
              key={t.title}
              onClick={() => { setInput(t.prompt); textareaRef.current?.focus() }}
              className="flex items-center border border-border rounded-full text-sm text-fg-muted hover:text-fg hover:bg-surface transition-colors"
              style={{ padding: "6px 14px", gap: "6px" }}
            >
              <span>{getTemplateIcon(t.title)}</span>
              <span>{t.title}</span>
            </button>
          ))}
        </div>
      </div>

      <p className="text-xs text-fg-subtle text-center" style={{ marginTop: "24px" }}>
        Saldivia RAG es IA y puede cometer errores. Por favor, verificá las respuestas.
      </p>
    </div>
  )
}
