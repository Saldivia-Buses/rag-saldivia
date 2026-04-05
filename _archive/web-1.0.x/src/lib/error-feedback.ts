/**
 * Error feedback system — lets users report errors to the server
 * with optional context about what they were doing.
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

/**
 * State setter for the ErrorFeedbackDialog.
 * Components call this to open the dialog; the dialog is rendered
 * wherever <ErrorFeedbackMount /> is placed.
 */
let _feedbackSetter: ((state: { error: string; context: string } | null) => void) | null = null

export function registerFeedbackSetter(setter: typeof _feedbackSetter) {
  _feedbackSetter = setter
}

/** Open the error feedback dialog from anywhere. */
export function showErrorFeedback(error: string, context: string) {
  _feedbackSetter?.({ error, context })
}
