export default function CollectionDetailLoading() {
  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="h-7 w-40 rounded bg-surface-2 animate-pulse" />
        <div className="h-5 w-16 rounded-full bg-surface-2 animate-pulse" />
      </div>
      {/* Stats */}
      <div className="grid grid-cols-3 gap-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="rounded-lg border border-border bg-surface p-4 space-y-2">
            <div className="h-3 w-16 rounded bg-surface-2 animate-pulse" />
            <div className="h-6 w-12 rounded bg-surface-2 animate-pulse" />
          </div>
        ))}
      </div>
      {/* History table */}
      <div className="rounded-lg border border-border bg-surface">
        <div className="p-4 border-b border-border">
          <div className="h-4 w-28 rounded bg-surface-2 animate-pulse" />
        </div>
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center gap-4 p-3 border-b border-border last:border-0">
            <div className="h-3 w-36 rounded bg-surface-2 animate-pulse" />
            <div className="h-3 w-20 rounded bg-surface-2 animate-pulse ml-auto" />
          </div>
        ))}
      </div>
    </div>
  )
}
