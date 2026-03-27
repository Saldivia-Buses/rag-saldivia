"use client"

import { useState, useRef, useCallback } from "react"
import { Upload, FileText, CheckCircle2, XCircle, Loader2 } from "lucide-react"

type UploadState = "idle" | "uploading" | "success" | "error"
type JobResult = {
  filename: string
  jobId: string
  collection: string
  state: UploadState
  error?: string
}

export function UploadClient({ collections }: { collections: string[] }) {
  const [selectedCollection, setSelectedCollection] = useState(collections[0] ?? "")
  const [jobs, setJobs] = useState<JobResult[]>([])
  const [isDragging, setIsDragging] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  async function uploadFile(file: File) {
    const jobEntry: JobResult = { filename: file.name, jobId: "", collection: selectedCollection, state: "uploading" }
    setJobs((prev) => [jobEntry, ...prev])
    try {
      const form = new FormData()
      form.append("file", file)
      form.append("collection", selectedCollection)
      const res = await fetch("/api/upload", { method: "POST", body: form })
      const data = await res.json()
      setJobs((prev) => prev.map((j) =>
        j.filename === file.name && j.state === "uploading"
          ? (!res.ok || !data.ok)
            ? { ...j, state: "error", error: data.error ?? "Error desconocido" }
            : { ...j, state: "success", jobId: data.data.jobId }
          : j
      ))
    } catch (err) {
      setJobs((prev) => prev.map((j) =>
        j.filename === file.name && j.state === "uploading"
          ? { ...j, state: "error", error: String(err) } : j
      ))
    }
  }

  function handleFiles(files: FileList | null) {
    if (!files || !selectedCollection) return
    Array.from(files).forEach(uploadFile)
  }

  const onDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault(); setIsDragging(false); handleFiles(e.dataTransfer.files)
  // handleFiles y uploadFile dependen de selectedCollection vía closure
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedCollection])

  return (
    <div className="p-6 max-w-2xl space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-fg">Subir documentos</h1>
        <p className="text-sm text-fg-muted mt-0.5">Ingestá documentos en una colección</p>
      </div>

      {/* Colección */}
      <div className="space-y-1.5">
        <label className="text-sm font-medium text-fg">Colección destino</label>
        {collections.length === 0 ? (
          <p className="text-sm text-fg-muted">No tenés colecciones con permiso de escritura. Contactá al administrador.</p>
        ) : (
          <select
            value={selectedCollection}
            onChange={(e) => setSelectedCollection(e.target.value)}
            className="h-9 w-full rounded-md border border-border bg-bg px-3 text-sm text-fg focus:outline-none focus:ring-1 focus:ring-ring"
          >
            {collections.map((c) => <option key={c} value={c}>{c}</option>)}
          </select>
        )}
      </div>

      {/* Drop zone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setIsDragging(true) }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={onDrop}
        onClick={() => inputRef.current?.click()}
        className={`rounded-xl border-2 border-dashed p-12 text-center cursor-pointer transition-colors ${
          isDragging ? "border-accent bg-accent-subtle" : "border-border hover:border-accent hover:bg-surface"
        }`}
      >
        <Upload size={32} className={`mx-auto mb-3 ${isDragging ? "text-accent" : "text-fg-subtle"}`} />
        <p className="font-medium text-sm text-fg">
          Arrastrá archivos acá o hacé click para seleccionar
        </p>
        <p className="text-xs mt-1 text-fg-subtle">PDF, DOCX, TXT — hasta 100MB por archivo</p>
        <input ref={inputRef} type="file" multiple accept=".pdf,.docx,.txt,.md" className="hidden"
          onChange={(e) => handleFiles(e.target.files)} />
      </div>

      {/* Jobs list */}
      {jobs.length > 0 && (
        <div className="space-y-2">
          <p className="text-sm font-semibold text-fg">Archivos subidos</p>
          {jobs.map((job, i) => (
            <div key={i} className="flex items-center gap-3 px-4 py-3 rounded-lg border border-border bg-surface">
              <FileText size={16} className="shrink-0 text-fg-subtle" />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-fg truncate">{job.filename}</p>
                <p className="text-xs text-fg-muted truncate">
                  {job.collection}{job.jobId && ` — job: ${job.jobId.slice(0, 8)}`}
                </p>
              </div>
              <div className="shrink-0">
                {job.state === "uploading" && <Loader2 size={16} className="animate-spin text-fg-muted" />}
                {job.state === "success" && <CheckCircle2 size={16} className="text-success" />}
                {job.state === "error" && (
                  <div className="flex items-center gap-1">
                    <XCircle size={16} className="text-destructive" />
                    <span className="text-xs text-destructive">{job.error}</span>
                  </div>
                )}
              </div>
            </div>
          ))}
          <p className="text-xs text-fg-subtle">
            Los documentos están en cola de procesamiento. Usá{" "}
            <code className="font-mono">rag ingest status</code> para ver el progreso.
          </p>
        </div>
      )}
    </div>
  )
}
