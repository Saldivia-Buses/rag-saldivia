import type { LucideIcon } from "lucide-react"
import { TrendingUp, TrendingDown, Minus } from "lucide-react"
import { cn } from "@/lib/utils"

interface StatCardProps {
  label: string
  value: string | number
  delta?: number
  deltaLabel?: string
  icon?: LucideIcon
  className?: string
}

export function StatCard({
  label,
  value,
  delta,
  deltaLabel,
  icon: Icon,
  className,
}: StatCardProps) {
  const isPositive = delta !== undefined && delta > 0
  const isNegative = delta !== undefined && delta < 0
  const isNeutral  = delta === 0

  return (
    <div
      className={cn(
        "rounded-xl border border-border bg-surface p-5 flex flex-col gap-3 hover:shadow-sm transition-shadow",
        className
      )}
    >
      <div className="flex items-start justify-between">
        <span className="text-xs font-semibold uppercase tracking-wide text-fg-subtle">
          {label}
        </span>
        {Icon && (
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent-subtle">
            <Icon className="h-4 w-4 text-accent" strokeWidth={1.5} />
          </div>
        )}
      </div>

      <p className="text-2xl font-semibold text-fg tabular-nums">
        {value}
      </p>

      {delta !== undefined && (
        <div className="flex items-center gap-1.5">
          <span
            className={cn(
              "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
              isPositive && "bg-success-subtle text-success",
              isNegative && "bg-destructive-subtle text-destructive",
              isNeutral  && "bg-surface-2 text-fg-muted"
            )}
          >
            {isPositive && <TrendingUp className="h-3 w-3" />}
            {isNegative && <TrendingDown className="h-3 w-3" />}
            {isNeutral  && <Minus className="h-3 w-3" />}
            {isPositive ? "+" : ""}{delta}%
          </span>
          {deltaLabel && (
            <span className="text-xs text-fg-subtle">{deltaLabel}</span>
          )}
        </div>
      )}
    </div>
  )
}
