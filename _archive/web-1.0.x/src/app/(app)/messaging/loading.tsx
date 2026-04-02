export default function MessagingLoading() {
  return (
    <div className="flex h-full">
      {/* Channel list skeleton */}
      <div className="w-64 shrink-0 border-r border-border bg-surface flex flex-col">
        <div className="p-3 border-b border-border flex items-center justify-between">
          <div className="h-3 w-20 rounded bg-surface-2 animate-pulse" />
          <div className="h-6 w-6 rounded bg-surface-2 animate-pulse" />
        </div>
        <div className="flex-1 p-2 flex flex-col gap-1">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-9 rounded-md bg-surface-2 animate-pulse" style={{ opacity: 1 - i * 0.1 }} />
          ))}
        </div>
      </div>
      {/* Message area skeleton */}
      <div className="flex-1 flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <div className="h-10 w-10 rounded-full bg-surface-2 animate-pulse" />
          <div className="h-3 w-40 rounded bg-surface-2 animate-pulse" />
        </div>
      </div>
    </div>
  )
}
