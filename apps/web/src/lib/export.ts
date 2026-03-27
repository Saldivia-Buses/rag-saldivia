/**
 * Utilidades de export de sesión de chat.
 * Usado por ExportSession.tsx (F1.9).
 */

import { formatDateTime } from "@/lib/utils"
import type { Citation } from "@rag-saldivia/shared"

type ExportMessage = {
  role: "user" | "assistant"
  content: string
  sources?: Citation[]
}

type ExportSession = {
  title: string
  collection: string
  createdAt: number
  messages: ExportMessage[]
}

/** Serializa la sesión a Markdown con fuentes citadas si existen */
export function exportToMarkdown(session: ExportSession): string {
  const date = formatDateTime(session.createdAt)
  const lines: string[] = [
    `# ${session.title}`,
    ``,
    `**Colección:** ${session.collection}  `,
    `**Fecha:** ${date}`,
    ``,
    `---`,
    ``,
  ]

  for (const msg of session.messages) {
    if (msg.role === "user") {
      lines.push(`**Usuario:** ${msg.content}`, ``)
    } else {
      lines.push(`**Asistente:** ${msg.content}`, ``)
      const sources = msg.sources
      if (sources && sources.length > 0) {
        lines.push(`*Fuentes:*`)
        for (const src of sources) {
          const name = src.document ?? "Documento"
          lines.push(`- ${name}`)
        }
        lines.push(``)
      }
    }
    lines.push(`---`, ``)
  }

  return lines.join("\n")
}

/** Descarga un string como archivo */
export function downloadFile(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

/** Abre el diálogo de impresión del navegador para exportar como PDF */
export function exportToPDF() {
  window.print()
}
