import { requireUser } from "@/lib/auth/current-user"
import Link from "next/link"
import { ArrowLeft } from "lucide-react"
import { DocumentGraphLazy } from "@/components/collections/DocumentGraphLazy"

export default async function CollectionGraphPage({
  params,
}: {
  params: Promise<{ name: string }>
}) {
  await requireUser()
  const { name } = await params
  const collection = decodeURIComponent(name)

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <Link href="/collections" className="flex items-center gap-1.5 text-sm mb-4 hover:opacity-70 transition-opacity" style={{ color: "var(--muted-foreground)" }}>
        <ArrowLeft size={14} /> Colecciones
      </Link>
      <div className="mb-4">
        <h1 className="text-xl font-semibold">Grafo de documentos</h1>
        <p className="text-sm mt-1" style={{ color: "var(--muted-foreground)" }}>
          Similitud semántica entre documentos de <strong>{collection}</strong>. Nodos cercanos tienen contenido relacionado.
        </p>
      </div>
      <DocumentGraphLazy collection={collection} />
      <p className="text-xs mt-3" style={{ color: "var(--muted-foreground)" }}>
        Simulación de similitud. En producción los embeddings vienen de Milvus via el RAG server.
      </p>
    </div>
  )
}
