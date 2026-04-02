import { eq, and } from "drizzle-orm"
import { getDb } from "../connection"
import { projects, projectSessions, projectCollections } from "../schema"
import type { NewProject } from "../schema"
import { randomUUID } from "crypto"

export async function createProject(data: Omit<NewProject, "id" | "createdAt" | "updatedAt">) {
  const db = getDb()
  const now = Date.now()
  const [row] = await db
    .insert(projects)
    .values({ id: randomUUID(), ...data, createdAt: now, updatedAt: now })
    .returning()
  return row!
}

export async function listProjects(userId: number) {
  const db = getDb()
  return db.select().from(projects).where(eq(projects.userId, userId))
}

export async function getProject(id: string) {
  const db = getDb()
  const rows = await db.select().from(projects).where(eq(projects.id, id)).limit(1)
  return rows[0] ?? null
}

export async function updateProject(id: string, data: Partial<Pick<NewProject, "name" | "description" | "instructions">>) {
  const db = getDb()
  await db.update(projects).set({ ...data, updatedAt: Date.now() }).where(eq(projects.id, id))
}

export async function deleteProject(id: string, userId: number) {
  const db = getDb()
  await db.delete(projects).where(and(eq(projects.id, id), eq(projects.userId, userId)))
}

export async function addSessionToProject(projectId: string, sessionId: string) {
  const db = getDb()
  await db.insert(projectSessions).values({ projectId, sessionId }).onConflictDoNothing()
}

export async function removeSessionFromProject(projectId: string, sessionId: string) {
  const db = getDb()
  await db.delete(projectSessions).where(
    and(eq(projectSessions.projectId, projectId), eq(projectSessions.sessionId, sessionId))
  )
}

export async function addCollectionToProject(projectId: string, collectionName: string) {
  const db = getDb()
  await db.insert(projectCollections).values({ projectId, collectionName }).onConflictDoNothing()
}

export async function getProjectBySession(sessionId: string) {
  const db = getDb()
  const rows = await db
    .select({ projectId: projectSessions.projectId })
    .from(projectSessions)
    .where(eq(projectSessions.sessionId, sessionId))
    .limit(1)
  if (!rows[0]) return null
  return getProject(rows[0].projectId)
}
