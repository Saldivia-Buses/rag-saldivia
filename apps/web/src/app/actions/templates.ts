/**
 * Server actions for prompt templates.
 */

"use server"

import { z } from "zod"
import { authAction, adminAction } from "@/lib/safe-action"
import { listActiveTemplates, createTemplate, deleteTemplate } from "@rag-saldivia/db"

export const actionListTemplates = authAction
  .action(async () => {
    return listActiveTemplates()
  })

export const actionCreateTemplate = adminAction
  .schema(z.object({
    title: z.string().min(1),
    prompt: z.string().min(1),
    focusMode: z.string().optional(),
  }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    return createTemplate({
      title: data.title,
      prompt: data.prompt,
      focusMode: data.focusMode ?? "detallado",
      createdBy: user.id,
    })
  })

export const actionDeleteTemplate = adminAction
  .schema(z.object({ id: z.number() }))
  .action(async ({ parsedInput: { id } }) => {
    await deleteTemplate(id)
  })
