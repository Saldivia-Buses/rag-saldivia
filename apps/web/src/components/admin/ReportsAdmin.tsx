"use client"

import { useEffect, useState } from "react"
import { Plus, Trash2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import type { DbScheduledReport } from "@rag-saldivia/db"

const SCHEDULE_LABELS: Record<string, string> = { daily: "Diario", weekly: "Semanal", monthly: "Mensual" }

export function ReportsAdmin() {
  const [reports, setReports] = useState<DbScheduledReport[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ query: "", collection: "", schedule: "daily", destination: "saved", email: "" })
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    fetch("/api/admin/reports")
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: DbScheduledReport[] }) => { if (d.ok) setReports(d.data ?? []) })
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    try {
      const res = await fetch("/api/admin/reports", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      })
      const d = await res.json() as { ok: boolean; data?: DbScheduledReport }
      if (d.ok && d.data) {
        setReports((p) => [...p, d.data!])
        setShowCreate(false)
        setForm({ query: "", collection: "", schedule: "daily", destination: "saved", email: "" })
      }
    } finally {
      setCreating(false)
    }
  }

  async function handleDelete(id: string) {
    await fetch(`/api/admin/reports?id=${id}`, { method: "DELETE" })
    setReports((p) => p.filter((r) => r.id !== id))
  }

  const inputClass = "w-full px-3 py-1.5 rounded-md border text-sm outline-none"
  const inputStyle = { borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }

  return (
    <div className="space-y-4">
      <Button size="sm" onClick={() => setShowCreate(!showCreate)} className="gap-1.5">
        <Plus size={13} /> Nuevo informe
      </Button>

      {showCreate && (
        <form onSubmit={handleCreate} className="p-4 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
          <div><label className="text-xs font-medium">Query</label><textarea value={form.query} onChange={(e) => setForm((p) => ({ ...p, query: e.target.value }))} className={`${inputClass} mt-1 min-h-[60px]`} style={inputStyle} placeholder="¿Cuáles son los contratos vigentes?" required /></div>
          <div><label className="text-xs font-medium">Colección</label><input value={form.collection} onChange={(e) => setForm((p) => ({ ...p, collection: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle} required /></div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-medium">Frecuencia</label>
              <select value={form.schedule} onChange={(e) => setForm((p) => ({ ...p, schedule: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle}>
                <option value="daily">Diario</option>
                <option value="weekly">Semanal</option>
                <option value="monthly">Mensual</option>
              </select>
            </div>
            <div>
              <label className="text-xs font-medium">Destino</label>
              <select value={form.destination} onChange={(e) => setForm((p) => ({ ...p, destination: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle}>
                <option value="saved">Guardados</option>
                <option value="email">Email</option>
              </select>
            </div>
          </div>
          {form.destination === "email" && <div><label className="text-xs font-medium">Email destino</label><input type="email" value={form.email} onChange={(e) => setForm((p) => ({ ...p, email: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle} /></div>}
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={creating}>{creating ? "Creando..." : "Crear informe"}</Button>
            <Button size="sm" variant="ghost" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      <div className="space-y-2">
        {reports.length === 0 && <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Sin informes configurados.</p>}
        {reports.map((r) => (
          <div key={r.id} className="p-3 rounded-lg border flex items-start justify-between gap-3" style={{ borderColor: "var(--border)" }}>
            <div className="min-w-0">
              <p className="text-sm font-medium truncate">{r.query}</p>
              <p className="text-xs mt-0.5" style={{ color: "var(--muted-foreground)" }}>
                {r.collection} · {SCHEDULE_LABELS[r.schedule]} · {r.destination === "email" ? r.email : "Guardados"}
              </p>
              {r.lastRun && <p className="text-xs mt-0.5" style={{ color: "var(--muted-foreground)" }}>Último: {new Date(r.lastRun).toLocaleDateString("es-AR")}</p>}
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Badge variant="outline" className="text-xs">{SCHEDULE_LABELS[r.schedule]}</Badge>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleDelete(r.id)} style={{ color: "var(--destructive)" }}><Trash2 size={13} /></Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
