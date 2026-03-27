"use client"

import { useState, useTransition } from "react"
import { Plus, Trash2, Brain } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import type { DbUserMemory } from "@rag-saldivia/db"
import { actionAddMemory, actionDeleteMemory } from "@/app/actions/settings"

export function MemoryClient({ entries: initial }: { entries: DbUserMemory[] }) {
  const [entries, setEntries] = useState<DbUserMemory[]>(initial)
  const [newKey, setNewKey] = useState("")
  const [newValue, setNewValue] = useState("")
  const [isPending, startTransition] = useTransition()

  function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!newKey.trim() || !newValue.trim()) return
    const key = newKey.trim()
    const value = newValue.trim()
    setNewKey("")
    setNewValue("")
    startTransition(async () => {
      await actionAddMemory(key, value)
      setEntries((p) => [
        ...p.filter((e) => e.key !== key),
        { id: Date.now(), userId: 0, key, value, source: "explicit", createdAt: Date.now(), updatedAt: Date.now() },
      ])
    })
  }

  function handleDelete(key: string) {
    startTransition(async () => {
      await actionDeleteMemory(key)
      setEntries((p) => p.filter((e) => e.key !== key))
    })
  }

  return (
    <div className="p-6 max-w-2xl space-y-6">
      <div>
        <h1 className="text-lg font-semibold text-fg flex items-center gap-2">
          <Brain size={18} className="text-accent" />
          Memoria del sistema
        </h1>
        <p className="text-sm text-fg-muted mt-0.5">
          El sistema inyecta estas preferencias en cada consulta para personalizar las respuestas.
        </p>
      </div>

      <Separator />

      <form onSubmit={handleAdd} className="flex gap-2 flex-wrap">
        <Input
          value={newKey}
          onChange={(e) => setNewKey(e.target.value)}
          placeholder="Clave (ej: idioma)"
          className="flex-1 min-w-[120px]"
        />
        <Input
          value={newValue}
          onChange={(e) => setNewValue(e.target.value)}
          placeholder="Valor (ej: siempre en español)"
          className="flex-[2] min-w-[200px]"
        />
        <Button size="sm" type="submit" disabled={isPending}>
          <Plus size={13} className="mr-1" /> Agregar
        </Button>
      </form>

      {entries.length === 0 ? (
        <p className="text-sm text-fg-muted">Sin preferencias guardadas.</p>
      ) : (
        <div className="space-y-2">
          {entries.map((e) => (
            <div key={e.key} className="flex items-center gap-3 p-3 rounded-xl border border-border bg-surface">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium text-fg">{e.key}</span>
                  <Badge variant="outline" className="text-xs">{e.source}</Badge>
                </div>
                <p className="text-sm text-fg-muted mt-0.5">{e.value}</p>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 shrink-0 text-destructive hover:text-destructive"
                onClick={() => handleDelete(e.key)}
                disabled={isPending}
              >
                <Trash2 size={13} />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
