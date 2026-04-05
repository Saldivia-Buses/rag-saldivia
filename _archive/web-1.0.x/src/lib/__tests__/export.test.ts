/**
 * Tests de las utilidades de export de sesión.
 * Corre con: bun test apps/web/src/lib/__tests__/export.test.ts
 */

import { describe, test, expect } from "bun:test"
import { exportToMarkdown } from "../export"
import type { Citation } from "@rag-saldivia/shared"

const BASE_SESSION = {
  title: "Análisis Legal",
  collection: "legal-docs",
  createdAt: new Date("2026-03-25T10:00:00Z").getTime(),
  messages: [] as Array<{ role: "user" | "assistant"; content: string; sources?: Citation[] }>,
}

describe("exportToMarkdown", () => {
  test("incluye el título de la sesión como heading", () => {
    const md = exportToMarkdown({ ...BASE_SESSION })
    expect(md).toContain("# Análisis Legal")
  })

  test("incluye el nombre de la colección", () => {
    const md = exportToMarkdown({ ...BASE_SESSION })
    expect(md).toContain("legal-docs")
  })

  test("serializa mensajes del usuario con prefijo 'Usuario:'", () => {
    const session = {
      ...BASE_SESSION,
      messages: [{ role: "user" as const, content: "¿Qué dice el contrato?" }],
    }
    const md = exportToMarkdown(session)
    expect(md).toContain("**Usuario:** ¿Qué dice el contrato?")
  })

  test("serializa mensajes del asistente con prefijo 'Asistente:'", () => {
    const session = {
      ...BASE_SESSION,
      messages: [{ role: "assistant" as const, content: "El contrato establece..." }],
    }
    const md = exportToMarkdown(session)
    expect(md).toContain("**Asistente:** El contrato establece...")
  })

  test("incluye fuentes cuando existen", () => {
    const session = {
      ...BASE_SESSION,
      messages: [
        {
          role: "assistant" as const,
          content: "Según el artículo 3...",
          sources: [{ document: "contrato-2024.pdf" }],
        },
      ],
    }
    const md = exportToMarkdown(session)
    expect(md).toContain("contrato-2024.pdf")
    expect(md).toContain("Fuentes:")
  })

  test("no incluye sección de fuentes cuando el array está vacío", () => {
    const session = {
      ...BASE_SESSION,
      messages: [
        { role: "assistant" as const, content: "Respuesta sin fuentes", sources: [] },
      ],
    }
    const md = exportToMarkdown(session)
    expect(md).not.toContain("Fuentes:")
  })

  test("maneja sesión con mensajes vacíos sin error", () => {
    const md = exportToMarkdown({ ...BASE_SESSION, messages: [] })
    expect(md).toContain("# Análisis Legal")
    expect(md).not.toContain("**Usuario:**")
    expect(md).not.toContain("**Asistente:**")
  })

  test("serializa múltiples mensajes en orden", () => {
    const session = {
      ...BASE_SESSION,
      messages: [
        { role: "user" as const, content: "Pregunta 1" },
        { role: "assistant" as const, content: "Respuesta 1" },
        { role: "user" as const, content: "Pregunta 2" },
      ],
    }
    const md = exportToMarkdown(session)
    const posUser1 = md.indexOf("Pregunta 1")
    const posAssist = md.indexOf("Respuesta 1")
    const posUser2 = md.indexOf("Pregunta 2")

    expect(posUser1).toBeLessThan(posAssist)
    expect(posAssist).toBeLessThan(posUser2)
  })

  test("retorna string no vacío en todos los casos", () => {
    const md = exportToMarkdown(BASE_SESSION)
    expect(typeof md).toBe("string")
    expect(md.length).toBeGreaterThan(0)
  })
})
