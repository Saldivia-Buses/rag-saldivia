import { SkeletonTable } from "@/components/ui/skeleton"

export default function CollectionsLoading() {
  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <div className="space-y-1">
        <div className="h-5 w-32 rounded bg-surface-2 animate-pulse" />
        <div className="h-3 w-24 rounded bg-surface-2 animate-pulse" />
      </div>
      <div className="rounded-lg border border-border overflow-hidden">
        <SkeletonTable rows={5} cols={2} />
      </div>
    </div>
  )
}
