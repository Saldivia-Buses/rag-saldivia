import { SkeletonTable } from "@/components/ui/skeleton"

export default function UsersLoading() {
  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div className="h-5 w-24 rounded bg-surface-2 animate-pulse" />
        <div className="h-8 w-32 rounded bg-surface-2 animate-pulse" />
      </div>
      <div className="rounded-lg border border-border overflow-hidden">
        <SkeletonTable rows={8} cols={4} />
      </div>
    </div>
  )
}
