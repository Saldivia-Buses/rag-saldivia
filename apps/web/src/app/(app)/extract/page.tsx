import { requireUser } from "@/lib/auth/current-user"
import { ExtractionWizard } from "@/components/extract/ExtractionWizard"
import { Table2 } from "lucide-react"

export default async function ExtractPage() {
  await requireUser()

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-6">
        <h1 className="text-xl font-semibold flex items-center gap-2">
          <Table2 size={20} style={{ color: "var(--accent)" }} />
          Extracción estructurada
        </h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Definí campos y el sistema extrae esos datos de todos los documentos de una colección.
          Exportable como CSV.
        </p>
      </div>
      <ExtractionWizard />
    </div>
  )
}
