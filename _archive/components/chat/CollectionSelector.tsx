"use client"

import { useEffect, useState } from "react"
import { ChevronDown, FolderOpen } from "lucide-react"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"

const STORAGE_KEY = "rag-selected-collections"

type Props = {
  defaultCollection: string
  availableCollections: string[]
  onCollectionsChange: (collections: string[]) => void
  disabled?: boolean
}

function loadSaved(defaultCollection: string): string[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw) as string[]
      if (Array.isArray(parsed) && parsed.length > 0) return parsed
    }
  } catch {
    // ignorar
  }
  return [defaultCollection]
}

export function CollectionSelector({ defaultCollection, availableCollections, onCollectionsChange, disabled }: Props) {
  const [selected, setSelected] = useState<string[]>([defaultCollection])
  const [open, setOpen] = useState(false)

  useEffect(() => {
    const saved = loadSaved(defaultCollection)
    setSelected(saved)
    onCollectionsChange(saved)
  // eslint-disable-next-line react-hooks/exhaustive-deps -- sync al montar o si cambia la colección por defecto
  }, [defaultCollection])

  function toggle(collection: string) {
    setSelected((prev) => {
      const next = prev.includes(collection)
        ? prev.filter((c) => c !== collection)
        : [...prev, collection]
      const result = next.length === 0 ? [defaultCollection] : next
      localStorage.setItem(STORAGE_KEY, JSON.stringify(result))
      onCollectionsChange(result)
      return result
    })
  }

  const label = selected.length === 1
    ? selected[0]
    : `${selected.length} colecciones`

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          disabled={disabled}
          className="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs border transition-colors hover:opacity-80 disabled:opacity-40"
          style={{ borderColor: "var(--border)", color: "var(--muted-foreground)" }}
        >
          <FolderOpen size={11} />
          <span className="max-w-[120px] truncate">{label}</span>
          <ChevronDown size={10} />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-56 p-2">
        <p className="text-xs font-medium px-1 pb-1" style={{ color: "var(--muted-foreground)" }}>
          Colecciones activas
        </p>
        <Separator className="mb-1" />
        {availableCollections.length === 0 ? (
          <p className="text-xs px-1 py-1" style={{ color: "var(--muted-foreground)" }}>
            Sin colecciones disponibles
          </p>
        ) : (
          availableCollections.map((col) => {
            const isSelected = selected.includes(col)
            return (
              <button
                key={col}
                onClick={() => toggle(col)}
                className="flex items-center justify-between w-full px-2 py-1.5 rounded-md text-sm transition-colors hover:opacity-80"
                style={{
                  background: isSelected ? "var(--accent)" : "transparent",
                  color: isSelected ? "white" : "var(--foreground)",
                }}
              >
                <span className="truncate">{col}</span>
                {isSelected && <Badge className="text-xs ml-1 shrink-0">✓</Badge>}
              </button>
            )
          })
        )}
      </PopoverContent>
    </Popover>
  )
}
