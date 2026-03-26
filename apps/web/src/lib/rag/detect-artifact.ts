/**
 * Detecta artifacts en el contenido de un mensaje del asistente.
 *
 * Dos mecanismos:
 * 1. Marcador explícito :::artifact{type="..." lang="..."}...:::
 * 2. Heurística: código >= 40 líneas, tabla >= 5 columnas
 */

export type ArtifactData = {
  type: "document" | "table" | "code"
  content: string
  language?: string
}

export function detectArtifact(content: string): ArtifactData | null {
  // Marcador explícito del servidor
  const artifactMatch = content.match(/:::artifact\{type="(\w+)"(?:\s+lang="(\w+)")?\}([\s\S]*?):::/)
  if (artifactMatch) {
    return {
      type: (artifactMatch[1] as ArtifactData["type"]) ?? "document",
      content: (artifactMatch[3] ?? "").trim(),
      language: artifactMatch[2],
    }
  }

  // Heurística: bloque de código >= 40 líneas
  const codeMatch = content.match(/```(\w*)\n([\s\S]{500,})```/)
  if (codeMatch) {
    const lines = (codeMatch[2] ?? "").split("\n").length
    if (lines >= 40) {
      return { type: "code", content: (codeMatch[2] ?? "").trim(), language: codeMatch[1] || undefined }
    }
  }

  // Heurística: tabla markdown con >= 5 columnas
  const tableMatch = content.match(/\|([^|\n]+\|){4,}/)
  if (tableMatch) {
    const cols = (tableMatch[0].match(/\|/g) ?? []).length - 1
    if (cols >= 5) {
      const tableStart = content.indexOf(tableMatch[0])
      const tableEnd = content.lastIndexOf("\n\n", tableStart + 200)
      return { type: "table", content: content.slice(tableStart, tableEnd > 0 ? tableEnd : undefined).trim() }
    }
  }

  return null
}
