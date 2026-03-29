/**
 * Tests de detección de artifacts en respuestas del asistente.
 * Corre con: bun test apps/web/src/lib/rag/__tests__/detect-artifact.test.ts
 */

import { describe, test, expect } from "bun:test"
import { detectArtifact } from "../detect-artifact"

// Helper para generar un bloque de código con N líneas
function makeCodeBlock(lang: string, lines: number): string {
  const content = Array.from({ length: lines }, (_, i) => `  // line ${i + 1}`).join("\n")
  return `\`\`\`${lang}\n${content}\n\`\`\``
}

// Helper para tabla markdown con N columnas
function makeTable(cols: number): string {
  const headers = Array.from({ length: cols }, (_, i) => `Col${i + 1}`).join(" | ")
  const separator = Array.from({ length: cols }, () => "---").join(" | ")
  const row = Array.from({ length: cols }, (_, i) => `val${i + 1}`).join(" | ")
  return `| ${headers} |\n| ${separator} |\n| ${row} |`
}

describe("detectArtifact — marcador explícito", () => {
  test("detecta marcador :::artifact con type='code' y lang", () => {
    const content = `:::artifact{type="code" lang="typescript"}
const x = 1
:::`
    const result = detectArtifact(content)
    expect(result).not.toBeNull()
    expect(result!.type).toBe("code")
    expect(result!.language).toBe("typescript")
    expect(result!.content).toContain("const x = 1")
  })

  test("detecta marcador :::artifact con type='document' sin lang", () => {
    const content = `:::artifact{type="document"}
Contenido del documento
:::`
    const result = detectArtifact(content)
    expect(result!.type).toBe("document")
    expect(result!.language).toBeUndefined()
  })

  test("detecta marcador :::artifact con type='table'", () => {
    const content = `:::artifact{type="table"}
| A | B |
|---|---|
| 1 | 2 |
:::`
    const result = detectArtifact(content)
    expect(result!.type).toBe("table")
  })

  test("trimea el contenido del artifact", () => {
    const content = `:::artifact{type="document"}

  Contenido con espacios  

:::`
    const result = detectArtifact(content)
    expect(result!.content).toBe("Contenido con espacios")
  })
})

describe("detectArtifact — heurística: bloque de código", () => {
  test("detecta bloque de código con exactamente 40 líneas", () => {
    const content = makeCodeBlock("typescript", 40)
    const result = detectArtifact(content)
    expect(result).not.toBeNull()
    expect(result!.type).toBe("code")
  })

  test("detecta bloque de código con más de 40 líneas", () => {
    const content = makeCodeBlock("python", 60)
    const result = detectArtifact(content)
    expect(result!.type).toBe("code")
    expect(result!.language).toBe("python")
  })

  test("NO detecta bloque de código con menos de 40 líneas", () => {
    const content = makeCodeBlock("javascript", 20)
    const result = detectArtifact(content)
    expect(result).toBeNull()
  })

  test("retorna language undefined para bloque sin lenguaje", () => {
    const content = makeCodeBlock("", 45)
    const result = detectArtifact(content)
    expect(result!.language).toBeUndefined()
  })
})

describe("detectArtifact — heurística: tabla markdown", () => {
  test("detecta tabla con exactamente 5 columnas", () => {
    const content = makeTable(5)
    const result = detectArtifact(content)
    expect(result).not.toBeNull()
    expect(result!.type).toBe("table")
  })

  test("detecta tabla con más de 5 columnas", () => {
    const content = makeTable(8)
    const result = detectArtifact(content)
    expect(result!.type).toBe("table")
  })

  test("NO detecta tabla con menos de 5 columnas", () => {
    const content = makeTable(4)
    const result = detectArtifact(content)
    expect(result).toBeNull()
  })
})

describe("detectArtifact — casos sin artifact", () => {
  test("retorna null para string vacío", () => {
    expect(detectArtifact("")).toBeNull()
  })

  test("retorna null para texto sin artifact", () => {
    const content = "Esta es una respuesta normal del asistente sin ningún artifact especial."
    expect(detectArtifact(content)).toBeNull()
  })

  test("retorna null para bloque de código corto con texto normal", () => {
    const content = "Aquí está el código:\n```typescript\nconst x = 1\n```\nEso es todo."
    expect(detectArtifact(content)).toBeNull()
  })

  test("el marcador explícito tiene prioridad sobre las heurísticas", () => {
    // Tiene tanto un marcador explícito como un bloque de código largo
    const codeBlock = makeCodeBlock("typescript", 50)
    const content = `:::artifact{type="document"}
Un documento
:::

${codeBlock}`
    const result = detectArtifact(content)
    expect(result!.type).toBe("document") // gana el marcador explícito
  })
})
