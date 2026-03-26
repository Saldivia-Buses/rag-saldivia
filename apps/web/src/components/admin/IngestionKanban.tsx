"use client"

import { useEffect, useRef, useState } from "react"
import { RefreshCw, XCircle, CheckCircle, Clock, AlertCircle, ChevronDown, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"

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

const COLUMNS = [
  { states: ["pending", "stalled"], label: "Pendiente",     icon: <Clock size={14} className="text-fg-muted" /> },
  { states: ["running"],            label: "En progreso",   icon: <RefreshCw size={14} className="animate-spin text-accent" /> },
  { states: ["done"],               label: "Completado",    icon: <CheckCircle size={14} className="text-success" /> },
  { states: ["error", "cancelled"], label: "Error",         icon: <XCircle size={14} className="text-destructive" /> },
]

function JobCard({ job, onRetry }: { job: Job; onRetry: (id: string) => void }) {
  const [expanded, setExpanded] = useState(false)
  const elapsed = Math.round((Date.now() - job.createdAt) / 1000)

  return (
    <div className="rounded-lg border border-border bg-bg p-3 text-sm space-y-2 hover:shadow-sm transition-shadow">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="font-medium text-fg truncate">{job.filename}</p>
          <p className="text-xs mt-0.5 text-fg-muted">
            {job.collection} · {job.tier} · {elapsed < 60 ? `${elapsed}s` : `${Math.round(elapsed / 60)}m`}
          </p>
        </div>
        {(job.state === "error" || job.state === "stalled") && (
          <Button variant="ghost" size="sm" className="h-6 text-xs shrink-0" onClick={() => onRetry(job.id)}>
            <RefreshCw size={11} className="mr-1" /> Retry
          </Button>
        )}
      </div>

      {job.state === "running" && (
        <div>
          <div className="flex justify-between text-xs mb-1 text-fg-muted">
            <span>Progreso</span><span>{job.progress}%</span>
          </div>
          <div className="h-1.5 rounded-full bg-surface-2 overflow-hidden">
            <div className="h-full rounded-full bg-accent transition-all" style={{ width: `${job.progress}%` }} />
          </div>
        </div>
      )}

      {job.error && (
        <div>
          <button onClick={() => setExpanded((e) => !e)} className="flex items-center gap-1 text-xs text-destructive">
            <AlertCircle size={11} />
            Ver error
            {expanded ? <ChevronDown size={10} /> : <ChevronRight size={10} />}
          </button>
          {expanded && (
            <p className="mt-1 text-xs rounded-md p-2 bg-destructive-subtle text-destructive">{job.error}</p>
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
      try { setJobs((JSON.parse(e.data) as { jobs: Job[] }).jobs) } catch { /* ignorar */ }
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
    <div className="p-6 space-y-5">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Ingesta</h1>
          <p className="text-sm text-fg-muted mt-0.5">Monitor de jobs en tiempo real</p>
        </div>
        <div className="flex items-center gap-2">
          <div className={`w-2 h-2 rounded-full ${connected ? "bg-success" : "bg-fg-subtle"}`} />
          <span className="text-xs text-fg-muted">{connected ? "En tiempo real" : "Reconectando..."}</span>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {COLUMNS.map((col) => {
          const colJobs = jobs.filter((j) => col.states.includes(j.state))
          return (
            <div key={col.label}>
              <div className="flex items-center gap-2 mb-3 pb-2 border-b border-border">
                {col.icon}
                <span className="text-sm font-semibold text-fg">{col.label}</span>
                <Badge variant="outline" className="ml-auto text-xs">{colJobs.length}</Badge>
              </div>
              <div className="space-y-2 min-h-[80px]">
                {colJobs.map((job) => <JobCard key={job.id} job={job} onRetry={handleRetry} />)}
                {colJobs.length === 0 && (
                  <p className="text-xs text-center py-6 text-fg-subtle">Sin jobs</p>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
