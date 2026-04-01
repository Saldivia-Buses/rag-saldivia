"use client"

import { ErrorRecoveryFromError } from "@/components/ui/error-recovery"

export default function CollectionsError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="flex items-center justify-center h-full p-8">
      <ErrorRecoveryFromError error={error} variant="page" reset={reset} />
    </div>
  )
}
