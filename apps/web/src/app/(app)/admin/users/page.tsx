import { requireAdmin } from "@/lib/auth/current-user"
import { listUsers, listAreas } from "@rag-saldivia/db"
import { UsersAdmin } from "@/components/admin/UsersAdmin"

export default async function AdminUsersPage() {
  await requireAdmin()
  const [users, areas] = await Promise.all([listUsers(), listAreas()])

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Usuarios</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Gestión de usuarios del sistema
        </p>
      </div>
      <UsersAdmin users={users} areas={areas} />
    </div>
  )
}
