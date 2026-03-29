"use client"

import { X, Copy, Check, Maximize2, Minimize2 } from "lucide-react"
import { useState } from "react"

export type Artifact = {
  type: "code" | "table" | "text"
  title: string
  content: string
  language?: string | undefined
}

export function ArtifactPanel({
  artifact,
  onClose,
}: {
  artifact: Artifact
  onClose: () => void
}) {
  const [copied, setCopied] = useState(false)
  const [expanded, setExpanded] = useState(false)

  function handleCopy() {
    navigator.clipboard.writeText(artifact.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div
      className={`shrink-0 border-l border-border bg-bg flex flex-col h-full transition-all ${
        expanded ? "w-full absolute inset-0 z-50" : ""
      }`}
      style={expanded ? {} : { width: "min(50%, 600px)" }}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between border-b border-border shrink-0"
        style={{ padding: "12px 16px" }}
      >
        <div className="flex items-center" style={{ gap: "8px" }}>
          <div
            className="text-xs font-medium uppercase tracking-wide text-accent"
            style={{
              backgroundColor: "var(--accent-subtle)",
              padding: "2px 8px",
              borderRadius: "4px",
            }}
          >
            {artifact.type === "code" ? artifact.language ?? "código" : artifact.type === "table" ? "tabla" : "texto"}
          </div>
          <span className="text-sm font-medium text-fg truncate">
            {artifact.title}
          </span>
        </div>
        <div className="flex items-center" style={{ gap: "4px" }}>
          <button
            onClick={handleCopy}
            className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
            title="Copiar"
          >
            {copied ? <Check size={14} /> : <Copy size={14} />}
          </button>
          <button
            onClick={() => setExpanded(!expanded)}
            className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
            title={expanded ? "Reducir" : "Expandir"}
          >
            {expanded ? <Minimize2 size={14} /> : <Maximize2 size={14} />}
          </button>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
            title="Cerrar"
          >
            <X size={14} />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto" style={{ padding: "16px" }}>
        {artifact.type === "code" ? (
          <pre
            className="text-sm font-mono text-fg leading-relaxed"
            style={{
              backgroundColor: "var(--surface)",
              padding: "16px",
              borderRadius: "8px",
              overflowX: "auto",
              whiteSpace: "pre",
              tabSize: 2,
            }}
          >
            <code>{artifact.content}</code>
          </pre>
        ) : artifact.type === "table" ? (
          <div
            className="text-sm text-fg"
            dangerouslySetInnerHTML={{ __html: artifact.content }}
          />
        ) : (
          <div className="text-sm text-fg leading-relaxed whitespace-pre-wrap">
            {artifact.content}
          </div>
        )}
      </div>
    </div>
  )
}
