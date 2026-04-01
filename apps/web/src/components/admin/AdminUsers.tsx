/**
 * AdminUsers — Full user management interface for admins.
 *
 * Features:
 *   - List all users with role, status, last login
 *   - Create new users (via CreateUserForm)
 *   - Toggle active/inactive status (optimistic)
 *   - Change user role (optimistic)
 *   - Reset password (via PasswordResetCell)
 *   - Delete user with confirmation (cannot delete self)
 *
 * Data flow: initialUsers (server prop) → local state → server actions → optimistic update
 * Used by: /admin/users page
 */

"use client"

import { useState, useTransition, useCallback } from "react"
import { Plus, Trash2, Power, Check, X } from "lucide-react"
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { actionUpdateUser, actionDeleteUser, actionListUsers } from "@/app/actions/admin"
import { UserRoleSelector } from "./UserRoleSelector"
import { CreateUserForm } from "./CreateUserForm"
import { PasswordResetCell } from "./PasswordResetCell"

type UserRow = {
  id: number; email: string; name: string; role: string
  active: boolean; createdAt: number; lastLogin: number | null
}

type RoleInfo = { id: number; name: string; color: string; level: number }

function formatDate(ts: number | null): string {
  if (!ts) return "Nunca"
  return new Date(ts).toLocaleDateString("es-AR", { day: "2-digit", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit" })
}

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

  function flashSuccess(msg: string) {
    setSuccessMsg(msg)
    setTimeout(() => setSuccessMsg(null), 3000)
  }

  const refreshUsers = useCallback(async () => {
    try { setUsers(await actionListUsers() as UserRow[]) } catch { /* keep current state */ }
  }, [])

  function handleToggleActive(user: UserRow) {
    const newActive = !user.active
    setUsers((prev) => prev.map((u) => (u.id === user.id ? { ...u, active: newActive } : u)))
    startTransition(async () => {
      try {
        await actionUpdateUser({ id: user.id, data: { active: newActive } })
        flashSuccess(`${user.name} ${newActive ? "activado" : "desactivado"}`)
      } catch {
        setUsers((prev) => prev.map((u) => (u.id === user.id ? { ...u, active: !newActive } : u)))
        setError("Error al cambiar estado")
      }
    })
  }

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
    setUsers((prev) => prev.filter((u) => u.id !== user.id))
    startTransition(async () => {
      try {
        await actionDeleteUser({ id: user.id })
        flashSuccess(`${user.name} eliminado`)
      } catch {
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
            <Plus size={16} /> Nuevo usuario
          </button>
        </div>

        {/* Feedback */}
        {error && (
          <div className="text-sm text-destructive flex items-center justify-between"
            style={{ marginBottom: "16px", padding: "10px 14px", borderRadius: "10px", background: "color-mix(in srgb, var(--destructive) 10%, transparent)" }}>
            {error}
            <button onClick={() => setError(null)} className="text-destructive hover:opacity-70"><X size={14} /></button>
          </div>
        )}
        {successMsg && (
          <div className="text-sm text-success flex items-center"
            style={{ marginBottom: "16px", padding: "10px 14px", borderRadius: "10px", gap: "8px", background: "color-mix(in srgb, var(--success) 10%, transparent)" }}>
            <Check size={14} /> {successMsg}
          </div>
        )}

        {/* Create form */}
        {showCreate && (
          <CreateUserForm
            onCreated={(msg, fresh) => { setShowCreate(false); flashSuccess(msg); setUsers(fresh) }}
            onCancel={() => { setShowCreate(false); setError(null) }}
            onError={setError}
          />
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
                  <tr key={user.id} className={`border-t border-border transition-colors ${isSelf ? "bg-accent/5" : isSuperAdmin ? "bg-accent/3" : "hover:bg-surface/50"}`}>
                    {/* User info */}
                    <td style={{ padding: "12px 16px" }}>
                      <div className="flex items-center" style={{ gap: "8px" }}>
                        <div
                          className="flex items-center justify-center rounded-full text-accent-fg font-medium text-xs"
                          style={{ width: "32px", height: "32px", background: "var(--accent-subtle)" }}
                        >
                          {user.name.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <div className="font-medium text-fg">{user.name}</div>
                          <div className="text-xs text-fg-subtle">{user.email}</div>
                        </div>
                      </div>
                    </td>

                    {/* Role */}
                    <td style={{ padding: "12px 16px" }}>
                      <UserRoleSelector
                        userId={user.id}
                        assignedRoleIds={userRoleMap[user.id] ?? []}
                        allRoles={allRoles} disabled={isProtected || isPending}
                        onUpdate={(ids: number[]) => setUserRoleMap((prev) => ({ ...prev, [user.id]: ids }))}
                      />
                    </td>

                    {/* Status */}
                    <td style={{ padding: "12px 16px" }}>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <button
                            onClick={() => handleToggleActive(user)}
                            disabled={isProtected || isPending}
                            className={`flex items-center rounded-full text-xs font-medium transition-colors ${
                              user.active ? "text-success" : "text-fg-subtle"
                            } ${isProtected ? "cursor-not-allowed opacity-50" : "hover:opacity-80"}`}
                            style={{ padding: "3px 10px", gap: "4px", background: user.active ? "color-mix(in srgb, var(--success) 10%, transparent)" : "var(--surface-2)" }}
                          >
                            <Power size={11} />
                            {user.active ? "Activo" : "Inactivo"}
                          </button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom" sideOffset={4}>
                          {isProtected ? "Protegido" : user.active ? "Desactivar" : "Activar"}
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
                        <PasswordResetCell userId={user.id} userName={user.name} onSuccess={flashSuccess} />
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <button
                              onClick={() => handleDeleteClick(user)} disabled={isProtected}
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
