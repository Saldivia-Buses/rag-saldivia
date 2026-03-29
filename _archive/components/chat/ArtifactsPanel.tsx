"use client"

import { useState } from "react"
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Download, Bookmark, Code, FileText, Table } from "lucide-react"
import { downloadFile } from "@/lib/export"
import type { ArtifactData } from "@/hooks/useRagStream"

type Props = {
  artifact: ArtifactData | null
  onClose: () => void
  onSave?: (content: string) => void
}

const TYPE_ICONS: Record<string, React.ReactNode> = {
  code: <Code size={14} />,
  table: <Table size={14} />,
  document: <FileText size={14} />,
}

const TYPE_LABELS: Record<string, string> = {
  code: "Código",
  table: "Tabla",
  document: "Documento",
}

export function ArtifactsPanel({ artifact, onClose, onSave }: Props) {
  const [saved, setSaved] = useState(false)

  function handleExport() {
    if (!artifact) return
    const ext = artifact.type === "code" ? (artifact.language ?? "txt") : artifact.type === "table" ? "md" : "md"
    downloadFile(artifact.content, `artifact.${ext}`, "text/plain")
  }

  function handleSave() {
    if (!artifact) return
    onSave?.(artifact.content)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <Sheet open={!!artifact} onOpenChange={(o) => !o && onClose()}>
      <SheetContent side="right" className="w-[500px] max-w-full flex flex-col p-0">
        <SheetHeader className="px-4 py-3 border-b shrink-0 flex-row items-center justify-between" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-2">
            {artifact ? TYPE_ICONS[artifact.type] : null}
            <SheetTitle className="text-sm font-medium">
              Artifact — {artifact ? TYPE_LABELS[artifact.type] : ""}
            </SheetTitle>
            {artifact?.language && (
              <Badge variant="outline" className="text-xs">{artifact.language}</Badge>
            )}
          </div>
          <div className="flex gap-1">
            <Button variant="ghost" size="sm" className="h-7 gap-1 text-xs" onClick={handleSave} disabled={saved}>
              <Bookmark size={12} />
              {saved ? "Guardado" : "Guardar"}
            </Button>
            <Button variant="ghost" size="sm" className="h-7 gap-1 text-xs" onClick={handleExport}>
              <Download size={12} />
              Exportar
            </Button>
          </div>
        </SheetHeader>

        <div className="flex-1 overflow-y-auto">
          {artifact?.type === "code" ? (
            <pre
              className="p-4 text-xs leading-relaxed overflow-x-auto"
              style={{ background: "var(--muted)", color: "var(--foreground)", fontFamily: "monospace" }}
            >
              {artifact.content}
            </pre>
          ) : (
            <div className="p-4 text-sm leading-relaxed whitespace-pre-wrap" style={{ color: "var(--foreground)" }}>
              {artifact?.content}
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
