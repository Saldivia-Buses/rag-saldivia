"use client"

import { useState, useTransition } from "react"
import { Check, X } from "lucide-react"
import type { DbArea } from "@rag-saldivia/db"
import { actionSetAreaCollections } from "@/app/actions/areas"

type AreaWithCollections = DbArea & {
  areaCollections?: Array<{ collectionName: string; permission: string }>
}

type Permission = "none" | "read" | "write"

export function PermissionsAdmin({
  areas,
  collections,
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
    <div className="grid grid-cols-[200px_1fr] gap-6">
      {/* Sidebar áreas */}
      <div className="space-y-1">
        <p className="text-xs font-medium uppercase tracking-wider mb-2" style={{ color: "var(--muted-foreground)" }}>Áreas</p>
        {areas.map((area) => (
          <button
            key={area.id}
            onClick={() => selectArea(area.id)}
            className="w-full text-left px-3 py-2 rounded-md text-sm transition-colors"
            style={{
              background: selectedArea === area.id ? "var(--accent)" : "transparent",
              color: selectedArea === area.id ? "var(--foreground)" : "var(--muted-foreground)",
            }}
          >
            {area.name}
          </button>
        ))}
      </div>

      {/* Permissions matrix */}
      <div className="space-y-4">
        {currentArea && (
          <div className="flex items-center justify-between">
            <div>
              <h3 className="font-medium">{currentArea.name}</h3>
              {currentArea.description && (
                <p className="text-xs" style={{ color: "var(--muted-foreground)" }}>{currentArea.description}</p>
              )}
            </div>
            <button
              onClick={handleSave}
              disabled={isPending}
              className="px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
              style={{ background: saved ? "#dcfce7" : "var(--primary)", color: saved ? "#166534" : "var(--primary-foreground)" }}
            >
              {isPending ? "Guardando..." : saved ? "✓ Guardado" : "Guardar cambios"}
            </button>
          </div>
        )}

        {collections.length === 0 ? (
          <p className="text-sm" style={{ color: "var(--muted-foreground)" }}>No hay colecciones disponibles</p>
        ) : (
          <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
            <table className="w-full text-sm">
              <thead>
                <tr style={{ background: "var(--muted)" }}>
                  <th className="text-left px-4 py-3 font-medium">Colección</th>
                  <th className="text-center px-4 py-3 font-medium w-24">Leer</th>
                  <th className="text-center px-4 py-3 font-medium w-24">Escribir</th>
                </tr>
              </thead>
              <tbody>
                {collections.map((col, i) => {
                  const perm = permissions[col] ?? "none"
                  return (
                    <tr key={col} style={{ borderTop: i > 0 ? "1px solid var(--border)" : undefined }}>
                      <td className="px-4 py-3 font-medium">{col}</td>
                      <td className="px-4 py-3 text-center">
                        <button
                          onClick={() => togglePermission(col, "read")}
                          className="w-8 h-8 rounded-md flex items-center justify-center mx-auto transition-colors"
                          style={{
                            background: (perm === "read" || perm === "write") ? "#dcfce7" : "var(--accent)",
                            color: (perm === "read" || perm === "write") ? "#166534" : "var(--muted-foreground)",
                          }}
                        >
                          {(perm === "read" || perm === "write") ? <Check size={14} /> : <X size={14} />}
                        </button>
                      </td>
                      <td className="px-4 py-3 text-center">
                        <button
                          onClick={() => togglePermission(col, "write")}
                          className="w-8 h-8 rounded-md flex items-center justify-center mx-auto transition-colors"
                          style={{
                            background: perm === "write" ? "#dbeafe" : "var(--accent)",
                            color: perm === "write" ? "#1d4ed8" : "var(--muted-foreground)",
                          }}
                        >
                          {perm === "write" ? <Check size={14} /> : <X size={14} />}
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
