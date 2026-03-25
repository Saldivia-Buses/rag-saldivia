/**
 * Tests de FOCUS_MODES.
 * Corre con: bun test packages/shared/src/__tests__/focus-modes.test.ts
 */

import { describe, test, expect } from "bun:test"
import { FOCUS_MODES, FOCUS_MODE_IDS } from "../schemas"

describe("FOCUS_MODES estructura", () => {
  test("tiene exactamente 4 modos", () => {
    expect(FOCUS_MODES.length).toBe(4)
  })

  test("contiene todos los IDs definidos en FOCUS_MODE_IDS", () => {
    const ids = FOCUS_MODES.map((m) => m.id)
    for (const id of FOCUS_MODE_IDS) {
      expect(ids).toContain(id)
    }
  })

  test("todos los IDs son únicos", () => {
    const ids = FOCUS_MODES.map((m) => m.id)
    const unique = new Set(ids)
    expect(unique.size).toBe(FOCUS_MODES.length)
  })

  test("todos los modos tienen label no vacío", () => {
    for (const mode of FOCUS_MODES) {
      expect(mode.label).toBeTruthy()
      expect(mode.label.length).toBeGreaterThan(0)
    }
  })

  test("todos los modos tienen systemPrompt no vacío", () => {
    for (const mode of FOCUS_MODES) {
      expect(mode.systemPrompt).toBeTruthy()
      expect(mode.systemPrompt.length).toBeGreaterThan(10)
    }
  })

  test("modo 'ejecutivo' tiene system prompt orientado a brevedad", () => {
    const ejecutivo = FOCUS_MODES.find((m) => m.id === "ejecutivo")
    expect(ejecutivo).toBeDefined()
    // El prompt ejecutivo debe mencionar algo sobre resumen o concisión
    const prompt = ejecutivo!.systemPrompt.toLowerCase()
    expect(prompt.includes("concis") || prompt.includes("summary") || prompt.includes("brief") || prompt.includes("executive")).toBe(true)
  })
})
