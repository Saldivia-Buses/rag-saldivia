/**
 * Política de retención de eventos.
 *
 * La tabla events crece indefinidamente sin esta función.
 * Se llama desde el worker de ingesta una vez al día.
 *
 * Variable de entorno: LOG_RETENTION_DAYS (default 90).
 */

import { lt } from "drizzle-orm"
import { getDb } from "../connection"
import { events } from "../schema"

export async function deleteOldEvents(olderThanDays?: number): Promise<number> {
  const days = olderThanDays ?? Number(process.env["LOG_RETENTION_DAYS"] ?? 90)
  const cutoff = Date.now() - days * 24 * 60 * 60 * 1000

  const result = await getDb()
    .delete(events)
    .where(lt(events.ts, cutoff))

  return result.rowsAffected
}
