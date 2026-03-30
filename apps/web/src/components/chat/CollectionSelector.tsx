"use client"

import { useState, useCallback } from "react"
import { ChevronDown, FolderOpen, Check } from "lucide-react"
import { useLocalStorage } from "@/hooks/useLocalStorage"

type Props = {
  defaultCollection: string
  availableCollections: string[]
  onCollectionsChange: (collections: string[]) => void
  disabled?: boolean
}

export function CollectionSelector({
  defaultCollection,
  availableCollections,
  onCollectionsChange,
  disabled,
}: Props) {
  const [selected, setSelected] = useLocalStorage<string[]>(
    "rag-selected-collections",
    [defaultCollection]
  )
  const [open, setOpen] = useState(false)

  const toggle = useCallback(
    (collection: string) => {
      setSelected((prev) => {
        const next = prev.includes(collection)
          ? prev.filter((c) => c !== collection)
          : [...prev, collection]
        const result = next.length === 0 ? [defaultCollection] : next
        onCollectionsChange(result)
        return result
      })
    },
    [defaultCollection, onCollectionsChange, setSelected]
  )

  const label =
    selected.length === 1 ? selected[0] : `${selected.length} colecciones`

  return (
    <div className="relative">
      <button
        disabled={disabled}
        onClick={() => setOpen(!open)}
        className="flex items-center rounded-full text-xs border border-border text-fg-subtle transition-colors hover:text-fg hover:border-fg-subtle disabled:opacity-40"
        style={{ padding: "4px 10px", gap: "4px" }}
      >
        <FolderOpen size={11} />
        <span className="max-w-[120px] truncate">{label}</span>
        <ChevronDown size={10} />
      </button>

      {open && (
        <>
          {/* Backdrop */}
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          {/* Dropdown */}
          <div
            className="absolute bottom-full left-0 z-20 rounded-lg border border-border bg-bg shadow-md"
            style={{ marginBottom: "4px", width: "220px", maxHeight: "240px", overflow: "auto" }}
          >
            <p
              className="text-xs font-medium text-fg-subtle"
              style={{ padding: "8px 10px 4px" }}
            >
              Colecciones activas
            </p>
            {availableCollections.length === 0 ? (
              <p className="text-xs text-fg-muted" style={{ padding: "8px 10px" }}>
                Sin colecciones
              </p>
            ) : (
              availableCollections.map((col) => {
                const isSelected = selected.includes(col)
                return (
                  <button
                    key={col}
                    onClick={() => toggle(col)}
                    className={`flex items-center justify-between w-full text-left text-sm transition-colors ${
                      isSelected
                        ? "text-accent font-medium"
                        : "text-fg-muted hover:text-fg hover:bg-surface"
                    }`}
                    style={{ padding: "6px 10px" }}
                  >
                    <span className="truncate">{col}</span>
                    {isSelected && <Check size={14} className="shrink-0 text-accent" />}
                  </button>
                )
              })
            )}
          </div>
        </>
      )}
    </div>
  )
}
