"use client"

import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { useState, useMemo, memo } from "react"
import { Check, Copy, Code, Eye, FileText } from "lucide-react"
import type { ParsedArtifact } from "@/lib/rag/artifact-parser"
import { extractArtifacts } from "@/lib/rag/artifact-parser"
import { TYPE_CONFIG } from "./ArtifactPanel"

// ── Stable markdown component renderers ──

const H1 = ({ children }: { children?: React.ReactNode }) => (
  <h2 className="text-lg font-semibold text-fg" style={{ marginTop: "24px", marginBottom: "8px" }}>{children}</h2>
)
const H2 = ({ children }: { children?: React.ReactNode }) => (
  <h3 className="text-base font-semibold text-fg" style={{ marginTop: "20px", marginBottom: "8px" }}>{children}</h3>
)
const H3 = ({ children }: { children?: React.ReactNode }) => (
  <h4 className="text-sm font-semibold text-fg" style={{ marginTop: "16px", marginBottom: "4px" }}>{children}</h4>
)
const P = ({ children }: { children?: React.ReactNode }) => (
  <p className="text-sm leading-relaxed text-fg" style={{ marginBottom: "12px" }}>{children}</p>
)
const Ul = ({ children }: { children?: React.ReactNode }) => (
  <ul className="text-sm text-fg" style={{ marginBottom: "12px", paddingLeft: "20px", listStyleType: "disc", display: "flex", flexDirection: "column" as const, gap: "4px" }}>{children}</ul>
)
const Ol = ({ children }: { children?: React.ReactNode }) => (
  <ol className="text-sm text-fg" style={{ marginBottom: "12px", paddingLeft: "20px", listStyleType: "decimal", display: "flex", flexDirection: "column" as const, gap: "4px" }}>{children}</ol>
)
const Li = ({ children }: { children?: React.ReactNode }) => (
  <li className="leading-relaxed">{children}</li>
)
const Strong = ({ children }: { children?: React.ReactNode }) => (
  <strong className="font-semibold">{children}</strong>
)
const Em = ({ children }: { children?: React.ReactNode }) => (
  <em className="italic">{children}</em>
)
const A = (props: React.AnchorHTMLAttributes<HTMLAnchorElement>) => (
  <a {...props} target="_blank" rel="noopener noreferrer" className="text-accent hover:underline">{props.children}</a>
)
const Blockquote = ({ children }: { children?: React.ReactNode }) => (
  <blockquote className="text-sm text-fg-muted italic" style={{ borderLeft: "3px solid var(--border-strong)", paddingLeft: "12px", marginBottom: "12px" }}>{children}</blockquote>
)
const Hr = () => <hr className="border-border" style={{ margin: "16px 0" }} />
const Table = ({ children }: { children?: React.ReactNode }) => (
  <div className="overflow-x-auto rounded-lg border border-border" style={{ marginBottom: "12px" }}>
    <table className="w-full text-sm">{children}</table>
  </div>
)
const Thead = ({ children }: { children?: React.ReactNode }) => (
  <thead style={{ backgroundColor: "var(--surface)" }}>{children}</thead>
)
const Th = ({ children }: { children?: React.ReactNode }) => (
  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "8px 12px", borderBottom: "1px solid var(--border)" }}>{children}</th>
)
const Td = ({ children }: { children?: React.ReactNode }) => (
  <td className="text-fg" style={{ padding: "8px 12px", borderBottom: "1px solid var(--border)" }}>{children}</td>
)
const Pre = ({ children }: { children?: React.ReactNode }) => <>{children}</>

const REMARK_PLUGINS = [remarkGfm]

// ── Artifact card (inline in messages) ──

export const ArtifactCard = memo(function ArtifactCard({
  artifact,
  onClick,
}: {
  artifact: ParsedArtifact
  onClick: () => void
}) {
  const config = TYPE_CONFIG[artifact.type] ?? TYPE_CONFIG.code!

  return (
    <div
      className="flex items-center rounded-xl border border-border hover:border-accent/30 hover:bg-surface/50 transition-all cursor-pointer group"
      style={{ margin: "12px 0", padding: "14px 16px", gap: "14px" }}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter") onClick()
      }}
    >
      {/* Icon */}
      <div
        className="shrink-0 flex items-center justify-center rounded-lg"
        style={{
          width: "44px",
          height: "44px",
          backgroundColor: `color-mix(in srgb, ${config.color} 12%, transparent)`,
        }}
      >
        {artifact.type === "html" || artifact.type === "svg" ? (
          <Eye size={20} style={{ color: config.color }} />
        ) : artifact.type === "mermaid" ? (
          <FileText size={20} style={{ color: config.color }} />
        ) : (
          <Code size={20} style={{ color: config.color }} />
        )}
      </div>
      {/* Info */}
      <div className="flex-1 min-w-0">
        <div className="text-sm font-medium text-fg truncate">{artifact.title}</div>
        <div className="text-xs text-fg-subtle" style={{ marginTop: "2px" }}>
          {config.label}
          {artifact.language && artifact.language !== artifact.type
            ? ` · ${artifact.language}`
            : ""}
          {artifact.isStreaming && (
            <span className="italic animate-pulse"> · generando...</span>
          )}
        </div>
      </div>
      {/* Arrow hint on hover */}
      <div className="shrink-0 text-fg-subtle opacity-0 group-hover:opacity-100 transition-opacity">
        <Eye size={16} />
      </div>
    </div>
  )
})

// ── Code block fallback (when no artifact callback) ──

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
    <div
      className="relative rounded-lg overflow-hidden"
      style={{ marginBottom: "12px", backgroundColor: "var(--surface)" }}
    >
      <div
        className="flex items-center justify-between"
        style={{
          padding: "6px 12px",
          borderBottom: "1px solid var(--border)",
        }}
      >
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

// ── Markdown renderer (no artifact awareness) ──

const MarkdownOnly = memo(function MarkdownOnly({ content }: { content: string }) {
  const components = useMemo(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (): any => ({
      h1: H1, h2: H2, h3: H3, p: P, ul: Ul, ol: Ol, li: Li,
      strong: Strong, em: Em, a: A, blockquote: Blockquote, hr: Hr,
      table: Table, thead: Thead, th: Th, td: Td, pre: Pre,
      code: ({ className, children }: { className?: string; children?: React.ReactNode }) => {
        const isBlock = className?.includes("language-")
        if (isBlock) return <CodeBlock className={className}>{children}</CodeBlock>
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
    }),
    []
  )

  return (
    <ReactMarkdown remarkPlugins={REMARK_PLUGINS} components={components}>
      {content}
    </ReactMarkdown>
  )
})

// ── Legacy markdown with code-block-as-artifact detection ──

const MarkdownWithCodeBlockArtifacts = memo(function MarkdownWithCodeBlockArtifacts({
  content,
  onOpenArtifact,
}: {
  content: string
  onOpenArtifact: (artifact: ParsedArtifact) => void
}) {
  const components = useMemo(
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (): any => ({
      h1: H1, h2: H2, h3: H3, p: P, ul: Ul, ol: Ol, li: Li,
      strong: Strong, em: Em, a: A, blockquote: Blockquote, hr: Hr,
      table: Table, thead: Thead, th: Th, td: Td, pre: Pre,
      code: ({ className, children }: { className?: string; children?: React.ReactNode }) => {
        const isBlock = className?.includes("language-")
        if (isBlock) {
          const lang = className?.replace("language-", "") ?? "text"
          const text = typeof children === "string" ? children : String(children ?? "")
          const isMermaid = lang === "mermaid"
          // Only create artifact cards for substantial blocks
          if (text.trim().split("\n").length >= 3) {
            const artifact: ParsedArtifact = {
              id: `cb_${Math.random().toString(36).slice(2, 8)}`,
              type: isMermaid ? "mermaid" : "code",
              title: isMermaid ? "Diagrama" : `Código ${lang}`,
              content: text,
              language: lang,
              version: 1,
            }
            return <ArtifactCard artifact={artifact} onClick={() => onOpenArtifact(artifact)} />
          }
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
    }),
    [onOpenArtifact]
  )

  return (
    <ReactMarkdown remarkPlugins={REMARK_PLUGINS} components={components}>
      {content}
    </ReactMarkdown>
  )
})

// ── Main export: renders message with artifact support ──

export const MarkdownMessage = memo(function MarkdownMessage({
  content,
  onOpenArtifact,
}: {
  content: string
  onOpenArtifact?: (artifact: ParsedArtifact) => void
}) {
  // Try extracting <artifact> tags first
  const { cleanText, artifacts } = useMemo(() => extractArtifacts(content), [content])

  // If artifact tags found, render with inline cards
  if (artifacts.length > 0 && onOpenArtifact) {
    const segments = cleanText.split(/\[ARTIFACT:([\w_]+)\]/)

    return (
      <>
        {segments.map((segment, i) => {
          if (i % 2 === 0) {
            const trimmed = segment.trim()
            return trimmed ? <MarkdownOnly key={i} content={trimmed} /> : null
          }
          const art = artifacts.find((a) => a.id === segment)
          return art ? (
            <ArtifactCard key={i} artifact={art} onClick={() => onOpenArtifact(art)} />
          ) : null
        })}
      </>
    )
  }

  // Fallback: detect code blocks as artifacts
  if (onOpenArtifact) {
    return (
      <MarkdownWithCodeBlockArtifacts
        content={content}
        onOpenArtifact={onOpenArtifact}
      />
    )
  }

  return <MarkdownOnly content={content} />
})
