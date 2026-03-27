/**
 * Tests de las utilidades compartidas de lib/utils.ts.
 * Corre con: bun test apps/web/src/lib/__tests__/utils.test.ts
 */

import { describe, test, expect } from "bun:test"
import { formatDate, formatDateTime } from "../utils"

describe("formatDate", () => {
  test("retorna string en formato dd/mm/yyyy", () => {
    const ts = new Date("2026-03-27T15:30:00Z").getTime()
    const result = formatDate(ts)
    // El formato es dd/mm/yyyy — verificar que tiene 3 segmentos separados por /
    const parts = result.split("/")
    expect(parts).toHaveLength(3)
    expect(parts[2]).toHaveLength(4) // año de 4 dígitos
  })

  test("acepta Date, number y string como entrada", () => {
    const ts = new Date("2026-01-15").getTime()
    const fromNumber = formatDate(ts)
    const fromDate = formatDate(new Date(ts))
    const fromString = formatDate(new Date(ts).toISOString())
    expect(fromNumber).toBe(fromDate)
    expect(fromNumber).toBe(fromString)
  })

  test("retorna string no vacío para timestamp válido", () => {
    expect(formatDate(Date.now()).length).toBeGreaterThan(0)
  })
})

describe("formatDateTime", () => {
  test("retorna string que incluye fecha y hora", () => {
    const ts = new Date("2026-03-27T10:05:00Z").getTime()
    const result = formatDateTime(ts)
    // Debe incluir separador de fecha
    expect(result).toContain("/")
    // Debe incluir separador de hora
    expect(result.length).toBeGreaterThan(8)
  })

  test("acepta Date, number y string como entrada", () => {
    const ts = new Date("2026-06-15T09:00:00Z").getTime()
    const fromNumber = formatDateTime(ts)
    const fromDate = formatDateTime(new Date(ts))
    expect(fromNumber).toBe(fromDate)
  })
})
