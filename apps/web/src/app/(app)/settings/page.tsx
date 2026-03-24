import { requireUser } from "@/lib/auth/current-user"
import { getUserById } from "@rag-saldivia/db"
import { SettingsClient } from "@/components/settings/SettingsClient"

export default async function SettingsPage() {
  const current = await requireUser()
  const user = await getUserById(current.id)
  if (!user) return null

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Configuración</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Perfil, contraseña y preferencias
        </p>
      </div>
      <SettingsClient user={user} />
    </div>
  )
}
