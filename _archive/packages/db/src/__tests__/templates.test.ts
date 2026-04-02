/**
 * Tests de queries de prompt templates contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/templates.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema, insertUser } from "./setup"
import { createTemplate, listActiveTemplates, deleteTemplate } from "../queries/templates"
import * as schema from "../schema"
import { eq } from "drizzle-orm"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM prompt_templates; DELETE FROM users;")
})

describe("createTemplate", () => {
  test("crea template con active=true por defecto", async () => {
    const admin = await insertUser(db, "admin@test.com", "admin")
    const tpl = await createTemplate({ title: "Resumen", prompt: "Resume en 3 puntos", createdBy: admin.id })
    expect(tpl.title).toBe("Resumen")
    expect(tpl.active).toBe(true)
    expect(tpl.focusMode).toBe("detallado")
  })

  test("acepta focusMode personalizado", async () => {
    const admin = await insertUser(db, "admin@test.com", "admin")
    const tpl = await createTemplate({ title: "Técnico", prompt: "...", createdBy: admin.id, focusMode: "técnico" })
    expect(tpl.focusMode).toBe("técnico")
  })
})

describe("listActiveTemplates", () => {
  test("retorna vacío si no hay templates", async () => {
    expect(await listActiveTemplates()).toHaveLength(0)
  })

  test("retorna solo los templates activos", async () => {
    const admin = await insertUser(db, "admin@test.com", "admin")
    const t1 = await createTemplate({ title: "Activo", prompt: "...", createdBy: admin.id })
    const t2 = await createTemplate({ title: "Otro", prompt: "...", createdBy: admin.id })
    // Desactivar t1
    await db.update(schema.promptTemplates).set({ active: false }).where(eq(schema.promptTemplates.id, t1.id))
    const list = await listActiveTemplates()
    expect(list).toHaveLength(1)
    expect(list[0]!.title).toBe("Otro")
  })
})

describe("deleteTemplate", () => {
  test("elimina el template y no aparece en la lista", async () => {
    const admin = await insertUser(db, "admin@test.com", "admin")
    const tpl = await createTemplate({ title: "Borrar", prompt: "...", createdBy: admin.id })
    await deleteTemplate(tpl.id)
    const list = await listActiveTemplates()
    expect(list.find((t) => t.id === tpl.id)).toBeUndefined()
  })
})
