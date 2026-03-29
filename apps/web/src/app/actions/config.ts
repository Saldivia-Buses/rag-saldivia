"use server"

import { requireAdmin } from "@/lib/auth/current-user"
import { saveRagParams } from "@rag-saldivia/config"
import { log } from "@rag-saldivia/logger/backend"
import type { RagParams } from "@rag-saldivia/shared"

export async function actionUpdateRagParams(params: Partial<RagParams>) {
  const admin = await requireAdmin()
  await saveRagParams(params)
  log.info("admin.config_changed", { params }, { userId: admin.id })
}

export async function actionResetRagParams() {
  const admin = await requireAdmin()
  // Eliminar overrides (vuelve a los defaults)
  const { unlink } = await import("fs/promises")
  const { join } = await import("path")
  const overridesPath = join(process.cwd(), "config", "admin-overrides.yaml")
  try {
    await unlink(overridesPath)
  } catch { /* ya no existía */ }
  log.info("admin.config_changed", { action: "reset" }, { userId: admin.id })
}
