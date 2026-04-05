"use client"

import { useState, useTransition, useOptimistic } from "react"
import { Plus, Trash2, Brain } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import type { DbUserMemory } from "@rag-saldivia/db"
import { actionAddMemory, actionDeleteMemory } from "@/app/actions/settings"

type MemoryAction =
  | { type: "add"; entry: DbUserMemory }
  | { type: "delete"; key: string }

export function MemoryClient({ entries: initial }: { entries: DbUserMemory[] }) {
  const [newKey, setNewKey] = useState("")
  const [newValue, setNewValue] = useState("")
  const [isPending, startTransition] = useTransition()
  const [optimisticEntries, applyOptimistic] = useOptimistic(
    initial,
    (state: DbUserMemory[], action: MemoryAction) => {
      if (action.type === "add") {
        return [...state.filter((e) => e.key !== action.entry.key), action.entry]
      }
      return state.filter((e) => e.key !== action.key)
    }
  )

  function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!newKey.trim() || !newValue.trim()) return
    const key = newKey.trim()
    const value = newValue.trim()
    setNewKey("")
    setNewValue("")
    startTransition(async () => {
      applyOptimistic({
        type: "add",
        entry: { id: 0, userId: 0, key, value, source: "explicit", createdAt: Date.now(), updatedAt: Date.now() },
      })
      await actionAddMemory({ key, value })
    })
  }

  function handleDelete(key: string) {
    startTransition(async () => {
      applyOptimistic({ type: "delete", key })
      await actionDeleteMemory({ key })
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

      {optimisticEntries.length === 0 ? (
        <p className="text-sm text-fg-muted">Sin preferencias guardadas.</p>
      ) : (
        <div className="space-y-2">
          {optimisticEntries.map((e) => (
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
