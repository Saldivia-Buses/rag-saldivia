"use client"

import { useState } from "react"
import type { DbEvent } from "@rag-saldivia/db"

const LEVEL_STYLES: Record<string, { bg: string; color: string }> = {
  ERROR: { bg: "#fef2f2", color: "#991b1b" },
  FATAL: { bg: "#fef2f2", color: "#7f1d1d" },
  WARN: { bg: "#fffbeb", color: "#92400e" },
  INFO: { bg: "#f0fdf4", color: "#166534" },
  DEBUG: { bg: "#eff6ff", color: "#1e40af" },
  TRACE: { bg: "#f9fafb", color: "#6b7280" },
}

export function AuditTable({
  events,
  isAdmin,
}: {
  events: DbEvent[]
  isAdmin: boolean
}) {
  const [filter, setFilter] = useState("")
  const [levelFilter, setLevelFilter] = useState("ALL")

  const filtered = events.filter((e) => {
    const matchesLevel = levelFilter === "ALL" || e.level === levelFilter
    const matchesSearch = !filter ||
      e.type.includes(filter) ||
      JSON.stringify(e.payload).toLowerCase().includes(filter.toLowerCase())
    return matchesLevel && matchesSearch
  })

  return (
    <div className="space-y-4">
      {/* Filtros */}
      <div className="flex gap-3">
        <input
          placeholder="Buscar por tipo o contenido..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="flex-1 px-3 py-2 rounded-md border text-sm"
          style={{ borderColor: "var(--border)", background: "var(--background)" }}
        />
        <select
          value={levelFilter}
          onChange={(e) => setLevelFilter(e.target.value)}
          className="px-3 py-2 rounded-md border text-sm"
          style={{ borderColor: "var(--border)", background: "var(--background)" }}
        >
          <option value="ALL">Todos los niveles</option>
          <option value="FATAL">FATAL</option>
          <option value="ERROR">ERROR</option>
          <option value="WARN">WARN</option>
          <option value="INFO">INFO</option>
          <option value="DEBUG">DEBUG</option>
        </select>
      </div>

      {/* Tabla */}
      <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
        <table className="w-full text-sm">
          <thead>
            <tr style={{ background: "var(--muted)" }}>
              <th className="text-left px-4 py-3 font-medium">Timestamp</th>
              <th className="text-left px-4 py-3 font-medium">Nivel</th>
              <th className="text-left px-4 py-3 font-medium">Tipo</th>
              {isAdmin && <th className="text-left px-4 py-3 font-medium">Usuario</th>}
              <th className="text-left px-4 py-3 font-medium">Detalle</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 && (
              <tr>
                <td colSpan={isAdmin ? 5 : 4} className="px-4 py-8 text-center" style={{ color: "var(--muted-foreground)" }}>
                  No hay eventos que coincidan
                </td>
              </tr>
            )}
            {filtered.map((event, i) => {
              const style = LEVEL_STYLES[event.level] ?? LEVEL_STYLES["INFO"]!
              const ts = new Date(event.ts).toISOString().replace("T", " ").slice(0, 19)
              const payload = event.payload as Record<string, unknown>

              return (
                <tr
                  key={event.id}
                  style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}
                >
                  <td className="px-4 py-3 font-mono text-xs" style={{ color: "var(--muted-foreground)" }}>
                    {ts}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className="px-2 py-0.5 rounded text-xs font-medium"
                      style={{ background: style.bg, color: style.color }}
                    >
                      {event.level}
                    </span>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs font-medium">{event.type}</td>
                  {isAdmin && (
                    <td className="px-4 py-3 text-xs" style={{ color: "var(--muted-foreground)" }}>
                      {event.userId ?? "—"}
                    </td>
                  )}
                  <td className="px-4 py-3 text-xs max-w-xs truncate" style={{ color: "var(--muted-foreground)" }}>
                    {JSON.stringify(payload).slice(0, 80)}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>

      <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
        Mostrando {filtered.length} de {events.length} eventos
      </p>
    </div>
  )
}
