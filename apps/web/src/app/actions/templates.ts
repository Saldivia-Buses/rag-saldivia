/**
 * Server actions for prompt templates.
 *
 * Templates are pre-defined prompts that appear in the chat empty state,
 * replacing hardcoded suggestion chips. Users click a template to start
 * a conversation with a pre-filled prompt.
 *
 * Data flow: DB (promptTemplates table) → server action → ChatInterface
 * Depends on: @rag-saldivia/db (templates queries), lib/auth/current-user.ts
 */

"use server"

import { listActiveTemplates, createTemplate, deleteTemplate } from "@rag-saldivia/db"
import { requireUser, requireAdmin } from "@/lib/auth/current-user"

/** List all active prompt templates. Any authenticated user. */
export async function actionListTemplates() {
  await requireUser()
  return listActiveTemplates()
}

/** Create a new prompt template. Admin only. */
export async function actionCreateTemplate(data: { title: string; prompt: string; focusMode?: string }) {
  const user = await requireAdmin()
  return createTemplate({
    title: data.title,
    prompt: data.prompt,
    focusMode: data.focusMode ?? "detallado",
    createdBy: user.id,
  })
}

/** Delete a prompt template by ID. Admin only. */
export async function actionDeleteTemplate(id: number) {
  await requireAdmin()
  await deleteTemplate(id)
}
