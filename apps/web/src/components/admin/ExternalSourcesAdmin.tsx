"use client"

import { useEffect, useState } from "react"
import { Plus, Trash2, Cloud } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { DbExternalSource } from "@rag-saldivia/db"

const PROVIDER_LABELS: Record<string, string> = {
  google_drive: "Google Drive",
  sharepoint: "SharePoint",
  confluence: "Confluence",
}

const SELECT_CLASS = "h-9 w-full rounded-md border border-border bg-bg px-3 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"

export function ExternalSourcesAdmin() {
  const [sources, setSources] = useState<DbExternalSource[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ provider: "google_drive", name: "", collectionDest: "", schedule: "daily" })
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    fetch("/api/admin/external-sources").then((r) => r.json())
      .then((d: { ok: boolean; data?: DbExternalSource[] }) => { if (d.ok) setSources(d.data ?? []) })
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

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Fuentes externas</h1>
          <p className="text-sm text-fg-muted mt-0.5">Sincronización automática desde servicios cloud</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <Plus className="h-3.5 w-3.5" /> Nueva fuente
        </Button>
      </div>

      <div className="rounded-lg bg-surface border border-border p-3 text-sm text-fg-muted">
        ℹ️ La integración OAuth requiere configurar las credenciales de cada provider en las variables de entorno.
      </div>

      {showCreate && (
        <form onSubmit={handleCreate} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Nueva fuente externa</h3>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Provider</label>
            <select value={form.provider} onChange={(e) => setForm((p) => ({ ...p, provider: e.target.value }))} className={SELECT_CLASS}>
              <option value="google_drive">Google Drive</option>
              <option value="sharepoint">SharePoint</option>
              <option value="confluence">Confluence</option>
            </select>
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Nombre</label>
            <Input value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} required />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Colección destino</label>
            <Input value={form.collectionDest} onChange={(e) => setForm((p) => ({ ...p, collectionDest: e.target.value }))} required />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Schedule</label>
            <select value={form.schedule} onChange={(e) => setForm((p) => ({ ...p, schedule: e.target.value }))} className={SELECT_CLASS}>
              <option value="hourly">Cada hora</option>
              <option value="daily">Diario</option>
              <option value="weekly">Semanal</option>
            </select>
          </div>
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={creating}>{creating ? "Creando..." : "Agregar"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {sources.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Cloud} />
          <EmptyPlaceholder.Title>Sin fuentes externas</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Conectá Google Drive, SharePoint o Confluence para sincronizar documentos.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="space-y-2">
          {sources.map((s) => (
            <div key={s.id} className="flex items-center justify-between p-4 rounded-xl border border-border bg-surface gap-3">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm text-fg">{s.name}</span>
                  <Badge variant="secondary">{PROVIDER_LABELS[s.provider] ?? s.provider}</Badge>
                  <Badge variant="outline">{s.schedule}</Badge>
                </div>
                <p className="text-xs text-fg-muted mt-0.5">
                  → {s.collectionDest} {s.lastSync ? `· Último sync: ${new Date(s.lastSync).toLocaleDateString("es-AR")}` : "· Sin sync aún"}
                </p>
              </div>
              <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0 text-destructive hover:text-destructive" onClick={() => handleDelete(s.id)}>
                <Trash2 size={13} />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
