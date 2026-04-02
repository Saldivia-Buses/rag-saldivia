import { describe, test, expect } from "bun:test"
import { getErrorRecovery, parseUseChatError } from "../error-recovery"

describe("getErrorRecovery", () => {
  // --- By code ---

  test("UNAVAILABLE returns server unavailable recovery", () => {
    const r = getErrorRecovery({ code: "UNAVAILABLE" })
    expect(r.title).toBe("Servidor no disponible")
    expect(r.icon).toBe("unavailable")
    expect(r.actions.some((a) => a.type === "retry")).toBe(true)
  })

  test("ECONNREFUSED in message maps to unavailable", () => {
    const r = getErrorRecovery({ message: "fetch failed: ECONNREFUSED" })
    expect(r.icon).toBe("unavailable")
  })

  test("TIMEOUT returns timeout recovery", () => {
    const r = getErrorRecovery({ code: "TIMEOUT" })
    expect(r.title).toBe("La consulta tardó demasiado")
    expect(r.icon).toBe("timeout")
  })

  test("RATE_LIMITED with details shows minutes and max count", () => {
    const r = getErrorRecovery({
      code: "RATE_LIMITED",
      details: { retryAfterMs: 1_800_000, maxCount: 10 },
    })
    expect(r.title).toBe("Límite de consultas alcanzado")
    expect(r.description).toContain("10")
    expect(r.suggestion).toContain("30 minutos")
    expect(r.icon).toBe("rate-limit")
  })

  test("RATE_LIMITED without details uses 60min fallback", () => {
    const r = getErrorRecovery({ code: "RATE_LIMITED" })
    expect(r.suggestion).toContain("60 minutos")
  })

  test("FORBIDDEN with collection keyword shows collection recovery", () => {
    const r = getErrorRecovery({ code: "FORBIDDEN", message: "Sin acceso a la colección 'test'" })
    expect(r.title).toBe("Sin acceso a la colección")
    expect(r.actions.some((a) => a.href === "/collections")).toBe(true)
  })

  test("FORBIDDEN without collection shows generic forbidden", () => {
    const r = getErrorRecovery({ code: "FORBIDDEN", message: "No autorizado" })
    expect(r.title).toBe("Acceso denegado")
  })

  test("UPSTREAM_ERROR returns upstream recovery with report action", () => {
    const r = getErrorRecovery({ code: "UPSTREAM_ERROR" })
    expect(r.title).toBe("Error en el servidor de búsqueda")
    expect(r.actions.some((a) => a.type === "report")).toBe(true)
    expect(r.actions.some((a) => a.type === "retry")).toBe(true)
  })

  // --- By status ---

  test("401 returns session expired", () => {
    const r = getErrorRecovery({ status: 401 })
    expect(r.title).toBe("Sesión expirada")
    expect(r.actions.some((a) => a.href === "/login")).toBe(true)
    expect(r.icon).toBe("auth")
  })

  test("404 returns not found", () => {
    const r = getErrorRecovery({ status: 404 })
    expect(r.title).toBe("No encontrado")
    expect(r.icon).toBe("not-found")
  })

  test("429 status maps to rate limited", () => {
    const r = getErrorRecovery({ status: 429 })
    expect(r.icon).toBe("rate-limit")
  })

  test("503 returns service unavailable", () => {
    const r = getErrorRecovery({ status: 503 })
    expect(r.title).toBe("Servicio no disponible")
  })

  // --- By message pattern ---

  test("SQLITE_BUSY maps to database busy", () => {
    const r = getErrorRecovery({ message: "SQLITE_BUSY: database is locked" })
    expect(r.title).toBe("Base de datos ocupada")
  })

  test("collection not found maps to not found", () => {
    const r = getErrorRecovery({ message: "Colección no encontrada" })
    expect(r.title).toBe("Colección no encontrada")
    expect(r.actions.some((a) => a.href === "/collections")).toBe(true)
  })

  // --- Fallback ---

  test("unknown error returns generic recovery", () => {
    const r = getErrorRecovery({ message: "something weird happened" })
    expect(r.title).toBe("Error inesperado")
    expect(r.icon).toBe("generic")
  })

  test("empty error returns generic recovery", () => {
    const r = getErrorRecovery({})
    expect(r.title).toBe("Error inesperado")
  })

  test("every recovery has at least one action", () => {
    const cases = [
      { code: "UNAVAILABLE" },
      { code: "TIMEOUT" },
      { code: "RATE_LIMITED" },
      { code: "FORBIDDEN" },
      { code: "UPSTREAM_ERROR" },
      { status: 401 },
      { status: 404 },
      {},
    ] as const

    for (const input of cases) {
      const r = getErrorRecovery(input)
      expect(r.actions.length).toBeGreaterThan(0)
    }
  })
})

describe("parseUseChatError", () => {
  test("extracts status from error.status", () => {
    const err = Object.assign(new Error("failed"), { status: 503 })
    const result = parseUseChatError(err)
    expect(result.status).toBe(503)
  })

  test("does not guess status from message text (avoids false positives)", () => {
    const err = new Error("Server returned 429")
    const result = parseUseChatError(err)
    expect(result.status).toBeUndefined()
  })

  test("uses error.status when available", () => {
    const err = Object.assign(new Error("rate limited"), { status: 429 })
    const result = parseUseChatError(err)
    expect(result.status).toBe(429)
    expect(result.code).toBe("RATE_LIMITED")
  })

  test("detects ECONNREFUSED as UNAVAILABLE", () => {
    const err = new Error("fetch failed: ECONNREFUSED")
    const result = parseUseChatError(err)
    expect(result.code).toBe("UNAVAILABLE")
  })

  test("detects timeout keyword", () => {
    const err = new Error("La consulta tardó demasiado")
    const result = parseUseChatError(err)
    expect(result.code).toBe("TIMEOUT")
  })

  test("detects límite keyword as RATE_LIMITED", () => {
    const err = new Error("Límite de 10 queries/hora alcanzado")
    const result = parseUseChatError(err)
    expect(result.code).toBe("RATE_LIMITED")
  })

  test("returns message as-is for unknown errors", () => {
    const err = new Error("something weird")
    const result = parseUseChatError(err)
    expect(result.message).toBe("something weird")
    expect(result.code).toBeUndefined()
  })
})
