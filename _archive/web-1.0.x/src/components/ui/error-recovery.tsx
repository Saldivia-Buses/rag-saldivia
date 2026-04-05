"use client"

import { useRouter } from "next/navigation"
import { AlertTriangle, Clock, ShieldX, Zap, ServerCrash, LogIn, Search, HelpCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { getErrorRecovery, parseUseChatError, type UserErrorRecovery, type ErrorAction } from "@/lib/error-recovery"
import { showErrorFeedback } from "@/lib/error-feedback"

const ICONS = {
  unavailable: ServerCrash,
  timeout: Clock,
  forbidden: ShieldX,
  "rate-limit": Zap,
  upstream: AlertTriangle,
  auth: LogIn,
  "not-found": Search,
  generic: HelpCircle,
} as const

type ErrorRecoveryProps = {
  recovery: UserErrorRecovery
  variant?: "inline" | "page" | undefined
  onRetry?: (() => void) | undefined
  onDismiss?: (() => void) | undefined
  className?: string | undefined
}

export function ErrorRecovery({
  recovery,
  variant = "page",
  onRetry,
  onDismiss,
  className = "",
}: ErrorRecoveryProps) {
  const router = useRouter()
  const Icon = ICONS[recovery.icon]

  function handleAction(action: ErrorAction) {
    switch (action.type) {
      case "retry":
        onRetry?.()
        break
      case "navigate":
        if (action.href) router.push(action.href)
        break
      case "dismiss":
        onDismiss?.()
        break
      case "report":
        showErrorFeedback(
          recovery.description,
          `${recovery.title}: ${recovery.suggestion}`,
        )
        break
    }
  }

  if (variant === "inline") {
    return (
      <div
        className={`text-sm ${className}`}
        style={{
          padding: "12px 16px",
          borderRadius: "12px",
          background: "var(--destructive-subtle, color-mix(in srgb, var(--destructive) 8%, transparent))",
          border: "1px solid color-mix(in srgb, var(--destructive) 20%, transparent)",
        }}
      >
        <div style={{ display: "flex", alignItems: "flex-start", gap: "10px" }}>
          <Icon className="text-destructive shrink-0" size={18} style={{ marginTop: "1px" }} />
          <div style={{ flex: 1, minWidth: 0 }}>
            <p className="font-medium text-fg" style={{ marginBottom: "2px" }}>{recovery.title}</p>
            <p className="text-fg-muted" style={{ marginBottom: "4px" }}>{recovery.description}</p>
            <p className="text-fg-subtle text-xs">{recovery.suggestion}</p>
          </div>
        </div>
        {recovery.actions.length > 0 && (
          <div style={{ display: "flex", gap: "8px", marginTop: "10px", paddingLeft: "28px" }}>
            {recovery.actions.map((action) => (
              <Button
                key={action.label}
                variant={action.type === "retry" ? "default" : "outline"}
                size="sm"
                onClick={() => handleAction(action)}
              >
                {action.label}
              </Button>
            ))}
          </div>
        )}
      </div>
    )
  }

  // variant === "page"
  return (
    <div
      className={`flex flex-col items-center justify-center text-center ${className}`}
      style={{ maxWidth: "420px", padding: "32px" }}
    >
      <div
        className="text-destructive"
        style={{
          width: "48px",
          height: "48px",
          borderRadius: "12px",
          background: "var(--destructive-subtle, color-mix(in srgb, var(--destructive) 8%, transparent))",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          marginBottom: "16px",
        }}
      >
        <Icon size={24} />
      </div>
      <h2 className="text-lg font-semibold text-fg" style={{ marginBottom: "8px" }}>{recovery.title}</h2>
      <p className="text-fg-muted text-sm" style={{ marginBottom: "4px" }}>{recovery.description}</p>
      <p className="text-fg-subtle text-xs" style={{ marginBottom: "20px" }}>{recovery.suggestion}</p>
      {recovery.actions.length > 0 && (
        <div style={{ display: "flex", gap: "8px", flexWrap: "wrap", justifyContent: "center" }}>
          {recovery.actions.map((action) => (
            <Button
              key={action.label}
              variant={action.type === "retry" ? "default" : "outline"}
              size="sm"
              onClick={() => handleAction(action)}
            >
              {action.label}
            </Button>
          ))}
        </div>
      )}
    </div>
  )
}

/**
 * Convenience wrapper that takes a raw Error and produces ErrorRecovery.
 * Use in error.tsx pages and ErrorBoundary where you get a plain Error.
 */
export function ErrorRecoveryFromError({
  error,
  variant = "page",
  onRetry,
  onDismiss,
  reset,
}: {
  error: Error & { digest?: string; status?: number; code?: string }
  variant?: "inline" | "page" | undefined
  onRetry?: (() => void) | undefined
  onDismiss?: (() => void) | undefined
  reset?: (() => void) | undefined
}) {
  const recovery = getErrorRecovery(parseUseChatError(error))
  const retryFn = reset ?? onRetry
  return (
    <ErrorRecovery
      recovery={recovery}
      variant={variant}
      {...(retryFn ? { onRetry: retryFn } : {})}
      {...(onDismiss ? { onDismiss } : {})}
    />
  )
}
