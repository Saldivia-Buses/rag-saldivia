export default function ChatDetailLoading() {
  return (
    <div className="flex h-full">
      {/* Session list skeleton */}
      <div className="w-64 shrink-0 border-r border-border bg-surface flex flex-col">
        <div className="p-3 border-b border-border flex items-center justify-between">
          <div className="h-3 w-16 rounded bg-surface-2 animate-pulse" />
          <div className="h-6 w-6 rounded bg-surface-2 animate-pulse" />
        </div>
        <div className="flex-1 p-2 space-y-1">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="h-9 rounded-md bg-surface-2 animate-pulse" style={{ opacity: 1 - i * 0.1 }} />
          ))}
        </div>
      </div>
      {/* Chat area skeleton */}
      <div className="flex-1 flex flex-col">
        <div className="flex-1 p-4 space-y-4">
          <div className="h-14 w-3/4 rounded-lg bg-surface-2 animate-pulse" />
          <div className="h-14 w-2/3 rounded-lg bg-surface-2 animate-pulse ml-auto" />
          <div className="h-14 w-3/4 rounded-lg bg-surface-2 animate-pulse" />
          <div className="h-20 w-1/2 rounded-lg bg-surface-2 animate-pulse ml-auto" />
        </div>
        <div className="p-4 border-t border-border">
          <div className="h-12 w-full rounded-lg bg-surface-2 animate-pulse" />
        </div>
      </div>
    </div>
  )
}
