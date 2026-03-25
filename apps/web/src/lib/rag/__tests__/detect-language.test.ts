/**
 * Tests de detección de idioma para multilenguaje automático.
 * Corre con: bun test apps/web/src/lib/rag/__tests__/detect-language.test.ts
 */

import { describe, test, expect } from "bun:test"
import { detectLanguageHint } from "../client"

const LANG_INSTRUCTION = "Respond in the same language as the user's message."

describe("detectLanguageHint — textos en español (no debe inyectar instrucción)", () => {
  test("texto puramente en español retorna string vacío", () => {
    expect(detectLanguageHint("¿Cuál es el resumen del contrato?")).toBe("")
  })

  test("pregunta técnica en español retorna string vacío", () => {
    expect(detectLanguageHint("Cómo funciona el proceso de ingesta de documentos")).toBe("")
  })

  test("texto muy corto retorna string vacío", () => {
    expect(detectLanguageHint("hola")).toBe("")
  })

  test("string vacío retorna string vacío", () => {
    expect(detectLanguageHint("")).toBe("")
  })

  test("string undefined-like (muy corto) retorna string vacío", () => {
    expect(detectLanguageHint("ok")).toBe("")
  })
})

describe("detectLanguageHint — textos en inglés (debe inyectar instrucción)", () => {
  test("pregunta en inglés con múltiples palabras clave retorna instrucción", () => {
    const hint = detectLanguageHint("What is the summary of the contract?")
    expect(hint).toBe(LANG_INSTRUCTION)
  })

  test("pregunta corta en inglés con una keyword retorna instrucción", () => {
    const hint = detectLanguageHint("How does this work?")
    expect(hint).toBe(LANG_INSTRUCTION)
  })

  test("instrucción en inglés retorna instrucción de idioma", () => {
    const hint = detectLanguageHint("Please list all the documents")
    expect(hint).toBe(LANG_INSTRUCTION)
  })

  test("pregunta técnica en inglés retorna instrucción", () => {
    const hint = detectLanguageHint("Can you find the relevant sections?")
    expect(hint).toBe(LANG_INSTRUCTION)
  })
})

describe("detectLanguageHint — caracteres no-latinos (CJK, cirílico, árabe)", () => {
  test("texto en chino retorna instrucción", () => {
    const hint = detectLanguageHint("你好，请告诉我")
    expect(hint).toBe(LANG_INSTRUCTION)
  })

  test("texto en japonés retorna instrucción", () => {
    const hint = detectLanguageHint("これは何ですか")
    expect(hint).toBe(LANG_INSTRUCTION)
  })

  test("texto en árabe retorna instrucción", () => {
    const hint = detectLanguageHint("ما هو ملخص العقد؟")
    expect(hint).toBe(LANG_INSTRUCTION)
  })

  test("texto en cirílico retorna instrucción", () => {
    const hint = detectLanguageHint("Что такое машинное обучение?")
    expect(hint).toBe(LANG_INSTRUCTION)
  })
})
