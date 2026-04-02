/**
 * Tests for area server actions (next-safe-action, adminAction).
 *
 * Mocks: next/headers, next/cache, @rag-saldivia/db, @rag-saldivia/logger
 * Runs with: bun test apps/web/src/__tests__/actions/areas.test.ts
 */

import { describe, test, expect, mock, beforeEach } from "bun:test"

// ── Mock functions (declared before mock.module) ─────────────────────────

const mockCreateArea = mock(() =>
  Promise.resolve({ id: 1, name: "operations", description: "", createdAt: Date.now() })
)
const mockUpdateArea = mock(() =>
  Promise.resolve({ id: 1, name: "updated-area", description: "new desc" })
)
const mockDeleteArea = mock(() => Promise.resolve())
const mockCountUsersInArea = mock(() => Promise.resolve(0))
const mockSetAreaCollections = mock(() => Promise.resolve())
const mockAddUserArea = mock(() => Promise.resolve())
const mockRemoveUserArea = mock(() => Promise.resolve())

// ── Module mocks ─────────────────────────────────────────────────────────

mock.module("@rag-saldivia/db", () => ({
  createArea: mockCreateArea,
  updateArea: mockUpdateArea,
  deleteArea: mockDeleteArea,
  countUsersInArea: mockCountUsersInArea,
  setAreaCollections: mockSetAreaCollections,
  addUserArea: mockAddUserArea,
  removeUserArea: mockRemoveUserArea,
  touchUserPresence: mock(() => Promise.resolve()),
}))

mock.module("@rag-saldivia/logger/backend", () => ({
  log: { info: () => {}, warn: () => {}, error: () => {} },
}))

// Mock next/headers — simulate authenticated admin via x-user-* headers
mock.module("next/headers", () => ({
  headers: () =>
    new Headers({
      "x-user-id": "1",
      "x-user-email": "admin@test.com",
      "x-user-name": "Admin",
      "x-user-role": "admin",
    }),
  cookies: () => ({ get: () => null, set: () => {}, delete: () => {} }),
}))

const mockRevalidatePath = mock(() => {})
mock.module("next/cache", () => ({
  revalidatePath: mockRevalidatePath,
}))

// ── Import actions AFTER mocks ───────────────────────────────────────────

import {
  actionCreateArea,
  actionUpdateArea,
  actionDeleteArea,
  actionSetAreaCollections,
  actionAddUserToArea,
  actionRemoveUserFromArea,
} from "@/app/actions/areas"

// ── Reset mocks between tests ────────────────────────────────────────────

beforeEach(() => {
  mockCreateArea.mockClear()
  mockUpdateArea.mockClear()
  mockDeleteArea.mockClear()
  mockCountUsersInArea.mockClear()
  mockSetAreaCollections.mockClear()
  mockAddUserArea.mockClear()
  mockRemoveUserArea.mockClear()
  mockRevalidatePath.mockClear()

  // Reset to defaults
  mockCountUsersInArea.mockImplementation(() => Promise.resolve(0))
})

// ── Tests ────────────────────────────────────────────────────────────────

describe("actionCreateArea", () => {
  test("valid input returns created area", async () => {
    const result = await actionCreateArea({ name: "operations" })

    expect(result?.data).toBeDefined()
    expect(result?.data?.id).toBe(1)
    expect(result?.data?.name).toBe("operations")
    expect(mockCreateArea).toHaveBeenCalledTimes(1)
    expect(mockCreateArea).toHaveBeenCalledWith("operations", "")
  })

  test("passes optional description", async () => {
    await actionCreateArea({ name: "logistics", description: "Logistics team" })

    expect(mockCreateArea).toHaveBeenCalledWith("logistics", "Logistics team")
  })

  test("revalidates admin paths after creation", async () => {
    await actionCreateArea({ name: "operations" })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/areas")
    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/permissions")
  })

  test("validation error when name is too short (less than 2 chars)", async () => {
    const result = await actionCreateArea({ name: "a" })

    expect(result?.validationErrors).toBeDefined()
    expect(mockCreateArea).not.toHaveBeenCalled()
  })

  test("validation error when name is missing", async () => {
    // @ts-expect-error -- intentionally passing invalid input
    const result = await actionCreateArea({})

    expect(result?.validationErrors).toBeDefined()
    expect(mockCreateArea).not.toHaveBeenCalled()
  })
})

describe("actionUpdateArea", () => {
  test("valid update returns updated area", async () => {
    const result = await actionUpdateArea({
      id: 1,
      data: { name: "updated-area", description: "new desc" },
    })

    expect(result?.data).toBeDefined()
    expect(result?.data?.name).toBe("updated-area")
    expect(mockUpdateArea).toHaveBeenCalledTimes(1)
  })

  test("passes cleaned data (strips undefined)", async () => {
    await actionUpdateArea({ id: 1, data: { name: "renamed" } })

    // The clean() function strips undefined values
    const callArgs = mockUpdateArea.mock.calls[0]
    expect(callArgs?.[0]).toBe(1)
    // Second arg should be a clean object without undefined values
    const data = callArgs?.[1] as Record<string, unknown>
    expect(data?.["name"]).toBe("renamed")
    expect("description" in (data ?? {})).toBe(false)
  })

  test("revalidates /admin/areas", async () => {
    await actionUpdateArea({ id: 1, data: { name: "renamed" } })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/areas")
  })
})

describe("actionDeleteArea", () => {
  test("succeeds when area has no users", async () => {
    mockCountUsersInArea.mockImplementation(() => Promise.resolve(0))

    const result = await actionDeleteArea({ id: 5 })

    expect(mockCountUsersInArea).toHaveBeenCalledWith(5)
    expect(mockDeleteArea).toHaveBeenCalledWith(5)
    // Server error should not be present
    expect(result?.serverError).toBeUndefined()
  })

  test("revalidates both /admin/areas and /admin/permissions on success", async () => {
    await actionDeleteArea({ id: 5 })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/areas")
    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/permissions")
  })

  test("returns server error when area has assigned users", async () => {
    mockCountUsersInArea.mockImplementation(() => Promise.resolve(2))

    const result = await actionDeleteArea({ id: 3 })

    expect(result?.serverError).toBeDefined()
    expect(result?.serverError).toContain("2")
    expect(result?.serverError).toContain("usuario")
    expect(mockDeleteArea).not.toHaveBeenCalled()
  })
})

describe("actionSetAreaCollections", () => {
  test("sets collections for area", async () => {
    const collections = [
      { name: "docs", permission: "read" as const },
      { name: "contracts", permission: "write" as const },
    ]

    const result = await actionSetAreaCollections({ areaId: 1, collections })

    expect(mockSetAreaCollections).toHaveBeenCalledWith(1, collections)
    expect(result?.serverError).toBeUndefined()
  })

  test("revalidates admin paths", async () => {
    await actionSetAreaCollections({
      areaId: 1,
      collections: [{ name: "docs", permission: "read" }],
    })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/areas")
    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/permissions")
  })

  test("accepts empty collections array", async () => {
    const result = await actionSetAreaCollections({ areaId: 1, collections: [] })

    expect(mockSetAreaCollections).toHaveBeenCalledWith(1, [])
    expect(result?.serverError).toBeUndefined()
  })

  test("validation error for invalid permission value", async () => {
    const result = await actionSetAreaCollections({
      areaId: 1,
      // @ts-expect-error -- intentionally passing invalid permission
      collections: [{ name: "docs", permission: "superadmin" }],
    })

    expect(result?.validationErrors).toBeDefined()
    expect(mockSetAreaCollections).not.toHaveBeenCalled()
  })
})

describe("actionAddUserToArea", () => {
  test("adds user to area", async () => {
    const result = await actionAddUserToArea({ userId: 42, areaId: 3 })

    expect(mockAddUserArea).toHaveBeenCalledWith(42, 3)
    expect(result?.serverError).toBeUndefined()
  })

  test("revalidates /admin/areas", async () => {
    await actionAddUserToArea({ userId: 42, areaId: 3 })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/areas")
  })
})

describe("actionRemoveUserFromArea", () => {
  test("removes user from area", async () => {
    const result = await actionRemoveUserFromArea({ userId: 42, areaId: 3 })

    expect(mockRemoveUserArea).toHaveBeenCalledWith(42, 3)
    expect(result?.serverError).toBeUndefined()
  })

  test("revalidates /admin/areas", async () => {
    await actionRemoveUserFromArea({ userId: 42, areaId: 3 })

    expect(mockRevalidatePath).toHaveBeenCalledWith("/admin/areas")
  })
})
