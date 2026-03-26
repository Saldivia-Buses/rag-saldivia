import type { LucideIcon } from "lucide-react"
import { cn } from "@/lib/utils"

interface EmptyPlaceholderProps {
  className?: string
  children: React.ReactNode
}

function EmptyPlaceholder({ className, children }: EmptyPlaceholderProps) {
  return (
    <div
      className={cn(
        "flex min-h-[320px] flex-col items-center justify-center gap-4 rounded-xl border border-dashed border-border bg-surface/50 px-8 py-12 text-center",
        className
      )}
    >
      {children}
    </div>
  )
}

function EmptyPlaceholderIcon({ icon: Icon }: { icon: LucideIcon }) {
  return (
    <div className="flex h-14 w-14 items-center justify-center rounded-full bg-accent-subtle">
      <Icon className="h-7 w-7 text-accent" strokeWidth={1.5} />
    </div>
  )
}

function EmptyPlaceholderTitle({
  className,
  children,
}: {
  className?: string
  children: React.ReactNode
}) {
  return (
    <h3 className={cn("text-base font-semibold text-fg", className)}>
      {children}
    </h3>
  )
}

function EmptyPlaceholderDescription({
  className,
  children,
}: {
  className?: string
  children: React.ReactNode
}) {
  return (
    <p className={cn("max-w-xs text-sm text-fg-muted leading-relaxed", className)}>
      {children}
    </p>
  )
}

EmptyPlaceholder.Icon = EmptyPlaceholderIcon
EmptyPlaceholder.Title = EmptyPlaceholderTitle
EmptyPlaceholder.Description = EmptyPlaceholderDescription

export { EmptyPlaceholder }
