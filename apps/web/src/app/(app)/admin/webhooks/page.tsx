import { requireAdmin } from "@/lib/auth/current-user"
import { WebhooksAdmin } from "@/components/admin/WebhooksAdmin"

export default async function WebhooksPage() {
  await requireAdmin()
  return (
    <div className="p-6 max-w-4xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold">Webhooks salientes</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Recibí notificaciones POST con firma HMAC-SHA256 cuando ocurran eventos en el sistema.
        </p>
      </div>
      <WebhooksAdmin />
    </div>
  )
}
