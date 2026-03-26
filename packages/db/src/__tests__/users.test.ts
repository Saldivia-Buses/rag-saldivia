/**
 * Tests de queries de usuarios contra SQLite en memoria.
 * Corre con: bun test packages/db/src/__tests__/users.test.ts
 */

import { describe, test, expect, beforeAll, afterEach } from "bun:test"
import { createClient } from "@libsql/client"
import { drizzle } from "drizzle-orm/libsql"
import * as schema from "../schema"

// Usar DB en memoria para tests
process.env["DATABASE_PATH"] = ":memory:"

// Crear conexión de test directamente (no usar el singleton de connection.ts)
const client = createClient({ url: ":memory:" })
const testDb = drizzle(client, { schema })

// Inicializar schema con SQL puro (igual que init.ts)
beforeAll(async () => {
  await client.executeMultiple(`
    CREATE TABLE IF NOT EXISTS areas (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL UNIQUE,
      description TEXT NOT NULL DEFAULT '',
      created_at INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      email TEXT NOT NULL UNIQUE,
      name TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'user',
      api_key_hash TEXT NOT NULL,
      password_hash TEXT,
      preferences TEXT NOT NULL DEFAULT '{}',
      active INTEGER NOT NULL DEFAULT 1,
      onboarding_completed INTEGER NOT NULL DEFAULT 0,
      sso_provider TEXT,
      sso_subject TEXT,
      created_at INTEGER NOT NULL,
      last_login INTEGER
    );
    CREATE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key_hash);
    CREATE TABLE IF NOT EXISTS user_areas (
      user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
      PRIMARY KEY (user_id, area_id)
    );
    CREATE TABLE IF NOT EXISTS area_collections (
      area_id INTEGER NOT NULL REFERENCES areas(id) ON DELETE CASCADE,
      collection_name TEXT NOT NULL,
      permission TEXT NOT NULL DEFAULT 'read',
      PRIMARY KEY (area_id, collection_name)
    );
    CREATE TABLE IF NOT EXISTS events (
      id TEXT PRIMARY KEY,
      ts INTEGER NOT NULL,
      source TEXT NOT NULL,
      level TEXT NOT NULL,
      type TEXT NOT NULL,
      user_id INTEGER REFERENCES users(id),
      session_id TEXT,
      payload TEXT NOT NULL DEFAULT '{}',
      sequence INTEGER NOT NULL
    );
  `)
})

// Limpiar tabla de usuarios entre tests para evitar contaminación
afterEach(async () => {
  await client.executeMultiple("DELETE FROM user_areas; DELETE FROM users;")
})

// Helpers — versiones locales que usan testDb en lugar del singleton
import { eq, and } from "drizzle-orm"
import { hashSync, compareSync } from "bcrypt-ts"

async function createTestUser(data: {
  email: string
  name: string
  password: string
  role?: "admin" | "area_manager" | "user"
  areaIds?: number[]
}) {
  const passwordHash = hashSync(data.password, 10)
  const apiKeyHash = Math.random().toString(36)

  const [user] = await testDb
    .insert(schema.users)
    .values({
      email: data.email.toLowerCase(),
      name: data.name,
      role: data.role ?? "user",
      apiKeyHash,
      passwordHash,
      preferences: {},
      active: true,
      createdAt: Date.now(),
    })
    .returning()

  if (!user) throw new Error("Failed to create user")

  if (data.areaIds && data.areaIds.length > 0) {
    await testDb.insert(schema.userAreas).values(
      data.areaIds.map((areaId) => ({ userId: user.id, areaId }))
    )
  }

  return user
}

async function verifyTestPassword(email: string, password: string) {
  const user = await testDb.query.users.findFirst({
    where: (u, { eq }) => eq(u.email, email.toLowerCase()),
  })
  if (!user || !user.active || !user.passwordHash) return null
  return compareSync(password, user.passwordHash) ? user : null
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe("createUser", () => {
  test("crea un usuario con campos correctos", async () => {
    const user = await createTestUser({
      email: "test@example.com",
      name: "Test User",
      password: "password123",
    })

    expect(user.email).toBe("test@example.com")
    expect(user.name).toBe("Test User")
    expect(user.role).toBe("user")
    expect(user.active).toBe(true)
    expect(user.passwordHash).toBeDefined()
    expect(user.id).toBeGreaterThan(0)
  })

  test("normaliza el email a minúsculas", async () => {
    const user = await createTestUser({
      email: "UPPER@EXAMPLE.COM",
      name: "Upper User",
      password: "password123",
    })
    expect(user.email).toBe("upper@example.com")
  })

  test("crea usuario con rol admin", async () => {
    const user = await createTestUser({
      email: "admin@example.com",
      name: "Admin User",
      password: "password123",
      role: "admin",
    })
    expect(user.role).toBe("admin")
  })

  test("falla con email duplicado", async () => {
    await createTestUser({
      email: "dup@example.com",
      name: "First",
      password: "password123",
    })

    await expect(
      createTestUser({
        email: "dup@example.com",
        name: "Second",
        password: "password123",
      })
    ).rejects.toThrow()
  })
})

describe("verifyPassword", () => {
  test("retorna usuario con credenciales correctas", async () => {
    await createTestUser({
      email: "auth@example.com",
      name: "Auth User",
      password: "secret123",
    })

    const result = await verifyTestPassword("auth@example.com", "secret123")
    expect(result).not.toBeNull()
    expect(result?.email).toBe("auth@example.com")
  })

  test("retorna null con password incorrecta", async () => {
    await createTestUser({
      email: "auth2@example.com",
      name: "Auth User 2",
      password: "correct-password",
    })

    const result = await verifyTestPassword("auth2@example.com", "wrong-password")
    expect(result).toBeNull()
  })

  test("retorna null para email inexistente", async () => {
    const result = await verifyTestPassword("noexiste@example.com", "password")
    expect(result).toBeNull()
  })

  test("retorna null para usuario inactivo", async () => {
    const user = await createTestUser({
      email: "inactive@example.com",
      name: "Inactive",
      password: "password123",
    })

    // Desactivar el usuario
    await testDb
      .update(schema.users)
      .set({ active: false })
      .where(eq(schema.users.id, user.id))

    const result = await verifyTestPassword("inactive@example.com", "password123")
    expect(result).toBeNull()
  })
})

describe("listUsers", () => {
  test("retorna array vacío cuando no hay usuarios", async () => {
    const users = await testDb.query.users.findMany()
    expect(users).toHaveLength(0)
  })

  test("retorna todos los usuarios creados", async () => {
    await createTestUser({ email: "a@example.com", name: "A", password: "pass1" })
    await createTestUser({ email: "b@example.com", name: "B", password: "pass2" })
    await createTestUser({ email: "c@example.com", name: "C", password: "pass3" })

    const users = await testDb.query.users.findMany()
    expect(users).toHaveLength(3)
  })

  test("los usuarios tienen los campos esperados", async () => {
    await createTestUser({ email: "list@example.com", name: "List User", password: "pass" })

    const users = await testDb.query.users.findMany()
    const user = users[0]!
    expect(user).toHaveProperty("id")
    expect(user).toHaveProperty("email")
    expect(user).toHaveProperty("name")
    expect(user).toHaveProperty("role")
    expect(user).toHaveProperty("active")
    expect(user).toHaveProperty("createdAt")
  })
})

describe("updateUser", () => {
  test("actualiza el nombre del usuario", async () => {
    const user = await createTestUser({
      email: "update@example.com",
      name: "Original Name",
      password: "pass",
    })

    const [updated] = await testDb
      .update(schema.users)
      .set({ name: "Updated Name" })
      .where(eq(schema.users.id, user.id))
      .returning()

    expect(updated?.name).toBe("Updated Name")
  })

  test("actualiza el rol del usuario", async () => {
    const user = await createTestUser({
      email: "roleupdate@example.com",
      name: "Role User",
      password: "pass",
      role: "user",
    })

    const [updated] = await testDb
      .update(schema.users)
      .set({ role: "area_manager" })
      .where(eq(schema.users.id, user.id))
      .returning()

    expect(updated?.role).toBe("area_manager")
  })

  test("desactivar usuario cambia active a false", async () => {
    const user = await createTestUser({
      email: "deactivate@example.com",
      name: "Active User",
      password: "pass",
    })

    expect(user.active).toBe(true)

    await testDb
      .update(schema.users)
      .set({ active: false })
      .where(eq(schema.users.id, user.id))

    const updated = await testDb.query.users.findFirst({
      where: (u, { eq }) => eq(u.id, user.id),
    })
    expect(updated?.active).toBe(false)
  })
})

describe("deleteUser", () => {
  test("elimina el usuario de la base de datos", async () => {
    const user = await createTestUser({
      email: "delete@example.com",
      name: "Delete Me",
      password: "pass",
    })

    await testDb.delete(schema.users).where(eq(schema.users.id, user.id))

    const deleted = await testDb.query.users.findFirst({
      where: (u, { eq }) => eq(u.id, user.id),
    })
    expect(deleted).toBeUndefined()
  })

  test("deleteUser elimina también filas en user_areas (CASCADE)", async () => {
    // Crear un área primero
    const [area] = await testDb
      .insert(schema.areas)
      .values({ name: "Test Area", description: "", createdAt: Date.now() })
      .returning()

    const user = await createTestUser({
      email: "cascade@example.com",
      name: "Cascade User",
      password: "pass",
      areaIds: [area!.id],
    })

    // Verificar que user_areas tiene la fila
    const beforeDelete = await testDb.query.userAreas.findMany({
      where: (ua, { eq }) => eq(ua.userId, user.id),
    })
    expect(beforeDelete).toHaveLength(1)

    // Borrar el usuario
    await testDb.delete(schema.users).where(eq(schema.users.id, user.id))

    // Verificar que user_areas ya no tiene la fila (CASCADE)
    const afterDelete = await testDb.query.userAreas.findMany({
      where: (ua, { eq }) => eq(ua.userId, user.id),
    })
    expect(afterDelete).toHaveLength(0)
  })
})
