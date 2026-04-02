/**
 * Tests de queries de compartición de sesiones contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/shares.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertSession } from "./setup"
import { createShare, getShareByToken, getShareWithSession, revokeShare, listSharesByUser } from "../queries/shares"
import * as schema from "../schema"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM session_shares; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("createShare", () => {
  test("crea share con token único y expiresAt en el futuro", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const share = await createShare(session.id, user.id)
    expect(share.token).toHaveLength(64)
    expect(share.expiresAt).toBeGreaterThan(Date.now())
  })

  test("tokens de shares distintos son únicos", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const s1 = await createShare(session.id, user.id)
    const s2 = await createShare(session.id, user.id)
    expect(s1.token).not.toBe(s2.token)
  })
})

describe("getShareByToken", () => {
  test("retorna share válido", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const share = await createShare(session.id, user.id)
    const found = await getShareByToken(share.token)
    expect(found!.id).toBe(share.id)
  })

  test("retorna null para token inexistente", async () => {
    expect(await getShareByToken("token-inexistente")).toBeNull()
  })

  test("retorna null para token expirado", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    // Insertar share expirado directamente con testDb
    await db.insert(schema.sessionShares).values({
      id: "exp-id",
      sessionId: session.id,
      userId: user.id,
      token: "expired-token",
      expiresAt: Date.now() - 1000,
      createdAt: Date.now(),
    })
    expect(await getShareByToken("expired-token")).toBeNull()
  })
})

describe("getShareWithSession", () => {
  test("retorna share con sesión y mensajes incluidos", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const share = await createShare(session.id, user.id)
    const result = await getShareWithSession(share.token)
    expect(result!.session.id).toBe(session.id)
    expect(Array.isArray(result!.messages)).toBe(true)
  })
})

describe("revokeShare", () => {
  test("revoca el share y el token deja de ser válido", async () => {
    const user = await insertUser(db)
    const session = await insertSession(db, user.id)
    const share = await createShare(session.id, user.id)
    await revokeShare(share.id, user.id)
    expect(await getShareByToken(share.token)).toBeNull()
  })

  test("no revoca shares de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const session = await insertSession(db, u1.id)
    const share = await createShare(session.id, u1.id)
    await revokeShare(share.id, u2.id) // intento de otro usuario
    expect(await getShareByToken(share.token)).not.toBeNull()
  })
})

describe("listSharesByUser", () => {
  test("retorna solo los shares del usuario especificado", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    const s1 = await insertSession(db, u1.id)
    const s2 = await insertSession(db, u2.id)
    await createShare(s1.id, u1.id)
    await createShare(s2.id, u2.id)
    const shares = await listSharesByUser(u1.id)
    expect(shares).toHaveLength(1)
    expect(shares[0]!.userId).toBe(u1.id)
  })
})
