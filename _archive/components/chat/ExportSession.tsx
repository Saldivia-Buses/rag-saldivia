"use client"

import { Download, FileText, File } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { exportToMarkdown, exportToPDF, downloadFile } from "@/lib/export"
import type { DbChatSession, DbChatMessage } from "@rag-saldivia/db"

type Props = {
  session: DbChatSession & { messages?: DbChatMessage[] }
}

export function ExportSession({ session }: Props) {
  function handleMarkdown() {
    const md = exportToMarkdown({
      title: session.title,
      collection: session.collection,
      createdAt: session.createdAt,
      messages: (session.messages ?? []).map((m) => ({
        role: m.role as "user" | "assistant",
        content: m.content,
        sources: (m.sources as import("@rag-saldivia/shared").Citation[]) ?? [],
      })),
    })
    const filename = `${session.title.replace(/\s+/g, "-").toLowerCase()}.md`
    downloadFile(md, filename, "text/markdown")
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8"
          title="Exportar sesión"
        >
          <Download size={15} />
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-44 p-1">
        <button
          onClick={handleMarkdown}
          className="flex items-center gap-2 w-full px-3 py-2 text-sm rounded-md transition-colors hover:opacity-80"
          style={{ color: "var(--foreground)" }}
        >
          <FileText size={14} />
          Markdown
        </button>
        <button
          onClick={exportToPDF}
          className="flex items-center gap-2 w-full px-3 py-2 text-sm rounded-md transition-colors hover:opacity-80"
          style={{ color: "var(--foreground)" }}
        >
          <File size={14} />
          PDF (imprimir)
        </button>
      </PopoverContent>
    </Popover>
  )
}
