/**
 * Parser de CHANGELOG.md.
 * Exportado para poder ser testado de forma aislada.
 */

export type ChangelogEntry = {
  version: string
  content: string
}

/**
 * Parsea el contenido crudo de un CHANGELOG.md y retorna las primeras N entradas.
 * Cada entrada empieza con `## [version]`.
 */
export function parseChangelog(raw: string, limit = 5): ChangelogEntry[] {
  const entries: ChangelogEntry[] = []
  const sections = raw.split(/^## /m).slice(1)

  for (const section of sections.slice(0, limit)) {
    const firstLine = section.split("\n")[0] ?? ""
    const versionMatch = firstLine.match(/\[([^\]]+)\]/)
    const version = versionMatch?.[1] ?? firstLine.trim()
    const content = section.split("\n").slice(1).join("\n").trim()
    if (version) entries.push({ version, content })
  }

  return entries
}
