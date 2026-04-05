"use client"

import { useState, useTransition, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { KeyRound, Plus, Trash2, Power, PowerOff } from "lucide-react"
import { toast } from "sonner"
import {
  actionCreateSsoProvider,
  actionDeleteSsoProvider,
  actionUpdateSsoProvider,
  actionListSsoProviders,
} from "@/app/actions/sso"
import type { SsoProviderType } from "@rag-saldivia/shared"

type ProviderRow = {
  id: number
  name: string
  type: string
  clientId: string
  active: boolean
  autoProvision: boolean
  defaultRole: string
  createdAt: number
}

const TYPE_LABELS: Record<string, string> = {
  google: "Google",
  microsoft: "Microsoft",
  github: "GitHub",
  oidc_generic: "OIDC Genérico",
  saml: "SAML 2.0",
}

export function AdminSso({ initialProviders }: { initialProviders: ProviderRow[] }) {
  const [providers, setProviders] = useState(initialProviders)
  const [showForm, setShowForm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<ProviderRow | null>(null)
  const [isPending, startTransition] = useTransition()

  const refresh = useCallback(() => {
    startTransition(async () => {
      const result = await actionListSsoProviders({})
      if (result?.data?.providers) setProviders(result.data.providers as ProviderRow[])
    })
  }, [startTransition])

  const handleDelete = useCallback((provider: ProviderRow) => {
    startTransition(async () => {
      await actionDeleteSsoProvider({ id: provider.id })
      toast.success(`Proveedor "${provider.name}" eliminado`)
      refresh()
    })
    setDeleteTarget(null)
  }, [startTransition, refresh])

  const handleToggle = useCallback((provider: ProviderRow) => {
    startTransition(async () => {
      await actionUpdateSsoProvider({ id: provider.id, active: !provider.active })
      toast.success(provider.active ? "Proveedor desactivado" : "Proveedor activado")
      refresh()
    })
  }, [startTransition, refresh])

  if (providers.length === 0 && !showForm) {
    return (
      <div>
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={KeyRound} />
          <EmptyPlaceholder.Title>Sin proveedores SSO</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            Configurá un proveedor de identidad para que los usuarios inicien sesión con Google, Microsoft, GitHub o SAML.
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
        <div style={{ marginTop: "16px", textAlign: "center" }}>
          <Button onClick={() => setShowForm(true)}>
            <Plus size={16} style={{ marginRight: "6px" }} /> Agregar proveedor
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <h2 className="text-lg font-semibold text-fg">Proveedores SSO</h2>
        <Button size="sm" onClick={() => setShowForm(true)} disabled={showForm}>
          <Plus size={16} style={{ marginRight: "6px" }} /> Agregar
        </Button>
      </div>

      {/* Create form */}
      {showForm && (
        <SsoProviderForm
          onSave={() => { setShowForm(false); refresh() }}
          onCancel={() => setShowForm(false)}
        />
      )}

      {/* Provider list */}
      <div className="flex flex-col gap-3">
        {providers.map((p) => (
          <div
            key={p.id}
            className="flex items-center justify-between rounded-xl border border-border bg-surface"
            style={{ padding: "16px 20px" }}
          >
            <div className="flex items-center gap-3">
              <Badge variant={p.active ? "success" : "secondary"}>
                {p.active ? "Activo" : "Inactivo"}
              </Badge>
              <div>
                <span className="font-medium text-fg">{p.name}</span>
                <span className="text-fg-muted text-sm" style={{ marginLeft: "8px" }}>
                  {TYPE_LABELS[p.type] ?? p.type}
                </span>
              </div>
            </div>
            <div style={{ display: "flex", gap: "8px" }}>
              <Button
                variant="ghost" size="icon"
                onClick={() => handleToggle(p)}
                disabled={isPending}
                title={p.active ? "Desactivar" : "Activar"}
              >
                {p.active ? <PowerOff size={16} /> : <Power size={16} />}
              </Button>
              <Button
                variant="ghost" size="icon"
                onClick={() => setDeleteTarget(p)}
                disabled={isPending}
                title="Eliminar"
              >
                <Trash2 size={16} className="text-destructive" />
              </Button>
            </div>
          </div>
        ))}
      </div>

      {/* Delete confirm */}
      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}
        title={`Eliminar "${deleteTarget?.name}"`}
        description="Se eliminará el proveedor SSO. Los usuarios que se autenticaron con este proveedor podrán seguir usando email/contraseña si tienen una configurada."
        onConfirm={() => deleteTarget && handleDelete(deleteTarget)}
        variant="destructive"
      />
    </div>
  )
}

// ── Create form ───────────────────────────────────────────────────────────

function SsoProviderForm({ onSave, onCancel }: { onSave: () => void; onCancel: () => void }) {
  const [type, setType] = useState<SsoProviderType>("google")
  const [name, setName] = useState("")
  const [clientId, setClientId] = useState("")
  const [clientSecret, setClientSecret] = useState("")
  const [tenantId, setTenantId] = useState("")
  const [issuerUrl, setIssuerUrl] = useState("")
  const [autoProvision, setAutoProvision] = useState(true)
  const [isPending, startTransition] = useTransition()

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    startTransition(async () => {
      const result = await actionCreateSsoProvider({
        name: name || TYPE_LABELS[type] || type,
        type,
        clientId,
        clientSecret,
        ...(tenantId ? { tenantId } : {}),
        ...(issuerUrl ? { issuerUrl } : {}),
        autoProvision,
      })
      if (result?.data) {
        toast.success("Proveedor SSO creado")
        onSave()
      }
    })
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="rounded-xl border border-border bg-surface flex flex-col gap-4"
      style={{ padding: "20px" }}
    >
      <h3 className="font-medium text-fg">Nuevo proveedor SSO</h3>

      {/* Type selector */}
      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">Tipo</label>
        <select
          value={type}
          onChange={(e) => setType(e.target.value as SsoProviderType)}
          className="h-10 rounded-lg border border-border bg-bg px-3 text-sm text-fg"
        >
          <option value="google">Google</option>
          <option value="microsoft">Microsoft / Azure AD</option>
          <option value="github">GitHub</option>
          <option value="oidc_generic">OIDC Genérico</option>
          <option value="saml">SAML 2.0</option>
        </select>
      </div>

      {/* Name */}
      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">Nombre (visible en login)</label>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={TYPE_LABELS[type] ?? "Mi proveedor"}
        />
      </div>

      {/* Client ID / Entity ID */}
      <div className="flex flex-col gap-2">
        <label className="text-sm text-fg-muted">
          {type === "saml" ? "Entity ID" : "Client ID"}
        </label>
        <Input
          value={clientId}
          onChange={(e) => setClientId(e.target.value)}
          placeholder={type === "saml" ? "https://app.example.com/saml" : "abc123.apps.googleusercontent.com"}
          required
        />
      </div>

      {/* Client Secret (not needed for SAML) */}
      {type !== "saml" && (
        <div className="flex flex-col gap-2">
          <label className="text-sm text-fg-muted">Client Secret</label>
          <Input
            type="password"
            value={clientSecret}
            onChange={(e) => setClientSecret(e.target.value)}
            placeholder="••••••••"
            required
          />
        </div>
      )}

      {/* Microsoft: Tenant ID */}
      {type === "microsoft" && (
        <div className="flex flex-col gap-2">
          <label className="text-sm text-fg-muted">Tenant ID</label>
          <Input
            value={tenantId}
            onChange={(e) => setTenantId(e.target.value)}
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
          />
        </div>
      )}

      {/* Generic OIDC / SAML: Issuer URL */}
      {(type === "oidc_generic" || type === "saml") && (
        <div className="flex flex-col gap-2">
          <label className="text-sm text-fg-muted">
            {type === "saml" ? "Entry Point (URL de login del IdP)" : "Issuer URL"}
          </label>
          <Input
            value={issuerUrl}
            onChange={(e) => setIssuerUrl(e.target.value)}
            placeholder={type === "saml" ? "https://idp.example.com/sso/saml" : "https://accounts.example.com"}
          />
        </div>
      )}

      {/* Auto-provision toggle */}
      <label className="flex items-center gap-2 text-sm text-fg">
        <input
          type="checkbox"
          checked={autoProvision}
          onChange={(e) => setAutoProvision(e.target.checked)}
          className="rounded"
        />
        Crear usuarios automáticamente en el primer login SSO
      </label>

      {/* Actions */}
      <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}>
        <Button type="button" variant="outline" onClick={onCancel} disabled={isPending}>
          Cancelar
        </Button>
        <Button type="submit" disabled={isPending || !clientId || (type !== "saml" && !clientSecret)}>
          {isPending ? "Creando..." : "Crear proveedor"}
        </Button>
      </div>
    </form>
  )
}
