/**
 * AdminRoles — Role management with permission matrix.
 *
 * Features:
 *   - List all roles as cards with colored accent, user count, level
 *   - Create custom roles with name, description, level, color
 *   - Select a role to view/edit its permission matrix
 *   - Delete custom roles (system roles protected)
 *
 * Data flow: server props → local state → server actions → optimistic update
 * Used by: /admin/roles page
 */

"use client"

import { useState, useEffect, useTransition } from "react"
import {
  Plus, Trash2, Check, X, Users, Crown, ShieldCheck, User,
  Eye, Pencil, Lock, Star, Zap, Briefcase, BookOpen, Headphones,
  Wrench, Globe, Bell, Heart, Layers, Database,
} from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip"
import { PermissionMatrix } from "./PermissionMatrix"
import {
  actionCreateRole,
  actionUpdateRole,
  actionDeleteRole,
  actionListRoles,
  actionGetRolePermissions,
} from "@/app/actions/roles"

type RoleWithCount = {
  id: number
  name: string
  description: string
  level: number
  color: string
  icon: string
  isSystem: boolean
  userCount: number
}

type Permission = {
  id: number
  key: string
  label: string
  category: string
  description: string
}

const ICON_OPTIONS = [
  { value: "user", label: "Usuario", Icon: User },
  { value: "shield", label: "Corona", Icon: Crown },
  { value: "user-cog", label: "Escudo", Icon: ShieldCheck },
  { value: "eye", label: "Ojo", Icon: Eye },
  { value: "pencil", label: "Lápiz", Icon: Pencil },
  { value: "lock", label: "Candado", Icon: Lock },
  { value: "star", label: "Estrella", Icon: Star },
  { value: "zap", label: "Rayo", Icon: Zap },
  { value: "briefcase", label: "Maletín", Icon: Briefcase },
  { value: "book", label: "Libro", Icon: BookOpen },
  { value: "headphones", label: "Soporte", Icon: Headphones },
  { value: "wrench", label: "Herramienta", Icon: Wrench },
  { value: "globe", label: "Globo", Icon: Globe },
  { value: "bell", label: "Campana", Icon: Bell },
  { value: "heart", label: "Corazón", Icon: Heart },
  { value: "layers", label: "Capas", Icon: Layers },
  { value: "database", label: "Base de datos", Icon: Database },
] as const

const COLOR_SWATCHES = [
  "#2563eb", // blue (accent)
  "#7c3aed", // violet
  "#db2777", // pink
  "#dc2626", // red
  "#ea580c", // orange
  "#d97706", // amber
  "#16a34a", // green
  "#0d9488", // teal
  "#0284c7", // sky
  "#6e6c69", // gray
  "#1e1e1e", // dark
  "#8b5cf6", // purple
] as const

function RoleIcon({ icon, size = 20 }: { icon: string; size?: number }) {
  const found = ICON_OPTIONS.find((o) => o.value === icon)
  const IconComponent = found?.Icon ?? User
  return <IconComponent size={size} strokeWidth={1.5} />
}

export function AdminRoles({
  initialRoles,
  initialPermissions,
}: {
  initialRoles: RoleWithCount[]
  initialPermissions: Permission[]
}) {
  const [roles, setRoles] = useState(initialRoles)
  const [selectedRoleId, setSelectedRoleId] = useState<number | null>(
    [...initialRoles].sort((a, b) => b.level - a.level)[0]?.id ?? null
  )
  const [activeKeys, setActiveKeys] = useState<string[]>([])
  const [keysLoaded, setKeysLoaded] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [editingRoleId, setEditingRoleId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: "", description: "", level: "20", color: "#6e6c69", icon: "user" })

  const selectedRole = roles.find((r) => r.id === selectedRoleId)
  const sortedRoles = [...roles].sort((a, b) => b.level - a.level)

  function flashSuccess(msg: string) {
    setSuccessMsg(msg)
    setTimeout(() => setSuccessMsg(null), 3000)
  }

  async function loadRolePermissions(roleId: number) {
    // Don't set keysLoaded=false to avoid collapsing the matrix (causes scroll jump)
    try {
      const result = await actionGetRolePermissions({ roleId })
      const keys = result?.data ?? []
      setActiveKeys(keys)
    } catch {
      setActiveKeys([])
    }
    setKeysLoaded(true)
  }

  function selectRole(roleId: number) {
    setSelectedRoleId(roleId)
    loadRolePermissions(roleId)
  }

  useEffect(() => {
    if (selectedRoleId) {
      loadRolePermissions(selectedRoleId)
    }
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function handleCreate() {
    if (!form.name.trim()) {
      setError("El nombre es obligatorio")
      return
    }
    const level = parseInt(form.level, 10)
    if (isNaN(level) || level < 1 || level > 99) {
      setError("El nivel debe estar entre 1 y 99")
      return
    }
    setError(null)

    startTransition(async () => {
      try {
        await actionCreateRole({
          name: form.name.trim(),
          description: form.description.trim(),
          level,
          color: form.color,
          icon: form.icon,
        })
        setForm({ name: "", description: "", level: "20", color: "#6e6c69", icon: "user" })
        setShowCreate(false)
        flashSuccess(`Rol "${form.name}" creado`)
        const fresh = await actionListRoles()
        setRoles(fresh as RoleWithCount[])
      } catch (err) {
        const msg = String(err)
        if (msg.includes("UNIQUE") || msg.includes("unique")) {
          setError("Ya existe un rol con ese nombre")
        } else {
          setError("Error al crear rol")
        }
      }
    })
  }

  function startEditing(role: RoleWithCount) {
    setEditingRoleId(role.id)
    setForm({
      name: role.name,
      description: role.description,
      level: String(role.level),
      color: role.color,
      icon: role.icon,
    })
    setShowCreate(false)
  }

  function handleSaveEdit() {
    if (!editingRoleId || !form.name.trim()) return
    const level = parseInt(form.level, 10)
    if (isNaN(level) || level < 1 || level > 99) {
      setError("El nivel debe estar entre 1 y 99")
      return
    }
    setError(null)

    startTransition(async () => {
      try {
        await actionUpdateRole({ id: editingRoleId, data: {
          name: form.name.trim(),
          description: form.description.trim(),
          level,
          color: form.color,
          icon: form.icon,
        } })
        setEditingRoleId(null)
        setForm({ name: "", description: "", level: "20", color: "#6e6c69", icon: "user" })
        flashSuccess("Rol actualizado")
        const fresh = await actionListRoles()
        setRoles(fresh as RoleWithCount[])
      } catch (err) {
        setError(String(err).includes("UNIQUE") ? "Ya existe un rol con ese nombre" : "Error al actualizar")
      }
    })
  }

  function handleDelete(role: RoleWithCount) {
    if (role.isSystem) return
    if (role.userCount > 0) {
      setError(`No se puede eliminar "${role.name}" porque tiene ${role.userCount} usuario(s) asignado(s)`)
      setTimeout(() => setError(null), 4000)
      return
    }
    if (!confirm(`¿Eliminar el rol "${role.name}"?\nEsta acción no se puede deshacer.`)) return

    startTransition(async () => {
      try {
        await actionDeleteRole({ id: role.id })
        flashSuccess(`Rol "${role.name}" eliminado`)
        const fresh = await actionListRoles()
        setRoles(fresh as RoleWithCount[])
        if (selectedRoleId === role.id) {
          setSelectedRoleId(null)
          setKeysLoaded(false)
        }
      } catch {
        setError("Error al eliminar rol")
      }
    })
  }

  return (
    <TooltipProvider delayDuration={200}>
      <div>
        {/* Header */}
        <div className="flex items-center justify-between" style={{ marginBottom: "24px" }}>
          <div>
            <h2 className="text-lg font-semibold text-fg">Roles del sistema</h2>
            <p className="text-xs text-fg-subtle" style={{ marginTop: "4px" }}>
              Los permisos se suman cuando un usuario tiene múltiples roles.
            </p>
          </div>
          <button
            onClick={() => { setShowCreate(true); setError(null) }}
            className="flex items-center rounded-lg bg-accent text-accent-fg text-sm font-medium hover:opacity-90 transition-opacity"
            style={{ padding: "8px 16px", gap: "8px" }}
          >
            <Plus size={16} />
            Nuevo rol
          </button>
        </div>

        {/* Feedback */}
        {error && (
          <div
            className="text-sm text-destructive flex items-center justify-between rounded-xl"
            style={{ marginBottom: "16px", padding: "10px 14px", background: "color-mix(in srgb, var(--destructive) 8%, var(--surface))" }}
          >
            {error}
            <button onClick={() => setError(null)} className="text-destructive hover:opacity-70"><X size={14} /></button>
          </div>
        )}
        {successMsg && (
          <div
            className="text-sm text-success flex items-center rounded-xl"
            style={{ marginBottom: "16px", padding: "10px 14px", gap: "8px", background: "color-mix(in srgb, var(--success) 8%, var(--surface))" }}
          >
            <Check size={14} /> {successMsg}
          </div>
        )}

        {/* Create/Edit form */}
        {(showCreate || editingRoleId) && (
          <div
            className="rounded-xl bg-surface"
            style={{ padding: "20px", marginBottom: "20px", border: `1px solid ${editingRoleId ? form.color : "var(--border)"}` }}
          >
            <h3 className="text-sm font-semibold text-fg" style={{ marginBottom: "16px" }}>
              {editingRoleId ? `Editar rol: ${form.name}` : "Crear rol personalizado"}
            </h3>
            <div className="grid grid-cols-2" style={{ gap: "12px" }}>
              <input
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                placeholder="Nombre del rol"
                className="rounded-lg bg-bg text-fg text-sm outline-none focus:ring-1 focus:ring-accent transition-shadow"
                style={{ padding: "9px 12px", border: "1px solid var(--border)" }}
              />
              <input
                value={form.description}
                onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                placeholder="Descripción (opcional)"
                className="rounded-lg bg-bg text-fg text-sm outline-none focus:ring-1 focus:ring-accent transition-shadow"
                style={{ padding: "9px 12px", border: "1px solid var(--border)" }}
              />
              <input
                value={form.level}
                onChange={(e) => setForm((f) => ({ ...f, level: e.target.value }))}
                placeholder="Nivel de prioridad (1-99)"
                type="number"
                min={1}
                max={99}
                className="rounded-lg bg-bg text-fg text-sm outline-none focus:ring-1 focus:ring-accent transition-shadow"
                style={{ padding: "9px 12px", border: "1px solid var(--border)" }}
              />
              <div className="flex items-center flex-wrap" style={{ gap: "8px" }}>
                <label className="text-sm text-fg-muted shrink-0">Color:</label>
                {/* Swatches + native picker */}
                <div className="flex items-center" style={{ gap: "4px" }}>
                  {COLOR_SWATCHES.map((c) => (
                    <button
                      key={c}
                      type="button"
                      onClick={() => setForm((f) => ({ ...f, color: c }))}
                      className="rounded-full shrink-0"
                      style={{
                        width: "20px",
                        height: "20px",
                        background: c,
                        outline: form.color === c ? "2px solid var(--fg)" : "none",
                        outlineOffset: "1px",
                      }}
                    />
                  ))}
                  <input
                    type="color"
                    value={form.color}
                    onChange={(e) => setForm((f) => ({ ...f, color: e.target.value }))}
                    className="rounded cursor-pointer shrink-0"
                    style={{ width: "20px", height: "20px", border: "1px solid var(--border)", padding: 0 }}
                    title="Color personalizado"
                  />
                </div>
                {/* Live preview */}
                <div
                  className="flex items-center rounded-lg shrink-0"
                  style={{
                    gap: "8px",
                    padding: "5px 12px 5px 8px",
                    color: form.color,
                    background: `color-mix(in srgb, ${form.color} 10%, transparent)`,
                    border: `1px solid color-mix(in srgb, ${form.color} 20%, transparent)`,
                    borderRadius: "8px",
                  }}
                >
                  <RoleIcon icon={form.icon} size={16} />
                  <span className="text-xs font-medium">{form.name || "Preview"}</span>
                </div>
              </div>
            </div>

            {/* Icon picker */}
            <div style={{ marginTop: "14px" }}>
              <label className="text-xs font-medium text-fg-muted" style={{ display: "block", marginBottom: "8px" }}>
                Ícono
              </label>
              <div className="flex flex-wrap" style={{ gap: "6px" }}>
                {ICON_OPTIONS.map((opt) => {
                  const isSelected = form.icon === opt.value
                  return (
                    <Tooltip key={opt.value}>
                      <TooltipTrigger asChild>
                        <button
                          type="button"
                          onClick={() => setForm((f) => ({ ...f, icon: opt.value }))}
                          className="flex items-center justify-center rounded-lg transition-all"
                          style={{
                            width: "36px",
                            height: "36px",
                            color: isSelected ? form.color : "var(--fg-subtle)",
                            background: isSelected
                              ? `color-mix(in srgb, ${form.color} 12%, transparent)`
                              : "transparent",
                            border: isSelected
                              ? `2px solid ${form.color}`
                              : "1px solid var(--border)",
                          }}
                        >
                          <opt.Icon size={16} strokeWidth={1.5} />
                        </button>
                      </TooltipTrigger>
                      <TooltipContent side="bottom" sideOffset={4}>{opt.label}</TooltipContent>
                    </Tooltip>
                  )
                })}
              </div>
            </div>
            <div className="flex justify-end" style={{ gap: "8px", marginTop: "16px" }}>
              <button
                onClick={() => { setShowCreate(false); setEditingRoleId(null); setError(null); setForm({ name: "", description: "", level: "20", color: "#6e6c69", icon: "user" }) }}
                className="text-sm text-fg-muted hover:text-fg rounded-lg transition-colors"
                style={{ padding: "8px 16px", border: "1px solid var(--border)" }}
              >
                Cancelar
              </button>
              <button
                onClick={editingRoleId ? handleSaveEdit : handleCreate}
                disabled={isPending || !form.name.trim()}
                className="text-sm font-medium rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity"
                style={{ padding: "8px 16px" }}
              >
                {isPending ? "Guardando..." : editingRoleId ? "Guardar" : "Crear"}
              </button>
            </div>
          </div>
        )}

        {/* Roles cards — system roles in a 3-col grid, custom below */}
        {(() => {
          const systemRoles = sortedRoles.filter((r) => r.isSystem)
          const customRoles = sortedRoles.filter((r) => !r.isSystem)
          return (
            <>
        <div className="grid grid-cols-1 md:grid-cols-3" style={{ gap: "14px", marginBottom: customRoles.length > 0 ? "14px" : "28px" }}>
          {systemRoles.map((role) => {
            const isSelected = selectedRoleId === role.id
            return (
              <button
                key={role.id}
                type="button"
                onClick={() => selectRole(role.id)}
                className="relative text-left rounded-xl transition-all"
                style={{
                  padding: 0,
                  overflow: "hidden",
                  border: isSelected ? `2px solid ${role.color}` : "1px solid var(--border)",
                  boxShadow: isSelected ? `0 0 0 1px color-mix(in srgb, ${role.color} 20%, transparent)` : "none",
                }}
              >
                {/* Colored top bar */}
                <div style={{ height: "4px", background: role.color }} />

                <div style={{ padding: "18px 18px 16px" }}>
                  {/* Icon + name */}
                  <div className="flex items-start justify-between" style={{ marginBottom: "8px" }}>
                    <div className="flex items-center" style={{ gap: "12px" }}>
                      <div
                        className="flex items-center justify-center rounded-lg shrink-0"
                        style={{
                          width: "40px",
                          height: "40px",
                          color: role.color,
                          background: `color-mix(in srgb, ${role.color} 10%, transparent)`,
                        }}
                      >
                        <RoleIcon icon={role.icon} size={22} />
                      </div>
                      <div>
                        <div className="font-semibold text-fg" style={{ fontSize: "14px" }}>{role.name}</div>
                        <div className="text-xs text-fg-subtle">Nivel {role.level}</div>
                      </div>
                    </div>
                    <div className="flex items-center shrink-0" style={{ gap: "4px" }}>
                      {role.isSystem && (
                        <span
                          className="text-[10px] font-medium text-fg-subtle rounded-full"
                          style={{ padding: "2px 8px", background: "var(--surface-2)" }}
                        >
                          sistema
                        </span>
                      )}
                      {/* Edit — always available */}
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <div
                            role="button"
                            onClick={(e) => { e.stopPropagation(); startEditing(role) }}
                            className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-accent hover:bg-accent/10 transition-colors"
                            style={{ width: "28px", height: "28px" }}
                          >
                            <Pencil size={13} />
                          </div>
                        </TooltipTrigger>
                        <TooltipContent side="bottom" sideOffset={4}>Editar rol</TooltipContent>
                      </Tooltip>
                      {/* Delete — only custom roles */}
                      {!role.isSystem && (
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div
                              role="button"
                              onClick={(e) => { e.stopPropagation(); handleDelete(role) }}
                              className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-destructive hover:bg-destructive/10 transition-colors"
                              style={{ width: "28px", height: "28px" }}
                            >
                              <Trash2 size={13} />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side="bottom" sideOffset={4}>Eliminar rol</TooltipContent>
                        </Tooltip>
                      )}
                    </div>
                  </div>

                  {/* Description */}
                  <p className="text-xs text-fg-muted" style={{ marginBottom: "14px", lineHeight: "1.5" }}>
                    {role.description || "Sin descripción"}
                  </p>

                  {/* User count */}
                  <div
                    className="flex items-center text-xs text-fg-subtle rounded-lg"
                    style={{ gap: "5px", padding: "6px 10px", background: "var(--surface-2)" }}
                  >
                    <Users size={12} />
                    {role.userCount} usuario{role.userCount !== 1 ? "s" : ""}
                  </div>
                </div>
              </button>
            )
          })}
        </div>

        {/* Custom roles */}
        {customRoles.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-3" style={{ gap: "14px", marginBottom: "28px" }}>
            {customRoles.map((role) => {
              const isSelected = selectedRoleId === role.id
              return (
                <button
                  key={role.id}
                  type="button"
                  onClick={() => selectRole(role.id)}
                  className="relative text-left rounded-xl"
                  style={{
                    padding: 0,
                    overflow: "hidden",
                    border: isSelected ? `2px solid ${role.color}` : "1px solid var(--border)",
                    boxShadow: isSelected ? `0 0 0 1px color-mix(in srgb, ${role.color} 20%, transparent)` : "none",
                  }}
                >
                  <div style={{ height: "4px", background: role.color }} />
                  <div style={{ padding: "18px 18px 16px" }}>
                    <div className="flex items-start justify-between" style={{ marginBottom: "8px" }}>
                      <div className="flex items-center" style={{ gap: "12px" }}>
                        <div
                          className="flex items-center justify-center rounded-lg shrink-0"
                          style={{ width: "40px", height: "40px", color: role.color, background: `color-mix(in srgb, ${role.color} 10%, transparent)` }}
                        >
                          <RoleIcon icon={role.icon} size={22} />
                        </div>
                        <div>
                          <div className="font-semibold text-fg" style={{ fontSize: "14px" }}>{role.name}</div>
                          <div className="text-xs text-fg-subtle">Nivel {role.level}</div>
                        </div>
                      </div>
                      <div className="flex items-center shrink-0" style={{ gap: "4px" }}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div role="button" onClick={(e) => { e.stopPropagation(); startEditing(role) }} className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-accent hover:bg-accent/10 transition-colors" style={{ width: "28px", height: "28px" }}>
                              <Pencil size={13} />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side="bottom" sideOffset={4}>Editar</TooltipContent>
                        </Tooltip>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div role="button" onClick={(e) => { e.stopPropagation(); handleDelete(role) }} className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-destructive hover:bg-destructive/10 transition-colors" style={{ width: "28px", height: "28px" }}>
                              <Trash2 size={13} />
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side="bottom" sideOffset={4}>Eliminar</TooltipContent>
                        </Tooltip>
                      </div>
                    </div>
                    <p className="text-xs text-fg-muted" style={{ marginBottom: "14px", lineHeight: "1.5" }}>{role.description || "Sin descripción"}</p>
                    <div className="flex items-center text-xs text-fg-subtle rounded-lg" style={{ gap: "5px", padding: "6px 10px", background: "var(--surface-2)" }}>
                      <Users size={12} />
                      {role.userCount} usuario{role.userCount !== 1 ? "s" : ""}
                    </div>
                  </div>
                  {isSelected && <div className="absolute bottom-0 left-0 right-0" style={{ height: "2px", background: role.color }} />}
                </button>
              )
            })}
          </div>
        )}
            </>
          )
        })()}

        {/* Permission matrix */}
        {selectedRole && keysLoaded && (
          <PermissionMatrix
            roleId={selectedRole.id}
            roleName={selectedRole.name}
            isSystem={selectedRole.isSystem}
            locked={selectedRole.isSystem && selectedRole.level >= 100}
            allPermissions={initialPermissions}
            activeKeys={activeKeys}
            onUpdate={setActiveKeys}
          />
        )}
      </div>
    </TooltipProvider>
  )
}
