/**
 * Tests de queries de memoria de usuario contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/memory.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { setMemory, getMemory, deleteMemory, getMemoryAsContext } from "../queries/memory"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM user_memory; DELETE FROM users;")
})

describe("setMemory", () => {
  test("inserta nueva entrada", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "idioma", "español")
    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.value).toBe("español")
    expect(entries[0]!.source).toBe("explicit")
  })

  test("upsert — actualiza el valor si la key ya existe", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "idioma", "inglés")
    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.value).toBe("inglés")
  })

  test("guarda source 'inferred'", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "tema", "tecnología", "inferred")
    const entries = await getMemory(user.id)
    expect(entries[0]!.source).toBe("inferred")
  })

  test("permite múltiples keys por usuario", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "formato", "conciso")
    expect(await getMemory(user.id)).toHaveLength(2)
  })
})

describe("getMemory", () => {
  test("retorna vacío si no hay memoria", async () => {
    const user = await insertUser(db)
    expect(await getMemory(user.id)).toHaveLength(0)
  })

  test("no retorna entradas de otros usuarios", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await setMemory(u1.id, "idioma", "español")
    await setMemory(u2.id, "idioma", "inglés")
    const entries = await getMemory(u1.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.value).toBe("español")
  })
})

describe("deleteMemory", () => {
  test("elimina la key especificada sin afectar otras", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "formato", "conciso")
    await deleteMemory(user.id, "idioma")
    const entries = await getMemory(user.id)
    expect(entries).toHaveLength(1)
    expect(entries[0]!.key).toBe("formato")
  })

  test("no elimina keys de otros usuarios", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await setMemory(u1.id, "idioma", "español")
    await setMemory(u2.id, "idioma", "inglés")
    await deleteMemory(u1.id, "idioma")
    expect(await getMemory(u2.id)).toHaveLength(1)
  })
})

describe("getMemoryAsContext", () => {
  test("retorna string vacío sin entradas", async () => {
    const user = await insertUser(db)
    expect(await getMemoryAsContext(user.id)).toBe("")
  })

  test("retorna formato 'User context and preferences: key: value'", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "idioma", "español")
    const ctx = await getMemoryAsContext(user.id)
    expect(ctx).toContain("User context and preferences:")
    expect(ctx).toContain("idioma: español")
  })

  test("incluye múltiples entradas separadas", async () => {
    const user = await insertUser(db)
    await setMemory(user.id, "idioma", "español")
    await setMemory(user.id, "formato", "conciso")
    const ctx = await getMemoryAsContext(user.id)
    expect(ctx).toContain("idioma: español")
    expect(ctx).toContain("formato: conciso")
  })
})
