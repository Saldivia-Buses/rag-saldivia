"use client"

import { useState, useTransition } from "react"
import { UserPlus, Trash2, UserCheck, UserX } from "lucide-react"
import type { DbUser, DbArea } from "@rag-saldivia/db"
import {
  actionCreateUser,
  actionDeleteUser,
  actionUpdateUser,
} from "@/app/actions/users"

type UserWithAreas = DbUser & {
  userAreas?: Array<{ area: DbArea }>
}

export function UsersAdmin({
  users: initialUsers,
  areas,
}: {
  users: UserWithAreas[]
  areas: DbArea[]
}) {
  const [users, setUsers] = useState(initialUsers)
  const [showCreate, setShowCreate] = useState(false)
  const [isPending, startTransition] = useTransition()

  // Form state
  const [newEmail, setNewEmail] = useState("")
  const [newName, setNewName] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [newRole, setNewRole] = useState<"admin" | "area_manager" | "user">("user")
  const [newAreaIds, setNewAreaIds] = useState<number[]>([])
  const [formError, setFormError] = useState<string | null>(null)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)
    startTransition(async () => {
      try {
        await actionCreateUser({
          email: newEmail,
          name: newName,
          password: newPassword,
          role: newRole,
          areaIds: newAreaIds,
        })
        setShowCreate(false)
        setNewEmail(""); setNewName(""); setNewPassword("")
        setNewAreaIds([])
      } catch (err) {
        setFormError(String(err))
      }
    })
  }

  async function handleToggleActive(user: UserWithAreas) {
    startTransition(async () => {
      await actionUpdateUser(user.id, { active: !user.active })
    })
  }

  async function handleDelete(id: number) {
    if (!confirm("¿Eliminar este usuario? Esta acción no se puede deshacer.")) return
    startTransition(async () => {
      await actionDeleteUser(id)
    })
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex justify-between items-center">
        <span className="text-sm" style={{ color: "var(--muted-foreground)" }}>
          {users.length} usuario(s)
        </span>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-medium"
          style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
        >
          <UserPlus size={15} />
          Nuevo usuario
        </button>
      </div>

      {/* Create form */}
      {showCreate && (
        <div className="p-4 rounded-lg border space-y-4" style={{ borderColor: "var(--border)" }}>
          <h3 className="font-medium text-sm">Crear usuario</h3>
          <form onSubmit={handleCreate} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <input
                placeholder="Nombre"
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                required
                className="px-3 py-2 rounded-md border text-sm"
                style={{ borderColor: "var(--border)" }}
              />
              <input
                placeholder="Email"
                type="email"
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
                required
                className="px-3 py-2 rounded-md border text-sm"
                style={{ borderColor: "var(--border)" }}
              />
              <input
                placeholder="Contraseña"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
                minLength={8}
                className="px-3 py-2 rounded-md border text-sm"
                style={{ borderColor: "var(--border)" }}
              />
              <select
                value={newRole}
                onChange={(e) => setNewRole(e.target.value as typeof newRole)}
                className="px-3 py-2 rounded-md border text-sm"
                style={{ borderColor: "var(--border)", background: "var(--background)" }}
              >
                <option value="user">Usuario</option>
                <option value="area_manager">Gerente de área</option>
                <option value="admin">Admin</option>
              </select>
            </div>

            {/* Areas */}
            <div>
              <p className="text-xs mb-2" style={{ color: "var(--muted-foreground)" }}>Áreas</p>
              <div className="flex flex-wrap gap-2">
                {areas.map((area) => (
                  <label key={area.id} className="flex items-center gap-1.5 text-sm">
                    <input
                      type="checkbox"
                      checked={newAreaIds.includes(area.id)}
                      onChange={(e) => {
                        setNewAreaIds(
                          e.target.checked
                            ? [...newAreaIds, area.id]
                            : newAreaIds.filter((id) => id !== area.id)
                        )
                      }}
                    />
                    {area.name}
                  </label>
                ))}
              </div>
            </div>

            {formError && (
              <p className="text-sm" style={{ color: "var(--destructive)" }}>{formError}</p>
            )}

            <div className="flex gap-2">
              <button
                type="submit"
                disabled={isPending}
                className="px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
                style={{ background: "var(--primary)", color: "var(--primary-foreground)" }}
              >
                {isPending ? "Creando..." : "Crear"}
              </button>
              <button
                type="button"
                onClick={() => setShowCreate(false)}
                className="px-4 py-2 rounded-md text-sm border"
                style={{ borderColor: "var(--border)" }}
              >
                Cancelar
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Users table */}
      <div className="rounded-lg border overflow-hidden" style={{ borderColor: "var(--border)" }}>
        <table className="w-full text-sm">
          <thead>
            <tr style={{ background: "var(--muted)" }}>
              <th className="text-left px-4 py-3 font-medium">Nombre</th>
              <th className="text-left px-4 py-3 font-medium">Email</th>
              <th className="text-left px-4 py-3 font-medium">Rol</th>
              <th className="text-left px-4 py-3 font-medium">Áreas</th>
              <th className="text-left px-4 py-3 font-medium">Estado</th>
              <th className="px-4 py-3" />
            </tr>
          </thead>
          <tbody>
            {users.map((user, i) => (
              <tr
                key={user.id}
                style={{
                  borderTop: i > 0 ? `1px solid var(--border)` : undefined,
                }}
              >
                <td className="px-4 py-3 font-medium">{user.name}</td>
                <td className="px-4 py-3" style={{ color: "var(--muted-foreground)" }}>{user.email}</td>
                <td className="px-4 py-3">
                  <span className="px-2 py-0.5 rounded text-xs" style={{ background: "var(--accent)" }}>
                    {user.role}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs" style={{ color: "var(--muted-foreground)" }}>
                  {user.userAreas?.map((ua) => ua.area.name).join(", ") || "—"}
                </td>
                <td className="px-4 py-3">
                  <span
                    className="px-2 py-0.5 rounded text-xs"
                    style={{
                      background: user.active ? "#dcfce7" : "#fee2e2",
                      color: user.active ? "#166534" : "#991b1b",
                    }}
                  >
                    {user.active ? "Activo" : "Inactivo"}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-1 justify-end">
                    <button
                      onClick={() => handleToggleActive(user)}
                      disabled={isPending}
                      title={user.active ? "Desactivar" : "Activar"}
                      className="p-1.5 rounded hover:opacity-80 disabled:opacity-40"
                    >
                      {user.active ? <UserX size={14} /> : <UserCheck size={14} />}
                    </button>
                    <button
                      onClick={() => handleDelete(user.id)}
                      disabled={isPending}
                      title="Eliminar"
                      className="p-1.5 rounded hover:opacity-80 disabled:opacity-40"
                      style={{ color: "var(--destructive)" }}
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
