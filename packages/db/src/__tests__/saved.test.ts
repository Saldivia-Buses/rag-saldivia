/**
 * Tests de queries de respuestas guardadas contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/saved.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertSession, insertMessage } from "./setup"
import { saveResponse, listSavedResponses, unsaveResponse, unsaveByMessageId, isSaved } from "../queries/saved"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM saved_responses; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("saveResponse", () => {
  test("guarda una respuesta con todos los campos", async () => {
    const user = await insertUser(db)
    const saved = await saveResponse({ userId: user.id, content: "Respuesta guardada", sessionTitle: "Sesión de prueba" })
    expect(saved.content).toBe("Respuesta guardada")
    expect(saved.sessionTitle).toBe("Sesión de prueba")
    expect(saved.userId).toBe(user.id)
  })

  test("guarda sin messageId (null)", async () => {
    const user = await insertUser(db)
    const saved = await saveResponse({ userId: user.id, content: "Sin mensaje" })
    expect(saved.messageId).toBeNull()
  })

  test("guarda con messageId FK válido", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const msg = await insertMessage(db, session.id, "assistant")
    const saved = await saveResponse({ userId: user.id, messageId: msg.id, content: "Con msg" })
    expect(saved.messageId).toBe(msg.id)
  })
})

describe("listSavedResponses", () => {
  test("retorna vacío si no hay guardados", async () => {
    const user = await insertUser(db)
    expect(await listSavedResponses(user.id)).toHaveLength(0)
  })

  test("retorna solo las respuestas del usuario especificado", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await saveResponse({ userId: u1.id, content: "De u1" })
    await saveResponse({ userId: u2.id, content: "De u2" })
    const list = await listSavedResponses(u1.id)
    expect(list).toHaveLength(1)
    expect(list[0]!.content).toBe("De u1")
  })

  test("retorna ordenado por createdAt descendente", async () => {
    const user = await insertUser(db)
    // Insertar con timestamps explícitos para garantizar el orden
    await db.insert(schema.savedResponses).values({ userId: user.id, content: "Primero", createdAt: 1000 })
    await db.insert(schema.savedResponses).values({ userId: user.id, content: "Segundo", createdAt: 2000 })
    const list = await listSavedResponses(user.id)
    expect(list[0]!.content).toBe("Segundo")
  })
})

describe("unsaveResponse", () => {
  test("elimina la respuesta del usuario correcto", async () => {
    const user = await insertUser(db)
    const saved = await saveResponse({ userId: user.id, content: "Borrar" })
    await unsaveResponse(saved.id, user.id)
    expect(await listSavedResponses(user.id)).toHaveLength(0)
  })

  test("no elimina respuestas de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const saved = await saveResponse({ userId: u1.id, content: "Protegida" })
    await unsaveResponse(saved.id, u2.id)
    expect(await listSavedResponses(u1.id)).toHaveLength(1)
  })
})

describe("unsaveByMessageId", () => {
  test("elimina por messageId del usuario correcto", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const msg = await insertMessage(db, session.id, "assistant")
    await saveResponse({ userId: user.id, messageId: msg.id, content: "Guardada" })
    await unsaveByMessageId(msg.id, user.id)
    expect(await listSavedResponses(user.id)).toHaveLength(0)
  })
})

describe("isSaved", () => {
  test("retorna true si el mensaje está guardado por el usuario", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const msg = await insertMessage(db, session.id, "assistant")
    await saveResponse({ userId: user.id, messageId: msg.id, content: "x" })
    expect(await isSaved(msg.id, user.id)).toBe(true)
  })

  test("retorna false si no está guardado", async () => {
    const user = await insertUser(db)
    expect(await isSaved(999, user.id)).toBe(false)
  })

  test("retorna false si otro usuario guardó el mismo mensaje", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const session = await insertSession(db, u1.id)
    const msg = await insertMessage(db, session.id, "assistant")
    await saveResponse({ userId: u1.id, messageId: msg.id, content: "x" })
    expect(await isSaved(msg.id, u2.id)).toBe(false)
  })
})
