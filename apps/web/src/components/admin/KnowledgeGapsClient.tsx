"use client"

import { useEffect, useState } from "react"
import { Upload, ExternalLink } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { downloadFile } from "@/lib/export"
import Link from "next/link"

type Gap = {
  messageId: number
  content: string
  sessionId: string
  sessionTitle: string
  collection: string
  timestamp: number
}

export function KnowledgeGapsClient() {
  const [gaps, setGaps] = useState<Gap[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch("/api/admin/knowledge-gaps")
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: Gap[] }) => {
        if (d.ok) setGaps(d.data ?? [])
      })
      .finally(() => setLoading(false))
  }, [])

  function exportCSV() {
    const rows = [
      ["Sesión", "Colección", "Respuesta", "Fecha"],
      ...gaps.map((g) => [
        g.sessionTitle,
        g.collection,
        `"${g.content.replace(/"/g, '""')}"`,
        new Date(g.timestamp).toLocaleDateString("es-AR"),
      ]),
    ]
    const csv = rows.map((r) => r.join(",")).join("\n")
    downloadFile(csv, "knowledge-gaps.csv", "text/csv")
  }

  if (loading) return <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Cargando...</p>

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
          {gaps.length} brecha{gaps.length !== 1 ? "s" : ""} detectada{gaps.length !== 1 ? "s" : ""}
        </p>
        <div className="flex gap-2">
          {gaps.length > 0 && (
            <Button variant="outline" size="sm" onClick={exportCSV} className="gap-1.5">
              Exportar CSV
            </Button>
          )}
          <Button size="sm" asChild className="gap-1.5">
            <Link href="/upload">
              <Upload size={13} /> Ingestar documentos
            </Link>
          </Button>
        </div>
      </div>

      {gaps.length === 0 ? (
        <div
          className="rounded-xl border p-12 text-center"
          style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
        >
          <p className="text-sm">¡Sin brechas detectadas! El RAG está respondiendo con confianza.</p>
        </div>
      ) : (
        <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
          <Table>
            <TableHeader>
              <TableRow style={{ background: "var(--muted)" }}>
                <TableHead>Sesión</TableHead>
                <TableHead>Colección</TableHead>
                <TableHead>Respuesta</TableHead>
                <TableHead>Fecha</TableHead>
                <TableHead className="text-right">Acción</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {gaps.map((gap) => (
                <TableRow key={gap.messageId}>
                  <TableCell className="max-w-[150px]">
                    <Link
                      href={`/chat/${gap.sessionId}`}
                      className="truncate block hover:opacity-80"
                      style={{ color: "var(--accent)" }}
                    >
                      {gap.sessionTitle}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <span className="text-xs">{gap.collection}</span>
                  </TableCell>
                  <TableCell className="max-w-[300px]">
                    <p className="text-xs truncate" style={{ color: "var(--muted-foreground)" }}>
                      {gap.content}
                    </p>
                  </TableCell>
                  <TableCell className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                    {new Date(gap.timestamp).toLocaleDateString("es-AR")}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm" className="h-7 text-xs gap-1" asChild>
                      <Link href="/upload">
                        <Upload size={11} /> Ingestar
                      </Link>
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
