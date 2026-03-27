"use client"

import { useOptimistic, useState, useTransition } from "react"
import { useForm } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { z } from "zod"
import { UserPlus, Trash2, UserCheck, UserX } from "lucide-react"
import type { DbUser, DbArea } from "@rag-saldivia/db"

const CreateUserSchema = z.object({
  email: z.string().email("Email inválido"),
  name: z.string().min(2, "El nombre debe tener al menos 2 caracteres"),
  password: z.string().min(8, "La contraseña debe tener al menos 8 caracteres"),
  role: z.enum(["admin", "area_manager", "user"]),
})
import {
  actionCreateUser,
  actionDeleteUser,
  actionUpdateUser,
} from "@/app/actions/users"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import {
  Table, TableBody, TableCell, TableHead,
  TableHeader, TableRow,
} from "@/components/ui/table"
import { EmptyPlaceholder } from "@/components/ui/empty-placeholder"
import { Users } from "lucide-react"

type UserWithAreas = DbUser & {
  userAreas?: Array<{ area: DbArea }>
}

const ROLE_VARIANT: Record<string, "default" | "secondary" | "outline"> = {
  admin: "default",
  area_manager: "secondary",
  user: "outline",
}

const ROLE_LABEL: Record<string, string> = {
  admin: "Admin",
  area_manager: "Gerente",
  user: "Usuario",
}

export function UsersAdmin({
  users: initialUsers,
  areas,
}: {
  users: UserWithAreas[]
  areas: DbArea[]
}) {
  const [optimisticUsers, applyOptimistic] = useOptimistic(
    initialUsers,
    (state, action: { type: "delete"; id: number }) => {
      if (action.type === "delete") return state.filter((u) => u.id !== action.id)
      return state
    }
  )
  const [showCreate, setShowCreate] = useState(false)
  const [isPending, startTransition] = useTransition()
  const [newAreaIds, setNewAreaIds] = useState<number[]>([])
  const [formError, setFormError] = useState<string | null>(null)

  const createForm = useForm<z.infer<typeof CreateUserSchema>>({
    resolver: zodResolver(CreateUserSchema),
    defaultValues: { email: "", name: "", password: "", role: "user" },
  })

  function handleCreate(data: z.infer<typeof CreateUserSchema>) {
    setFormError(null)
    startTransition(async () => {
      try {
        await actionCreateUser({ ...data, areaIds: newAreaIds })
        setShowCreate(false)
        createForm.reset()
        setNewAreaIds([])
      } catch (err) {
        setFormError(String(err))
      }
    })
  }

  async function handleToggleActive(user: UserWithAreas) {
    startTransition(async () => { await actionUpdateUser(user.id, { active: !user.active }) })
  }

  function handleDelete(id: number) {
    if (!confirm("¿Eliminar este usuario? Esta acción no se puede deshacer.")) return
    startTransition(async () => {
      applyOptimistic({ type: "delete", id })
      await actionDeleteUser(id)
    })
  }

  return (
    <div className="p-6 space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-fg">Usuarios</h1>
          <p className="text-sm text-fg-muted mt-0.5">{optimisticUsers.length} usuario{optimisticUsers.length !== 1 ? "s" : ""} registrado{optimisticUsers.length !== 1 ? "s" : ""}</p>
        </div>
        <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
          <UserPlus className="h-3.5 w-3.5" />
          Nuevo usuario
        </Button>
      </div>

      {/* Formulario de creación */}
      {showCreate && (
        <div className="rounded-xl border border-border bg-surface p-5 space-y-4">
          <h3 className="text-sm font-semibold text-fg">Crear usuario</h3>
          <form onSubmit={createForm.handleSubmit(handleCreate)} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <Input placeholder="Nombre completo" {...createForm.register("name")} />
              <Input placeholder="Email" type="email" {...createForm.register("email")} />
              <Input placeholder="Contraseña (mín. 8 caracteres)" type="password" {...createForm.register("password")} />
              <select
                {...createForm.register("role")}
                className="h-9 w-full rounded-md border border-border bg-bg px-3 text-sm text-fg focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              >
                <option value="user">Usuario</option>
                <option value="area_manager">Gerente de área</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            {(createForm.formState.errors.email || createForm.formState.errors.name || createForm.formState.errors.password) && (
              <p className="text-xs text-destructive">
                {createForm.formState.errors.email?.message ?? createForm.formState.errors.name?.message ?? createForm.formState.errors.password?.message}
              </p>
            )}

            {areas.length > 0 && (
              <div className="space-y-1.5">
                <p className="text-xs font-medium text-fg-muted">Áreas de acceso</p>
                <div className="flex flex-wrap gap-3">
                  {areas.map((area) => (
                    <label key={area.id} className="flex items-center gap-1.5 text-sm text-fg cursor-pointer">
                      <input
                        type="checkbox"
                        checked={newAreaIds.includes(area.id)}
                        onChange={(e) => setNewAreaIds(e.target.checked ? [...newAreaIds, area.id] : newAreaIds.filter((id) => id !== area.id))}
                        className="accent-accent"
                      />
                      {area.name}
                    </label>
                  ))}
                </div>
              </div>
            )}

            {formError && (
              <p className="text-sm text-destructive">{formError}</p>
            )}

            <div className="flex gap-2 pt-1">
              <Button type="submit" size="sm" disabled={isPending}>
                {isPending ? "Creando..." : "Crear usuario"}
              </Button>
              <Button type="button" size="sm" variant="outline" onClick={() => setShowCreate(false)}>
                Cancelar
              </Button>
            </div>
          </form>
        </div>
      )}

      {/* Tabla de usuarios */}
      {optimisticUsers.length === 0 ? (
        <EmptyPlaceholder>
          <EmptyPlaceholder.Icon icon={Users} />
          <EmptyPlaceholder.Title>Sin usuarios</EmptyPlaceholder.Title>
          <EmptyPlaceholder.Description>Creá el primer usuario del sistema.</EmptyPlaceholder.Description>
        </EmptyPlaceholder>
      ) : (
        <div className="rounded-xl border border-border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Rol</TableHead>
                <TableHead>Áreas</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {optimisticUsers.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="font-medium text-fg">{user.name}</TableCell>
                  <TableCell className="text-fg-muted">{user.email}</TableCell>
                  <TableCell>
                    <Badge variant={ROLE_VARIANT[user.role] ?? "outline"}>
                      {ROLE_LABEL[user.role] ?? user.role}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-fg-subtle text-xs">
                    {user.userAreas?.map((ua) => ua.area.name).join(", ") || "—"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={user.active ? "success" : "destructive"}>
                      {user.active ? "Activo" : "Inactivo"}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1 justify-end">
                      <Button
                        variant="ghost" size="icon"
                        className="h-7 w-7"
                        title={user.active ? "Desactivar" : "Activar"}
                        onClick={() => handleToggleActive(user)}
                        disabled={isPending}
                      >
                        {user.active ? <UserX className="h-3.5 w-3.5" /> : <UserCheck className="h-3.5 w-3.5" />}
                      </Button>
                      <Button
                        variant="ghost" size="icon"
                        className="h-7 w-7 text-destructive hover:text-destructive"
                        title="Eliminar"
                        onClick={() => handleDelete(user.id)}
                        disabled={isPending}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
