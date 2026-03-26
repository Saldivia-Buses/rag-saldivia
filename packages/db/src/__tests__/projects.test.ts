/**
 * Tests de queries de proyectos contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/projects.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import { eq, and } from "drizzle-orm"
import { randomUUID } from "crypto"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      email TEXT NOT NULL UNIQUE,
      name TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'user',
      api_key_hash TEXT NOT NULL,
      password_hash TEXT,
      preferences TEXT NOT NULL DEFAULT '{}',
      active INTEGER NOT NULL DEFAULT 1,
      onboarding_completed INTEGER NOT NULL DEFAULT 0,
      sso_provider TEXT,
      sso_subject TEXT,
      created_at INTEGER NOT NULL,
      last_login INTEGER
    );
    CREATE TABLE IF NOT EXISTS chat_sessions (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      title TEXT NOT NULL,
      collection TEXT NOT NULL,
      crossdoc INTEGER NOT NULL DEFAULT 0,
      forked_from TEXT,
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS projects (
      id TEXT PRIMARY KEY,
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      name TEXT NOT NULL,
      description TEXT NOT NULL DEFAULT '',
      instructions TEXT NOT NULL DEFAULT '',
      created_at INTEGER NOT NULL,
      updated_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS project_sessions (
      project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
      session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
      PRIMARY KEY (project_id, session_id)
    );
    CREATE TABLE IF NOT EXISTS project_collections (
      project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
      collection_name TEXT NOT NULL,
      PRIMARY KEY (project_id, collection_name)
    );
    CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);
  `)
})

afterEach(async () => {
  await client.executeMultiple(
    "DELETE FROM project_collections; DELETE FROM project_sessions; DELETE FROM projects; DELETE FROM chat_sessions; DELETE FROM users;"
  )
})

// ── Helpers ──────────────────────────────────────────────────────────────────

async function createUser(email = "user@test.com") {
  const [user] = await testDb
    .insert(schema.users)
    .values({ email, name: "Test", role: "user", apiKeyHash: "hash", preferences: {}, active: true, createdAt: Date.now() })
    .returning()
  return user!
}

async function createSession(userId: number, id = "sess-1") {
  const now = Date.now()
  const [session] = await testDb
    .insert(schema.chatSessions)
    .values({ id, userId, title: "Test", collection: "col", crossdoc: false, createdAt: now, updatedAt: now })
    .returning()
  return session!
}

async function createProject(userId: number, data: { name: string; description?: string; instructions?: string }) {
  const now = Date.now()
  const [row] = await testDb
    .insert(schema.projects)
    .values({ id: randomUUID(), userId, name: data.name, description: data.description ?? "", instructions: data.instructions ?? "", createdAt: now, updatedAt: now })
    .returning()
  return row!
}

async function listProjects(userId: number) {
  return testDb.select().from(schema.projects).where(eq(schema.projects.userId, userId))
}

async function getProject(id: string) {
  const rows = await testDb.select().from(schema.projects).where(eq(schema.projects.id, id)).limit(1)
  return rows[0] ?? null
}

async function updateProject(id: string, data: Partial<{ name: string; description: string; instructions: string }>) {
  await testDb.update(schema.projects).set({ ...data, updatedAt: Date.now() }).where(eq(schema.projects.id, id))
}

async function deleteProject(id: string, userId: number) {
  await testDb.delete(schema.projects).where(and(eq(schema.projects.id, id), eq(schema.projects.userId, userId)))
}

async function addSessionToProject(projectId: string, sessionId: string) {
  await testDb.insert(schema.projectSessions).values({ projectId, sessionId }).onConflictDoNothing()
}

async function addCollectionToProject(projectId: string, collectionName: string) {
  await testDb.insert(schema.projectCollections).values({ projectId, collectionName }).onConflictDoNothing()
}

async function getProjectBySession(sessionId: string) {
  const rows = await testDb
    .select({ projectId: schema.projectSessions.projectId })
    .from(schema.projectSessions)
    .where(eq(schema.projectSessions.sessionId, sessionId))
    .limit(1)
  if (!rows[0]) return null
  return getProject(rows[0].projectId)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("createProject", () => {
  test("crea un proyecto con los campos correctos", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "Proyecto Alpha", description: "Descripción", instructions: "Instrucciones" })

    expect(project.name).toBe("Proyecto Alpha")
    expect(project.description).toBe("Descripción")
    expect(project.instructions).toBe("Instrucciones")
    expect(project.userId).toBe(user.id)
    expect(project.id).toHaveLength(36)
    expect(project.createdAt).toBeGreaterThan(0)
  })

  test("description e instructions tienen string vacío por defecto", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "Minimal" })
    expect(project.description).toBe("")
    expect(project.instructions).toBe("")
  })
})

describe("listProjects", () => {
  test("retorna vacío si el usuario no tiene proyectos", async () => {
    const user = await createUser()
    const projects = await listProjects(user.id)
    expect(projects).toHaveLength(0)
  })

  test("retorna solo proyectos del usuario especificado", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    await createProject(u1.id, { name: "P1" })
    await createProject(u2.id, { name: "P2" })

    const projects = await listProjects(u1.id)
    expect(projects).toHaveLength(1)
    expect(projects[0]!.name).toBe("P1")
  })
})

describe("getProject", () => {
  test("retorna el proyecto por id", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "Buscado" })
    const found = await getProject(project.id)
    expect(found!.name).toBe("Buscado")
  })

  test("retorna null si no existe", async () => {
    const found = await getProject("id-inexistente")
    expect(found).toBeNull()
  })
})

describe("updateProject", () => {
  test("actualiza los campos del proyecto", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "Viejo" })

    await updateProject(project.id, { name: "Nuevo", description: "Nueva descripción" })

    const updated = await getProject(project.id)
    expect(updated!.name).toBe("Nuevo")
    expect(updated!.description).toBe("Nueva descripción")
  })
})

describe("deleteProject", () => {
  test("elimina el proyecto del usuario correcto", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "Borrar" })

    await deleteProject(project.id, user.id)

    const found = await getProject(project.id)
    expect(found).toBeNull()
  })

  test("no elimina proyectos de otro usuario", async () => {
    const u1 = await createUser("a@test.com")
    const u2 = await createUser("b@test.com")
    const project = await createProject(u1.id, { name: "Protegido" })

    await deleteProject(project.id, u2.id)

    const found = await getProject(project.id)
    expect(found).not.toBeNull()
  })
})

describe("addSessionToProject / addCollectionToProject", () => {
  test("asocia una sesión a un proyecto", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "P" })
    await createSession(user.id, "sess-p1")

    await addSessionToProject(project.id, "sess-p1")

    const sessions = await testDb.select().from(schema.projectSessions).where(eq(schema.projectSessions.projectId, project.id))
    expect(sessions).toHaveLength(1)
    expect(sessions[0]!.sessionId).toBe("sess-p1")
  })

  test("addSessionToProject es idempotente", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "P" })
    await createSession(user.id, "sess-idem")

    await addSessionToProject(project.id, "sess-idem")
    await addSessionToProject(project.id, "sess-idem") // segunda llamada

    const sessions = await testDb.select().from(schema.projectSessions).where(eq(schema.projectSessions.projectId, project.id))
    expect(sessions).toHaveLength(1)
  })

  test("asocia una colección a un proyecto", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "P" })

    await addCollectionToProject(project.id, "mi-coleccion")

    const cols = await testDb.select().from(schema.projectCollections).where(eq(schema.projectCollections.projectId, project.id))
    expect(cols).toHaveLength(1)
    expect(cols[0]!.collectionName).toBe("mi-coleccion")
  })
})

describe("getProjectBySession", () => {
  test("retorna el proyecto asociado a una sesión", async () => {
    const user = await createUser()
    const project = await createProject(user.id, { name: "Proyecto" })
    await createSession(user.id, "sess-find")
    await addSessionToProject(project.id, "sess-find")

    const found = await getProjectBySession("sess-find")
    expect(found!.id).toBe(project.id)
  })

  test("retorna null si la sesión no pertenece a ningún proyecto", async () => {
    const user = await createUser()
    await createSession(user.id, "sess-orphan")

    const found = await getProjectBySession("sess-orphan")
    expect(found).toBeNull()
  })
})
