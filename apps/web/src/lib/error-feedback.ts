/**
 * Error feedback system — lets users report errors to the server
 * with optional context about what they were doing.
 *
 * Usage:
 *   import { reportError } from "@/lib/error-feedback"
 *   reportError({ error: "Something failed", context: "Creating a role", comment: "I clicked save" })
 */

export type ErrorReport = {
  error: string
  context: string
  comment?: string | undefined
}

export async function reportError(report: ErrorReport): Promise<boolean> {
  try {
    const res = await fetch("/api/feedback", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(report),
    })
    return res.ok
  } catch {
    return false
  }
}
