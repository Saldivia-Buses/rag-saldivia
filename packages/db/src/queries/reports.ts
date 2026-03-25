import { eq, and, lte } from "drizzle-orm"
import { getDb } from "../connection"
import { scheduledReports } from "../schema"
import type { NewScheduledReport } from "../schema"
import { randomUUID } from "crypto"

function calcNextRun(schedule: "daily" | "weekly" | "monthly"): number {
  const now = Date.now()
  switch (schedule) {
    case "daily": return now + 24 * 60 * 60 * 1000
    case "weekly": return now + 7 * 24 * 60 * 60 * 1000
    case "monthly": return now + 30 * 24 * 60 * 60 * 1000
  }
}

export async function createReport(data: Omit<NewScheduledReport, "id" | "createdAt" | "nextRun" | "lastRun" | "active">) {
  const db = getDb()
  const [row] = await db
    .insert(scheduledReports)
    .values({
      id: randomUUID(),
      ...data,
      active: true,
      nextRun: calcNextRun(data.schedule),
      createdAt: Date.now(),
    })
    .returning()
  return row!
}

export async function listActiveReports() {
  const db = getDb()
  const now = Date.now()
  return db
    .select()
    .from(scheduledReports)
    .where(and(eq(scheduledReports.active, true), lte(scheduledReports.nextRun, now)))
}

export async function listReportsByUser(userId: number) {
  const db = getDb()
  return db
    .select()
    .from(scheduledReports)
    .where(eq(scheduledReports.userId, userId))
}

export async function updateLastRun(id: string, schedule: "daily" | "weekly" | "monthly") {
  const db = getDb()
  await db
    .update(scheduledReports)
    .set({ lastRun: Date.now(), nextRun: calcNextRun(schedule) })
    .where(eq(scheduledReports.id, id))
}

export async function deleteReport(id: string, userId: number) {
  const db = getDb()
  await db
    .delete(scheduledReports)
    .where(and(eq(scheduledReports.id, id), eq(scheduledReports.userId, userId)))
}
