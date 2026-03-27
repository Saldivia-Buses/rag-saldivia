export default function ChatLoading() {
  return (
    <div className="flex h-full">
      {/* Skeleton del panel de sesiones */}
      <div className="w-64 shrink-0 border-r border-border bg-surface flex flex-col">
        <div className="p-3 border-b border-border flex items-center justify-between">
          <div className="h-3 w-16 rounded bg-surface-2 animate-pulse" />
          <div className="h-6 w-6 rounded bg-surface-2 animate-pulse" />
        </div>
        <div className="flex-1 p-2 space-y-1">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="h-9 rounded-md bg-surface-2 animate-pulse" style={{ opacity: 1 - i * 0.08 }} />
          ))}
        </div>
      </div>
      {/* Área central vacía */}
      <div className="flex-1 flex items-center justify-center">
        <div className="space-y-3 text-center">
          <div className="h-14 w-14 rounded-full bg-surface-2 animate-pulse mx-auto" />
          <div className="h-3 w-48 rounded bg-surface-2 animate-pulse mx-auto" />
          <div className="h-3 w-36 rounded bg-surface-2 animate-pulse mx-auto" />
        </div>
      </div>
    </div>
  )
}
