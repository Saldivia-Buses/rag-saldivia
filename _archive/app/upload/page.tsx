import { requireUser } from "@/lib/auth/current-user"
import { getUserCollections } from "@rag-saldivia/db"
import { UploadClient } from "@/components/upload/UploadClient"

export default async function UploadPage() {
  const user = await requireUser()
  const collections = await getUserCollections(user.id)
  const writableCollections = collections
    .filter((c) => c.permission === "write" || c.permission === "admin" || user.role === "admin")
    .map((c) => c.name)

  return (
    <div className="p-6 max-w-2xl mx-auto space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Subir documentos</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          PDF, Word o texto plano — máximo 100MB por archivo
        </p>
      </div>
      <UploadClient collections={writableCollections} />
    </div>
  )
}
