"use client"

import { useState } from "react"
import { Plus, Trash2, Play, Download } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { Table2 } from "lucide-react"
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table"
import { downloadFile } from "@/lib/export"

type Field = { name: string; description: string }
type Row = Record<string, string>

const STEPS = ["Colección", "Campos", "Resultados"] as const

function StepDot({ n, active, done }: { n: number; active: boolean; done: boolean }) {
  return (
    <div className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-semibold shrink-0 ${
      done || active ? "bg-accent text-accent-fg" : "bg-surface-2 text-fg-subtle"
    }`}>{n}</div>
  )
}

export function ExtractionWizard() {
  const [step, setStep] = useState<1 | 2 | 3>(1)
  const [collection, setCollection] = useState("")
  const [fields, setFields] = useState<Field[]>([{ name: "", description: "" }])
  const [results, setResults] = useState<Row[]>([])
  const [columns, setColumns] = useState<string[]>([])
  const [running, setRunning] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function addField() { setFields((p) => [...p, { name: "", description: "" }]) }
  function removeField(i: number) { setFields((p) => p.filter((_, idx) => idx !== i)) }
  function updateField(i: number, key: keyof Field, value: string) {
    setFields((p) => p.map((f, idx) => idx === i ? { ...f, [key]: value } : f))
  }

  async function handleExtract() {
    const validFields = fields.filter((f) => f.name.trim())
    if (!collection || validFields.length === 0) return
    setRunning(true); setError(null)
    try {
      const res = await fetch("/api/extract", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ collection, fields: validFields }),
      })
      const d = await res.json() as { ok: boolean; data?: Row[]; fields?: string[]; error?: string }
      if (!d.ok) { setError(d.error ?? "Error"); return }
      setResults(d.data ?? []); setColumns(d.fields ?? []); setStep(3)
    } catch (err) { setError(String(err)) }
    finally { setRunning(false) }
  }

  function exportCSV() {
    if (results.length === 0) return
    const rows = results.map((r) => columns.map((c) => `"${String(r[c] ?? "").replace(/"/g, '""')}"`).join(","))
    downloadFile([columns.join(","), ...rows].join("\n"), "extraccion.csv", "text/csv")
  }

  return (
    <div className="p-6 space-y-6 max-w-3xl">
      <div>
        <h1 className="text-lg font-semibold text-fg">Extracción estructurada</h1>
        <p className="text-sm text-fg-muted mt-0.5">Extraé datos tabulares de documentos usando el RAG</p>
      </div>

      {/* Steps */}
      <div className="flex items-center gap-3 text-sm">
        {STEPS.map((label, idx) => {
          const n = idx + 1
          return (
            <div key={n} className="flex items-center gap-2">
              <StepDot n={n} active={step === n} done={step > n} />
              <span className={step === n ? "text-fg font-medium" : "text-fg-muted"}>{label}</span>
              {idx < 2 && <span className="text-border">→</span>}
            </div>
          )
        })}
      </div>

      {/* Step 1 */}
      {step === 1 && (
        <div className="space-y-4">
          <h2 className="text-sm font-semibold text-fg">Seleccioná la colección</h2>
          <Input
            value={collection} onChange={(e) => setCollection(e.target.value)}
            placeholder="nombre-de-coleccion" className="max-w-sm"
          />
          <Button size="sm" onClick={() => setStep(2)} disabled={!collection.trim()}>
            Siguiente →
          </Button>
        </div>
      )}

      {/* Step 2 */}
      {step === 2 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-fg">Definí los campos a extraer</h2>
            <Badge variant="secondary">{collection}</Badge>
          </div>
          <div className="space-y-2">
            {fields.map((f, i) => (
              <div key={i} className="flex gap-2 items-center">
                <Input value={f.name} onChange={(e) => updateField(i, "name", e.target.value)}
                  placeholder="Nombre del campo (ej: Fecha)" className="flex-1" />
                <Input value={f.description} onChange={(e) => updateField(i, "description", e.target.value)}
                  placeholder="Descripción (ej: Fecha de firma)" className="flex-[2]" />
                <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0 text-destructive hover:text-destructive"
                  onClick={() => removeField(i)}>
                  <Trash2 size={13} />
                </Button>
              </div>
            ))}
          </div>
          <div className="flex gap-2 flex-wrap">
            <Button size="sm" variant="outline" onClick={addField}><Plus size={13} /> Campo</Button>
            <Button size="sm" variant="ghost" onClick={() => setStep(1)}>← Atrás</Button>
            <Button size="sm" onClick={handleExtract}
              disabled={running || fields.every((f) => !f.name.trim())}>
              <Play size={13} className="mr-1" />
              {running ? "Extrayendo..." : "Extraer"}
            </Button>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
      )}

      {/* Step 3 */}
      {step === 3 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-fg">Resultados — {results.length} filas</h2>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => setStep(2)}>← Volver</Button>
              <Button size="sm" onClick={exportCSV} disabled={results.length === 0}>
                <Download size={13} className="mr-1" /> Exportar CSV
              </Button>
            </div>
          </div>

          {results.length === 0 ? (
            <EmptyPlaceholder>
              <EmptyPlaceholder.Icon icon={Table2} />
              <EmptyPlaceholder.Title>Sin resultados</EmptyPlaceholder.Title>
              <EmptyPlaceholder.Description>El RAG no encontró datos para los campos especificados.</EmptyPlaceholder.Description>
            </EmptyPlaceholder>
          ) : (
            <div className="rounded-xl border border-border overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>{columns.map((c) => <TableHead key={c}>{c}</TableHead>)}</TableRow>
                </TableHeader>
                <TableBody>
                  {results.map((row, i) => (
                    <TableRow key={i}>
                      {columns.map((c) => (
                        <TableCell key={c} className="max-w-[200px] truncate">{row[c] ?? "—"}</TableCell>
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
