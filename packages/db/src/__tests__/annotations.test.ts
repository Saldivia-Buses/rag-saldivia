/**
 * Tests de queries de anotaciones contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/annotations.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertSession } from "./setup"
import { saveAnnotation, listAnnotationsBySession, deleteAnnotation } from "../queries/annotations"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM annotations; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("saveAnnotation", () => {
  test("crea anotación con texto seleccionado", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const ann = await saveAnnotation({ userId: user.id, sessionId: session.id, selectedText: "texto importante" })
    expect(ann.selectedText).toBe("texto importante")
    expect(ann.note).toBeNull()
  })

  test("guarda nota cuando se provee", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const ann = await saveAnnotation({ userId: user.id, sessionId: session.id, selectedText: "frase", note: "recordar" })
    expect(ann.note).toBe("recordar")
  })
})

describe("listAnnotationsBySession", () => {
  test("retorna vacío si no hay anotaciones", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    expect(await listAnnotationsBySession(session.id, user.id)).toHaveLength(0)
  })

  test("filtra por sessionId Y userId simultáneamente", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const s1 = await insertSession(db, u1.id)
    const s2 = await insertSession(db, u2.id)
    await saveAnnotation({ userId: u1.id, sessionId: s1.id, selectedText: "de u1" })
    await saveAnnotation({ userId: u2.id, sessionId: s2.id, selectedText: "de u2" })
    const result = await listAnnotationsBySession(s1.id, u1.id)
    expect(result).toHaveLength(1)
    expect(result[0]!.selectedText).toBe("de u1")
  })

  test("no retorna anotaciones de otro usuario en la misma sesión", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const session = await insertSession(db, u1.id)
    await saveAnnotation({ userId: u1.id, sessionId: session.id, selectedText: "de u1" })
    expect(await listAnnotationsBySession(session.id, u2.id)).toHaveLength(0)
  })
})

describe("deleteAnnotation", () => {
  test("elimina la anotación del usuario correcto", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const ann = await saveAnnotation({ userId: user.id, sessionId: session.id, selectedText: "borrar" })
    await deleteAnnotation(ann.id, user.id)
    expect(await listAnnotationsBySession(session.id, user.id)).toHaveLength(0)
  })

  test("no elimina anotaciones de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const session = await insertSession(db, u1.id)
    const ann = await saveAnnotation({ userId: u1.id, sessionId: session.id, selectedText: "protegida" })
    await deleteAnnotation(ann.id, u2.id)
    expect(await listAnnotationsBySession(session.id, u1.id)).toHaveLength(1)
  })
})
