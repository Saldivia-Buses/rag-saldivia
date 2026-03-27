"use client"

import { useEffect, useState } from "react"
import { GitCommit, Upload } from "lucide-react"
import type { DbCollectionHistory } from "@rag-saldivia/db"
import { formatDateTime } from "@/lib/utils"

type Props = {
  collection: string
}

export function CollectionHistory({ collection }: Props) {
  const [history, setHistory] = useState<DbCollectionHistory[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch(`/api/collections/${encodeURIComponent(collection)}/history`)
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: DbCollectionHistory[] }) => {
        if (d.ok) setHistory(d.data ?? [])
      })
      .finally(() => setLoading(false))
  }, [collection])

  if (loading) return <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Cargando historial...</p>
  if (history.length === 0) return <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Sin historial de ingestas.</p>

  return (
    <div className="space-y-2">
      {history.map((entry) => (
        <div key={entry.id} className="flex items-start gap-3">
          <div
            className="w-7 h-7 rounded-full flex items-center justify-center shrink-0 mt-0.5"
            style={{ background: "var(--muted)" }}
          >
            {entry.action === "added"
              ? <Upload size={12} style={{ color: "var(--accent)" }} />
              : <GitCommit size={12} style={{ color: "var(--muted-foreground)" }} />}
          </div>
          <div>
            <p className="text-sm">
              <span className="font-medium">
                {entry.action === "added" ? "Agregado" : "Eliminado"}
              </span>
              {entry.filename && <span className="ml-1 text-xs" style={{ color: "var(--muted-foreground)" }}>{entry.filename}</span>}
            </p>
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              {formatDateTime(entry.createdAt)}
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}
