"use client"

import { useOptimistic, useState, useTransition } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Plus, Trash2, Cloud } from "lucide-react"

const ExternalSourceSchema = z.object({
  provider: z.enum(["google_drive", "sharepoint", "confluence"]),
  name: z.string().min(2, "El nombre debe tener al menos 2 caracteres"),
  collectionDest: z.string().min(1, "La colección destino es requerida"),
  schedule: z.enum(["hourly", "daily", "weekly"]),
})
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { DbExternalSource } from "@rag-saldivia/db"
import { formatDate } from "@/lib/utils"
import { actionCreateExternalSource, actionDeleteExternalSource } from "@/app/actions/external-sources"

const PROVIDER_LABELS: Record<string, string> = {
  google_drive: "Google Drive",
  sharepoint: "SharePoint",
  confluence: "Confluence",
}

const SELECT_CLASS = "h-9 w-full rounded-md border border-border bg-bg px-3 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"

export function ExternalSourcesAdmin({ initialSources }: { initialSources: DbExternalSource[] }) {
  const [optimisticSources, applyOptimistic] = useOptimistic(
    initialSources,
    (state, action: { type: "delete"; id: string }) => {
      if (action.type === "delete") return state.filter((s) => s.id !== action.id)
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [showCreate, setShowCreate] = useState(false)

  const sourceForm = useForm<z.infer<typeof ExternalSourceSchema>>({
    resolver: zodResolver(ExternalSourceSchema),
    defaultValues: { provider: "google_drive", name: "", collectionDest: "", schedule: "daily" },
  })

  function handleCreate(data: z.infer<typeof ExternalSourceSchema>) {
    startTransition(async () => {
      const source = await actionCreateExternalSource(data)
      if (source) { setShowCreate(false); sourceForm.reset() }
    })
  }

  function handleDelete(id: string) {
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      await actionDeleteExternalSource(id)
    })
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
        <form onSubmit={sourceForm.handleSubmit(handleCreate)} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Nueva fuente externa</h3>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Provider</label>
            <select {...sourceForm.register("provider")} className={SELECT_CLASS}>
              <option value="google_drive">Google Drive</option>
              <option value="sharepoint">SharePoint</option>
              <option value="confluence">Confluence</option>
            </select>
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Nombre</label>
            <Input {...sourceForm.register("name")} />
            {sourceForm.formState.errors.name && <p className="text-xs text-destructive">{sourceForm.formState.errors.name.message}</p>}
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Colección destino</label>
            <Input {...sourceForm.register("collectionDest")} />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Schedule</label>
            <select {...sourceForm.register("schedule")} className={SELECT_CLASS}>
              <option value="hourly">Cada hora</option>
              <option value="daily">Diario</option>
              <option value="weekly">Semanal</option>
            </select>
          </div>
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={isPending}>{isPending ? "Creando..." : "Agregar"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {optimisticSources.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Cloud} />
          <EmptyPlaceholder.Title>Sin fuentes externas</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Conectá Google Drive, SharePoint o Confluence para sincronizar documentos.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="space-y-2">
          {optimisticSources.map((s) => (
            <div key={s.id} className="flex items-center justify-between p-4 rounded-xl border border-border bg-surface gap-3">
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm text-fg">{s.name}</span>
                  <Badge variant="secondary">{PROVIDER_LABELS[s.provider] ?? s.provider}</Badge>
                  <Badge variant="outline">{s.schedule}</Badge>
                </div>
                <p className="text-xs text-fg-muted mt-0.5">
                  → {s.collectionDest} {s.lastSync ? `· Último sync: ${formatDate(s.lastSync)}` : "· Sin sync aún"}
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
