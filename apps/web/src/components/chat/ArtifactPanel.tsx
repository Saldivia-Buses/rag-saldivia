"use client"

import { Copy, Check, Code, Eye, Download } from "lucide-react"
import { useState, useEffect, useRef, useCallback, memo } from "react"

export type Artifact = {
  type: "code" | "table" | "text" | "mermaid"
  title: string
  content: string
  language?: string | undefined
}

// ── Syntax highlighting con shiki (lazy + cached) ──

const highlighterCache = new Map<string, string>()

function useIsDark() {
  const [dark, setDark] = useState(() =>
    typeof document !== "undefined" && document.documentElement.classList.contains("dark")
  )
  useEffect(() => {
    const obs = new MutationObserver(() => {
      setDark(document.documentElement.classList.contains("dark"))
    })
    obs.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] })
    return () => obs.disconnect()
  }, [])
  return dark
}

const HighlightedCode = memo(function HighlightedCode({ code, language }: { code: string; language: string }) {
  const isDark = useIsDark()
  const theme = isDark ? "github-dark" : "github-light"
  const cacheKey = `${theme}:${language}:${code}`
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
          themes: [theme],
          langs: [language],
        })
        if (cancelled) return
        const result = highlighter.codeToHtml(code, { lang: language, theme })
        highlighterCache.set(cacheKey, result)
        setHtml(result)
      } catch {
        // Lenguaje no soportado
      }
    })
    return () => { cancelled = true }
  }, [code, language, theme, cacheKey])

  if (html) {
    return (
      <div
        className="text-sm [&_pre]:!bg-transparent [&_pre]:!p-0 [&_code]:!text-sm [&_code]:!whitespace-pre-wrap [&_code]:!break-all"
        style={{ width: "100%", minWidth: 0 }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    )
  }

  return (
    <pre className="text-sm font-mono text-fg whitespace-pre-wrap break-all" style={{ width: "100%", minWidth: 0 }}>
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
      mermaid.initialize({
        startOnLoad: false,
        theme: "base",
        themeVariables: {
          primaryColor: "#2563eb",
          primaryTextColor: "#ffffff",
          primaryBorderColor: "#1d4ed8",
          secondaryColor: "#dbeafe",
          secondaryTextColor: "#141413",
          secondaryBorderColor: "#93c5fd",
          tertiaryColor: "#f0eee8",
          tertiaryTextColor: "#141413",
          tertiaryBorderColor: "#e0ddd6",
          lineColor: "#6e6c69",
          textColor: "#f5f4f0",
          mainBkg: "#2563eb",
          nodeBorder: "#1d4ed8",
          clusterBkg: "#242320",
          clusterBorder: "#3a3935",
          titleColor: "#f5f4f0",
          edgeLabelBackground: "#242320",
          nodeTextColor: "#ffffff",
          fontFamily: "Instrument Sans, system-ui, sans-serif",
          fontSize: "14px",
        },
      })
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
        className="flex items-center justify-center overflow-auto [&_svg]:max-w-full"
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

const ICON_BTN = "flex items-center justify-center rounded-lg transition-colors"
const ICON_BTN_SIZE = { width: "42px", height: "42px" } as const

export function ArtifactPanel({
  artifacts,
  activeIndex,
  onSelect,
  onClose: _onClose,
  panelWidth,
  onWidthChange,
  isResizing: isResizingProp,
  onResizeStart,
  onResizeEnd,
}: {
  artifacts: Artifact[]
  activeIndex: number
  onSelect: (index: number) => void
  onClose: () => void
  panelWidth: number
  onWidthChange: (w: number) => void
  isResizing: boolean
  onResizeStart: () => void
  onResizeEnd: () => void
}) {
  const [viewMode, setViewMode] = useState<ViewMode>("preview")
  const [copied, setCopied] = useState(false)
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

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    resizing.current = true
    onResizeStart()
    const startX = e.clientX
    const startWidth = panelWidth
    document.body.style.cursor = "col-resize"
    document.body.style.userSelect = "none"

    function onMouseMove(ev: MouseEvent) {
      if (!resizing.current) return
      const delta = startX - ev.clientX
      onWidthChange(Math.max(320, Math.min(900, startWidth + delta)))
    }

    function onMouseUp() {
      resizing.current = false
      onResizeEnd()
      document.body.style.cursor = ""
      document.body.style.userSelect = ""
      document.removeEventListener("mousemove", onMouseMove)
      document.removeEventListener("mouseup", onMouseUp)
    }

    document.addEventListener("mousemove", onMouseMove)
    document.addEventListener("mouseup", onMouseUp)
  }, [panelWidth, onWidthChange, onResizeStart, onResizeEnd])

  if (!artifact) return null

  const typeLabel = isMermaid ? "MERMAID" : (artifact.language ?? artifact.type).toUpperCase()

  return (
    <div
      className={`shrink-0 flex h-full overflow-hidden ${isResizingProp ? "" : "transition-[width] duration-200"}`}
      style={{ width: `${panelWidth}px` }}
    >
      {/* Resize handle */}
      <div
        className="shrink-0 cursor-col-resize group flex items-center justify-center hover:bg-accent/10 transition-colors"
        style={{ width: "8px" }}
        onMouseDown={handleMouseDown}
      >
        <div
          className="rounded-full bg-border group-hover:bg-accent transition-colors"
          style={{ width: "3px", height: "40px" }}
        />
      </div>

      <div className="flex-1 flex flex-col bg-bg overflow-hidden border-l border-border">
        {/* Header */}
        <div
          className="flex items-center justify-between border-b border-border shrink-0"
          style={{ height: "48px", padding: "0 8px 0 12px" }}
        >
          <div className="flex items-center min-w-0" style={{ gap: "8px" }}>
            {hasPreview && (
              <div className="flex rounded-lg border border-border overflow-hidden shrink-0" style={{ height: "36px" }}>
                <button
                  onClick={() => setViewMode("preview")}
                  className={`flex items-center justify-center transition-colors ${
                    viewMode === "preview" ? "bg-surface-2 text-fg" : "text-fg-subtle hover:text-fg hover:bg-surface"
                  }`}
                  style={{ width: "38px" }}
                  title="Preview"
                >
                  <Eye size={18} />
                </button>
                <button
                  onClick={() => setViewMode("code")}
                  className={`flex items-center justify-center border-l border-border transition-colors ${
                    viewMode === "code" ? "bg-surface-2 text-fg" : "text-fg-subtle hover:text-fg hover:bg-surface"
                  }`}
                  style={{ width: "38px" }}
                  title="Código"
                >
                  <Code size={18} />
                </button>
              </div>
            )}

            <span
              className="shrink-0 text-xs font-medium uppercase tracking-wide text-accent"
              style={{ backgroundColor: "var(--accent-subtle)", padding: "3px 8px", borderRadius: "4px" }}
            >
              {typeLabel}
            </span>
            <span className="text-sm text-fg truncate">{artifact.title}</span>
          </div>

          <div className="flex items-center shrink-0" style={{ gap: "4px" }}>
            <button onClick={handleDownload} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE} title="Descargar">
              <Download size={16} />
            </button>
            <button onClick={handleCopy} className={`${ICON_BTN} text-fg-subtle hover:text-fg hover:bg-surface-2`} style={ICON_BTN_SIZE} title="Copiar">
              {copied ? <Check size={16} /> : <Copy size={16} />}
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto">
          {isMermaid && viewMode === "preview" ? (
            <MermaidPreview content={artifact.content} />
          ) : (
            <div style={{ padding: "16px" }} className="min-w-0 w-full">
              <HighlightedCode
                code={artifact.content}
                language={artifact.language ?? "text"}
              />
            </div>
          )}
        </div>

        {/* Artifact tabs */}
        {artifacts.length > 1 && (
          <div
            className="border-t border-border shrink-0 overflow-x-auto flex"
            style={{ padding: "8px 12px", gap: "4px" }}
          >
            {artifacts.map((a, i) => (
              <button
                key={i}
                onClick={() => onSelect(i)}
                className={`shrink-0 text-xs rounded-lg transition-colors ${
                  i === activeIndex
                    ? "bg-accent-subtle text-accent font-medium"
                    : "text-fg-subtle hover:text-fg hover:bg-surface"
                }`}
                style={{ padding: "6px 12px" }}
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
