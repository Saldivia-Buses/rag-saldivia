/**
 * Tests del parser de CHANGELOG.
 * Corre con: bun test apps/web/src/lib/__tests__/changelog.test.ts
 */

import { describe, test, expect } from "bun:test"
import { parseChangelog } from "../changelog"

const SAMPLE_CHANGELOG = `# Changelog

## [Unreleased]

### Added
- Nueva feature A
- Nueva feature B

### Fixed
- Bug corregido

## [0.2.0] - 2026-03-20

### Added
- Feature de la v0.2

## [0.1.0] - 2026-03-01

### Added
- Versión inicial
`

describe("parseChangelog", () => {
  test("parsea la sección [Unreleased] correctamente", () => {
    const entries = parseChangelog(SAMPLE_CHANGELOG)
    expect(entries[0]!.version).toBe("Unreleased")
  })

  test("parsea versiones con formato [vX.Y.Z]", () => {
    const entries = parseChangelog(SAMPLE_CHANGELOG)
    const versions = entries.map((e) => e.version)
    expect(versions).toContain("0.2.0")
    expect(versions).toContain("0.1.0")
  })

  test("incluye el contenido de cada sección", () => {
    const entries = parseChangelog(SAMPLE_CHANGELOG)
    const unreleased = entries.find((e) => e.version === "Unreleased")
    expect(unreleased).toBeDefined()
    expect(unreleased!.content).toContain("Nueva feature A")
  })

  test("respeta el límite de entradas", () => {
    const entries = parseChangelog(SAMPLE_CHANGELOG, 2)
    expect(entries.length).toBeLessThanOrEqual(2)
  })

  test("retorna array vacío para changelog vacío", () => {
    const entries = parseChangelog("")
    expect(entries).toBeArray()
    expect(entries.length).toBe(0)
  })

  test("mantiene el orden original del changelog", () => {
    const entries = parseChangelog(SAMPLE_CHANGELOG)
    expect(entries[0]!.version).toBe("Unreleased")
    expect(entries[1]!.version).toBe("0.2.0")
    expect(entries[2]!.version).toBe("0.1.0")
  })
})
