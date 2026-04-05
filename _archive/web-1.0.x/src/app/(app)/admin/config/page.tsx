import { loadRagParams } from "@rag-saldivia/config"
import { RagParamsSchema } from "@rag-saldivia/shared"
import { AdminRagConfig } from "@/components/admin/AdminRagConfig"

export default function ConfigPage() {
  const params = loadRagParams()
  const defaults = RagParamsSchema.parse({})

  return <AdminRagConfig params={params} defaults={defaults} />
}
