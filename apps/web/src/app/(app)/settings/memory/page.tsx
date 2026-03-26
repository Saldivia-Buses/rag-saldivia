"use client"

import { useEffect, useState } from "react"
import { Plus, Trash2, Brain } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import type { DbUserMemory } from "@rag-saldivia/db"

export default function MemoryPage() {
  const [entries, setEntries] = useState<DbUserMemory[]>([])
  const [newKey, setNewKey] = useState("")
  const [newValue, setNewValue] = useState("")
  const [adding, setAdding] = useState(false)

  useEffect(() => {
    fetch("/api/memory")
      .then((r) => r.json())
      .then((d: { ok: boolean; data?: DbUserMemory[] }) => { if (d.ok) setEntries(d.data ?? []) })
  }, [])

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!newKey.trim() || !newValue.trim()) return
    setAdding(true)
    try {
      await fetch("/api/memory", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ key: newKey, value: newValue }) })
      setEntries((p) => [...p.filter((e) => e.key !== newKey), { id: Date.now(), userId: 0, key: newKey, value: newValue, source: "explicit", createdAt: Date.now(), updatedAt: Date.now() }])
      setNewKey(""); setNewValue("")
    } finally { setAdding(false) }
  }

  async function handleDelete(key: string) {
    await fetch(`/api/memory?key=${encodeURIComponent(key)}`, { method: "DELETE" })
    setEntries((p) => p.filter((e) => e.key !== key))
  }

  const inputClass = "px-3 py-1.5 rounded-md border text-sm outline-none"
  const inputStyle = { borderColor: "var(--border)", background: "var(--background)", color: "var(--foreground)" }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div>
        <h2 className="text-lg font-semibold flex items-center gap-2">
          <Brain size={18} style={{ color: "var(--accent)" }} />
          Memoria del sistema
        </h2>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          El sistema inyecta estas preferencias en cada consulta para personalizar las respuestas.
        </p>
      </div>
      <Separator />

      <form onSubmit={handleAdd} className="flex gap-2 flex-wrap">
        <input value={newKey} onChange={(e) => setNewKey(e.target.value)} placeholder="Clave (ej: idioma)" className={`${inputClass} flex-1 min-w-[120px]`} style={inputStyle} />
        <input value={newValue} onChange={(e) => setNewValue(e.target.value)} placeholder="Valor (ej: siempre en español)" className={`${inputClass} flex-1 min-w-[200px]`} style={inputStyle} />
        <Button size="sm" type="submit" disabled={adding} className="gap-1.5"><Plus size={13} /> Agregar</Button>
      </form>

      {entries.length === 0 ? (
        <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>Sin preferencias guardadas.</p>
      ) : (
        <div className="space-y-2">
          {entries.map((e) => (
            <div key={e.key} className="flex items-center gap-3 p-3 rounded-lg border" style={{ borderColor: "var(--border)" }}>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">{e.key}</span>
                  <Badge variant="outline" className="text-xs">{e.source}</Badge>
                </div>
                <p className="text-sm mt-0.5" style={{ color: "var(--muted-foreground)" }}>{e.value}</p>
              </div>
              <Button variant="ghost" size="icon" className="h-7 w-7 shrink-0" onClick={() => handleDelete(e.key)} style={{ color: "var(--destructive)" }}>
                <Trash2 size={13} />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
