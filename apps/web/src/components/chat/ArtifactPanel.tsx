"use client"

import { X, Copy, Check, Code, Eye, Download } from "lucide-react"
import { useState, useEffect, useRef, useCallback, memo } from "react"

export type Artifact = {
  type: "code" | "table" | "text" | "mermaid"
  title: string
  content: string
  language?: string | undefined
}

// ── Syntax highlighting con shiki (lazy) ──

// Cache global de highlighter para no recrear en cada render
let highlighterCache: Map<string, string> = new Map()

const HighlightedCode = memo(function HighlightedCode({ code, language }: { code: string; language: string }) {
  const cacheKey = `${language}:${code}`
  const [html, setHtml] = useState<string | null>(() => highlighterCache.get(cacheKey) ?? null)

  useEffect(() => {
    if (highlighterCache.has(cacheKey)) {
      setHtml(highlighterCache.get(cacheKey)!)
      return
    }
    let cancelled = false
    import("shiki").then(async ({ createHighlighter }) => {
      try {
        const highlighter = await createHighlighter({
          themes: ["github-dark"],
          langs: [language],
        })
        if (cancelled) return
        const result = highlighter.codeToHtml(code, {
          lang: language,
          theme: "github-dark",
        })
        highlighterCache.set(cacheKey, result)
        setHtml(result)
      } catch {
        // Lenguaje no soportado
      }
    })
    return () => { cancelled = true }
  }, [code, language, cacheKey])

  if (html) {
    return (
      <div
        className="text-sm [&_pre]:!bg-transparent [&_pre]:!p-0 [&_code]:!text-sm"
        dangerouslySetInnerHTML={{ __html: html }}
      />
    )
  }

  // Fallback sin highlighting
  return (
    <pre className="text-sm font-mono text-fg whitespace-pre overflow-x-auto">
      <code>{code}</code>
    </pre>
  )
})

// ── Mermaid rendering ──

const MermaidPreview = memo(function MermaidPreview({ content }: { content: string }) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [svg, setSvg] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    import("mermaid").then(async (mod) => {
      const mermaid = mod.default
      mermaid.initialize({ startOnLoad: false, theme: "dark" })
      try {
        const id = `mermaid-${Date.now()}`
        const { svg: result } = await mermaid.render(id, content)
        if (!cancelled) setSvg(result)
      } catch (err) {
        if (!cancelled) setError(String(err))
      }
    })
    return () => { cancelled = true }
  }, [content])

  if (error) {
    return (
      <div className="text-sm text-destructive" style={{ padding: "16px" }}>
        Error renderizando diagrama: {error}
      </div>
    )
  }

  if (svg) {
    return (
      <div
        ref={containerRef}
        className="flex items-center justify-center"
        style={{ padding: "24px", minHeight: "200px" }}
        dangerouslySetInnerHTML={{ __html: svg }}
      />
    )
  }

  return (
    <div className="flex items-center justify-center text-fg-subtle text-sm" style={{ padding: "48px" }}>
      Renderizando diagrama...
    </div>
  )
})

// ── Panel principal ──

type ViewMode = "preview" | "code"

export function ArtifactPanel({
  artifacts,
  activeIndex,
  onSelect,
  onClose,
}: {
  artifacts: Artifact[]
  activeIndex: number
  onSelect: (index: number) => void
  onClose: () => void
}) {
  const [viewMode, setViewMode] = useState<ViewMode>("preview")
  const [copied, setCopied] = useState(false)
  const [panelWidth, setPanelWidth] = useState(500)
  const resizing = useRef(false)

  const artifact = artifacts[activeIndex] as Artifact | undefined

  const isMermaid = artifact?.type === "mermaid" || artifact?.language === "mermaid"
  const hasPreview = isMermaid

  function handleCopy() {
    if (!artifact) return
    navigator.clipboard.writeText(artifact.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function handleDownload() {
    if (!artifact) return
    const ext = artifact.language ?? (artifact.type === "mermaid" ? "mmd" : "txt")
    const blob = new Blob([artifact.content], { type: "text/plain" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `artifact.${ext}`
    a.click()
    URL.revokeObjectURL(url)
  }

  // ── Resize drag ──
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    resizing.current = true
    const startX = e.clientX
    const startWidth = panelWidth

    function onMouseMove(ev: MouseEvent) {
      if (!resizing.current) return
      const delta = startX - ev.clientX
      setPanelWidth(Math.max(300, Math.min(900, startWidth + delta)))
    }

    function onMouseUp() {
      resizing.current = false
      document.removeEventListener("mousemove", onMouseMove)
      document.removeEventListener("mouseup", onMouseUp)
    }

    document.addEventListener("mousemove", onMouseMove)
    document.addEventListener("mouseup", onMouseUp)
  }, [panelWidth])

  if (!artifact) return null

  return (
    <div className="shrink-0 flex h-full" style={{ width: `${panelWidth}px` }}>
      {/* Resize handle */}
      <div
        className="w-1 cursor-col-resize hover:bg-accent/30 transition-colors shrink-0"
        onMouseDown={handleMouseDown}
        style={{ borderLeft: "1px solid var(--border)" }}
      />

      <div className="flex-1 flex flex-col bg-bg overflow-hidden">
        {/* Header */}
        <div
          className="flex items-center justify-between border-b border-border shrink-0"
          style={{ padding: "8px 12px" }}
        >
          <div className="flex items-center" style={{ gap: "6px" }}>
            {/* View mode toggle */}
            {hasPreview && (
              <div
                className="flex rounded-md border border-border overflow-hidden"
                style={{ height: "28px" }}
              >
                <button
                  onClick={() => setViewMode("preview")}
                  className={`flex items-center justify-center transition-colors ${
                    viewMode === "preview" ? "bg-surface-2 text-fg" : "text-fg-subtle hover:text-fg"
                  }`}
                  style={{ width: "32px" }}
                  title="Preview"
                >
                  <Eye size={14} />
                </button>
                <button
                  onClick={() => setViewMode("code")}
                  className={`flex items-center justify-center transition-colors ${
                    viewMode === "code" ? "bg-surface-2 text-fg" : "text-fg-subtle hover:text-fg"
                  }`}
                  style={{ width: "32px", borderLeft: "1px solid var(--border)" }}
                  title="Código"
                >
                  <Code size={14} />
                </button>
              </div>
            )}

            {/* Type badge */}
            <span
              className="text-xs font-medium uppercase tracking-wide text-accent"
              style={{
                backgroundColor: "var(--accent-subtle)",
                padding: "2px 8px",
                borderRadius: "4px",
              }}
            >
              {isMermaid ? "mermaid" : artifact.language ?? artifact.type}
            </span>

            <span className="text-sm text-fg truncate">{artifact.title}</span>
          </div>

          <div className="flex items-center" style={{ gap: "2px" }}>
            <button
              onClick={handleDownload}
              className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
              title="Descargar"
            >
              <Download size={14} />
            </button>
            <button
              onClick={handleCopy}
              className="p-1.5 rounded-md text-fg-subtle hover:text-fg hover:bg-surface transition-colors"
              title="Copiar"
            >
              {copied ? <Check size={14} /> : <Copy size={14} />}
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
          {isMermaid && viewMode === "preview" ? (
            <MermaidPreview content={artifact.content} />
          ) : (
            <div
              style={{
                backgroundColor: "var(--surface)",
                borderRadius: "8px",
                padding: "16px",
                overflow: "auto",
              }}
            >
              <HighlightedCode
                code={artifact.content}
                language={artifact.language ?? "text"}
              />
            </div>
          )}
        </div>

        {/* Artifact list (if multiple) */}
        {artifacts.length > 1 && (
          <div
            className="border-t border-border shrink-0 overflow-x-auto flex"
            style={{ padding: "6px 12px", gap: "4px" }}
          >
            {artifacts.map((a, i) => (
              <button
                key={i}
                onClick={() => onSelect(i)}
                className={`shrink-0 text-xs rounded-md transition-colors ${
                  i === activeIndex
                    ? "bg-accent-subtle text-accent font-medium"
                    : "text-fg-subtle hover:text-fg hover:bg-surface"
                }`}
                style={{ padding: "4px 10px" }}
              >
                {a.language ?? a.type} #{i + 1}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
