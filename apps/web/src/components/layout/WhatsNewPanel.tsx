"use client"

import { useEffect, useState } from "react"
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { marked } from "marked"

const STORAGE_KEY = "last_seen_version"

type Entry = { version: string; content: string }

type Changelog = { version: string; entries: Entry[] }

type Props = {
  open: boolean
  onClose: () => void
  changelog: Changelog
}

export function WhatsNewPanel({ open, onClose, changelog }: Props) {
  useEffect(() => {
    if (open && changelog.version) {
      localStorage.setItem(STORAGE_KEY, changelog.version)
    }
  }, [open, changelog.version])

  return (
    <Sheet open={open} onOpenChange={(o) => !o && onClose()}>
      <SheetContent side="right" className="w-80 overflow-y-auto">
        <SheetHeader>
          <SheetTitle>¿Qué hay de nuevo?</SheetTitle>
          {changelog.version && (
            <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>
              v{changelog.version}
            </p>
          )}
        </SheetHeader>
        <div className="mt-4 space-y-6">
          {changelog.entries.map((entry) => (
            <div key={entry.version}>
              <p className="text-xs font-semibold mb-2 uppercase tracking-wide" style={{ color: "var(--accent)" }}>
                {entry.version}
              </p>
              <div
                className="text-sm prose prose-sm max-w-none"
                style={{ color: "var(--foreground)" }}
                dangerouslySetInnerHTML={{ __html: marked.parse(entry.content) as string }}
              />
            </div>
          ))}
          {changelog.entries.length === 0 && (
            <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>
              Sin entradas de changelog.
            </p>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}

/** Retorna true si la versión actual es mayor a la última vista */
export function useHasNewVersion(currentVersion: string): boolean {
  const [hasNew, setHasNew] = useState(false)

  useEffect(() => {
    const lastSeen = localStorage.getItem(STORAGE_KEY)
    setHasNew(lastSeen !== currentVersion)
  }, [currentVersion])

  return hasNew
}
