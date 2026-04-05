export default function AdminLoading() {
  return (
    <div className="space-y-6 p-6">
      {/* Stat cards skeleton */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-lg border border-border bg-surface p-4 space-y-2">
            <div className="h-3 w-20 rounded bg-surface-2 animate-pulse" />
            <div className="h-8 w-16 rounded bg-surface-2 animate-pulse" />
          </div>
        ))}
      </div>
      {/* Table skeleton */}
      <div className="rounded-lg border border-border bg-surface">
        <div className="p-4 border-b border-border">
          <div className="h-4 w-32 rounded bg-surface-2 animate-pulse" />
        </div>
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center gap-4 p-4 border-b border-border last:border-0">
            <div className="h-8 w-8 rounded-full bg-surface-2 animate-pulse" />
            <div className="h-3 w-32 rounded bg-surface-2 animate-pulse" />
            <div className="h-3 w-24 rounded bg-surface-2 animate-pulse ml-auto" />
          </div>
        ))}
      </div>
    </div>
  )
}
