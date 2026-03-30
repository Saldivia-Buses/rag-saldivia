/**
 * AdminUsers — Full user management interface for admins.
 *
 * Features:
 *   - List all users with role, status, last login
 *   - Create new users with email, name, password, role
 *   - Toggle active/inactive status (optimistic)
 *   - Change user role (optimistic)
 *   - Reset password with inline input
 *   - Delete user with confirmation (cannot delete self)
 *
 * Uses optimistic updates: local state updates immediately, then syncs with server.
 * On error, reverts to previous state and shows error message.
 *
 * Data flow: initialUsers (server prop) → local state → server actions → optimistic update
 * Used by: /admin/users page
 * Depends on: actions/admin.ts, ui/tooltip
 */

"use client"

import { useState, useTransition, useCallback } from "react"
import { Plus, Trash2, KeyRound, Power, Check, X, Wand2, Copy } from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import {
  actionCreateUser,
  actionUpdateUser,
  actionResetPassword,
  actionDeleteUser,
  actionListUsers,
} from "@/app/actions/admin"
import { UserRoleSelector } from "./UserRoleSelector"

// ── Types ──

type UserRow = {
  id: number
  email: string
  name: string
  role: string
  active: boolean
  createdAt: number
  lastLogin: number | null
}

type RoleInfo = {
  id: number
  name: string
  color: string
  level: number
}

// ── Helpers ──

function formatDate(ts: number | null): string {
  if (!ts) return "Nunca"
  return new Date(ts).toLocaleDateString("es-AR", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

// ── Component ──

export function AdminUsers({
  initialUsers,
  currentUserId,
  allRoles = [],
  userRoleMap: initialRoleMap = {},
}: {
  initialUsers: UserRow[]
  currentUserId: number
  allRoles?: RoleInfo[]
  userRoleMap?: Record<number, number[]>
}) {
  const [users, setUsers] = useState<UserRow[]>(initialUsers)
  const [userRoleMap, setUserRoleMap] = useState<Record<number, number[]>>(initialRoleMap)
  const [showCreate, setShowCreate] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [error, setError] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<UserRow | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)

  // Reset password state
  const [resetingId, setResetingId] = useState<number | null>(null)
  const [newPassword, setNewPassword] = useState("")
  const [resetError, setResetError] = useState<string | null>(null)
  const [generatedPassword, setGeneratedPassword] = useState<{ userId: number; password: string } | null>(null)

  // Create form state
  const [form, setForm] = useState({ email: "", name: "", password: "", role: "user" })

  /** Refresh user list from server */
  const refreshUsers = useCallback(async () => {
    try {
      const fresh = await actionListUsers()
      setUsers(fresh as UserRow[])
    } catch {
      // Keep current state if refresh fails
    }
  }, [])

  /** Flash a success message for 3 seconds */
  function flashSuccess(msg: string) {
    setSuccessMsg(msg)
    setTimeout(() => setSuccessMsg(null), 3000)
  }

  // ── Create ──

  function handleCreate() {
    if (!form.email || !form.name || !form.password) {
      setError("Completá todos los campos")
      return
    }
    if (form.password.length < 6) {
      setError("La contraseña debe tener al menos 6 caracteres")
      return
    }
    setError(null)
    startTransition(async () => {
      try {
        await actionCreateUser({
          email: form.email,
          name: form.name,
          password: form.password,
          role: form.role as "admin" | "area_manager" | "user",
        })
        setForm({ email: "", name: "", password: "", role: "user" })
        setShowCreate(false)
        flashSuccess(`Usuario ${form.email} creado`)
        await refreshUsers()
      } catch (err) {
        const msg = String(err)
        if (msg.includes("UNIQUE") || msg.includes("unique")) {
          setError("Ya existe un usuario con ese email")
        } else {
          setError("Error al crear usuario")
        }
      }
    })
  }

  // ── Toggle active ──

  function handleToggleActive(user: UserRow) {
    // Optimistic update
    const newActive = !user.active
    setUsers((prev) => prev.map((u) => (u.id === user.id ? { ...u, active: newActive } : u)))

    startTransition(async () => {
      try {
        await actionUpdateUser({ id: user.id, data: { active: newActive } })
        flashSuccess(`${user.name} ${newActive ? "activado" : "desactivado"}`)
      } catch {
        // Revert on error
        setUsers((prev) => prev.map((u) => (u.id === user.id ? { ...u, active: !newActive } : u)))
        setError("Error al cambiar estado")
      }
    })
  }

  // ── Change role ──

  function handleChangeRole(userId: number, newRole: string) {
    const user = users.find((u) => u.id === userId)
    if (!user) return
    const oldRole = user.role

    // Optimistic update
    setUsers((prev) => prev.map((u) => (u.id === userId ? { ...u, role: newRole } : u)))

    startTransition(async () => {
      try {
        await actionUpdateUser({ id: userId, data: { role: newRole as "admin" | "area_manager" | "user" } })
        flashSuccess(`Rol de ${user.name} cambiado a ${newRole}`)
      } catch {
        // Revert
        setUsers((prev) => prev.map((u) => (u.id === userId ? { ...u, role: oldRole } : u)))
        setError("Error al cambiar rol")
      }
    })
  }

  // ── Reset password ──

  function generatePassword() {
    const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
    let pw = ""
    for (let i = 0; i < 12; i++) pw += chars[Math.floor(Math.random() * chars.length)]
    return pw
  }

  function handleResetPassword(userId: number) {
    if (!newPassword) {
      setResetError("Ingresá una contraseña")
      return
    }
    if (newPassword.length < 6) {
      setResetError("Mínimo 6 caracteres")
      return
    }
    setResetError(null)
    startTransition(async () => {
      try {
        await actionResetPassword({ userId, newPassword })
        setGeneratedPassword({ userId, password: newPassword })
        setResetingId(null)
        setNewPassword("")
        const user = users.find((u) => u.id === userId)
        flashSuccess(`Contraseña de ${user?.name ?? "usuario"} reseteada`)
      } catch {
        setResetError("Error al resetear contraseña")
      }
    })
  }

  function handleAutoReset(userId: number) {
    const pw = generatePassword()
    setNewPassword(pw)
    setResetingId(userId)
  }

  // ── Delete ──

  function handleDeleteClick(user: UserRow) {
    if (user.id === currentUserId) {
      setError("No podés eliminarte a vos mismo")
      setTimeout(() => setError(null), 3000)
      return
    }
    setDeleteTarget(user)
  }

  const confirmDeleteUser = useCallback(() => {
    if (!deleteTarget) return
    const user = deleteTarget
    setDeleteTarget(null)

    // Optimistic: remove from list
    setUsers((prev) => prev.filter((u) => u.id !== user.id))

    startTransition(async () => {
      try {
        await actionDeleteUser({ id: user.id })
        flashSuccess(`${user.name} eliminado`)
      } catch {
        // Revert
        await refreshUsers()
        setError("Error al eliminar usuario")
      }
    })
  }, [deleteTarget, refreshUsers, startTransition])

  return (
    <TooltipProvider delayDuration={200}>
      <div>
          {/* Header */}
          <div className="flex items-center justify-between" style={{ marginBottom: "24px" }}>
            <div>
              <h1 className="text-xl font-semibold text-fg">Gestión de usuarios</h1>
              <p className="text-sm text-fg-subtle" style={{ marginTop: "4px" }}>
                {users.length} usuario{users.length !== 1 ? "s" : ""} registrado{users.length !== 1 ? "s" : ""}
              </p>
            </div>
            <button
              onClick={() => { setShowCreate(true); setError(null) }}
              className="flex items-center rounded-lg bg-accent text-accent-fg text-sm font-medium hover:opacity-90 transition-opacity"
              style={{ padding: "8px 16px", gap: "8px" }}
            >
              <Plus size={16} />
              Nuevo usuario
            </button>
          </div>

          {/* Global feedback messages */}
          {error && (
            <div
              className="text-sm text-destructive flex items-center justify-between"
              style={{ marginBottom: "16px", padding: "10px 14px", borderRadius: "10px", background: "color-mix(in srgb, var(--destructive) 10%, transparent)" }}
            >
              {error}
              <button onClick={() => setError(null)} className="text-destructive hover:opacity-70"><X size={14} /></button>
            </div>
          )}
          {successMsg && (
            <div
              className="text-sm text-success flex items-center"
              style={{ marginBottom: "16px", padding: "10px 14px", borderRadius: "10px", gap: "8px", background: "color-mix(in srgb, var(--success) 10%, transparent)" }}
            >
              <Check size={14} /> {successMsg}
            </div>
          )}

          {/* Create user form */}
          {showCreate && (
            <div
              className="border border-border rounded-xl bg-surface"
              style={{ padding: "20px", marginBottom: "20px" }}
            >
              <h2 className="text-sm font-semibold text-fg" style={{ marginBottom: "16px" }}>
                Crear usuario
              </h2>
              <div className="grid grid-cols-2" style={{ gap: "12px" }}>
                <input
                  value={form.name}
                  onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                  placeholder="Nombre completo"
                  className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
                  style={{ padding: "8px 12px" }}
                />
                <input
                  value={form.email}
                  onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
                  placeholder="Email"
                  type="email"
                  className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
                  style={{ padding: "8px 12px" }}
                />
                <input
                  value={form.password}
                  onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
                  placeholder="Contraseña (mín. 6 caracteres)"
                  type="password"
                  className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
                  style={{ padding: "8px 12px" }}
                  onKeyDown={(e) => { if (e.key === "Enter") handleCreate() }}
                />
                <select
                  value={form.role}
                  onChange={(e) => setForm((f) => ({ ...f, role: e.target.value }))}
                  className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
                  style={{ padding: "8px 12px" }}
                >
                  <option value="user">Usuario</option>
                  <option value="area_manager">Manager</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
              <div className="flex justify-end" style={{ gap: "8px", marginTop: "16px" }}>
                <button
                  onClick={() => { setShowCreate(false); setError(null) }}
                  className="text-sm text-fg-muted hover:text-fg rounded-lg border border-border transition-colors"
                  style={{ padding: "8px 16px" }}
                >
                  Cancelar
                </button>
                <button
                  onClick={handleCreate}
                  disabled={isPending || !form.email || !form.name || !form.password}
                  className="text-sm font-medium rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity"
                  style={{ padding: "8px 16px" }}
                >
                  {isPending ? "Creando..." : "Crear"}
                </button>
              </div>
            </div>
          )}

          {/* Users table */}
          <div className="border border-border rounded-xl overflow-visible">
            <table className="w-full text-sm">
              <thead>
                <tr style={{ backgroundColor: "var(--surface)" }}>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>Usuario</th>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>Rol</th>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>Estado</th>
                  <th className="text-left text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>Último acceso</th>
                  <th className="text-right text-xs font-semibold text-fg-muted" style={{ padding: "10px 16px" }}>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => {
                  const isSelf = user.id === currentUserId
                  const isSuperAdmin = user.id === 1
                  const isProtected = isSelf || isSuperAdmin

                  return (
                    <tr
                      key={user.id}
                      className={`border-t border-border transition-colors ${isSelf ? "bg-accent/5" : isSuperAdmin ? "bg-accent/3" : "hover:bg-surface/50"}`}
                    >
                      {/* User info */}
                      <td style={{ padding: "12px 16px" }}>
                        <div className="flex items-center" style={{ gap: "8px" }}>
                          <div>
                            <div className="font-medium text-fg">
                              {user.name}
                              {isSelf && <span className="text-xs text-accent font-normal" style={{ marginLeft: "6px" }}>(vos)</span>}
                              {isSuperAdmin && !isSelf && <span className="text-xs text-accent font-normal" style={{ marginLeft: "6px" }}>sudo</span>}
                            </div>
                            <div className="text-xs text-fg-subtle">{user.email}</div>
                          </div>
                        </div>
                      </td>

                      {/* Role selector */}
                      <td style={{ padding: "12px 16px" }}>
                        {allRoles.length > 0 ? (
                          <UserRoleSelector
                            userId={user.id}
                            allRoles={allRoles}
                            assignedRoleIds={userRoleMap[user.id] ?? []}
                            disabled={isProtected}
                            onUpdate={(newIds) => {
                              setUserRoleMap((prev) => ({ ...prev, [user.id]: newIds }))
                            }}
                          />
                        ) : (
                          <select
                            value={user.role}
                            onChange={(e) => handleChangeRole(user.id, e.target.value)}
                            disabled={isProtected}
                            className="text-xs font-medium rounded-md bg-transparent border-0 cursor-pointer disabled:cursor-not-allowed disabled:opacity-50"
                            style={{ color: "var(--fg-subtle)", padding: "2px 4px" }}
                          >
                            <option value="user">Usuario</option>
                            <option value="area_manager">Manager</option>
                            <option value="admin">Admin</option>
                          </select>
                        )}
                      </td>

                      {/* Active status */}
                      <td style={{ padding: "12px 16px" }}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button
                              onClick={() => !isProtected && handleToggleActive(user)}
                              disabled={isProtected}
                              className={`flex items-center text-xs font-medium rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
                                user.active ? "text-success" : "text-fg-subtle"
                              }`}
                              style={{ gap: "4px", padding: "2px 8px" }}
                            >
                              <Power size={12} />
                              {user.active ? "Activo" : "Inactivo"}
                            </button>
                          </TooltipTrigger>
                          <TooltipContent side="bottom" sideOffset={4}>
                            {isSuperAdmin ? "Admin protegido" : isSelf ? "No podés desactivarte a vos mismo" : user.active ? "Desactivar usuario" : "Activar usuario"}
                          </TooltipContent>
                        </Tooltip>
                      </td>

                      {/* Last login */}
                      <td className="text-fg-subtle text-xs" style={{ padding: "12px 16px" }}>
                        {formatDate(user.lastLogin)}
                      </td>

                      {/* Actions */}
                      <td style={{ padding: "12px 16px" }}>
                        <div className="flex items-center justify-end" style={{ gap: "4px" }}>
                          {/* Reset password */}
                          {resetingId === user.id ? (
                            <div className="flex items-center" style={{ gap: "4px" }}>
                              <div>
                                <div className="flex items-center" style={{ gap: "3px" }}>
                                  <input
                                    value={newPassword}
                                    onChange={(e) => { setNewPassword(e.target.value); setResetError(null) }}
                                    placeholder="Nueva contraseña"
                                    type="text"
                                    className={`text-xs rounded bg-bg text-fg outline-none focus:ring-1 focus:ring-accent transition-shadow font-mono ${resetError ? "ring-1 ring-destructive" : ""}`}
                                    style={{ padding: "4px 8px", width: "140px", border: "1px solid var(--border)" }}
                                    autoFocus
                                    onKeyDown={(e) => {
                                      if (e.key === "Enter") handleResetPassword(user.id)
                                      if (e.key === "Escape") { setResetingId(null); setNewPassword(""); setResetError(null) }
                                    }}
                                  />
                                  <Tooltip>
                                    <TooltipTrigger asChild>
                                      <button
                                        type="button"
                                        onClick={() => setNewPassword(generatePassword())}
                                        className="flex items-center justify-center rounded text-fg-subtle hover:text-accent transition-colors"
                                        style={{ width: "24px", height: "24px" }}
                                      >
                                        <Wand2 size={12} />
                                      </button>
                                    </TooltipTrigger>
                                    <TooltipContent side="bottom" sideOffset={4}>Generar</TooltipContent>
                                  </Tooltip>
                                </div>
                                {resetError && <div className="text-[10px] text-destructive" style={{ marginTop: "2px" }}>{resetError}</div>}
                              </div>
                              <button onClick={() => handleResetPassword(user.id)} className="text-xs text-accent hover:underline font-medium">OK</button>
                              <button onClick={() => { setResetingId(null); setNewPassword(""); setResetError(null) }} className="text-xs text-fg-subtle hover:text-fg">
                                <X size={12} />
                              </button>
                            </div>
                          ) : generatedPassword?.userId === user.id ? (
                            <div className="flex items-center" style={{ gap: "4px" }}>
                              <code
                                className="text-xs font-mono text-fg bg-surface-2 rounded select-all"
                                style={{ padding: "3px 8px" }}
                              >
                                {generatedPassword.password}
                              </code>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    type="button"
                                    onClick={() => {
                                      navigator.clipboard.writeText(generatedPassword.password)
                                      flashSuccess("Contraseña copiada")
                                    }}
                                    className="flex items-center justify-center rounded text-fg-subtle hover:text-accent transition-colors"
                                    style={{ width: "24px", height: "24px" }}
                                  >
                                    <Copy size={12} />
                                  </button>
                                </TooltipTrigger>
                                <TooltipContent side="bottom" sideOffset={4}>Copiar</TooltipContent>
                              </Tooltip>
                              <button
                                onClick={() => setGeneratedPassword(null)}
                                className="text-xs text-fg-subtle hover:text-fg"
                              >
                                <X size={12} />
                              </button>
                            </div>
                          ) : (
                            <div className="flex items-center" style={{ gap: "2px" }}>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    onClick={() => handleAutoReset(user.id)}
                                    className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-accent hover:bg-accent/10 transition-colors"
                                    style={{ width: "32px", height: "32px" }}
                                  >
                                    <Wand2 size={14} />
                                  </button>
                                </TooltipTrigger>
                                <TooltipContent side="bottom" sideOffset={4}>Generar contraseña</TooltipContent>
                              </Tooltip>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    onClick={() => { setResetingId(user.id); setResetError(null) }}
                                    className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-fg hover:bg-surface-2 transition-colors"
                                    style={{ width: "32px", height: "32px" }}
                                  >
                                    <KeyRound size={14} />
                                  </button>
                                </TooltipTrigger>
                                <TooltipContent side="bottom" sideOffset={4}>Elegir contraseña</TooltipContent>
                              </Tooltip>
                            </div>
                          )}

                          {/* Delete */}
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <button
                                onClick={() => handleDeleteClick(user)}
                                disabled={isProtected}
                                className="flex items-center justify-center rounded-lg text-fg-subtle hover:text-destructive hover:bg-destructive/10 transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:text-fg-subtle disabled:hover:bg-transparent"
                                style={{ width: "32px", height: "32px" }}
                              >
                                <Trash2 size={14} />
                              </button>
                            </TooltipTrigger>
                            <TooltipContent side="bottom" sideOffset={4}>
                              {isSuperAdmin ? "Admin protegido" : isSelf ? "No podés eliminarte" : "Eliminar usuario"}
                            </TooltipContent>
                          </Tooltip>
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
      </div>

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null) }}
        title={`¿Eliminar a ${deleteTarget?.name}?`}
        description={`${deleteTarget?.email} — Esta acción no se puede deshacer.`}
        onConfirm={confirmDeleteUser}
      />
    </TooltipProvider>
  )
}
