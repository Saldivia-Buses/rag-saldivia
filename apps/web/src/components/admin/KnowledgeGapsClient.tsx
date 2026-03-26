"use client"

import { useEffect, useState } from "react"
import { Upload } from "lucide-react"
import { Button } from "@/components/ui/button"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { SkeletonTable } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { downloadFile } from "@/lib/export"
import Link from "next/link"
import { SearchX } from "lucide-react"

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
      .then((d: { ok: boolean; data?: Gap[] }) => { if (d.ok) setGaps(d.data ?? []) })
      .finally(() => setLoading(false))
  }, [])

  function exportCSV() {
    const rows = [
      ["Sesión", "Colección", "Respuesta", "Fecha"],
      ...gaps.map((g) => [g.sessionTitle, g.collection, `"${g.content.replace(/"/g, '""')}"`, new Date(g.timestamp).toLocaleDateString("es-AR")]),
    ]
    downloadFile(rows.map((r) => r.join(",")).join("\n"), "knowledge-gaps.csv", "text/csv")
  }

  if (loading) return (
    <div className="p-6 space-y-4">
      <div className="h-5 w-48 bg-surface-2 animate-pulse rounded" />
      <SkeletonTable rows={4} cols={5} />
    </div>
  )

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Brechas de conocimiento</h1>
          <p className="text-sm text-fg-muted mt-0.5">{gaps.length} brecha{gaps.length !== 1 ? "s" : ""} detectada{gaps.length !== 1 ? "s" : ""}</p>
        </div>
        <div className="flex gap-2">
          {gaps.length > 0 && (
            <Button variant="outline" size="sm" onClick={exportCSV}>Exportar CSV</Button>
          )}
          <Button size="sm" asChild>
            <Link href="/upload"><Upload className="h-3.5 w-3.5" /> Ingestar documentos</Link>
          </Button>
        </div>
      </div>

      {gaps.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={SearchX} />
          <EmptyPlaceholder.Title>Sin brechas detectadas</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>El RAG está respondiendo con confianza a todas las preguntas.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="rounded-xl border border-border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
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
                    <Link href={`/chat/${gap.sessionId}`} className="text-accent hover:opacity-80 truncate block">
                      {gap.sessionTitle}
                    </Link>
                  </TableCell>
                  <TableCell className="text-xs text-fg-muted">{gap.collection}</TableCell>
                  <TableCell className="max-w-[300px]">
                    <p className="text-xs text-fg-muted truncate">{gap.content}</p>
                  </TableCell>
                  <TableCell className="text-xs text-fg-muted">
                    {new Date(gap.timestamp).toLocaleDateString("es-AR")}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button variant="ghost" size="sm" className="h-7 text-xs" asChild>
                      <Link href="/upload"><Upload size={11} /> Ingestar</Link>
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
