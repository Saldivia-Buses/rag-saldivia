"use client"

import { useState, useTransition } from "react"
import { actionCreateUser, actionListUsers } from "@/app/actions/admin"

type UserRow = {
  id: number; email: string; name: string; role: string
  active: boolean; createdAt: number; lastLogin: number | null
}

export function CreateUserForm({
  onCreated,
  onCancel,
  onError,
}: {
  onCreated: (msg: string, users: UserRow[]) => void
  onCancel: () => void
  onError: (msg: string) => void
}) {
  const [form, setForm] = useState({ email: "", name: "", password: "", role: "user" })
  const [isPending, startTransition] = useTransition()

  function handleCreate() {
    if (!form.email || !form.name || !form.password) { onError("Completá todos los campos"); return }
    if (form.password.length < 6) { onError("La contraseña debe tener al menos 6 caracteres"); return }

    startTransition(async () => {
      try {
        await actionCreateUser({
          email: form.email, name: form.name, password: form.password,
          role: form.role as "admin" | "area_manager" | "user",
        })
        const msg = `Usuario ${form.email} creado`
        setForm({ email: "", name: "", password: "", role: "user" })
        const fresh = await actionListUsers()
        onCreated(msg, fresh as UserRow[])
      } catch (err) {
        const msg = String(err)
        onError(msg.includes("UNIQUE") || msg.includes("unique") ? "Ya existe un usuario con ese email" : "Error al crear usuario")
      }
    })
  }

  return (
    <div className="border border-border rounded-xl bg-surface" style={{ padding: "20px", marginBottom: "20px" }}>
      <h2 className="text-sm font-semibold text-fg" style={{ marginBottom: "16px" }}>Crear usuario</h2>
      <div className="grid grid-cols-2" style={{ gap: "12px" }}>
        <input
          value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
          placeholder="Nombre completo"
          className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
          style={{ padding: "8px 12px" }}
        />
        <input
          value={form.email} onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
          placeholder="Email" type="email"
          className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
          style={{ padding: "8px 12px" }}
        />
        <input
          value={form.password} onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
          placeholder="Contraseña (mín. 6 caracteres)" type="password"
          className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
          style={{ padding: "8px 12px" }}
          onKeyDown={(e) => { if (e.key === "Enter") handleCreate() }}
        />
        <select
          value={form.role} onChange={(e) => setForm((f) => ({ ...f, role: e.target.value }))}
          className="rounded-lg border border-border bg-bg text-fg text-sm outline-none focus:border-accent transition-colors"
          style={{ padding: "8px 12px" }}
        >
          <option value="user">Usuario</option>
          <option value="area_manager">Manager</option>
          <option value="admin">Admin</option>
        </select>
      </div>
      <div className="flex justify-end" style={{ gap: "8px", marginTop: "16px" }}>
        <button onClick={onCancel}
          className="text-sm text-fg-muted hover:text-fg rounded-lg border border-border transition-colors"
          style={{ padding: "8px 16px" }}>
          Cancelar
        </button>
        <button onClick={handleCreate}
          disabled={isPending || !form.email || !form.name || !form.password}
          className="text-sm font-medium rounded-lg bg-accent text-accent-fg disabled:opacity-30 transition-opacity"
          style={{ padding: "8px 16px" }}>
          {isPending ? "Creando..." : "Crear"}
        </button>
      </div>
    </div>
  )
}
