/**
 * Tests for lib/auth/permissions.ts — checkPermission, requirePermission, getCurrentPermissions.
 *
 * These functions are thin wrappers over @rag-saldivia/db RBAC queries + current-user.
 * We mock the DB layer and next/headers to test the glue logic in isolation.
 *
 * Corre con: bun test apps/web/src/lib/auth/__tests__/permissions.test.ts
 */

import { describe, test, expect, mock, beforeEach } from "bun:test"

// ── Mocks ──────────────────────────────────────────────────────────────────

// Suppress logger side effects from transitive imports
mock.module("@rag-saldivia/logger/backend", () => ({
  log: { info: () => {}, warn: () => {}, error: () => {}, debug: () => {} },
}))

// Track mock state so we can reconfigure per-test
let mockHasPermission = mock(() => Promise.resolve(false))
let mockGetUserEffectivePermissions = mock(() => Promise.resolve(new Set<string>()))

mock.module("@rag-saldivia/db", () => ({
  hasPermission: (userId: number, permKey: string) => mockHasPermission(userId, permKey),
  getUserEffectivePermissions: (userId: number) => mockGetUserEffectivePermissions(userId),
  touchUserPresence: async () => {},
}))

// Mock next/headers — default: authenticated admin user
let mockHeaders = new Headers({
  "x-user-id": "42",
  "x-user-email": "admin@test.com",
  "x-user-name": "Admin User",
  "x-user-role": "admin",
})

mock.module("next/headers", () => ({
  headers: () => mockHeaders,
  cookies: () => ({ get: () => null, set: () => {}, delete: () => {} }),
}))

// Import after mocks are registered
const { checkPermission, requirePermission, getCurrentPermissions } = await import(
  "../permissions.js"
)

// ── Helpers ────────────────────────────────────────────────────────────────

function setAuthHeaders(overrides: Partial<Record<string, string>> = {}) {
  const defaults: Record<string, string> = {
    "x-user-id": "42",
    "x-user-email": "admin@test.com",
    "x-user-name": "Admin User",
    "x-user-role": "admin",
  }
  mockHeaders = new Headers({ ...defaults, ...overrides })
}

function clearAuthHeaders() {
  mockHeaders = new Headers()
}

// ── Tests ──────────────────────────────────────────────────────────────────

describe("checkPermission", () => {
  beforeEach(() => {
    mockHasPermission.mockReset()
  })

  test("returns true when user has the permission", async () => {
    mockHasPermission.mockResolvedValue(true)

    const result = await checkPermission(42, "users.manage")

    expect(result).toBe(true)
    expect(mockHasPermission).toHaveBeenCalledWith(42, "users.manage")
  })

  test("returns false when user lacks the permission", async () => {
    mockHasPermission.mockResolvedValue(false)

    const result = await checkPermission(42, "admin.config")

    expect(result).toBe(false)
    expect(mockHasPermission).toHaveBeenCalledWith(42, "admin.config")
  })

  test("passes through the exact userId and permKey to db layer", async () => {
    mockHasPermission.mockResolvedValue(false)

    await checkPermission(999, "collections.write")

    expect(mockHasPermission).toHaveBeenCalledWith(999, "collections.write")
  })

  test("propagates db errors", async () => {
    mockHasPermission.mockRejectedValue(new Error("DB connection failed"))

    await expect(checkPermission(1, "any.perm")).rejects.toThrow("DB connection failed")
  })
})

describe("requirePermission", () => {
  beforeEach(() => {
    mockHasPermission.mockReset()
    setAuthHeaders()
  })

  test("returns the current user when they have the permission", async () => {
    mockHasPermission.mockResolvedValue(true)

    const user = await requirePermission("users.manage")

    expect(user).toEqual({
      id: 42,
      email: "admin@test.com",
      name: "Admin User",
      role: "admin",
    })
    expect(mockHasPermission).toHaveBeenCalledWith(42, "users.manage")
  })

  test("throws when user lacks the permission", async () => {
    mockHasPermission.mockResolvedValue(false)

    await expect(requirePermission("admin.config")).rejects.toThrow(
      "Missing permission: admin.config"
    )
  })

  test("error message includes the permission key", async () => {
    mockHasPermission.mockResolvedValue(false)

    await expect(requirePermission("collections.delete")).rejects.toThrow(
      "collections.delete"
    )
  })

  test("redirects to /login when no authenticated user (no headers)", async () => {
    clearAuthHeaders()

    // next/navigation redirect mock throws NEXT_REDIRECT
    await expect(requirePermission("any.perm")).rejects.toThrow("NEXT_REDIRECT")
  })

  test("redirects when user headers are partial (missing x-user-name)", async () => {
    mockHeaders = new Headers({
      "x-user-id": "42",
      "x-user-email": "admin@test.com",
      // missing x-user-name and x-user-role
    })

    await expect(requirePermission("any.perm")).rejects.toThrow("NEXT_REDIRECT")
  })
})

describe("getCurrentPermissions", () => {
  beforeEach(() => {
    mockGetUserEffectivePermissions.mockReset()
    setAuthHeaders()
  })

  test("returns the effective permission set for the current user", async () => {
    const perms = new Set(["users.read", "collections.read", "collections.write"])
    mockGetUserEffectivePermissions.mockResolvedValue(perms)

    const result = await getCurrentPermissions()

    expect(result).toBeInstanceOf(Set)
    expect(result.has("users.read")).toBe(true)
    expect(result.has("collections.read")).toBe(true)
    expect(result.has("collections.write")).toBe(true)
    expect(result.size).toBe(3)
    expect(mockGetUserEffectivePermissions).toHaveBeenCalledWith(42)
  })

  test("returns empty set when user has no permissions", async () => {
    mockGetUserEffectivePermissions.mockResolvedValue(new Set())

    const result = await getCurrentPermissions()

    expect(result.size).toBe(0)
  })

  test("redirects to /login when no authenticated user", async () => {
    clearAuthHeaders()

    await expect(getCurrentPermissions()).rejects.toThrow("NEXT_REDIRECT")
  })
})
