"use client"

import { useState, useTransition, useOptimistic, useCallback } from "react"
import { Plus, Pencil, Trash2, Users, ChevronDown, ChevronRight, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { Map } from "lucide-react"
import {
  actionCreateArea,
  actionUpdateArea,
  actionDeleteArea,
  actionAddUserToArea,
  actionRemoveUserFromArea,
} from "@/app/actions/areas"

// ── Types ──

type AreaMember = { userId: number; areaId: number; user: { id: number; name: string; email: string } }
type AreaCollection = { areaId: number; collectionName: string; permission: string }
type Area = {
  id: number
  name: string
  description: string | null
  createdAt: number
  areaCollections: AreaCollection[]
  userAreas: AreaMember[]
}
type SimpleUser = { id: number; name: string; email: string }

type AreaAction =
  | { type: "create"; area: Area }
  | { type: "delete"; id: number }

const PERM_COLORS: Record<string, string> = {
  read: "var(--accent)",
  write: "var(--success)",
  admin: "var(--warning)",
}

// ── Component ──

export function AdminAreas({
  areas: initialAreas,
  allUsers,
}: {
  areas: Area[]
  allUsers: SimpleUser[]
}) {
  const [optimisticAreas, applyOptimistic] = useOptimistic(
    initialAreas,
    (state: Area[], action: AreaAction) => {
      if (action.type === "delete") return state.filter((a) => a.id !== action.id)
      if (action.type === "create") return [...state, action.area]
      return state
    }
  )
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)

  // Create form
  const [showCreate, setShowCreate] = useState(false)
  const [createName, setCreateName] = useState("")
  const [createDesc, setCreateDesc] = useState("")

  // Edit state
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState("")
  const [editDesc, setEditDesc] = useState("")

  // Expand/members
  const [expandedId, setExpandedId] = useState<number | null>(null)

  // Delete confirm
  const [deleteTarget, setDeleteTarget] = useState<Area | null>(null)

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!createName.trim()) return
    const name = createName.trim()
    const description = createDesc.trim()
    setError(null)
    setCreateName("")
    setCreateDesc("")
    setShowCreate(false)
    startTransition(async () => {
      applyOptimistic({
        type: "create",
        area: { id: 0, name, description, createdAt: Date.now(), areaCollections: [], userAreas: [] },
      })
      try {
        await actionCreateArea({ name, description: description || undefined })
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }

  function startEdit(area: Area) {
    setEditingId(area.id)
    setEditName(area.name)
    setEditDesc(area.description ?? "")
  }

  function handleUpdate(id: number) {
    setError(null)
    startTransition(async () => {
      try {
        await actionUpdateArea({ id, data: { name: editName.trim(), description: editDesc.trim() } })
        setEditingId(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }

  const confirmDelete = useCallback(() => {
    if (!deleteTarget) return
    const id = deleteTarget.id
    setDeleteTarget(null)
    setError(null)
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      try {
        await actionDeleteArea({ id })
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }, [deleteTarget, applyOptimistic, startTransition])

  function handleAddMember(areaId: number, userId: number) {
    startTransition(async () => {
      try {
        await actionAddUserToArea({ userId, areaId })
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }

  function handleRemoveMember(areaId: number, userId: number) {
    startTransition(async () => {
      try {
        await actionRemoveUserFromArea({ userId, areaId })
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
      }
    })
  }

  return (
    <div className="flex flex-col" style={{ gap: "20px" }}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-fg-muted">
            {optimisticAreas.length} área{optimisticAreas.length !== 1 ? "s" : ""}
          </p>
        </div>
        <Button onClick={() => setShowCreate(!showCreate)} style={{ gap: "6px" }}>
          <Plus size={16} /> Nueva área
        </Button>
      </div>

      {error && (
        <p className="text-sm text-destructive" style={{ padding: "8px 12px", borderRadius: "8px", background: "color-mix(in srgb, var(--destructive) 8%, transparent)" }}>
          {error}
        </p>
      )}

      {/* Create form */}
      {showCreate && (
        <form
          onSubmit={handleCreate}
          className="rounded-xl border border-border bg-surface"
          style={{ padding: "20px" }}
        >
          <p className="text-sm font-semibold text-fg" style={{ marginBottom: "12px" }}>Nueva área</p>
          <div className="flex flex-col" style={{ gap: "8px" }}>
            <Input
              autoFocus
              value={createName}
              onChange={(e) => setCreateName(e.target.value)}
              placeholder="Nombre del área"
            />
            <Input
              value={createDesc}
              onChange={(e) => setCreateDesc(e.target.value)}
              placeholder="Descripción (opcional)"
            />
          </div>
          <div className="flex" style={{ gap: "8px", marginTop: "12px" }}>
            <Button type="submit" disabled={isPending || !createName.trim()}>
              {isPending ? "Creando..." : "Crear"}
            </Button>
            <Button variant="ghost" type="button" onClick={() => setShowCreate(false)}>
              Cancelar
            </Button>
          </div>
        </form>
      )}

      {/* Areas list */}
      {optimisticAreas.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Map} />
          <EmptyPlaceholder.Title>Sin áreas</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>
            Creá un área para agrupar usuarios y colecciones.
          </EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="flex flex-col" style={{ gap: "8px" }}>
          {optimisticAreas.map((area) => {
            const isEditing = editingId === area.id
            const isExpanded = expandedId === area.id
            const memberIds = new Set(area.userAreas.map((ua) => ua.userId))
            const nonMembers = allUsers.filter((u) => !memberIds.has(u.id))

            return (
              <div
                key={area.id}
                className="rounded-xl border border-border bg-surface"
              >
                {/* Area row */}
                <div className="flex items-center" style={{ padding: "16px 20px", gap: "12px" }}>
                  {/* Expand toggle */}
                  <button
                    onClick={() => setExpandedId(isExpanded ? null : area.id)}
                    className="shrink-0 text-fg-subtle hover:text-fg transition-colors"
                  >
                    {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                  </button>

                  {/* Name + description */}
                  <div className="flex-1 min-w-0">
                    {isEditing ? (
                      <div className="flex" style={{ gap: "8px" }}>
                        <Input
                          value={editName}
                          onChange={(e) => setEditName(e.target.value)}
                          className="h-8 text-sm"
                          style={{ maxWidth: "200px" }}
                        />
                        <Input
                          value={editDesc}
                          onChange={(e) => setEditDesc(e.target.value)}
                          className="h-8 text-sm"
                          placeholder="Descripción"
                        />
                        <Button size="sm" onClick={() => handleUpdate(area.id)} disabled={isPending}>
                          Guardar
                        </Button>
                        <Button size="sm" variant="ghost" onClick={() => setEditingId(null)}>
                          Cancelar
                        </Button>
                      </div>
                    ) : (
                      <>
                        <span className="font-medium text-fg">{area.name}</span>
                        {area.description && (
                          <span className="text-sm text-fg-muted" style={{ marginLeft: "8px" }}>
                            — {area.description}
                          </span>
                        )}
                      </>
                    )}
                  </div>

                  {/* Collection badges */}
                  <div className="flex items-center shrink-0" style={{ gap: "4px" }}>
                    {area.areaCollections.map((ac) => (
                      <Badge
                        key={ac.collectionName}
                        variant="outline"
                        className="text-xs"
                        style={{ color: PERM_COLORS[ac.permission] ?? "var(--fg-subtle)" }}
                      >
                        {ac.collectionName}
                      </Badge>
                    ))}
                  </div>

                  {/* Member count */}
                  <span className="shrink-0 flex items-center text-xs text-fg-subtle" style={{ gap: "4px" }}>
                    <Users size={13} />
                    {area.userAreas.length}
                  </span>

                  {/* Actions */}
                  {!isEditing && (
                    <div className="flex shrink-0" style={{ gap: "2px" }}>
                      <button
                        onClick={() => startEdit(area)}
                        className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                        style={{ width: "32px", height: "32px" }}
                        title="Editar"
                      >
                        <Pencil size={14} />
                      </button>
                      <button
                        onClick={() => setDeleteTarget(area)}
                        className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-destructive hover:bg-surface-2 transition-colors"
                        style={{ width: "32px", height: "32px" }}
                        title="Eliminar"
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  )}
                </div>

                {/* Expanded: members section */}
                {isExpanded && (
                  <div
                    className="border-t border-border"
                    style={{ padding: "16px 20px 16px 52px" }}
                  >
                    <p className="text-xs font-semibold text-fg-subtle uppercase tracking-wider" style={{ marginBottom: "8px" }}>
                      Miembros
                    </p>

                    {area.userAreas.length === 0 ? (
                      <p className="text-sm text-fg-muted">Sin miembros asignados.</p>
                    ) : (
                      <div className="flex flex-wrap" style={{ gap: "6px", marginBottom: "12px" }}>
                        {area.userAreas.map((ua) => (
                          <span
                            key={ua.userId}
                            className="inline-flex items-center rounded-full border border-border bg-bg text-sm text-fg"
                            style={{ padding: "4px 10px", gap: "6px" }}
                          >
                            {ua.user.name}
                            <button
                              onClick={() => handleRemoveMember(area.id, ua.userId)}
                              className="text-fg-subtle hover:text-destructive transition-colors"
                              title="Quitar del área"
                            >
                              <X size={12} />
                            </button>
                          </span>
                        ))}
                      </div>
                    )}

                    {/* Add member dropdown */}
                    {nonMembers.length > 0 && (
                      <AddMemberDropdown
                        users={nonMembers}
                        onAdd={(userId) => handleAddMember(area.id, userId)}
                        disabled={isPending}
                      />
                    )}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null) }}
        title={`¿Eliminar el área "${deleteTarget?.name}"?`}
        description={
          deleteTarget && deleteTarget.userAreas.length > 0
            ? `Esta área tiene ${deleteTarget.userAreas.length} miembro(s). Primero quitá los miembros.`
            : "Se eliminarán las asignaciones de colecciones de esta área."
        }
        confirmLabel="Eliminar"
        onConfirm={confirmDelete}
      />
    </div>
  )
}

// ── Add member dropdown ──

function AddMemberDropdown({
  users,
  onAdd,
  disabled,
}: {
  users: SimpleUser[]
  onAdd: (userId: number) => void
  disabled: boolean
}) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState("")

  const filtered = search.trim()
    ? users.filter(
        (u) =>
          u.name.toLowerCase().includes(search.toLowerCase()) ||
          u.email.toLowerCase().includes(search.toLowerCase())
      )
    : users

  return (
    <div className="relative" style={{ maxWidth: "300px" }}>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setOpen(!open)}
        style={{ gap: "4px" }}
        disabled={disabled}
      >
        <Plus size={14} /> Agregar miembro
      </Button>

      {open && (
        <div
          className="absolute top-full left-0 z-10 rounded-lg border border-border bg-bg shadow-md"
          style={{ marginTop: "4px", width: "280px", maxHeight: "200px", overflow: "hidden" }}
        >
          <div style={{ padding: "8px" }}>
            <Input
              autoFocus
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar usuario..."
              className="h-8 text-sm"
            />
          </div>
          <div style={{ maxHeight: "140px", overflowY: "auto" }}>
            {filtered.length === 0 ? (
              <p className="text-xs text-fg-muted text-center" style={{ padding: "12px" }}>
                Sin resultados
              </p>
            ) : (
              filtered.map((u) => (
                <button
                  key={u.id}
                  onClick={() => {
                    onAdd(u.id)
                    setOpen(false)
                    setSearch("")
                  }}
                  className="w-full text-left text-sm text-fg hover:bg-surface-2 transition-colors"
                  style={{ padding: "8px 12px" }}
                >
                  <span className="font-medium">{u.name}</span>
                  <span className="text-fg-subtle" style={{ marginLeft: "6px" }}>
                    {u.email}
                  </span>
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  )
}
