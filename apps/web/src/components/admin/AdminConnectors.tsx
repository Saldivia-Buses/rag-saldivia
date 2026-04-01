"use client"

import { useState, useTransition, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Plug, Plus, Trash2, Power, PowerOff, RefreshCw } from "lucide-react"
import { toast } from "sonner"
import {
  actionListConnectors,
  actionCreateConnector,
  actionDeleteConnector,
  actionToggleConnector,
  actionSyncNow,
} from "@/app/actions/connectors"
import type { ConnectorProvider } from "@rag-saldivia/shared"

type ConnectorRow = {
  id: string
  provider: string
  name: string
  collectionDest: string
  schedule: string
  active: boolean
  lastSync: number | null
  docCount: number
}

const PROVIDER_LABELS: Record<string, string> = {
  google_drive: "Google Drive",
  sharepoint: "SharePoint",
  confluence: "Confluence",
  web_crawler: "Web Crawler",
}

function timeAgo(ts: number | null): string {
  if (!ts) return "Nunca"
  const diff = Date.now() - ts
  if (diff < 60_000) return "Hace segundos"
  if (diff < 3600_000) return `Hace ${Math.floor(diff / 60_000)} min`
  if (diff < 86400_000) return `Hace ${Math.floor(diff / 3600_000)}h`
  return `Hace ${Math.floor(diff / 86400_000)}d`
}

export function AdminConnectors({ initialConnectors }: { initialConnectors: ConnectorRow[] }) {
  const [connectors, setConnectors] = useState(initialConnectors)
  const [showForm, setShowForm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ConnectorRow | null>(null)
  const [isPending, startTransition] = useTransition()

  const refresh = useCallback(() => {
    startTransition(async () => {
      const result = await actionListConnectors({})
      if (result?.data?.connectors) setConnectors(result.data.connectors as ConnectorRow[])
    })
  }, [startTransition])

  const handleDelete = useCallback((c: ConnectorRow) => {
    startTransition(async () => {
      await actionDeleteConnector({ id: c.id })
      toast.success(`Conector "${c.name}" eliminado`)
      refresh()
    })
    setDeleteTarget(null)
  }, [startTransition, refresh])

  const handleToggle = useCallback((c: ConnectorRow) => {
    startTransition(async () => {
      await actionToggleConnector({ id: c.id, active: !c.active })
      toast.success(c.active ? "Conector desactivado" : "Conector activado")
      refresh()
    })
  }, [startTransition, refresh])

  const handleSync = useCallback((c: ConnectorRow) => {
    startTransition(async () => {
      await actionSyncNow({ id: c.id, provider: c.provider, collectionDest: c.collectionDest })
      toast.success("Sincronización encolada")
    })
  }, [startTransition])

  if (connectors.length === 0 && !showForm) {
    return (
      <div>
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Plug} />
          <EmptyPlaceholder.Title>Sin conectores</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            Configurá un conector para sincronizar documentos desde Google Drive, SharePoint, Confluence o URLs web.
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
        <div style={{ marginTop: "16px", textAlign: "center" }}>
          <Button onClick={() => setShowForm(true)}>
            <Plus size={16} style={{ marginRight: "6px" }} /> Agregar conector
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <h2 className="text-lg font-semibold text-fg">Conectores</h2>
        <Button size="sm" onClick={() => setShowForm(true)} disabled={showForm}>
          <Plus size={16} style={{ marginRight: "6px" }} /> Agregar
        </Button>
      </div>

      {showForm && (
        <ConnectorForm
          onSave={() => { setShowForm(false); refresh() }}
          onCancel={() => setShowForm(false)}
        />
      )}

      <div className="flex flex-col gap-3">
        {connectors.map((c) => (
          <div
            key={c.id}
            className="flex items-center justify-between rounded-xl border border-border bg-surface"
            style={{ padding: "16px 20px" }}
          >
            <div className="flex items-center gap-3" style={{ flex: 1, minWidth: 0 }}>
              <Badge variant={c.active ? "success" : "secondary"}>
                {c.active ? "Activo" : "Inactivo"}
              </Badge>
              <div style={{ minWidth: 0 }}>
                <span className="font-medium text-fg">{c.name}</span>
                <span className="text-fg-muted text-sm" style={{ marginLeft: "8px" }}>
                  {PROVIDER_LABELS[c.provider] ?? c.provider}
                </span>
                <div className="text-xs text-fg-subtle" style={{ marginTop: "2px" }}>
                  {c.docCount} docs · {c.schedule} · Última sync: {timeAgo(c.lastSync)} · Destino: {c.collectionDest}
                </div>
              </div>
            </div>
            <div style={{ display: "flex", gap: "4px", flexShrink: 0 }}>
              <Button variant="ghost" size="icon" onClick={() => handleSync(c)} disabled={isPending || !c.active} title="Sincronizar ahora">
                <RefreshCw size={16} />
              </Button>
              <Button variant="ghost" size="icon" onClick={() => handleToggle(c)} disabled={isPending} title={c.active ? "Desactivar" : "Activar"}>
                {c.active ? <PowerOff size={16} /> : <Power size={16} />}
              </Button>
              <Button variant="ghost" size="icon" onClick={() => setDeleteTarget(c)} disabled={isPending} title="Eliminar">
                <Trash2 size={16} className="text-destructive" />
              </Button>
            </div>
          </div>
        ))}
      </div>

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={`Eliminar "${deleteTarget?.name}"`}
        description="Se eliminará el conector y todos los registros de sincronización. Los documentos ya ingestados no se borran."
        onConfirm={() => deleteTarget && handleDelete(deleteTarget)}
        variant="destructive"
      />
    </div>
  )
}

// ── Create form ───────────────────────────────────────────────────────────

function ConnectorForm({ onSave, onCancel }: { onSave: () => void; onCancel: () => void }) {
  const [provider, setProvider] = useState<ConnectorProvider>("google_drive")
  const [name, setName] = useState("")
  const [collectionDest, setCollectionDest] = useState("")
  const [schedule, setSchedule] = useState<"hourly" | "daily" | "weekly">("daily")
  const [isPending, startTransition] = useTransition()

  // Provider-specific fields
  const [baseUrl, setBaseUrl] = useState("")
  const [email, setEmail] = useState("")
  const [apiToken, setApiToken] = useState("")
  const [spaceKey, setSpaceKey] = useState("")
  const [urls, setUrls] = useState("")
  const [maxDepth, setMaxDepth] = useState("2")

  function buildCredentials(): string {
    switch (provider) {
      case "confluence":
        return JSON.stringify({ baseUrl, email, apiToken, ...(spaceKey ? { spaceKey } : {}) })
      case "web_crawler":
        return JSON.stringify({
          urls: urls.split("\n").map((u) => u.trim()).filter(Boolean),
          maxDepth: parseInt(maxDepth) || 2,
        })
      case "google_drive":
      case "sharepoint":
        // OAuth flow — credentials come from callback, not form
        return "{}"
    }
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    startTransition(async () => {
      const creds = buildCredentials()
      const result = await actionCreateConnector({
        provider,
        name: name || PROVIDER_LABELS[provider] || provider,
        collectionDest,
        schedule,
        credentials: creds,
      })
      if (result?.data) {
        toast.success("Conector creado")
        onSave()
      }
    })
  }

  const needsOAuth = provider === "google_drive" || provider === "sharepoint"

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-xl border border-border bg-surface flex flex-col gap-4"
      style={{ padding: "20px" }}
    >
      <h3 className="font-medium text-fg">Nuevo conector</h3>

      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">Tipo</label>
        <select
          value={provider}
          onChange={(e) => setProvider(e.target.value as ConnectorProvider)}
          className="h-10 rounded-lg border border-border bg-bg px-3 text-sm text-fg"
        >
          <option value="google_drive">Google Drive</option>
          <option value="sharepoint">SharePoint / OneDrive</option>
          <option value="confluence">Confluence</option>
          <option value="web_crawler">Web Crawler</option>
        </select>
      </div>

      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">Nombre</label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder={PROVIDER_LABELS[provider]} />
      </div>

      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">Colección destino</label>
        <Input value={collectionDest} onChange={(e) => setCollectionDest(e.target.value)} placeholder="nombre-coleccion" required />
      </div>

      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">Frecuencia de sincronización</label>
        <select
          value={schedule}
          onChange={(e) => setSchedule(e.target.value as "hourly" | "daily" | "weekly")}
          className="h-10 rounded-lg border border-border bg-bg px-3 text-sm text-fg"
        >
          <option value="hourly">Cada hora</option>
          <option value="daily">Diario (2 AM)</option>
          <option value="weekly">Semanal (domingo 2 AM)</option>
        </select>
      </div>

      {/* OAuth notice for Google/SharePoint */}
      {needsOAuth && (
        <div className="text-sm text-fg-muted bg-bg rounded-lg" style={{ padding: "12px" }}>
          Después de crear el conector, se te redirigirá para autorizar el acceso con {PROVIDER_LABELS[provider]}.
          Necesitás configurar las variables de entorno <code className="text-xs">
            {provider === "google_drive" ? "GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET" : "AZURE_CLIENT_ID / AZURE_CLIENT_SECRET"}
          </code> en el servidor.
        </div>
      )}

      {/* Confluence fields */}
      {provider === "confluence" && (
        <>
          <div className="flex flex-col gap-2">
            <label className="text-sm text-fg-muted">URL base de Confluence</label>
            <Input value={baseUrl} onChange={(e) => setBaseUrl(e.target.value)} placeholder="https://tu-empresa.atlassian.net" required />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm text-fg-muted">Email</label>
            <Input value={email} onChange={(e) => setEmail(e.target.value)} placeholder="usuario@empresa.com" required />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm text-fg-muted">API Token</label>
            <Input type="password" value={apiToken} onChange={(e) => setApiToken(e.target.value)} placeholder="••••••••" required />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm text-fg-muted">Space Key (opcional)</label>
            <Input value={spaceKey} onChange={(e) => setSpaceKey(e.target.value)} placeholder="TEAM" />
          </div>
        </>
      )}

      {/* Web Crawler fields */}
      {provider === "web_crawler" && (
        <>
          <div className="flex flex-col gap-2">
            <label className="text-sm text-fg-muted">URLs a crawlear (una por línea)</label>
            <Textarea value={urls} onChange={(e) => setUrls(e.target.value)} placeholder={"https://docs.empresa.com\nhttps://wiki.empresa.com"} rows={3} />
          </div>
          <div className="flex flex-col gap-2">
            <label className="text-sm text-fg-muted">Profundidad máxima</label>
            <Input type="number" value={maxDepth} onChange={(e) => setMaxDepth(e.target.value)} min="0" max="5" />
          </div>
        </>
      )}

      <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}>
        <Button type="button" variant="outline" onClick={onCancel} disabled={isPending}>Cancelar</Button>
        <Button type="submit" disabled={isPending || !collectionDest}>
          {isPending ? "Creando..." : "Crear conector"}
        </Button>
      </div>
    </form>
  )
}
