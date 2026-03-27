"use client"

import dynamic from "next/dynamic"
import { Skeleton } from "@/components/ui/skeleton"

const DocumentGraph = dynamic(
  () => import("@/components/collections/DocumentGraph").then((m) => ({ default: m.DocumentGraph })),
  { ssr: false, loading: () => <Skeleton className="h-96 w-full rounded-xl" /> }
)

export function DocumentGraphLazy({ collection }: { collection: string }) {
  return <DocumentGraph collection={collection} />
}
