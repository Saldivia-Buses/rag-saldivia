/**
 * Parser for `<artifact>` tags embedded in LLM-generated text.
 *
 * Data flow:
 *   LLM response (raw text with `<artifact ...>` tags)
 *     -> extractArtifacts / extractStreamingArtifact
 *       -> ParsedArtifact[] consumed by ChatInterface to render rich panels
 *
 * The LLM wraps code, diagrams, tables, etc. in `<artifact>` XML tags with
 * attributes (id, type, title, language, version). This module:
 *   1. Extracts complete artifacts, replacing them with `[ARTIFACT:id]` markers
 *   2. Detects partial (still-streaming) artifacts for live preview
 *   3. Strips artifact tags for plain-text display
 *   4. Falls back to markdown code blocks (3+ lines) when no tags are present
 *
 * Used by: ChatInterface (message rendering), MessageBubble
 * Depends on: nothing (pure string parsing, no external deps)
 */

export type ParsedArtifact = {
  id: string
  type: "code" | "html" | "svg" | "mermaid" | "table" | "text"
  title: string
  content: string
  language?: string | undefined
  version: number
  isStreaming?: boolean | undefined
}

// ── Regexes ──

/** Matches complete `<artifact attrs>content</artifact>` blocks (global, greedy inner). */
const ARTIFACT_RE = /<artifact\s+([^>]*)>([\s\S]*?)<\/artifact>/g

/** Extracts key="value" pairs from the opening tag's attribute string. */
const ATTR_RE = /(\w+)="([^"]*)"/g

/** Parse HTML-style attribute string into a plain object. */
function parseAttrs(s: string): Record<string, string> {
  const out: Record<string, string> = {}
  let m: RegExpExecArray | null
  ATTR_RE.lastIndex = 0
  while ((m = ATTR_RE.exec(s))) out[m[1]!] = m[2]!
  return out
}

/** Module-scoped counter for auto-generating unique artifact IDs when the LLM omits one. */
let _id = 0
function nextId() { return `art_${++_id}` }

/**
 * Extract all complete `<artifact>` tags from text.
 *
 * Replaces each matched tag with an `[ARTIFACT:id]` placeholder so the
 * rendered markdown stays clean, while the extracted artifacts are passed
 * separately to the artifact viewer component.
 *
 * @returns cleanText - original text with artifact tags replaced by markers
 * @returns artifacts - array of parsed artifacts with content and metadata
 */
export function extractArtifacts(text: string): {
  cleanText: string
  artifacts: ParsedArtifact[]
} {
  const artifacts: ParsedArtifact[] = []
  ARTIFACT_RE.lastIndex = 0

  const cleanText = text.replace(ARTIFACT_RE, (_, attrs: string, content: string) => {
    const a = parseAttrs(attrs)
    const id = a.id ?? nextId()
    const type = (a.type as ParsedArtifact["type"]) ?? "code"
    artifacts.push({
      id,
      type,
      title: a.title ?? (type === "mermaid" ? "Diagrama" : "Sin título"),
      content: content.trim(),
      language: a.language ?? (type === "mermaid" ? "mermaid" : undefined),
      version: Number(a.version) || 1,
    })
    return `\n\n[ARTIFACT:${id}]\n\n`
  })

  return { cleanText, artifacts }
}

/**
 * Extract a partial (still-streaming) artifact from text.
 *
 * During streaming, the closing `</artifact>` tag hasn't arrived yet.
 * This function finds the open tag and treats everything after it as
 * in-progress content, enabling live preview in the UI.
 * Already-complete artifacts are stripped first to avoid false matches.
 */
export function extractStreamingArtifact(text: string): ParsedArtifact | null {
  ARTIFACT_RE.lastIndex = 0
  const stripped = text.replace(ARTIFACT_RE, "")

  const match = stripped.match(/<artifact\s+([^>]*)>([\s\S]*)$/s)
  if (!match) return null

  const attrs = parseAttrs(match[1]!)
  const type = (attrs.type as ParsedArtifact["type"]) ?? "code"
  return {
    id: attrs.id ?? "streaming",
    type,
    title: attrs.title ?? "Generando...",
    content: match[2] ?? "",
    language: attrs.language ?? (type === "mermaid" ? "mermaid" : undefined),
    version: 1,
    isStreaming: true,
  }
}

/**
 * Strip all artifact tags (complete and partial) from text.
 *
 * Used during streaming to show only the prose portion of the LLM response,
 * while artifact content is rendered separately in dedicated panels.
 */
export function stripArtifactTags(text: string): string {
  ARTIFACT_RE.lastIndex = 0
  let clean = text.replace(ARTIFACT_RE, "\n\n")

  // Strip partial opening tag + content at end (still streaming)
  const partial = clean.match(/<artifact[\s\S]*$/s)
  if (partial?.index != null) {
    clean = clean.slice(0, partial.index)
  }

  return clean.trim()
}

/**
 * Fallback: extract markdown fenced code blocks as artifacts.
 *
 * Only blocks with 3+ lines are promoted to artifacts — short snippets
 * are left inline. This covers LLM responses that don't use `<artifact>`
 * tags but still contain substantial code worth rendering in a panel.
 */
export function extractCodeBlocks(text: string): ParsedArtifact[] {
  const artifacts: ParsedArtifact[] = []
  const re = /```(\w+)?\n([\s\S]*?)```/g
  let m: RegExpExecArray | null
  while ((m = re.exec(text))) {
    const lang = m[1] || "text"
    const content = m[2] ?? ""
    if (content.trim().split("\n").length < 3) continue
    artifacts.push({
      id: nextId(),
      type: lang === "mermaid" ? "mermaid" : "code",
      title: lang === "mermaid" ? "Diagrama" : `Código ${lang}`,
      content: content.trimEnd(),
      language: lang,
      version: 1,
    })
  }
  return artifacts
}
