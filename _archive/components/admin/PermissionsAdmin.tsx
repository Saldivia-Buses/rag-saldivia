"use client"

import { useState, useTransition } from "react"
import { Check, X } from "lucide-react"
import type { DbArea } from "@rag-saldivia/db"
import { actionSetAreaCollections } from "@/app/actions/areas"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

type AreaWithCollections = DbArea & {
  areaCollections?: Array<{ collectionName: string; permission: string }>
}
type Permission = "none" | "read" | "write"

export function PermissionsAdmin({
  areas, collections,
}: {
  areas: AreaWithCollections[]
  collections: string[]
}) {
  const [selectedArea, setSelectedArea] = useState<number | null>(areas[0]?.id ?? null)
  const [permissions, setPermissions] = useState<Record<string, Permission>>(() => {
    const area = areas.find((a) => a.id === selectedArea)
    const result: Record<string, Permission> = {}
    for (const col of collections) {
      const existing = area?.areaCollections?.find((ac) => ac.collectionName === col)
      result[col] = (existing?.permission ?? "none") as Permission
    }
    return result
  })
  const [saved, setSaved] = useState(false)
  const [isPending, startTransition] = useTransition()

  function selectArea(id: number) {
    setSelectedArea(id)
    const area = areas.find((a) => a.id === id)
    const result: Record<string, Permission> = {}
    for (const col of collections) {
      const existing = area?.areaCollections?.find((ac) => ac.collectionName === col)
      result[col] = (existing?.permission ?? "none") as Permission
    }
    setPermissions(result)
    setSaved(false)
  }

  function togglePermission(col: string, level: "read" | "write") {
    setPermissions((prev) => {
      const current = prev[col] ?? "none"
      if (current === level) return { ...prev, [col]: "none" }
      if (level === "write" && current === "none") return { ...prev, [col]: "write" }
      if (level === "read" && current === "write") return { ...prev, [col]: "read" }
      return { ...prev, [col]: level }
    })
    setSaved(false)
  }

  async function handleSave() {
    if (!selectedArea) return
    startTransition(async () => {
      const toSave = Object.entries(permissions)
        .filter(([, p]) => p !== "none")
        .map(([name, permission]) => ({ name, permission: permission as "read" | "write" | "admin" }))
      await actionSetAreaCollections(selectedArea, toSave)
      setSaved(true)
    })
  }

  const currentArea = areas.find((a) => a.id === selectedArea)

  return (
    <div className="p-6">
      <div className="mb-5">
        <h1 className="text-lg font-semibold text-fg">Permisos</h1>
        <p className="text-sm text-fg-muted mt-0.5">Asigná colecciones a cada área</p>
      </div>

      <div className="grid grid-cols-[180px_1fr] gap-6">
        {/* Sidebar áreas */}
        <div className="space-y-0.5">
          <p className="text-xs font-semibold uppercase tracking-wider text-fg-subtle mb-2">Áreas</p>
          {areas.map((area) => (
            <button
              key={area.id}
              onClick={() => selectArea(area.id)}
              className={`w-full text-left px-3 py-2 rounded-md text-sm transition-colors ${
                selectedArea === area.id
                  ? "bg-accent-subtle text-accent font-medium"
                  : "text-fg-muted hover:bg-surface-2 hover:text-fg"
              }`}
            >
              {area.name}
            </button>
          ))}
        </div>

        {/* Matrix */}
        <div className="space-y-4">
          {currentArea && (
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-fg">{currentArea.name}</h3>
                {currentArea.description && (
                  <p className="text-xs text-fg-muted">{currentArea.description}</p>
                )}
              </div>
              <Button
                size="sm"
                variant={saved ? "outline" : "default"}
                onClick={handleSave}
                disabled={isPending}
                className={saved ? "text-success border-success" : ""}
              >
                {isPending ? "Guardando..." : saved ? "✓ Guardado" : "Guardar cambios"}
              </Button>
            </div>
          )}

          {collections.length === 0 ? (
            <p className="text-sm text-fg-muted">No hay colecciones disponibles</p>
          ) : (
            <div className="rounded-xl border border-border overflow-hidden">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Colección</TableHead>
                    <TableHead className="text-center w-28">Leer</TableHead>
                    <TableHead className="text-center w-28">Escribir</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {collections.map((col) => {
                    const perm = permissions[col] ?? "none"
                    const hasRead  = perm === "read"  || perm === "write"
                    const hasWrite = perm === "write"
                    return (
                      <TableRow key={col}>
                        <TableCell className="font-medium text-fg">{col}</TableCell>
                        <TableCell className="text-center">
                          <button
                            onClick={() => togglePermission(col, "read")}
                            className={`w-8 h-8 rounded-md flex items-center justify-center mx-auto transition-colors ${
                              hasRead ? "bg-success-subtle text-success" : "bg-surface-2 text-fg-subtle hover:bg-surface"
                            }`}
                          >
                            {hasRead ? <Check size={14} /> : <X size={14} />}
                          </button>
                        </TableCell>
                        <TableCell className="text-center">
                          <button
                            onClick={() => togglePermission(col, "write")}
                            className={`w-8 h-8 rounded-md flex items-center justify-center mx-auto transition-colors ${
                              hasWrite ? "bg-accent-subtle text-accent" : "bg-surface-2 text-fg-subtle hover:bg-surface"
                            }`}
                          >
                            {hasWrite ? <Check size={14} /> : <X size={14} />}
                          </button>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
