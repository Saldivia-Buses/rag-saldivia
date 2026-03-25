"use client"

import { useRef, useState } from "react"
import { Upload, X } from "lucide-react"

/**
 * Drop zone sobre el área del chat.
 *
 * Nota de viabilidad (F2.35): el Blueprint v2.5.0 no soporta colecciones efímeras
 * con TTL en Milvus. Por lo tanto, al soltar un PDF se usa el flujo de upload normal
 * redirigiendo a /upload con la sesión pre-seleccionada como parámetro.
 */

type Props = {
  sessionId: string
  children: React.ReactNode
}

export function ChatDropZone({ sessionId, children }: Props) {
  const [isDragging, setIsDragging] = useState(false)
  const [uploadedFile, setUploadedFile] = useState<string | null>(null)
  const dragCountRef = useRef(0)

  function onDragEnter(e: React.DragEvent) {
    e.preventDefault()
    dragCountRef.current++
    if (dragCountRef.current === 1) setIsDragging(true)
  }

  function onDragLeave(e: React.DragEvent) {
    e.preventDefault()
    dragCountRef.current--
    if (dragCountRef.current === 0) setIsDragging(false)
  }

  function onDragOver(e: React.DragEvent) {
    e.preventDefault()
  }

  async function onDrop(e: React.DragEvent) {
    e.preventDefault()
    setIsDragging(false)
    dragCountRef.current = 0

    const file = e.dataTransfer.files[0]
    if (!file) return

    const allowed = ["application/pdf", "text/plain", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"]
    if (!allowed.includes(file.type) && !file.name.endsWith(".pdf") && !file.name.endsWith(".txt")) {
      alert("Solo se aceptan archivos PDF, TXT o DOCX")
      return
    }

    // Subir el archivo usando el endpoint de upload normal
    const formData = new FormData()
    formData.append("file", file)
    formData.append("collection", "tecpia") // colección por defecto — en el futuro puede ser la de la sesión

    try {
      const res = await fetch("/api/upload", { method: "POST", body: formData })
      if (res.ok) {
        setUploadedFile(file.name)
        setTimeout(() => setUploadedFile(null), 4000)
      }
    } catch {
      alert("Error al subir el archivo")
    }
  }

  return (
    <div
      className="flex-1 flex flex-col min-h-0 relative"
      onDragEnter={onDragEnter}
      onDragLeave={onDragLeave}
      onDragOver={onDragOver}
      onDrop={onDrop}
    >
      {children}

      {/* Overlay mientras arrastra */}
      {isDragging && (
        <div
          className="absolute inset-0 z-50 flex items-center justify-center rounded-lg"
          style={{ background: "rgba(124, 106, 245, 0.12)", border: "2px dashed var(--accent)" }}
        >
          <div className="text-center">
            <Upload size={32} style={{ color: "var(--accent)", margin: "0 auto 8px" }} />
            <p className="text-sm font-medium" style={{ color: "var(--accent)" }}>
              Soltá el archivo para subirlo
            </p>
            <p className="text-xs mt-1" style={{ color: "var(--muted-foreground)" }}>
              PDF, TXT, DOCX
            </p>
          </div>
        </div>
      )}

      {/* Confirmación de upload */}
      {uploadedFile && (
        <div
          className="absolute bottom-24 left-1/2 -translate-x-1/2 px-4 py-2 rounded-full text-sm flex items-center gap-2 shadow-lg z-50"
          style={{ background: "var(--accent)", color: "white" }}
        >
          <Upload size={14} />
          {uploadedFile} subido correctamente
          <button onClick={() => setUploadedFile(null)}>
            <X size={13} />
          </button>
        </div>
      )}
    </div>
  )
}
