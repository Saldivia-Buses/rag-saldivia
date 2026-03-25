"use client"

import { useEffect, useRef, useState } from "react"
import { RefreshCw, XCircle, CheckCircle, Clock, AlertCircle, ChevronDown, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"

type Job = {
  id: string
  filename: string
  collection: string
  state: "pending" | "running" | "stalled" | "done" | "error" | "cancelled"
  progress: number
  tier: string
  retryCount: number
  createdAt: number
  error?: string | null
}

const STATE_LABELS: Record<string, string> = {
  pending: "Pendiente",
  running: "En progreso",
  done: "Completado",
  error: "Error",
  stalled: "Bloqueado",
  cancelled: "Cancelado",
}

const COLUMNS = [
  { states: ["pending", "stalled"], label: "Pendiente", icon: <Clock size={14} /> },
  { states: ["running"], label: "En progreso", icon: <RefreshCw size={14} className="animate-spin" /> },
  { states: ["done"], label: "Completado", icon: <CheckCircle size={14} /> },
  { states: ["error", "cancelled"], label: "Error / Cancelado", icon: <XCircle size={14} /> },
]

function JobCard({ job, onRetry }: { job: Job; onRetry: (id: string) => void }) {
  const [showError, setShowError] = useState(false)
  const [expanded, setExpanded] = useState(false)
  const elapsed = Math.round((Date.now() - job.createdAt) / 1000)

  return (
    <div
      className="rounded-lg border p-3 text-sm space-y-2"
      style={{ borderColor: "var(--border)", background: "var(--background)" }}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="font-medium truncate">{job.filename}</p>
          <p className="text-xs mt-0.5" style={{ color: "var(--muted-foreground)" }}>
            {job.collection} · {job.tier} · {elapsed < 60 ? `${elapsed}s` : `${Math.round(elapsed / 60)}m`}
          </p>
        </div>
        {(job.state === "error" || job.state === "stalled") && (
          <Button
            variant="ghost"
            size="sm"
            className="h-6 text-xs shrink-0"
            onClick={() => onRetry(job.id)}
            title="Reintentar"
          >
            <RefreshCw size={11} className="mr-1" /> Retry
          </Button>
        )}
      </div>

      {job.state === "running" && (
        <div>
          <div className="flex justify-between text-xs mb-1" style={{ color: "var(--muted-foreground)" }}>
            <span>Progreso</span>
            <span>{job.progress}%</span>
          </div>
          <div className="h-1.5 rounded-full overflow-hidden" style={{ background: "var(--muted)" }}>
            <div
              className="h-full rounded-full transition-all"
              style={{ width: `${job.progress}%`, background: "var(--accent)" }}
            />
          </div>
        </div>
      )}

      {job.error && (
        <div>
          <button
            onClick={() => setExpanded((e) => !e)}
            className="flex items-center gap-1 text-xs"
            style={{ color: "var(--destructive)" }}
          >
            <AlertCircle size={11} />
            Ver error
            {expanded ? <ChevronDown size={10} /> : <ChevronRight size={10} />}
          </button>
          {expanded && (
            <p className="mt-1 text-xs rounded p-2" style={{ background: "var(--muted)", color: "var(--destructive)" }}>
              {job.error}
            </p>
          )}
        </div>
      )}
    </div>
  )
}

export function IngestionKanban() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [connected, setConnected] = useState(false)
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    const es = new EventSource("/api/admin/ingestion/stream")
    esRef.current = es

    es.onopen = () => setConnected(true)
    es.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data) as { jobs: Job[] }
        setJobs(data.jobs)
      } catch { /* ignorar */ }
    }
    es.onerror = () => setConnected(false)

    return () => { es.close() }
  }, [])

  async function handleRetry(id: string) {
    await fetch(`/api/admin/ingestion/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action: "retry" }),
    })
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <div
          className="w-2 h-2 rounded-full"
          style={{ background: connected ? "#22c55e" : "var(--muted-foreground)" }}
          title={connected ? "Actualización en tiempo real" : "Desconectado"}
        />
        <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
          {connected ? "En tiempo real" : "Reconectando..."}
        </span>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {COLUMNS.map((col) => {
          const colJobs = jobs.filter((j) => col.states.includes(j.state))
          return (
            <div key={col.label}>
              <div className="flex items-center gap-2 mb-3">
                {col.icon}
                <span className="text-sm font-medium">{col.label}</span>
                <Badge variant="outline" className="ml-auto text-xs">{colJobs.length}</Badge>
              </div>
              <div className="space-y-2 min-h-[80px]">
                {colJobs.map((job) => (
                  <JobCard key={job.id} job={job} onRetry={handleRetry} />
                ))}
                {colJobs.length === 0 && (
                  <p className="text-xs text-center py-4" style={{ color: "var(--muted-foreground)" }}>
                    Sin jobs
                  </p>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
