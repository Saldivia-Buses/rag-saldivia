export default function AdminUsersLoading() {
  return (
    <div className="space-y-4 p-6">
      <div className="flex items-center justify-between">
        <div className="h-6 w-32 rounded bg-surface-2 animate-pulse" />
        <div className="h-9 w-28 rounded bg-surface-2 animate-pulse" />
      </div>
      <div className="rounded-lg border border-border bg-surface">
        {/* Table header */}
        <div className="flex items-center gap-4 p-3 border-b border-border bg-surface-2/50">
          <div className="h-3 w-24 rounded bg-surface-2 animate-pulse" />
          <div className="h-3 w-32 rounded bg-surface-2 animate-pulse" />
          <div className="h-3 w-16 rounded bg-surface-2 animate-pulse ml-auto" />
        </div>
        {/* Table rows */}
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="flex items-center gap-4 p-3 border-b border-border last:border-0">
            <div className="h-8 w-8 rounded-full bg-surface-2 animate-pulse" />
            <div className="space-y-1 flex-1">
              <div className="h-3 w-28 rounded bg-surface-2 animate-pulse" />
              <div className="h-2 w-40 rounded bg-surface-2 animate-pulse" />
            </div>
            <div className="h-5 w-14 rounded-full bg-surface-2 animate-pulse" />
          </div>
        ))}
      </div>
    </div>
  )
}
