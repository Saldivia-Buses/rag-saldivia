/**
 * Tests de queries de etiquetas de sesión contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/tags.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertSession } from "./setup"
import { addTag, removeTag, listTagsBySession, listTagsByUser } from "../queries/tags"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM session_tags; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("addTag", () => {
  test("agrega tag normalizado a minúsculas", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    await addTag(session.id, "URGENTE")
    expect(await listTagsBySession(session.id)).toContain("urgente")
  })

  test("es idempotente — no duplica", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    await addTag(session.id, "trabajo")
    await addTag(session.id, "trabajo")
    expect((await listTagsBySession(session.id)).filter((t) => t === "trabajo")).toHaveLength(1)
  })

  test("agrega múltiples tags distintos", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    await addTag(session.id, "urgente")
    await addTag(session.id, "trabajo")
    expect(await listTagsBySession(session.id)).toHaveLength(2)
  })
})

describe("removeTag", () => {
  test("elimina solo el tag especificado — no borra los demás", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    await addTag(session.id, "borrar")
    await addTag(session.id, "mantener")
    await removeTag(session.id, "borrar")
    const tags = await listTagsBySession(session.id)
    expect(tags).not.toContain("borrar")
    expect(tags).toContain("mantener")  // ← verifica el bug histórico de removeTag
  })
})

describe("listTagsByUser", () => {
  test("retorna tags únicos de todas las sesiones del usuario", async () => {
    const user = await insertUser(db)
    const s1 = await insertSession(db, user.id)
    const s2 = await insertSession(db, user.id)
    await addTag(s1.id, "trabajo")
    await addTag(s2.id, "trabajo") // mismo tag
    await addTag(s2.id, "urgente")
    const tags = await listTagsByUser(user.id)
    expect(tags).toHaveLength(2)
    expect(tags).toContain("trabajo")
    expect(tags).toContain("urgente")
  })

  test("no retorna tags de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const s1 = await insertSession(db, u1.id)
    const s2 = await insertSession(db, u2.id)
    await addTag(s1.id, "privado")
    await addTag(s2.id, "ajeno")
    expect(await listTagsByUser(u1.id)).not.toContain("ajeno")
  })
})
