import { requireUser } from "@/lib/auth/current-user"
import { redirect } from "next/navigation"
import { queryEvents } from "@rag-saldivia/db"
import { AuditTable } from "@/components/audit/AuditTable"

export default async function AuditPage() {
  const user = await requireUser()

  if (user.role === "user") redirect("/")

  // Admins ven todos los eventos, area_managers solo los suyos
  const events = await queryEvents({
    userId: user.role === "admin" ? undefined : user.id,
    limit: 100,
    order: "desc",
  })

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit Log</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Registro de eventos del sistema — Black Box
        </p>
      </div>
      <AuditTable events={events} isAdmin={user.role === "admin"} />
    </div>
  )
}
