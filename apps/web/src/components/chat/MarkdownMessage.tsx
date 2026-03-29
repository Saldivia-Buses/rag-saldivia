"use client"

import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { useState, memo } from "react"
import { Check, Copy, FileText } from "lucide-react"
import type { Artifact } from "./ArtifactPanel"

/**
 * Renderiza texto markdown con estilos del design system.
 * Soporta: headers, listas, tablas (GFM), code blocks, links, bold, italic.
 */
export const MarkdownMessage = memo(function MarkdownMessage({ content, onOpenArtifact }: { content: string; onOpenArtifact?: (artifact: Artifact) => void }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        h1: ({ children }) => (
          <h2 className="text-lg font-semibold text-fg" style={{ marginTop: "24px", marginBottom: "8px" }}>
            {children}
          </h2>
        ),
        h2: ({ children }) => (
          <h3 className="text-base font-semibold text-fg" style={{ marginTop: "20px", marginBottom: "8px" }}>
            {children}
          </h3>
        ),
        h3: ({ children }) => (
          <h4 className="text-sm font-semibold text-fg" style={{ marginTop: "16px", marginBottom: "4px" }}>
            {children}
          </h4>
        ),
        p: ({ children }) => (
          <p className="text-sm leading-relaxed text-fg" style={{ marginBottom: "12px" }}>
            {children}
          </p>
        ),
        ul: ({ children }) => (
          <ul className="text-sm text-fg" style={{ marginBottom: "12px", paddingLeft: "20px", listStyleType: "disc", display: "flex", flexDirection: "column", gap: "4px" }}>
            {children}
          </ul>
        ),
        ol: ({ children }) => (
          <ol className="text-sm text-fg" style={{ marginBottom: "12px", paddingLeft: "20px", listStyleType: "decimal", display: "flex", flexDirection: "column", gap: "4px" }}>
            {children}
          </ol>
        ),
        li: ({ children }) => (
          <li className="leading-relaxed">{children}</li>
        ),
        strong: ({ children }) => (
          <strong className="font-semibold">{children}</strong>
        ),
        em: ({ children }) => (
          <em className="italic">{children}</em>
        ),
        a: ({ href, children }) => (
          <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className="text-accent hover:underline"
          >
            {children}
          </a>
        ),
        blockquote: ({ children }) => (
          <blockquote
            className="text-sm text-fg-muted italic"
            style={{ borderLeft: "3px solid var(--border-strong)", paddingLeft: "12px", marginBottom: "12px" }}
          >
            {children}
          </blockquote>
        ),
        hr: () => <hr className="border-border" style={{ margin: "16px 0" }} />,
        table: ({ children }) => (
          <div className="overflow-x-auto rounded-lg border border-border" style={{ marginBottom: "12px" }}>
            <table className="w-full text-sm">{children}</table>
          </div>
        ),
        thead: ({ children }) => (
          <thead style={{ backgroundColor: "var(--surface)" }}>{children}</thead>
        ),
        th: ({ children }) => (
          <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "8px 12px", borderBottom: "1px solid var(--border)" }}>
            {children}
          </th>
        ),
        td: ({ children }) => (
          <td className="text-fg" style={{ padding: "8px 12px", borderBottom: "1px solid var(--border)" }}>
            {children}
          </td>
        ),
        code: ({ className, children }) => {
          const isBlock = className?.includes("language-")
          if (isBlock && onOpenArtifact) {
            const isMermaid = className?.includes("language-mermaid")
            return <ArtifactCard className={className} onOpenArtifact={onOpenArtifact} isMermaid={isMermaid}>{children}</ArtifactCard>
          }
          if (isBlock) {
            return <CodeBlock className={className}>{children}</CodeBlock>
          }
          return (
            <code
              className="text-sm font-mono text-accent"
              style={{
                backgroundColor: "var(--surface)",
                padding: "2px 6px",
                borderRadius: "4px",
              }}
            >
              {children}
            </code>
          )
        },
        pre: ({ children }) => <>{children}</>,
      }}
    >
      {content}
    </ReactMarkdown>
  )
})

const ArtifactCard = memo(function ArtifactCard({ className, children, onOpenArtifact, isMermaid }: { className?: string | undefined; children: React.ReactNode; onOpenArtifact: (artifact: Artifact) => void; isMermaid?: boolean | undefined }) {
  const lang = className?.replace("language-", "") ?? "código"
  const text = typeof children === "string" ? children : String(children ?? "")
  const title = isMermaid ? "Diagram" : (lang ? lang.charAt(0).toUpperCase() + lang.slice(1) : "Código")
  const subtitle = isMermaid ? "Diagrama · MERMAID" : lang.toUpperCase()

  function handleOpen() {
    onOpenArtifact({
      type: isMermaid ? "mermaid" : "code",
      title: isMermaid ? "Diagrama" : (lang ? `Código ${lang}` : "Código"),
      content: text,
      language: lang || undefined,
    })
  }

  function handleDownload(e: React.MouseEvent) {
    e.stopPropagation()
    const ext = isMermaid ? "mmd" : (lang || "txt")
    const blob = new Blob([text], { type: "text/plain" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `${title.toLowerCase()}.${ext}`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div
      className="flex items-center rounded-xl border border-border hover:bg-surface transition-colors cursor-pointer"
      style={{ marginBottom: "12px", padding: "12px 16px", gap: "12px" }}
      onClick={handleOpen}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => { if (e.key === "Enter") handleOpen() }}
    >
      {/* Document icon */}
      <div
        className="shrink-0 flex items-center justify-center rounded-lg bg-surface"
        style={{ width: "40px", height: "40px" }}
      >
        <FileText size={18} className="text-fg-subtle" />
      </div>

      {/* Title + subtitle */}
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-fg truncate">{title}</div>
        <div className="text-xs text-fg-subtle truncate">{subtitle}</div>
      </div>

      {/* Download button */}
      <button
        onClick={handleDownload}
        className="shrink-0 text-xs text-fg-muted hover:text-fg border border-border rounded-lg transition-colors"
        style={{ padding: "6px 12px" }}
      >
        Descargar
      </button>
    </div>
  )
})

function CodeBlock({ className, children }: { className?: string | undefined; children: React.ReactNode }) {
  const [copied, setCopied] = useState(false)
  const lang = className?.replace("language-", "") ?? ""
  const text = typeof children === "string" ? children : String(children ?? "")

  function handleCopy() {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative rounded-lg overflow-hidden" style={{ marginBottom: "12px", backgroundColor: "var(--surface)" }}>
      <div className="flex items-center justify-between" style={{ padding: "6px 12px", borderBottom: "1px solid var(--border)" }}>
        <span className="text-xs text-fg-subtle">{lang}</span>
        <button
          onClick={handleCopy}
          className="text-fg-subtle hover:text-fg transition-colors"
          title="Copiar código"
        >
          {copied ? <Check size={14} /> : <Copy size={14} />}
        </button>
      </div>
      <pre style={{ padding: "12px", overflowX: "auto" }}>
        <code className="text-sm font-mono text-fg">{children}</code>
      </pre>
    </div>
  )
}
