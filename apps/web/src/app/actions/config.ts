"use server"

import { z } from "zod"
import { adminAction } from "@/lib/safe-action"
import { saveRagParams } from "@rag-saldivia/config"
import { log } from "@rag-saldivia/logger/backend"

export const actionUpdateRagParams = adminAction
  .schema(z.record(z.string(), z.unknown()))
  .action(async ({ parsedInput: params, ctx: { user } }) => {
    await saveRagParams(params)
    log.info("admin.config_changed", { params }, { userId: user.id })
  })

export const actionResetRagParams = adminAction
  .action(async ({ ctx: { user } }) => {
    const { unlink } = await import("fs/promises")
    const { join } = await import("path")
    const overridesPath = join(process.cwd(), "config", "admin-overrides.yaml")
    try {
      await unlink(overridesPath)
    } catch { /* ya no existía */ }
    log.info("admin.config_changed", { action: "reset" }, { userId: user.id })
  })
