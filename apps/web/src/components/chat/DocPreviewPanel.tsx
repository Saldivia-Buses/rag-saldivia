"use client"

import { useState } from "react"
import { FileText, AlertCircle, ChevronLeft, ChevronRight } from "lucide-react"
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Button } from "@/components/ui/button"

type Props = {
  documentName: string | null
  highlightText?: string
  onClose: () => void
}

/**
 * Panel lateral para preview de documentos (F3.40).
 *
 * react-pdf se importa dinámicamente para evitar errores en SSR
 * (PDF.js requiere window/canvas — solo disponible en el browser).
 *
 * Si el RAG server no expone el endpoint de documentos, muestra
 * el fragmento de texto del source sin renderizado PDF.
 */
export function DocPreviewPanel({ documentName, highlightText, onClose }: Props) {
  const [pdfAvailable, setPdfAvailable] = useState<boolean | null>(null)
  const [numPages, setNumPages] = useState<number>(0)
  const [currentPage, setCurrentPage] = useState(1)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [PdfComponents, setPdfComponents] = useState<{ Document: any; Page: any } | null>(null)

  const open = !!documentName

  // Carga dinámica de react-pdf solo en el browser
  if (open && pdfAvailable === null && typeof window !== "undefined") {
    import("react-pdf").then((mod) => {
      try {
        mod.pdfjs.GlobalWorkerOptions.workerSrc = `//unpkg.com/pdfjs-dist@${mod.pdfjs.version}/build/pdf.worker.min.mjs`
      } catch { /* ignorar */ }
      setPdfComponents({ Document: mod.Document, Page: mod.Page })
    }).catch(() => {
      setPdfAvailable(false)
    })
  }

  const pdfUrl = documentName
    ? `/api/rag/document/${encodeURIComponent(documentName)}`
    : null

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent side="right" className="w-[480px] max-w-full flex flex-col p-0">
        <SheetHeader className="px-4 py-3 border-b shrink-0" style={{ borderColor: "var(--border)" }}>
          <div className="flex items-center gap-2">
            <FileText size={16} style={{ color: "var(--accent)" }} />
            <SheetTitle className="text-sm font-medium truncate flex-1">
              {documentName ?? "Documento"}
            </SheetTitle>
          </div>
          {highlightText && (
            <p className="text-xs mt-1 line-clamp-2" style={{ color: "var(--muted-foreground)" }}>
              Fragmento: &quot;{highlightText}&quot;
            </p>
          )}
        </SheetHeader>

        <div className="flex-1 overflow-y-auto">
          {!pdfUrl ? null : pdfAvailable === false ? (
            // Fallback: mostrar el fragmento de texto
            <div className="p-4 space-y-3">
              <div
                className="flex items-start gap-2 p-3 rounded-lg text-sm"
                style={{ background: "var(--muted)", color: "var(--muted-foreground)" }}
              >
                <AlertCircle size={14} className="shrink-0 mt-0.5" />
                <p>El servidor RAG no expone preview de documentos. Mostrando el fragmento relevante.</p>
              </div>
              {highlightText && (
                <div className="p-4 rounded-lg border text-sm" style={{ borderColor: "var(--border)" }}>
                  <p className="leading-relaxed">{highlightText}</p>
                </div>
              )}
            </div>
          ) : PdfComponents ? (
            // Renderizado PDF con react-pdf
            <div className="flex flex-col items-center py-4 px-2 gap-4">
              <PdfComponents.Document
                file={pdfUrl}
                onLoadSuccess={({ numPages: n }: { numPages: number }) => { setNumPages(n); setPdfAvailable(true) }}
                onLoadError={() => setPdfAvailable(false)}
              >
                <PdfComponents.Page
                  pageNumber={currentPage}
                  width={430}
                />
              </PdfComponents.Document>

              {numPages > 1 && (
                <div className="flex items-center gap-3 shrink-0">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    disabled={currentPage <= 1}
                    onClick={() => setCurrentPage((p) => p - 1)}
                  >
                    <ChevronLeft size={14} />
                  </Button>
                  <span className="text-xs" style={{ color: "var(--muted-foreground)" }}>
                    {currentPage} / {numPages}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    disabled={currentPage >= numPages}
                    onClick={() => setCurrentPage((p) => p + 1)}
                  >
                    <ChevronRight size={14} />
                  </Button>
                </div>
              )}
            </div>
          ) : (
            <div className="p-8 text-center text-sm" style={{ color: "var(--muted-foreground)" }}>
              Cargando documento...
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
