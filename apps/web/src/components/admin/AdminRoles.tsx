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

import { useState, useEffect, useTransition, useCallback } from "react"
import { Plus, Check, X } from "lucide-react"
import { TooltipProvider } from "@/components/ui/tooltip"
import { ConfirmDialog } from "@/components/ui/confirm-dialog"
import { PermissionMatrix } from "./PermissionMatrix"
import { RoleCard, type RoleWithCount } from "./RoleCard"
import { RoleForm, EMPTY_FORM, type RoleFormData } from "./RoleForm"
import {
  actionCreateRole,
  actionUpdateRole,
  actionDeleteRole,
  actionListRoles,
  actionGetRolePermissions,
} from "@/app/actions/roles"

type Permission = {
  id: number
  key: string
  label: string
  category: string
  description: string
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
  const [deleteTarget, setDeleteTarget] = useState<RoleWithCount | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [editingRoleId, setEditingRoleId] = useState<number | null>(null)
  const [form, setForm] = useState<RoleFormData>(EMPTY_FORM)

  const selectedRole = roles.find((r) => r.id === selectedRoleId)
  const sortedRoles = [...roles].sort((a, b) => b.level - a.level)
  const systemRoles = sortedRoles.filter((r) => r.isSystem)
  const customRoles = sortedRoles.filter((r) => !r.isSystem)

  function flashSuccess(msg: string) {
    setSuccessMsg(msg)
    setTimeout(() => setSuccessMsg(null), 3000)
  }

  async function loadRolePermissions(roleId: number) {
    try {
      const result = await actionGetRolePermissions({ roleId })
      setActiveKeys(result?.data ?? [])
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
    if (selectedRoleId) loadRolePermissions(selectedRoleId)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  function handleCreate() {
    if (!form.name.trim()) { setError("El nombre es obligatorio"); return }
    const level = parseInt(form.level, 10)
    if (isNaN(level) || level < 1 || level > 99) { setError("El nivel debe estar entre 1 y 99"); return }
    setError(null)

    startTransition(async () => {
      try {
        await actionCreateRole({ name: form.name.trim(), description: form.description.trim(), level, color: form.color, icon: form.icon })
        setForm(EMPTY_FORM); setShowCreate(false); flashSuccess(`Rol "${form.name}" creado`)
        setRoles(await actionListRoles() as RoleWithCount[])
      } catch (err) {
        const msg = String(err)
        setError(msg.includes("UNIQUE") || msg.includes("unique") ? "Ya existe un rol con ese nombre" : "Error al crear rol")
      }
    })
  }

  function startEditing(role: RoleWithCount) {
    setEditingRoleId(role.id)
    setForm({ name: role.name, description: role.description, level: String(role.level), color: role.color, icon: role.icon })
    setShowCreate(false)
  }

  function handleSaveEdit() {
    if (!editingRoleId || !form.name.trim()) return
    const level = parseInt(form.level, 10)
    if (isNaN(level) || level < 1 || level > 99) { setError("El nivel debe estar entre 1 y 99"); return }
    setError(null)

    startTransition(async () => {
      try {
        await actionUpdateRole({ id: editingRoleId, data: { name: form.name.trim(), description: form.description.trim(), level, color: form.color, icon: form.icon } })
        setEditingRoleId(null); setForm(EMPTY_FORM); flashSuccess("Rol actualizado")
        setRoles(await actionListRoles() as RoleWithCount[])
      } catch (err) {
        setError(String(err).includes("UNIQUE") ? "Ya existe un rol con ese nombre" : "Error al actualizar")
      }
    })
  }

  function handleDeleteClick(role: RoleWithCount) {
    if (role.isSystem) return
    if (role.userCount > 0) {
      setError(`No se puede eliminar "${role.name}" porque tiene ${role.userCount} usuario(s) asignado(s)`)
      setTimeout(() => setError(null), 4000)
      return
    }
    setDeleteTarget(role)
  }

  const confirmDelete = useCallback(() => {
    if (!deleteTarget) return
    const role = deleteTarget
    setDeleteTarget(null)
    startTransition(async () => {
      try {
        await actionDeleteRole({ id: role.id })
        flashSuccess(`Rol "${role.name}" eliminado`)
        setRoles(await actionListRoles() as RoleWithCount[])
        if (selectedRoleId === role.id) { setSelectedRoleId(null); setKeysLoaded(false) }
      } catch { setError("Error al eliminar rol") }
    })
  }, [deleteTarget, startTransition, selectedRoleId])

  function cancelForm() {
    setShowCreate(false); setEditingRoleId(null); setError(null); setForm(EMPTY_FORM)
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
            <Plus size={16} /> Nuevo rol
          </button>
        </div>

        {/* Feedback */}
        {error && (
          <div className="text-sm text-destructive flex items-center justify-between rounded-xl"
            style={{ marginBottom: "16px", padding: "10px 14px", background: "color-mix(in srgb, var(--destructive) 8%, var(--surface))" }}>
            {error}
            <button onClick={() => setError(null)} className="text-destructive hover:opacity-70"><X size={14} /></button>
          </div>
        )}
        {successMsg && (
          <div className="text-sm text-success flex items-center rounded-xl"
            style={{ marginBottom: "16px", padding: "10px 14px", gap: "8px", background: "color-mix(in srgb, var(--success) 8%, var(--surface))" }}>
            <Check size={14} /> {successMsg}
          </div>
        )}

        {/* Create/Edit form */}
        {(showCreate || editingRoleId) && (
          <RoleForm
            form={form} onChange={setForm}
            onSubmit={editingRoleId ? handleSaveEdit : handleCreate}
            onCancel={cancelForm}
            isEditing={!!editingRoleId} isPending={isPending}
          />
        )}

        {/* Role cards */}
        <div className="grid grid-cols-1 md:grid-cols-3" style={{ gap: "14px", marginBottom: customRoles.length > 0 ? "14px" : "28px" }}>
          {systemRoles.map((role) => (
            <RoleCard key={role.id} role={role} isSelected={selectedRoleId === role.id}
              onSelect={() => selectRole(role.id)} onEdit={() => startEditing(role)} onDelete={() => handleDeleteClick(role)} />
          ))}
        </div>
        {customRoles.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-3" style={{ gap: "14px", marginBottom: "28px" }}>
            {customRoles.map((role) => (
              <RoleCard key={role.id} role={role} isSelected={selectedRoleId === role.id}
                onSelect={() => selectRole(role.id)} onEdit={() => startEditing(role)} onDelete={() => handleDeleteClick(role)} />
            ))}
          </div>
        )}

        {/* Permission matrix */}
        {selectedRole && keysLoaded && (
          <PermissionMatrix
            roleId={selectedRole.id} roleName={selectedRole.name}
            isSystem={selectedRole.isSystem} locked={selectedRole.isSystem && selectedRole.level >= 100}
            allPermissions={initialPermissions} activeKeys={activeKeys} onUpdate={setActiveKeys}
          />
        )}
      </div>

      <ConfirmDialog
        open={deleteTarget !== null}
        onOpenChange={(o) => { if (!o) setDeleteTarget(null) }}
        title={`¿Eliminar el rol "${deleteTarget?.name}"?`}
        description="Esta acción no se puede deshacer."
        onConfirm={confirmDelete}
      />
    </TooltipProvider>
  )
}
