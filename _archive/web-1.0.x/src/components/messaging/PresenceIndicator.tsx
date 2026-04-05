/**
 * Presence dot — green (online), yellow (away), gray (offline).
 */
"use client"

import { cn } from "@/lib/utils"

const STATUS_COLORS = {
  online: "bg-success",
  away: "bg-warning",
  offline: "bg-fg-subtle/40",
} as const

export type PresenceStatus = "online" | "away" | "offline"

export function PresenceIndicator({
  status,
  className,
  size = "sm",
}: {
  status: PresenceStatus
  className?: string
  size?: "sm" | "md"
}) {
  const dims = size === "sm" ? "h-2.5 w-2.5" : "h-3 w-3"

  return (
    <span
      className={cn(
        "inline-block rounded-full border-2 border-bg shrink-0",
        dims,
        STATUS_COLORS[status],
        className,
      )}
      title={status === "online" ? "En línea" : status === "away" ? "Ausente" : "Desconectado"}
      aria-label={status === "online" ? "En línea" : status === "away" ? "Ausente" : "Desconectado"}
    />
  )
}
