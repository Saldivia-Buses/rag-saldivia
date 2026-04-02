"use client"

import { useState, useTransition } from "react"
import { X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { actionSetAreaCollections } from "@/app/actions/areas"

type AreaCollection = { collectionName: string; permission: string }
type Area = {
  id: number
  name: string
  description: string | null
  areaCollections: AreaCollection[]
}

type PermLevel = "none" | "read" | "write" | "admin"

const PERM_ORDER: PermLevel[] = ["none", "read", "write", "admin"]
const PERM_STYLES: Record<PermLevel, { bg: string; fg: string }> = {
  none: { bg: "var(--surface-2)", fg: "var(--fg-subtle)" },
  read: { bg: "color-mix(in srgb, var(--accent) 15%, transparent)", fg: "var(--accent)" },
  write: { bg: "color-mix(in srgb, var(--success) 15%, transparent)", fg: "var(--success)" },
  admin: { bg: "color-mix(in srgb, var(--warning) 15%, transparent)", fg: "var(--warning)" },
}

function cyclePerm(current: PermLevel): PermLevel {
  const idx = PERM_ORDER.indexOf(current)
  return PERM_ORDER[(idx + 1) % PERM_ORDER.length]!
}

export function AdminPermissions({
  areas,
  collections,
  preselectedAreaId,
}: {
  areas: Area[]
  collections: string[]
  preselectedAreaId?: number
}) {
  function buildPerms(areaId: number | null): Record<string, PermLevel> {
    const area = areas.find((a) => a.id === areaId)
    const result: Record<string, PermLevel> = {}
    for (const col of collections) {
      const existing = area?.areaCollections?.find((ac) => ac.collectionName === col)
      result[col] = (existing?.permission ?? "none") as PermLevel
    }
    return result
  }

  const [selectedArea, setSelectedArea] = useState<number | null>(preselectedAreaId ?? areas[0]?.id ?? null)
  const [perms, setPerms] = useState<Record<string, PermLevel>>(() => buildPerms(selectedArea))
  const [dirty, setDirty] = useState(false)
  const [saved, setSaved] = useState(false)
  const [isPending, startTransition] = useTransition()

  function selectArea(id: number) {
    setSelectedArea(id)
    setPerms(buildPerms(id))
    setDirty(false)
    setSaved(false)
  }

  function togglePerm(col: string) {
    setPerms((prev) => ({ ...prev, [col]: cyclePerm(prev[col] ?? "none") }))
    setDirty(true)
    setSaved(false)
  }

  function bulkSetAll(level: PermLevel) {
    setPerms(() => {
      const next: Record<string, PermLevel> = {}
      for (const col of collections) next[col] = level
      return next
    })
    setDirty(true)
    setSaved(false)
  }

  function handleSave() {
    if (!selectedArea) return
    const areaId = selectedArea
    startTransition(async () => {
      const toSave = Object.entries(perms)
        .filter(([, p]) => p !== "none")
        .map(([name, permission]) => ({ name, permission: permission as "read" | "write" | "admin" }))
      await actionSetAreaCollections({ areaId, collections: toSave })
      setDirty(false)
      setSaved(true)
    })
  }

  const currentArea = areas.find((a) => a.id === selectedArea)

  return (
    <div className="flex" style={{ gap: "24px", minHeight: "300px" }}>
      {/* Sidebar: areas */}
      <div className="shrink-0" style={{ width: "180px" }}>
        <p className="text-xs font-semibold uppercase tracking-wider text-fg-subtle" style={{ marginBottom: "8px" }}>
          Áreas
        </p>
        <div className="flex flex-col" style={{ gap: "2px" }}>
          {areas.map((area) => (
            <button
              key={area.id}
              onClick={() => selectArea(area.id)}
              className={`text-left text-sm rounded-lg transition-colors ${
                selectedArea === area.id
                  ? "bg-accent-subtle text-accent font-medium"
                  : "text-fg-muted hover:bg-surface-2 hover:text-fg"
              }`}
              style={{ padding: "8px 12px" }}
            >
              {area.name}
              <span className="text-xs text-fg-subtle" style={{ marginLeft: "6px" }}>
                ({area.areaCollections.length})
              </span>
            </button>
          ))}
        </div>
        {areas.length === 0 && (
          <p className="text-sm text-fg-muted" style={{ padding: "12px 0" }}>
            Sin áreas. Creá una en la pestaña Áreas.
          </p>
        )}
      </div>

      {/* Matrix */}
      <div className="flex-1 min-w-0">
        {currentArea ? (
          <div className="flex flex-col" style={{ gap: "16px" }}>
            {/* Area header */}
            <div className="flex items-center justify-between">
              <div>
                <h3 className="font-semibold text-fg">{currentArea.name}</h3>
                {currentArea.description && (
                  <p className="text-xs text-fg-muted" style={{ marginTop: "2px" }}>
                    {currentArea.description}
                  </p>
                )}
              </div>
              <div className="flex items-center" style={{ gap: "8px" }}>
                {dirty && (
                  <Button size="sm" onClick={handleSave} disabled={isPending}>
                    {isPending ? "Guardando..." : "Guardar cambios"}
                  </Button>
                )}
                {saved && !dirty && (
                  <Badge variant="outline" className="text-success border-success">
                    Guardado
                  </Badge>
                )}
              </div>
            </div>

            {/* Bulk actions */}
            <div className="flex items-center" style={{ gap: "6px" }}>
              <span className="text-xs text-fg-subtle">Bulk:</span>
              <Button variant="ghost" size="sm" onClick={() => bulkSetAll("read")} className="text-xs h-7">
                Todas read
              </Button>
              <Button variant="ghost" size="sm" onClick={() => bulkSetAll("write")} className="text-xs h-7">
                Todas write
              </Button>
              <Button variant="ghost" size="sm" onClick={() => bulkSetAll("none")} className="text-xs h-7 text-destructive">
                Quitar todas
              </Button>
            </div>

            {/* Collection grid */}
            {collections.length === 0 ? (
              <p className="text-sm text-fg-muted">No hay colecciones disponibles.</p>
            ) : (
              <div className="rounded-xl border border-border overflow-hidden">
                <table className="w-full text-sm">
                  <thead style={{ backgroundColor: "var(--surface)" }}>
                    <tr>
                      <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>
                        Colección
                      </th>
                      <th className="text-center text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px", width: "100px" }}>
                        Permiso
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {collections.map((col) => {
                      const level = perms[col] ?? "none"
                      const style = PERM_STYLES[level]
                      return (
                        <tr key={col} style={{ borderTop: "1px solid var(--border)" }}>
                          <td className="font-medium text-fg" style={{ padding: "10px 16px" }}>
                            {col}
                          </td>
                          <td style={{ padding: "10px 16px", textAlign: "center" }}>
                            <button
                              onClick={() => togglePerm(col)}
                              className="inline-flex items-center justify-center rounded-lg text-xs font-medium transition-colors"
                              style={{
                                width: "72px",
                                height: "30px",
                                backgroundColor: style.bg,
                                color: style.fg,
                              }}
                            >
                              {level === "none" ? <X size={14} /> : level}
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
        ) : (
          <p className="text-sm text-fg-muted" style={{ padding: "48px 0", textAlign: "center" }}>
            Seleccioná un área para ver sus permisos.
          </p>
        )}
      </div>
    </div>
  )
}
