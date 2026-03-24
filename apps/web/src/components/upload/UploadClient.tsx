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
    const jobEntry: JobResult = {
      filename: file.name,
      jobId: "",
      collection: selectedCollection,
      state: "uploading",
    }
    setJobs((prev) => [jobEntry, ...prev])

    try {
      const form = new FormData()
      form.append("file", file)
      form.append("collection", selectedCollection)

      const res = await fetch("/api/upload", { method: "POST", body: form })
      const data = await res.json()

      if (!res.ok || !data.ok) {
        setJobs((prev) =>
          prev.map((j) =>
            j.filename === file.name && j.state === "uploading"
              ? { ...j, state: "error", error: data.error ?? "Error desconocido" }
              : j
          )
        )
        return
      }

      setJobs((prev) =>
        prev.map((j) =>
          j.filename === file.name && j.state === "uploading"
            ? { ...j, state: "success", jobId: data.data.jobId }
            : j
        )
      )
    } catch (err) {
      setJobs((prev) =>
        prev.map((j) =>
          j.filename === file.name && j.state === "uploading"
            ? { ...j, state: "error", error: String(err) }
            : j
        )
      )
    }
  }

  function handleFiles(files: FileList | null) {
    if (!files || !selectedCollection) return
    Array.from(files).forEach(uploadFile)
  }

  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setIsDragging(false)
      handleFiles(e.dataTransfer.files)
    },
    [selectedCollection]
  )

  return (
    <div className="space-y-6">
      {/* Colección */}
      <div className="space-y-1">
        <label className="text-sm font-medium">Colección destino</label>
        {collections.length === 0 ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
            No tenés colecciones con permiso de escritura. Contactá al administrador.
          </p>
        ) : (
          <select
            value={selectedCollection}
            onChange={(e) => setSelectedCollection(e.target.value)}
            className="w-full px-3 py-2 rounded-md border text-sm"
            style={{ borderColor: "var(--border)", background: "var(--background)" }}
          >
            {collections.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        )}
      </div>

      {/* Drop zone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setIsDragging(true) }}
        onDragLeave={() => setIsDragging(false)}
        onDrop={onDrop}
        onClick={() => inputRef.current?.click()}
        className="border-2 border-dashed rounded-xl p-10 text-center cursor-pointer transition-colors"
        style={{
          borderColor: isDragging ? "var(--primary)" : "var(--border)",
          background: isDragging ? "var(--accent)" : "transparent",
        }}
      >
        <Upload size={32} className="mx-auto mb-3 opacity-40" />
        <p className="font-medium text-sm">
          Arrastrá archivos acá o hacé click para seleccionar
        </p>
        <p className="text-xs mt-1" style={{ color: "var(--muted-foreground)" }}>
          PDF, DOCX, TXT — hasta 100MB por archivo
        </p>
        <input
          ref={inputRef}
          type="file"
          multiple
          accept=".pdf,.docx,.txt,.md"
          className="hidden"
          onChange={(e) => handleFiles(e.target.files)}
        />
      </div>

      {/* Jobs list */}
      {jobs.length > 0 && (
        <div className="space-y-2">
          <p className="text-sm font-medium">Archivos subidos</p>
          {jobs.map((job, i) => (
            <div
              key={i}
              className="flex items-center gap-3 px-4 py-3 rounded-lg border"
              style={{ borderColor: "var(--border)" }}
            >
              <FileText size={16} className="flex-shrink-0 opacity-60" />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{job.filename}</p>
                <p className="text-xs truncate" style={{ color: "var(--muted-foreground)" }}>
                  {job.collection}
                  {job.jobId && ` — job: ${job.jobId.slice(0, 8)}`}
                </p>
              </div>
              <div className="flex-shrink-0">
                {job.state === "uploading" && <Loader2 size={16} className="animate-spin" />}
                {job.state === "success" && <CheckCircle2 size={16} style={{ color: "#16a34a" }} />}
                {job.state === "error" && (
                  <div className="flex items-center gap-1">
                    <XCircle size={16} style={{ color: "var(--destructive)" }} />
                    <span className="text-xs" style={{ color: "var(--destructive)" }}>{job.error}</span>
                  </div>
                )}
              </div>
            </div>
          ))}
          <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
            Los documentos están en cola de procesamiento. Usá{" "}
            <code className="font-mono">rag ingest status</code> para ver el progreso.
          </p>
        </div>
      )}
    </div>
  )
}
