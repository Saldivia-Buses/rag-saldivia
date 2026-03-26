"use client"

import { useState } from "react"
import type { DbEvent } from "@rag-saldivia/db"
import { Input } from "@/components/ui/input"
import { Badge, type BadgeVariant } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

const LEVEL_BADGE: Record<string, BadgeVariant> = {
  ERROR: "destructive",
  FATAL: "destructive",
  WARN:  "warning",
  INFO:  "success",
  DEBUG: "secondary",
  TRACE: "outline",
}

export function AuditTable({ events, isAdmin }: { events: DbEvent[]; isAdmin: boolean }) {
  const [filter, setFilter] = useState("")
  const [levelFilter, setLevelFilter] = useState("ALL")

  const filtered = events.filter((e) => {
    const matchesLevel  = levelFilter === "ALL" || e.level === levelFilter
    const matchesSearch = !filter || e.type.includes(filter) ||
      JSON.stringify(e.payload).toLowerCase().includes(filter.toLowerCase())
    return matchesLevel && matchesSearch
  })

  return (
    <div className="p-6 space-y-5">
      <div>
        <h1 className="text-lg font-semibold text-fg">Auditoría</h1>
        <p className="text-sm text-fg-muted mt-0.5">Historial de eventos del sistema</p>
      </div>

      {/* Filtros */}
      <div className="flex gap-3">
        <Input
          placeholder="Buscar por tipo o contenido..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="flex-1"
        />
        <select
          value={levelFilter}
          onChange={(e) => setLevelFilter(e.target.value)}
          className="h-9 rounded-md border border-border bg-bg px-3 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"
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
      <div className="rounded-xl border border-border overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>Nivel</TableHead>
              <TableHead>Tipo</TableHead>
              {isAdmin && <TableHead>Usuario</TableHead>}
              <TableHead>Detalle</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={isAdmin ? 5 : 4} className="text-center py-8 text-fg-muted">
                  No hay eventos que coincidan
                </TableCell>
              </TableRow>
            )}
            {filtered.map((event) => {
              const ts = new Date(event.ts).toISOString().replace("T", " ").slice(0, 19)
              const payload = event.payload as Record<string, unknown>
              const variant = LEVEL_BADGE[event.level] ?? "outline"

              return (
                <TableRow key={event.id}>
                  <TableCell className="font-mono text-xs text-fg-muted">{ts}</TableCell>
                  <TableCell>
                    <Badge variant={variant}>{event.level}</Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs font-medium text-fg">{event.type}</TableCell>
                  {isAdmin && (
                    <TableCell className="text-xs text-fg-muted">{event.userId ?? "—"}</TableCell>
                  )}
                  <TableCell className="text-xs text-fg-muted max-w-xs truncate">
                    {JSON.stringify(payload).slice(0, 80)}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>

      <p className="text-xs text-fg-subtle">
        Mostrando {filtered.length} de {events.length} eventos
      </p>
    </div>
  )
}
