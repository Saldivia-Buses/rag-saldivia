"use client"

import { useOptimistic, useState, useTransition } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { Plus, Trash2, FileText } from "lucide-react"

const ReportSchema = z.object({
  query: z.string().min(5, "La query debe tener al menos 5 caracteres"),
  collection: z.string().min(1, "La colección es requerida"),
  schedule: z.enum(["daily", "weekly", "monthly"]),
  destination: z.enum(["saved", "email"]),
  email: z.string().optional(),
})
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import type { DbScheduledReport } from "@rag-saldivia/db"
import { formatDate } from "@/lib/utils"
import { actionCreateReport, actionDeleteReport } from "@/app/actions/reports"

const SCHEDULE_LABELS: Record<string, string> = { daily: "Diario", weekly: "Semanal", monthly: "Mensual" }

const SELECT_CLASS = "h-9 w-full rounded-md border border-border bg-bg px-3 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"

export function ReportsAdmin({ initialReports }: { initialReports: DbScheduledReport[] }) {
  const [optimisticReports, applyOptimistic] = useOptimistic(
    initialReports,
    (state, action: { type: "delete"; id: string }) => {
      if (action.type === "delete") return state.filter((r) => r.id !== action.id)
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [showCreate, setShowCreate] = useState(false)

  const reportForm = useForm<z.infer<typeof ReportSchema>>({
    resolver: zodResolver(ReportSchema),
    defaultValues: { query: "", collection: "", schedule: "daily", destination: "saved", email: "" },
  })

  function handleCreate(data: z.infer<typeof ReportSchema>) {
    startTransition(async () => {
      const report = await actionCreateReport({
        query: data.query,
        collection: data.collection,
        schedule: data.schedule,
        destination: data.destination,
        email: data.email || undefined,
      })
      if (report) { setShowCreate(false); reportForm.reset() }
    })
  }

  function handleDelete(id: string) {
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      await actionDeleteReport(id)
    })
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
        <form onSubmit={reportForm.handleSubmit(handleCreate)} className="rounded-xl border border-border bg-surface p-5 space-y-3">
          <h3 className="text-sm font-semibold text-fg">Nuevo informe</h3>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Query</label>
            <Textarea {...reportForm.register("query")} placeholder="¿Cuáles son los contratos vigentes?" className="min-h-[60px]" />
            {reportForm.formState.errors.query && <p className="text-xs text-destructive">{reportForm.formState.errors.query.message}</p>}
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-fg-muted">Colección</label>
            <Input {...reportForm.register("collection")} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-fg-muted">Frecuencia</label>
              <select {...reportForm.register("schedule")} className={SELECT_CLASS}>
                <option value="daily">Diario</option><option value="weekly">Semanal</option><option value="monthly">Mensual</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-fg-muted">Destino</label>
              <select {...reportForm.register("destination")} className={SELECT_CLASS}>
                <option value="saved">Guardados</option><option value="email">Email</option>
              </select>
            </div>
          </div>
          {reportForm.watch("destination") === "email" && (
            <div className="space-y-1.5">
              <label className="text-xs font-medium text-fg-muted">Email destino</label>
              <Input type="email" {...reportForm.register("email")} />
            </div>
          )}
          <div className="flex gap-2 pt-1">
            <Button size="sm" type="submit" disabled={isPending}>{isPending ? "Creando..." : "Crear informe"}</Button>
            <Button size="sm" variant="outline" type="button" onClick={() => setShowCreate(false)}>Cancelar</Button>
          </div>
        </form>
      )}

      {optimisticReports.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={FileText} />
          <EmptyPlaceholder.Title>Sin informes configurados</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Creá un informe para recibir respuestas automáticas periódicas.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="space-y-2">
          {optimisticReports.map((r) => (
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
