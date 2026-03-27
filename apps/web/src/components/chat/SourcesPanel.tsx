"use client"

import { useState } from "react"
import { ChevronDown, ChevronRight, FileText, Eye } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { DocPreviewPanel } from "@/components/chat/DocPreviewPanel"

type Source = {
  document_name?: string
  content?: string
  score?: number
  [key: string]: unknown
}

type Props = {
  sources: unknown[]
}

export function SourcesPanel({ sources }: Props) {
  const [expanded, setExpanded] = useState(true)
  const [previewDoc, setPreviewDoc] = useState<{ name: string; snippet?: string } | null>(null)

  const typed = sources as Source[]
  if (!typed || typed.length === 0) return null

  return (
    <>
      <div
        className="mt-2 rounded-lg border text-xs"
        style={{ borderColor: "var(--border)" }}
      >
        <button
          onClick={() => setExpanded((e) => !e)}
          className="flex items-center gap-1.5 w-full px-3 py-2 text-left transition-colors hover:opacity-80"
          style={{ color: "var(--muted-foreground)" }}
        >
          {expanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
          <FileText size={12} />
          <span className="font-medium">
            {typed.length} fuente{typed.length !== 1 ? "s" : ""}
          </span>
        </button>

        {expanded && (
          <div
            className="border-t divide-y"
            style={{ borderColor: "var(--border)" }}
          >
            {typed.map((src, i) => {
              const name = src.document_name ?? (src.filename as string | undefined) ?? `Documento ${i + 1}`
              const snippet = src.content as string | undefined
              const score = src.score as number | undefined

              return (
                <div key={i} className="px-3 py-2 space-y-1">
                  <div className="flex items-center justify-between gap-2">
                    <button
                      onClick={() => setPreviewDoc({ name, ...(snippet !== undefined ? { snippet } : {}) })}
                      className="font-medium truncate text-left hover:opacity-70 transition-opacity flex items-center gap-1"
                      style={{ color: "var(--accent)" }}
                      title="Ver documento"
                    >
                      {name}
                      <Eye size={10} className="shrink-0 opacity-60" />
                    </button>
                    {score !== undefined && (
                      <Badge variant="outline" className="shrink-0 text-xs">
                        {(score * 100).toFixed(0)}%
                      </Badge>
                    )}
                  </div>
                  {snippet && (
                    <p
                      className="leading-relaxed line-clamp-2"
                      style={{ color: "var(--muted-foreground)" }}
                    >
                      {snippet}
                    </p>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>

      {/* Panel de preview — F3.40 */}
      <DocPreviewPanel
        documentName={previewDoc?.name ?? null}
        {...(previewDoc?.snippet !== undefined ? { highlightText: previewDoc.snippet } : {})}
        onClose={() => setPreviewDoc(null)}
      />
    </>
  )
}
