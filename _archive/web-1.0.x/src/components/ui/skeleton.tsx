import { cn } from "@/lib/utils"

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-surface-2", className)}
      {...props}
    />
  )
}

function SkeletonText({ lines = 3, className }: { lines?: number; className?: string }) {
  return (
    <div className={cn("space-y-2", className)}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          className="h-3"
          style={{ width: `${i === lines - 1 ? 65 : 85 + (i % 2) * 10}%` }}
        />
      ))}
    </div>
  )
}

function SkeletonAvatar({ size = "md" }: { size?: "sm" | "md" | "lg" }) {
  const sizes = { sm: "h-6 w-6", md: "h-9 w-9", lg: "h-12 w-12" }
  return <Skeleton className={cn("rounded-full shrink-0", sizes[size])} />
}

function SkeletonCard({ className }: { className?: string }) {
  return (
    <div className={cn("rounded-lg border border-border bg-surface p-4 space-y-3", className)}>
      <div className="flex items-center gap-3">
        <SkeletonAvatar size="sm" />
        <Skeleton className="h-3 w-32" />
      </div>
      <SkeletonText lines={2} />
    </div>
  )
}

function SkeletonTable({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="space-y-0">
      {/* Header */}
      <div className="flex gap-4 px-3 py-2 border-b border-border bg-surface">
        {Array.from({ length: cols }).map((_, i) => (
          <Skeleton key={i} className="h-3" style={{ width: `${60 + i * 20}px` }} />
        ))}
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="flex gap-4 px-3 py-2.5 border-b border-border items-center"
        >
          {Array.from({ length: cols }).map((_, j) => (
            <Skeleton key={j} className="h-3" style={{ width: `${50 + j * 25}px` }} />
          ))}
        </div>
      ))}
    </div>
  )
}

export { Skeleton, SkeletonText, SkeletonAvatar, SkeletonCard, SkeletonTable }
