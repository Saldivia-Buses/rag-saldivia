"use client"

import { useEffect, useState } from "react"
import { Plus, Trash2, Copy, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import type { DbWebhook } from "@rag-saldivia/db"

const EVENT_OPTIONS = [
  "ingestion.completed",
  "ingestion.error",
  "query.completed",
  "query.low_confidence",
  "user.created",
]

export function WebhooksAdmin() {
  const [hooks, setHooks] = useState<DbWebhook[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [url, setUrl] = useState("")
  const [selectedEvents, setSelectedEvents] = useState<string[]>([])
  const [creating, setCreating] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  useEffect(() => {
    fetch("/api/admin/webhooks").then((r) => r.json()).then((d: { ok: boolean; data?: DbWebhook[] }) => { if (d.ok) setHooks(d.data ?? []) })
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!url || selectedEvents.length === 0) return
    setCreating(true)
    try {
      const res = await fetch("/api/admin/webhooks", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ url, events: selectedEvents }) })
      const d = await res.json() as { ok: boolean; data?: DbWebhook }
      if (d.ok && d.data) { setHooks((p) => [...p, d.data!]); setShowCreate(false); setUrl(""); setSelectedEvents([]) }
    } finally { setCreating(false) }
  }

  async function handleDelete(id: string) {
    await fetch(`/api/admin/webhooks?id=${id}`, { method: "DELETE" })
    setHooks((p) => p.filter((h) => h.id !== id))
  }

  async function copySecret(secret: string, id: string) {
    await navigator.clipboard.writeText(secret)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
  }

  const inputClass = "w-full px-3 py-1.5 rounded-md border text-sm outline-none"
  const inputStyle = { borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }

  return (
    <div className="space-y-4">
      <Button size="sm" onClick={() => setShowCreate(!showCreate)} className="gap-1.5"><Plus size={13} /> Nuevo webhook</Button>

      {showCreate && (
        <form onSubmit={handleCreate} className="p-4 rounded-xl border space-y-3" style={{ borderColor: "var(--border)" }}>
          <div><label className="text-xs font-medium">URL destino</label><input value={url} onChange={(e) => setUrl(e.target.value)} className={`${inputClass} mt-1`} style={inputStyle} placeholder="https://mi-sistema.com/webhook" required /></div>
          <div>
            <label className="text-xs font-medium">Eventos a escuchar</label>
            <div className="flex flex-wrap gap-2 mt-1">
              {EVENT_OPTIONS.map((ev) => (
                <button key={ev} type="button" onClick={() => setSelectedEvents((p) => p.includes(ev) ? p.filter((e) => e !== ev) : [...p, ev])} className="px-2 py-1 rounded-full text-xs border transition-colors" style={{ background: selectedEvents.includes(ev) ? "var(--accent)" : "transparent", color: selectedEvents.includes(ev) ? "white" : "var(--muted-foreground)", borderColor: "var(--border)" }}>{ev}</button>
              ))}
            </div>
          </div>
          <div className="flex gap-2"><Button size="sm" type="submit" disabled={creating || selectedEvents.length === 0}>{creating ? "Creando..." : "Crear webhook"}</Button><Button size="sm" variant="ghost" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button></div>
        </form>
      )}

      <div className="space-y-2">
        {hooks.length === 0 && <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Sin webhooks configurados.</p>}
        {hooks.map((h) => (
          <div key={h.id} className="p-3 rounded-lg border space-y-2" style={{ borderColor: "var(--border)" }}>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <p className="text-sm font-medium truncate">{h.url}</p>
                <div className="flex flex-wrap gap-1 mt-1">
                  {(h.events as string[]).map((ev) => <Badge key={ev} variant="outline" className="text-xs">{ev}</Badge>)}
                </div>
              </div>
              <div className="flex gap-1 shrink-0">
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => copySecret(h.secret, h.id)} title="Copiar secret">{copiedId === h.id ? <Check size={12} /> : <Copy size={12} />}</Button>
                <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => handleDelete(h.id)} style={{ color: "var(--destructive)" }}><Trash2 size={13} /></Button>
              </div>
            </div>
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>Secret: {h.secret.slice(0, 8)}... · Firma: X-Signature: sha256=HMAC</p>
          </div>
        ))}
      </div>
    </div>
  )
}
