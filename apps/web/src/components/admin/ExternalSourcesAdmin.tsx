"use client"

import { useEffect, useState } from "react"
import { Plus, Trash2, RefreshCw, Cloud } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import type { DbExternalSource } from "@rag-saldivia/db"

const PROVIDER_LABELS: Record<string, string> = {
  google_drive: "Google Drive",
  sharepoint: "SharePoint",
  confluence: "Confluence",
}

export function ExternalSourcesAdmin() {
  const [sources, setSources] = useState<DbExternalSource[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ provider: "google_drive", name: "", collectionDest: "", schedule: "daily" })
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    fetch("/api/admin/external-sources").then((r) => r.json()).then((d: { ok: boolean; data?: DbExternalSource[] }) => { if (d.ok) setSources(d.data ?? []) })
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setCreating(true)
    try {
      const res = await fetch("/api/admin/external-sources", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(form) })
      const d = await res.json() as { ok: boolean; data?: DbExternalSource }
      if (d.ok && d.data) { setSources((p) => [...p, d.data!]); setShowCreate(false) }
    } finally { setCreating(false) }
  }

  async function handleDelete(id: string) {
    await fetch(`/api/admin/external-sources?id=${id}`, { method: "DELETE" })
    setSources((p) => p.filter((s) => s.id !== id))
  }

  const inputClass = "w-full px-3 py-1.5 rounded-md border text-sm outline-none"
  const inputStyle = { borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }

  return (
    <div className="space-y-4">
      <div className="p-3 rounded-lg text-sm" style={{ background: "var(--muted)", color: "var(--muted-foreground)" }}>
        ℹ️ La integración OAuth completa requiere configurar las credenciales de cada provider en las variables de entorno. El worker intentará sincronizar automáticamente según el schedule.
      </div>

      <Button size="sm" onClick={() => setShowCreate(!showCreate)} className="gap-1.5"><Plus size={13} /> Nueva fuente</Button>

      {showCreate && (
        <form onSubmit={handleCreate} className="p-4 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
          <div>
            <label className="text-xs font-medium">Provider</label>
            <select value={form.provider} onChange={(e) => setForm((p) => ({ ...p, provider: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle}>
              <option value="google_drive">Google Drive</option>
              <option value="sharepoint">SharePoint</option>
              <option value="confluence">Confluence</option>
            </select>
          </div>
          <div><label className="text-xs font-medium">Nombre</label><input value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle} required /></div>
          <div><label className="text-xs font-medium">Colección destino</label><input value={form.collectionDest} onChange={(e) => setForm((p) => ({ ...p, collectionDest: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle} required /></div>
          <div>
            <label className="text-xs font-medium">Schedule</label>
            <select value={form.schedule} onChange={(e) => setForm((p) => ({ ...p, schedule: e.target.value }))} className={`${inputClass} mt-1`} style={inputStyle}>
              <option value="hourly">Cada hora</option><option value="daily">Diario</option><option value="weekly">Semanal</option>
            </select>
          </div>
          <div className="flex gap-2"><Button size="sm" type="submit" disabled={creating}>{creating ? "Creando..." : "Agregar"}</Button><Button size="sm" variant="ghost" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button></div>
        </form>
      )}

      {sources.length === 0 ? (
        <div className="rounded-xl border p-10 text-center" style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}>
          <Cloud size={28} className="mx-auto mb-2 opacity-40" />
          <p className="text-sm">Sin fuentes externas configuradas.</p>
        </div>
      ) : (
        <div className="space-y-2">
          {sources.map((s) => (
            <div key={s.id} className="flex items-center justify-between p-3 rounded-lg border gap-3" style={{ borderColor: "var(--border)" }}>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm">{s.name}</span>
                  <Badge variant="outline" className="text-xs">{PROVIDER_LABELS[s.provider] ?? s.provider}</Badge>
                  <Badge variant="outline" className="text-xs">{s.schedule}</Badge>
                </div>
                <p className="text-xs mt-0.5" style={{ color: "var(--muted-foreground)" }}>
                  → {s.collectionDest} {s.lastSync ? `· Último sync: ${new Date(s.lastSync).toLocaleDateString("es-AR")}` : "· Sin sync aún"}
                </p>
              </div>
              <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => handleDelete(s.id)} style={{ color: "var(--destructive)" }}><Trash2 size={13} /></Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
