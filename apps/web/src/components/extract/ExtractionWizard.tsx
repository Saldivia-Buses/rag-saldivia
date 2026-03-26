"use client"

import { useState } from "react"
import { Plus, Trash2, Play, Download, Table2 } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { downloadFile } from "@/lib/export"

type Field = { name: string; description: string }
type Row = Record<string, string>

export function ExtractionWizard() {
  const [step, setStep] = useState<1 | 2 | 3>(1)
  const [collection, setCollection] = useState("")
  const [fields, setFields] = useState<Field[]>([{ name: "", description: "" }])
  const [results, setResults] = useState<Row[]>([])
  const [columns, setColumns] = useState<string[]>([])
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function addField() {
    setFields((p) => [...p, { name: "", description: "" }])
  }

  function removeField(i: number) {
    setFields((p) => p.filter((_, idx) => idx !== i))
  }

  function updateField(i: number, key: keyof Field, value: string) {
    setFields((p) => p.map((f, idx) => idx === i ? { ...f, [key]: value } : f))
  }

  async function handleExtract() {
    const validFields = fields.filter((f) => f.name.trim())
    if (!collection || validFields.length === 0) return
    setRunning(true)
    setError(null)
    try {
      const res = await fetch("/api/extract", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ collection, fields: validFields }),
      })
      const d = await res.json() as { ok: boolean; data?: Row[]; fields?: string[]; error?: string }
      if (!d.ok) { setError(d.error ?? "Error"); return }
      setResults(d.data ?? [])
      setColumns(d.fields ?? [])
      setStep(3)
    } catch (err) {
      setError(String(err))
    } finally {
      setRunning(false)
    }
  }

  function exportCSV() {
    if (results.length === 0) return
    const header = columns.join(",")
    const rows = results.map((r) => columns.map((c) => `"${String(r[c] ?? "").replace(/"/g, '""')}"`).join(","))
    downloadFile([header, ...rows].join("\n"), "extraccion.csv", "text/csv")
  }

  const inputClass = "w-full px-3 py-1.5 rounded-md border text-sm outline-none"
  const inputStyle = { borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }

  return (
    <div className="space-y-6">
      {/* Steps indicator */}
      <div className="flex items-center gap-2 text-sm">
        {[1, 2, 3].map((s) => (
          <div key={s} className="flex items-center gap-2">
            <div
              className="w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium"
              style={{ background: step >= s ? "var(--accent)" : "var(--muted)", color: step >= s ? "white" : "var(--muted-foreground)" }}
            >
              {s}
            </div>
            <span style={{ color: step === s ? "var(--foreground)" : "var(--muted-foreground)" }}>
              {s === 1 ? "Colección" : s === 2 ? "Campos" : "Resultados"}
            </span>
            {s < 3 && <span style={{ color: "var(--border)" }}>→</span>}
          </div>
        ))}
      </div>

      {/* Step 1: Seleccionar colección */}
      {step === 1 && (
        <div className="space-y-3">
          <h2 className="font-medium">Paso 1: Seleccioná la colección</h2>
          <input
            value={collection}
            onChange={(e) => setCollection(e.target.value)}
            placeholder="nombre-de-coleccion"
            className={`${inputClass} max-w-sm`}
            style={inputStyle}
          />
          <Button size="sm" onClick={() => setStep(2)} disabled={!collection.trim()}>
            Siguiente →
          </Button>
        </div>
      )}

      {/* Step 2: Definir campos */}
      {step === 2 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-medium">Paso 2: Definí los campos a extraer</h2>
            <Badge variant="outline">{collection}</Badge>
          </div>
          <div className="space-y-2">
            {fields.map((f, i) => (
              <div key={i} className="flex gap-2 items-center">
                <input
                  value={f.name}
                  onChange={(e) => updateField(i, "name", e.target.value)}
                  placeholder="Nombre del campo (ej: Fecha)"
                  className={`${inputClass} flex-1`}
                  style={inputStyle}
                />
                <input
                  value={f.description}
                  onChange={(e) => updateField(i, "description", e.target.value)}
                  placeholder="Descripción (ej: Fecha de firma del contrato)"
                  className={`${inputClass} flex-[2]`}
                  style={inputStyle}
                />
                <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0" onClick={() => removeField(i)} style={{ color: "var(--destructive)" }}>
                  <Trash2 size={13} />
                </Button>
              </div>
            ))}
          </div>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={addField} className="gap-1.5"><Plus size={13} /> Campo</Button>
            <Button size="sm" variant="ghost" onClick={() => setStep(1)}>← Atrás</Button>
            <Button
              size="sm"
              onClick={handleExtract}
              disabled={running || fields.every((f) => !f.name.trim())}
              className="gap-1.5"
            >
              <Play size={13} />
              {running ? "Extrayendo..." : "Extraer"}
            </Button>
          </div>
          {error && <p className="text-sm" style={{ color: "var(--destructive)" }}>{error}</p>}
        </div>
      )}

      {/* Step 3: Resultados */}
      {step === 3 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-medium">Resultados — {results.length} filas</h2>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => setStep(2)} className="gap-1.5">← Volver</Button>
              <Button size="sm" onClick={exportCSV} className="gap-1.5" disabled={results.length === 0}>
                <Download size={13} /> Exportar CSV
              </Button>
            </div>
          </div>

          {results.length === 0 ? (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Sin resultados. El RAG no encontró datos para los campos especificados.</p>
          ) : (
            <div className="rounded-lg border overflow-x-auto" style={{ borderColor: "var(--border)" }}>
              <Table>
                <TableHeader>
                  <TableRow style={{ background: "var(--muted)" }}>
                    {columns.map((c) => <TableHead key={c}>{c}</TableHead>)}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {results.map((row, i) => (
                    <TableRow key={i}>
                      {columns.map((c) => (
                        <TableCell key={c} className="text-sm max-w-[200px] truncate">
                          {row[c] ?? "—"}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
