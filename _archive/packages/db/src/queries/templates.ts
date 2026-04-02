import { eq, desc } from "drizzle-orm"
import { getDb } from "../connection"
import { promptTemplates } from "../schema"
import type { NewPromptTemplate } from "../schema"

export async function listActiveTemplates() {
  const db = getDb()
  return db
    .select()
    .from(promptTemplates)
    .where(eq(promptTemplates.active, true))
    .orderBy(desc(promptTemplates.createdAt))
}

export async function createTemplate(data: Omit<NewPromptTemplate, "createdAt" | "active">) {
  const db = getDb()
  const [row] = await db
    .insert(promptTemplates)
    .values({ ...data, active: true, createdAt: Date.now() })
    .returning()
  return row!
}

export async function deleteTemplate(id: number) {
  const db = getDb()
  await db.delete(promptTemplates).where(eq(promptTemplates.id, id))
}
