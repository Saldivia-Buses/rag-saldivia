"use client"

import { useEffect, useState } from "react"
import { Plus, Trash2, FileText } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { DbScheduledReport } from "@rag-saldivia/db"
import { formatDate } from "@/lib/utils"

const SCHEDULE_LABELS: Record<string, string> = { daily: "Diario", weekly: "Semanal", monthly: "Mensual" }

const SELECT_CLASS = "h-9 w-full rounded-md border border-border bg-bg px-3 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"

export function ReportsAdmin() {
  const [reports, setReports] = useState<DbScheduledReport[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ query: "", collection: "", schedule: "daily", destination: "saved", email: "" })
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    fetch("/api/admin/reports").then((r) => r.json())
      .then((d: { ok: boolean; data?: DbScheduledReport[] }) => { if (d.ok) setReports(d.data ?? []) })
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    try {
      const res = await fetch("/api/admin/reports", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(form) })
      const d = await res.json() as { ok: boolean; data?: DbScheduledReport }
      if (d.ok && d.data) { setReports((p) => [...p, d.data!]); setShowCreate(false); setForm({ query: "", collection: "", schedule: "daily", destination: "saved", email: "" }) }
    } finally { setCreating(false) }
  }

  async function handleDelete(id: string) {
    await fetch(`/api/admin/reports?id=${id}`, { method: "DELETE" })
    setReports((p) => p.filter((r) => r.id !== id))
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Informes programados</h1>
          <p className="text-sm text-fg-muted mt-0.5">Queries automáticas con entrega periódica</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <Plus className="h-3.5 w-3.5" /> Nuevo informe
        </Button>
      </div>

      {showCreate && (
        <form onSubmit={handleCreate} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Nuevo informe</h3>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Query</label>
            <Textarea value={form.query} onChange={(e) => setForm((p) => ({ ...p, query: e.target.value }))} placeholder="¿Cuáles son los contratos vigentes?" required className="min-h-[60px]" />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Colección</label>
            <Input value={form.collection} onChange={(e) => setForm((p) => ({ ...p, collection: e.target.value }))} required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-fg-muted">Frecuencia</label>
              <select value={form.schedule} onChange={(e) => setForm((p) => ({ ...p, schedule: e.target.value }))} className={SELECT_CLASS}>
                <option value="daily">Diario</option><option value="weekly">Semanal</option><option value="monthly">Mensual</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-fg-muted">Destino</label>
              <select value={form.destination} onChange={(e) => setForm((p) => ({ ...p, destination: e.target.value }))} className={SELECT_CLASS}>
                <option value="saved">Guardados</option><option value="email">Email</option>
              </select>
            </div>
          </div>
          {form.destination === "email" && (
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-fg-muted">Email destino</label>
              <Input type="email" value={form.email} onChange={(e) => setForm((p) => ({ ...p, email: e.target.value }))} />
            </div>
          )}
          <div className="flex gap-2 pt-1">
            <Button size="sm" type="submit" disabled={creating}>{creating ? "Creando..." : "Crear informe"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {reports.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FileText} />
          <EmptyPlaceholder.Title>Sin informes configurados</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Creá un informe para recibir respuestas automáticas periódicas.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="space-y-2">
          {reports.map((r) => (
            <div key={r.id} className="p-4 rounded-xl border border-border bg-surface flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="text-sm font-medium text-fg truncate">{r.query}</p>
                <p className="text-xs text-fg-muted mt-0.5">
                  {r.collection} · {SCHEDULE_LABELS[r.schedule]} · {r.destination === "email" ? r.email : "Guardados"}
                </p>
                {r.lastRun && <p className="text-xs text-fg-subtle mt-0.5">Último: {formatDate(r.lastRun)}</p>}
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Badge variant="outline">{SCHEDULE_LABELS[r.schedule]}</Badge>
                <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" onClick={() => handleDelete(r.id)}>
                  <Trash2 size={13} />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
