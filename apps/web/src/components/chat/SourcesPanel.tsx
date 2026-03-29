"use client"

import { useState } from "react"
import { ChevronDown, ChevronRight, FileText } from "lucide-react"
import { Badge } from "@/components/ui/badge"

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

  const typed = sources as Source[]
  if (!typed || typed.length === 0) return null

  return (
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
                  <span
                    className="font-medium truncate"
                    style={{ color: "var(--accent)" }}
                  >
                    {name}
                  </span>
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
  )
}
