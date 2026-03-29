"use client"

import { useOptimistic, useState, useTransition } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Plus, Trash2, Copy, Check, Webhook } from "lucide-react"

const WebhookSchema = z.object({
  url: z.string().url("URL inválida — debe comenzar con https://"),
})
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { DbWebhook } from "@rag-saldivia/db"
import { actionCreateWebhook, actionDeleteWebhook } from "@/app/actions/webhooks"

const EVENT_OPTIONS = [
  "ingestion.completed", "ingestion.error",
  "query.completed", "query.low_confidence", "user.created",
]

export function WebhooksAdmin({ initialHooks }: { initialHooks: DbWebhook[] }) {
  const [optimisticHooks, applyOptimistic] = useOptimistic(
    initialHooks,
    (state, action: { type: "delete"; id: string } | { type: "create"; hook: DbWebhook }) => {
      if (action.type === "delete") return state.filter((h) => h.id !== action.id)
      if (action.type === "create") return [...state, action.hook]
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [showCreate, setShowCreate] = useState(false)
  const [selectedEvents, setSelectedEvents] = useState<string[]>([])
  const [copiedId, setCopiedId] = useState<string | null>(null)

  const urlForm = useForm<{ url: string }>({
    resolver: zodResolver(WebhookSchema),
    defaultValues: { url: "" },
  })

  function handleCreate(data: { url: string }) {
    if (selectedEvents.length === 0) return
    const events = selectedEvents
    setShowCreate(false); setSelectedEvents([])
    urlForm.reset()
    startTransition(async () => {
      const tempHook = { id: crypto.randomUUID(), userId: 0, url: data.url, events, secret: "...", active: true, createdAt: Date.now() } as DbWebhook
      applyOptimistic({ type: "create", hook: tempHook })
      await actionCreateWebhook({ url: data.url, events })
    })
  }

  function handleDelete(id: string) {
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      await actionDeleteWebhook(id)
    })
  }

  async function copySecret(secret: string, id: string) {
    await navigator.clipboard.writeText(secret)
    setCopiedId(id); setTimeout(() => setCopiedId(null), 2000)
  }

  return (
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Webhooks</h1>
          <p className="text-sm text-fg-muted mt-0.5">Notificaciones HTTP para eventos del sistema</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <Plus className="h-3.5 w-3.5" /> Nuevo webhook
        </Button>
      </div>

      {showCreate && (
        <form onSubmit={urlForm.handleSubmit(handleCreate)} className="rounded-xl border border-border bg-surface p-5 space-y-4">
          <h3 className="text-sm font-semibold text-fg">Nuevo webhook</h3>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">URL destino</label>
            <Input {...urlForm.register("url")} placeholder="https://mi-sistema.com/webhook" />
            {urlForm.formState.errors.url && <p className="text-xs text-destructive">{urlForm.formState.errors.url.message}</p>}
          </div>
          <div className="space-y-2">
            <label className="text-xs font-medium text-fg-muted">Eventos a escuchar</label>
            <div className="flex flex-wrap gap-2">
              {EVENT_OPTIONS.map((ev) => (
                <button
                  key={ev} type="button"
                  onClick={() => setSelectedEvents((p) => p.includes(ev) ? p.filter((e) => e !== ev) : [...p, ev])}
                  className={`px-2.5 py-1 rounded-full text-xs border transition-colors ${
                    selectedEvents.includes(ev)
                      ? "bg-accent text-accent-fg border-accent"
                      : "border-border text-fg-muted hover:border-accent"
                  }`}
                >
                  {ev}
                </button>
              ))}
            </div>
          </div>
          <div className="flex gap-2">
            <Button size="sm" type="submit" disabled={isPending || selectedEvents.length === 0}>{isPending ? "Creando..." : "Crear webhook"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {optimisticHooks.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Webhook} />
          <EmptyPlaceholder.Title>Sin webhooks configurados</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Creá un webhook para recibir notificaciones de eventos.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="space-y-2">
          {optimisticHooks.map((h) => (
            <div key={h.id} className="p-4 rounded-xl border border-border bg-surface space-y-2">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-sm font-medium text-fg truncate">{h.url}</p>
                  <div className="flex flex-wrap gap-1 mt-1.5">
                    {(h.events as string[]).map((ev) => <Badge key={ev} variant="outline" className="text-xs">{ev}</Badge>)}
                  </div>
                </div>
                <div className="flex gap-1 shrink-0">
                  <Button variant="ghost" size="icon" className="h-7 w-7" title="Copiar secret" onClick={() => copySecret(h.secret, h.id)}>
                    {copiedId === h.id ? <Check size={12} /> : <Copy size={12} />}
                  </Button>
                  <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive" onClick={() => handleDelete(h.id)}>
                    <Trash2 size={13} />
                  </Button>
                </div>
              </div>
              <p className="text-xs text-fg-subtle">Secret: <code className="font-mono">{h.secret.slice(0, 8)}...</code> · Firma: X-Signature: sha256=HMAC</p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
