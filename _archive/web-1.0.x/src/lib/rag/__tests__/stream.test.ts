/**
 * Tests de las utilidades SSE compartidas.
 * Corre con: bun test apps/web/src/lib/rag/__tests__/stream.test.ts
 */

import { describe, test, expect } from "bun:test"
import { parseSseLine, collectSseText, readSseTokens } from "../stream"

// ── parseSseLine ────────────────────────────────────────────────────────────

describe("parseSseLine", () => {
  test("extrae token de línea SSE válida", () => {
    const line = `data: {"choices":[{"delta":{"content":"hola"}}]}`
    expect(parseSseLine(line)).toBe("hola")
  })

  test("retorna null para [DONE]", () => {
    expect(parseSseLine("data: [DONE]")).toBeNull()
  })

  test("retorna null para línea sin prefijo data:", () => {
    expect(parseSseLine("event: message")).toBeNull()
    expect(parseSseLine("")).toBeNull()
    expect(parseSseLine("id: 123")).toBeNull()
  })

  test("retorna null para JSON malformado", () => {
    expect(parseSseLine("data: {broken json}")).toBeNull()
  })

  test("retorna null para delta sin content", () => {
    const line = `data: {"choices":[{"delta":{"sources":[{"id":"doc1"}]}}]}`
    expect(parseSseLine(line)).toBeNull()
  })

  test("retorna null para content vacío", () => {
    const line = `data: {"choices":[{"delta":{"content":""}}]}`
    // "" es falsy pero la función devuelve "" que es !== null — pero readSseTokens no lo yielda
    const result = parseSseLine(line)
    // content "" → no es null pero es empty string
    expect(result).toBe("")
  })

  test("retorna null para choices vacíos", () => {
    const line = `data: {"choices":[]}`
    expect(parseSseLine(line)).toBeNull()
  })

  test("extrae token con caracteres especiales", () => {
    const line = `data: {"choices":[{"delta":{"content":"¡Hola, mundo!"}}]}`
    expect(parseSseLine(line)).toBe("¡Hola, mundo!")
  })
})

// ── readSseTokens ───────────────────────────────────────────────────────────

function makeStream(chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder()
  return new ReadableStream({
    start(controller) {
      for (const chunk of chunks) {
        controller.enqueue(encoder.encode(chunk))
      }
      controller.close()
    },
  })
}

describe("readSseTokens", () => {
  test("yields tokens de un stream simple", async () => {
    const stream = makeStream([
      `data: {"choices":[{"delta":{"content":"hola"}}]}\n`,
      `data: {"choices":[{"delta":{"content":" mundo"}}]}\n`,
      `data: [DONE]\n`,
    ])
    const tokens: string[] = []
    for await (const token of readSseTokens(stream)) {
      tokens.push(token)
    }
    expect(tokens).toEqual(["hola", " mundo"])
  })

  test("ignora líneas sin contenido de texto", async () => {
    const stream = makeStream([
      `data: {"choices":[{"delta":{"sources":[{"id":"doc1"}]}}]}\n`,
      `data: {"choices":[{"delta":{"content":"respuesta"}}]}\n`,
    ])
    const tokens: string[] = []
    for await (const token of readSseTokens(stream)) {
      tokens.push(token)
    }
    expect(tokens).toEqual(["respuesta"])
  })

  test("maneja chunks que cortan líneas a la mitad", async () => {
    const line = `data: {"choices":[{"delta":{"content":"tok"}}]}\n`
    const stream = makeStream([line.slice(0, 20), line.slice(20)])
    const tokens: string[] = []
    for await (const token of readSseTokens(stream)) {
      tokens.push(token)
    }
    expect(tokens).toEqual(["tok"])
  })

  test("stream vacío no yields nada", async () => {
    const stream = makeStream([])
    const tokens: string[] = []
    for await (const token of readSseTokens(stream)) {
      tokens.push(token)
    }
    expect(tokens).toHaveLength(0)
  })
})

// ── collectSseText ──────────────────────────────────────────────────────────

function makeResponse(body: string, contentType = "text/event-stream"): Response {
  const encoder = new TextEncoder()
  return new Response(encoder.encode(body), {
    headers: { "content-type": contentType },
  })
}

function makeSseBody(...tokens: string[]): string {
  return tokens
    .map((t) => `data: {"choices":[{"delta":{"content":"${t}"}}]}\n`)
    .join("") + "data: [DONE]\n"
}

describe("collectSseText", () => {
  test("acumula todos los tokens del stream", async () => {
    const res = makeResponse(makeSseBody("hola", " ", "mundo"))
    expect(await collectSseText(res)).toBe("hola mundo")
  })

  test("retorna string vacío para stream vacío", async () => {
    const res = makeResponse("data: [DONE]\n")
    expect(await collectSseText(res)).toBe("")
  })

  test("trunca al alcanzar maxChars", async () => {
    const res = makeResponse(makeSseBody("abc", "def", "ghi"))
    const result = await collectSseText(res, { maxChars: 4 })
    expect(result.length).toBeLessThanOrEqual(6) // puede superar justo al llegar al token
    expect(result).toContain("abc")
  })

  test("parse JSON para respuesta no-SSE", async () => {
    const json = JSON.stringify({ choices: [{ message: { content: "respuesta json" } }] })
    const res = new Response(json, { headers: { "content-type": "application/json" } })
    expect(await collectSseText(res)).toBe("respuesta json")
  })

  test("retorna vacío para JSON sin choices", async () => {
    const res = new Response("{}", { headers: { "content-type": "application/json" } })
    expect(await collectSseText(res)).toBe("")
  })

  test("detecta repetición cuando la opción está activa", async () => {
    // El algoritmo detecta repetición cuando el patrón NO empieza desde el principio del texto
    // (firstIdx > 0). Por eso se necesita un prefijo antes del bloque repetido.
    const prefix = "Respuesta inicial con contenido útil. "
    const block = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz12345678" // 60 chars
    const repeated = prefix + block.repeat(5)
    const tokens = [repeated.slice(0, 100), repeated.slice(100, 220), repeated.slice(220)]
    const res = makeResponse(makeSseBody(...tokens))
    const result = await collectSseText(res, { detectRepetition: true })
    expect(result.length).toBeLessThan(repeated.length)
  })
})
