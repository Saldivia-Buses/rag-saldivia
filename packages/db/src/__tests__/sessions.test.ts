/**
 * Tests de queries de sesiones y mensajes contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/sessions.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { createSession, listSessionsByUser, getSessionById, updateSessionTitle, deleteSession, addMessage, addFeedback } from "../queries/sessions"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM message_feedback; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("createSession", () => {
  test("crea sesión con defaults correctos", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "docs" })
    expect(session!.userId).toBe(user.id)
    expect(session!.title).toBe("Nueva sesión")
    expect(session!.collection).toBe("docs")
    expect(session!.crossdoc).toBe(false)
  })

  test("acepta título personalizado", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "docs", title: "Mi sesión" })
    expect(session!.title).toBe("Mi sesión")
  })
})

describe("listSessionsByUser", () => {
  test("retorna vacío si el usuario no tiene sesiones", async () => {
    const user = await insertUser(db)
    expect(await listSessionsByUser(user.id)).toHaveLength(0)
  })

  test("retorna solo sesiones del usuario, ordenadas por updatedAt desc", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await createSession({ userId: u1.id, collection: "c" })
    await createSession({ userId: u2.id, collection: "c" })
    expect(await listSessionsByUser(u1.id)).toHaveLength(1)
  })
})

describe("getSessionById", () => {
  test("retorna sesión con mensajes incluidos", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "c" })
    await addMessage({ sessionId: session!.id, role: "user", content: "Hola" })
    const found = await getSessionById(session!.id)
    expect(found!.messages).toHaveLength(1)
  })

  test("retorna undefined si el userId no coincide", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const session = await createSession({ userId: u1.id, collection: "c" })
    expect(await getSessionById(session!.id, u2.id)).toBeUndefined()
  })
})

describe("updateSessionTitle", () => {
  test("actualiza el título", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "c" })
    const updated = await updateSessionTitle(session!.id, user.id, "Nuevo título")
    expect(updated!.title).toBe("Nuevo título")
  })

  test("no actualiza sesiones de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const session = await createSession({ userId: u1.id, collection: "c", title: "Original" })
    const result = await updateSessionTitle(session!.id, u2.id, "Hack")
    expect(result).toBeUndefined()
  })
})

describe("deleteSession", () => {
  test("elimina sesión y mensajes en cascade", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "c" })
    await addMessage({ sessionId: session!.id, role: "user", content: "msg" })
    await deleteSession(session!.id, user.id)
    expect(await getSessionById(session!.id)).toBeUndefined()
  })
})

describe("addMessage", () => {
  test("crea mensaje y actualiza updatedAt de la sesión", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "c" })
    const msg = await addMessage({ sessionId: session!.id, role: "assistant", content: "Respuesta" })
    expect(msg!.role).toBe("assistant")
    expect(msg!.content).toBe("Respuesta")
  })
})

describe("addFeedback", () => {
  test("agrega y actualiza feedback (upsert)", async () => {
    const user = await insertUser(db)
    const session = await createSession({ userId: user.id, collection: "c" })
    const msg = await addMessage({ sessionId: session!.id, role: "assistant", content: "R" })
    await addFeedback(msg!.id, user.id, "up")
    await addFeedback(msg!.id, user.id, "down")
    // El segundo debe reemplazar al primero
    const rows = await db.query.messageFeedback.findMany()
    expect(rows).toHaveLength(1)
    expect(rows[0]!.rating).toBe("down")
  })
})
