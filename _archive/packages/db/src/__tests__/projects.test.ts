/**
 * Tests de queries de proyectos contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/projects.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertSession } from "./setup"
import { createProject, listProjects, getProject, updateProject, deleteProject, addSessionToProject, addCollectionToProject, getProjectBySession } from "../queries/projects"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM project_collections; DELETE FROM project_sessions; DELETE FROM projects; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("createProject", () => {
  test("crea proyecto con campos correctos", async () => {
    const user = await insertUser(db)
    const project = await createProject({ userId: user.id, name: "Alpha", description: "Desc", instructions: "Instrucciones" })
    expect(project.name).toBe("Alpha")
    expect(project.userId).toBe(user.id)
    expect(project.id).toHaveLength(36)
  })
})

describe("listProjects", () => {
  test("retorna solo proyectos del usuario especificado", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await createProject({ userId: u1.id, name: "P1" })
    await createProject({ userId: u2.id, name: "P2" })
    const list = await listProjects(u1.id)
    expect(list).toHaveLength(1)
    expect(list[0]!.name).toBe("P1")
  })
})

describe("getProject", () => {
  test("retorna proyecto por id", async () => {
    const user = await insertUser(db)
    const project = await createProject({ userId: user.id, name: "Buscado" })
    expect((await getProject(project.id))!.name).toBe("Buscado")
  })

  test("retorna null si no existe", async () => {
    expect(await getProject("id-inexistente")).toBeNull()
  })
})

describe("updateProject", () => {
  test("actualiza nombre y descripción", async () => {
    const user = await insertUser(db)
    const project = await createProject({ userId: user.id, name: "Viejo" })
    await updateProject(project.id, { name: "Nuevo" })
    expect((await getProject(project.id))!.name).toBe("Nuevo")
  })
})

describe("deleteProject", () => {
  test("no elimina proyectos de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const project = await createProject({ userId: u1.id, name: "Protegido" })
    await deleteProject(project.id, u2.id)
    expect(await getProject(project.id)).not.toBeNull()
  })

  test("elimina el proyecto del usuario correcto", async () => {
    const user = await insertUser(db)
    const project = await createProject({ userId: user.id, name: "Borrar" })
    await deleteProject(project.id, user.id)
    expect(await getProject(project.id)).toBeNull()
  })
})

describe("addSessionToProject / addCollectionToProject / getProjectBySession", () => {
  test("asocia sesión y es idempotente", async () => {
    const user = await insertUser(db)
    const project = await createProject({ userId: user.id, name: "P" })
    const session = await insertSession(db, user.id)
    await addSessionToProject(project.id, session.id)
    await addSessionToProject(project.id, session.id) // idempotente
    expect(await getProjectBySession(session.id)).not.toBeNull()
    expect((await getProjectBySession(session.id))!.id).toBe(project.id)
  })

  test("asocia colección", async () => {
    const user = await insertUser(db)
    const project = await createProject({ userId: user.id, name: "P" })
    await addCollectionToProject(project.id, "mi-coleccion")
    const cols = await db.query.projectCollections.findMany()
    expect(cols.some((c) => c.collectionName === "mi-coleccion")).toBe(true)
  })

  test("getProjectBySession retorna null para sesión sin proyecto", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    expect(await getProjectBySession(session.id)).toBeNull()
  })
})
