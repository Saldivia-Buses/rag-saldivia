"use client"

import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { useState } from "react"
import { Check, Copy, PanelRightOpen } from "lucide-react"
import type { Artifact } from "./ArtifactPanel"

/**
 * Renderiza texto markdown con estilos del design system.
 * Soporta: headers, listas, tablas (GFM), code blocks, links, bold, italic.
 */
export function MarkdownMessage({ content, onOpenArtifact }: { content: string; onOpenArtifact?: (artifact: Artifact) => void }) {
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
          if (isBlock) {
            return <CodeBlock className={className} onOpenArtifact={onOpenArtifact}>{children}</CodeBlock>
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
}

function CodeBlock({ className, children, onOpenArtifact }: { className?: string | undefined; children: React.ReactNode; onOpenArtifact?: ((artifact: Artifact) => void) | undefined }) {
  const [copied, setCopied] = useState(false)
  const lang = className?.replace("language-", "") ?? ""
  const text = typeof children === "string" ? children : String(children ?? "")

  function handleCopy() {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function handleOpen() {
    onOpenArtifact?.({
      type: "code",
      title: lang ? `Código ${lang}` : "Código",
      content: text,
      language: lang || undefined,
    })
  }

  return (
    <div className="relative rounded-lg overflow-hidden" style={{ marginBottom: "12px", backgroundColor: "var(--surface)" }}>
      <div className="flex items-center justify-between" style={{ padding: "6px 12px", borderBottom: "1px solid var(--border)" }}>
        <span className="text-xs text-fg-subtle">{lang}</span>
        <div className="flex items-center" style={{ gap: "4px" }}>
          {onOpenArtifact && (
            <button
              onClick={handleOpen}
              className="text-fg-subtle hover:text-accent transition-colors"
              title="Abrir en panel"
            >
              <PanelRightOpen size={14} />
            </button>
          )}
          <button
            onClick={handleCopy}
            className="text-fg-subtle hover:text-fg transition-colors"
            title="Copiar código"
          >
            {copied ? <Check size={14} /> : <Copy size={14} />}
          </button>
        </div>
      </div>
      <pre style={{ padding: "12px", overflowX: "auto" }}>
        <code className="text-sm font-mono text-fg">{children}</code>
      </pre>
    </div>
  )
}
