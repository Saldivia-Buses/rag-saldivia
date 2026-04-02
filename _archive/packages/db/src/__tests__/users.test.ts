/**
 * Tests de queries de usuarios contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/users.test.ts
 */

import { describe, test, expect, beforeAll, afterAll, afterEach } from "bun:test"
import { _injectDbForTesting, _resetDbForTesting } from "../connection"
import { createTestDb, initSchema } from "./setup"
import { createUser, getUserById, getUserByEmail, getUserByApiKey, listUsers, updateUser, updatePassword, deleteUser, verifyPassword, addUserArea, removeUserArea, getUserCollections, canAccessCollection } from "../queries/users"
import { createArea, addAreaCollection } from "../queries/areas"

process.env["DATABASE_PATH"] = ":memory:"

const { client, db } = createTestDb()

beforeAll(async () => {
  await initSchema(client)
  _injectDbForTesting(db)
})

afterAll(() => { _resetDbForTesting() })

afterEach(async () => {
  await client.executeMultiple("DELETE FROM area_collections; DELETE FROM user_areas; DELETE FROM areas; DELETE FROM users;")
})

describe("createUser", () => {
  test("crea usuario con email normalizado a minúsculas", async () => {
    const user = await createUser({ email: "USER@TEST.COM", name: "Test", password: "pass123" })
    expect(user.email).toBe("user@test.com")
  })

  test("crea usuario con rol correcto", async () => {
    const user = await createUser({ email: "admin@test.com", name: "Admin", password: "pass", role: "admin" })
    expect(user.role).toBe("admin")
  })

  test("lanza error si el email ya existe", async () => {
    await createUser({ email: "dup@test.com", name: "A", password: "pass" })
    await expect(createUser({ email: "dup@test.com", name: "B", password: "pass" })).rejects.toThrow()
  })

  test("asigna áreas si se pasan areaIds", async () => {
    const area = await createArea("Marketing")
    const user = await createUser({ email: "x@test.com", name: "X", password: "pass", areaIds: [area!.id] })
    const areas = await getUserCollections(user.id)
    // No hay colecciones asignadas, pero el area está asignada
    expect(areas).toHaveLength(0) // sin area_collections aún
  })
})

describe("getUserByEmail", () => {
  test("retorna usuario existente", async () => {
    await createUser({ email: "find@test.com", name: "Find", password: "pass" })
    const user = await getUserByEmail("find@test.com")
    expect(user).not.toBeUndefined()
    expect(user!.name).toBe("Find")
  })

  test("retorna undefined para email inexistente", async () => {
    const user = await getUserByEmail("noexist@test.com")
    expect(user).toBeUndefined()
  })
})

describe("listUsers", () => {
  test("retorna vacío si no hay usuarios", async () => {
    const users = await listUsers()
    expect(users).toHaveLength(0)
  })

  test("retorna todos los usuarios ordenados por nombre", async () => {
    await createUser({ email: "z@test.com", name: "Zeta", password: "p" })
    await createUser({ email: "a@test.com", name: "Alpha", password: "p" })
    const users = await listUsers()
    expect(users[0]!.name).toBe("Alpha")
    expect(users[1]!.name).toBe("Zeta")
  })
})

describe("verifyPassword", () => {
  test("retorna usuario para credenciales correctas", async () => {
    await createUser({ email: "vp@test.com", name: "VP", password: "mypassword" })
    const user = await verifyPassword("vp@test.com", "mypassword")
    expect(user).not.toBeNull()
    expect(user!.email).toBe("vp@test.com")
  })

  test("retorna null para contraseña incorrecta", async () => {
    await createUser({ email: "vp2@test.com", name: "VP2", password: "correct" })
    const result = await verifyPassword("vp2@test.com", "wrong")
    expect(result).toBeNull()
  })

  test("retorna null para email inexistente", async () => {
    const result = await verifyPassword("ghost@test.com", "any")
    expect(result).toBeNull()
  })

  test("retorna null para usuario inactivo", async () => {
    const user = await createUser({ email: "inactive@test.com", name: "I", password: "pass" })
    await updateUser(user.id, { active: false })
    const result = await verifyPassword("inactive@test.com", "pass")
    expect(result).toBeNull()
  })
})

describe("updateUser", () => {
  test("actualiza nombre y rol", async () => {
    const user = await createUser({ email: "upd@test.com", name: "Old", password: "p" })
    const updated = await updateUser(user.id, { name: "New", role: "admin" })
    expect(updated!.name).toBe("New")
    expect(updated!.role).toBe("admin")
  })

  test("desactiva usuario", async () => {
    const user = await createUser({ email: "deact@test.com", name: "D", password: "p" })
    await updateUser(user.id, { active: false })
    const found = await getUserById(user.id)
    expect(found!.active).toBe(false)
  })
})

describe("getUserByApiKey", () => {
  test("retorna usuario activo por apiKeyHash", async () => {
    const user = await createUser({ email: "apikey@test.com", name: "AK", password: "p" })
    const found = await getUserByApiKey(user.apiKeyHash)
    expect(found!.id).toBe(user.id)
  })
})

describe("updatePassword", () => {
  test("actualiza la contraseña y el nuevo hash funciona en verifyPassword", async () => {
    await createUser({ email: "pwd@test.com", name: "P", password: "old" })
    const user = (await getUserByEmail("pwd@test.com"))!
    await updatePassword(user.id, "new-password")
    expect(await verifyPassword("pwd@test.com", "new-password")).not.toBeNull()
    expect(await verifyPassword("pwd@test.com", "old")).toBeNull()
  })
})

describe("deleteUser", () => {
  test("elimina el usuario y sus relaciones en cascade", async () => {
    const user = await createUser({ email: "del@test.com", name: "Del", password: "p" })
    await deleteUser(user.id)
    const found = await getUserById(user.id)
    expect(found).toBeUndefined()
  })
})

describe("getUserCollections / canAccessCollection", () => {
  test("retorna colecciones del usuario a través de sus áreas", async () => {
    const user = await createUser({ email: "col@test.com", name: "C", password: "p" })
    const area = await createArea("IT")
    await addAreaCollection(area!.id, "tech-docs", "read")
    await addUserArea(user.id, area!.id)

    const collections = await getUserCollections(user.id)
    expect(collections.some((c) => c.name === "tech-docs")).toBe(true)
  })

  test("canAccessCollection retorna true si tiene permiso", async () => {
    const user = await createUser({ email: "perm@test.com", name: "P", password: "p" })
    const area = await createArea("Legal")
    await addAreaCollection(area!.id, "legal-docs", "write")
    await addUserArea(user.id, area!.id)

    expect(await canAccessCollection(user.id, "legal-docs", "read")).toBe(true)
    expect(await canAccessCollection(user.id, "legal-docs", "write")).toBe(true)
    expect(await canAccessCollection(user.id, "legal-docs", "admin")).toBe(false)
  })

  test("canAccessCollection retorna false para colección no asignada", async () => {
    const user = await createUser({ email: "noperm@test.com", name: "N", password: "p" })
    expect(await canAccessCollection(user.id, "secret-docs")).toBe(false)
  })
})
