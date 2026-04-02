/**
 * Tests de búsqueda universal contra SQLite en memoria.
 * universalSearch tiene FTS5 con fallback LIKE. En :memory: sin setup FTS5,
 * activa el path LIKE. Los tests verifican ese path + edge cases.
 * Corre con: bun test packages/db/src/__tests__/search.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser, insertSession } from "./setup"
import { universalSearch, sanitizeFts5Query } from "../queries/search"
import { saveResponse } from "../queries/saved"
import { createTemplate } from "../queries/templates"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM saved_responses; DELETE FROM prompt_templates; DELETE FROM chat_messages; DELETE FROM chat_sessions; DELETE FROM users;")
})

describe("sanitizeFts5Query", () => {
  test("wraps plain text in quotes", () => {
    expect(sanitizeFts5Query("hello world")).toBe('"hello world"')
  })

  test("escapes double quotes inside input", () => {
    expect(sanitizeFts5Query('say "hello"')).toBe('"say ""hello"""')
  })

  test("neutralizes AND/OR/NOT operators", () => {
    // Wrapped in quotes, FTS5 treats these as literal words
    expect(sanitizeFts5Query("foo AND bar")).toBe('"foo AND bar"')
    expect(sanitizeFts5Query("NOT secret")).toBe('"NOT secret"')
    expect(sanitizeFts5Query("a OR b")).toBe('"a OR b"')
  })

  test("neutralizes wildcard *", () => {
    expect(sanitizeFts5Query("test*")).toBe('"test*"')
  })

  test("neutralizes NEAR operator", () => {
    expect(sanitizeFts5Query("foo NEAR bar")).toBe('"foo NEAR bar"')
  })

  test("handles empty string", () => {
    expect(sanitizeFts5Query("")).toBe('""')
  })
})

describe("universalSearch — casos edge", () => {
  test("retorna vacío para query vacío", async () => {
    const user = await insertUser(db)
    expect(await universalSearch("", user.id)).toHaveLength(0)
  })

  test("retorna vacío para query de 1 caracter", async () => {
    const user = await insertUser(db)
    expect(await universalSearch("a", user.id)).toHaveLength(0)
  })

  test("retorna vacío para query de solo espacios", async () => {
    const user = await insertUser(db)
    expect(await universalSearch("   ", user.id)).toHaveLength(0)
  })
})

describe("universalSearch — sesiones", () => {
  test("encuentra sesión por título (LIKE path)", async () => {
    const user = await insertUser(db)
    await insertSession(db, user.id, crypto.randomUUID(), "Análisis de mercado")
    await insertSession(db, user.id, crypto.randomUUID(), "Estrategia comercial")
    const results = await universalSearch("mercado", user.id)
    expect(results.some((r) => r.type === "session" && r.title === "Análisis de mercado")).toBe(true)
    expect(results.every((r) => r.title !== "Estrategia comercial")).toBe(true)
  })

  test("no retorna sesiones de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await insertSession(db, u1.id, crypto.randomUUID(), "Proyecto confidencial")
    await insertSession(db, u2.id, crypto.randomUUID(), "Proyecto público")
    const results = await universalSearch("Proyecto", u1.id)
    expect(results.filter((r) => r.type === "session")).toHaveLength(1)
    expect(results[0]!.title).toBe("Proyecto confidencial")
  })
})

describe("universalSearch — templates", () => {
  test("encuentra templates por título", async () => {
    const user = await insertUser(db)
    await createTemplate({ title: "Resumen ejecutivo", prompt: "...", createdBy: user.id })
    const results = await universalSearch("ejecutivo", user.id)
    expect(results.some((r) => r.type === "template")).toBe(true)
  })

  test("encuentra templates por contenido del prompt", async () => {
    const user = await insertUser(db)
    await createTemplate({ title: "Análisis", prompt: "Proporciona aspectos técnicos detallados", createdBy: user.id })
    const results = await universalSearch("técnicos", user.id)
    expect(results.some((r) => r.type === "template")).toBe(true)
  })
})

describe("universalSearch — respuestas guardadas", () => {
  test("encuentra respuestas guardadas por contenido", async () => {
    const user = await insertUser(db)
    await saveResponse({ userId: user.id, content: "La inteligencia artificial transformará la industria" })
    const results = await universalSearch("inteligencia", user.id)
    expect(results.some((r) => r.type === "saved")).toBe(true)
  })

  test("no retorna guardados de otro usuario", async () => {
    const u1 = await insertUser(db, "a@test.com")
    const u2 = await insertUser(db, "b@test.com")
    await saveResponse({ userId: u2.id, content: "Contenido exclusivo de u2" })
    expect((await universalSearch("exclusivo", u1.id)).filter((r) => r.type === "saved")).toHaveLength(0)
  })
})
