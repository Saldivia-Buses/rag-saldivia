/**
 * Tests de dispatchWebhook — firma HMAC, headers, manejo de errores.
 * Corre con: bun test apps/web/src/lib/__tests__/webhook.test.ts
 */

import { describe, test, expect, beforeEach, mock, spyOn } from "bun:test"
import { createHmac } from "crypto"

// Mock del logger para evitar efectos secundarios en tests
mock.module("@rag-saldivia/logger/backend", () => ({
  log: {
    info: () => {},
    warn: () => {},
    error: () => {},
  },
}))

// Mock de @rag-saldivia/db para evitar inicialización de DB en tests
mock.module("@rag-saldivia/db", () => ({
  listWebhooksByEvent: async () => [],
}))

import { dispatchWebhook } from "../webhook"
import type { DbWebhook } from "@rag-saldivia/db"

const MOCK_WEBHOOK: DbWebhook = {
  id: "wh-test-123",
  userId: 1,
  url: "https://example.com/hook",
  events: ["ingestion.completed"],
  secret: "supersecret",
  active: true,
  createdAt: Date.now(),
}

describe("dispatchWebhook — headers", () => {
  test("envía header X-Signature con firma HMAC-SHA256 correcta", async () => {
    let capturedHeaders: Record<string, string> = {}

    const mockFetch = spyOn(globalThis, "fetch").mockImplementationOnce(async (_, init) => {
      capturedHeaders = Object.fromEntries(new Headers(init?.headers as HeadersInit).entries())
      return new Response(null, { status: 200 })
    })

    const payload = { event: "ingestion.completed", jobId: "job-1" }
    await dispatchWebhook(MOCK_WEBHOOK, payload)

    const body = JSON.stringify({ ...payload, timestamp: expect.any(Number) })
    expect(capturedHeaders["x-signature"]).toMatch(/^sha256=[a-f0-9]{64}$/)
    expect(capturedHeaders["x-webhook-id"]).toBe("wh-test-123")
    expect(capturedHeaders["content-type"]).toBe("application/json")

    mockFetch.mockRestore()
  })

  test("la firma HMAC-SHA256 puede verificarse con el secret", async () => {
    let capturedSignature = ""
    let capturedBody = ""

    const mockFetch = spyOn(globalThis, "fetch").mockImplementationOnce(async (_, init) => {
      capturedBody = init?.body as string
      const headers = new Headers(init?.headers as HeadersInit)
      capturedSignature = headers.get("x-signature") ?? ""
      return new Response(null, { status: 200 })
    })

    await dispatchWebhook(MOCK_WEBHOOK, { evento: "test" })

    const expectedSig = "sha256=" + createHmac("sha256", MOCK_WEBHOOK.secret).update(capturedBody).digest("hex")
    expect(capturedSignature).toBe(expectedSig)

    mockFetch.mockRestore()
  })

  test("el body incluye el payload original más 'timestamp'", async () => {
    let parsedBody: Record<string, unknown> = {}

    const mockFetch = spyOn(globalThis, "fetch").mockImplementationOnce(async (_, init) => {
      parsedBody = JSON.parse(init?.body as string)
      return new Response(null, { status: 200 })
    })

    await dispatchWebhook(MOCK_WEBHOOK, { jobId: "abc", status: "done" })

    expect(parsedBody["jobId"]).toBe("abc")
    expect(parsedBody["status"]).toBe("done")
    expect(typeof parsedBody["timestamp"]).toBe("number")

    mockFetch.mockRestore()
  })
})

describe("dispatchWebhook — manejo de errores", () => {
  test("no lanza si fetch retorna status no-ok (4xx)", async () => {
    const mockFetch = spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(null, { status: 400, statusText: "Bad Request" })
    )

    await expect(dispatchWebhook(MOCK_WEBHOOK, { event: "test" })).resolves.toBeUndefined()
    mockFetch.mockRestore()
  })

  test("no lanza si fetch retorna 500", async () => {
    const mockFetch = spyOn(globalThis, "fetch").mockResolvedValueOnce(
      new Response(null, { status: 500 })
    )

    await expect(dispatchWebhook(MOCK_WEBHOOK, { event: "test" })).resolves.toBeUndefined()
    mockFetch.mockRestore()
  })

  test("no lanza si fetch lanza una excepción (timeout, red)", async () => {
    const mockFetch = spyOn(globalThis, "fetch").mockRejectedValueOnce(
      new Error("Network error")
    )

    await expect(dispatchWebhook(MOCK_WEBHOOK, { event: "test" })).resolves.toBeUndefined()
    mockFetch.mockRestore()
  })

  test("no lanza si fetch lanza AbortError (timeout de 5s)", async () => {
    const abortError = new DOMException("The operation was aborted", "AbortError")
    const mockFetch = spyOn(globalThis, "fetch").mockRejectedValueOnce(abortError)

    await expect(dispatchWebhook(MOCK_WEBHOOK, { event: "test" })).resolves.toBeUndefined()
    mockFetch.mockRestore()
  })
})

describe("dispatchWebhook — método HTTP", () => {
  test("usa método POST", async () => {
    let capturedMethod = ""

    const mockFetch = spyOn(globalThis, "fetch").mockImplementationOnce(async (_, init) => {
      capturedMethod = init?.method ?? ""
      return new Response(null, { status: 200 })
    })

    await dispatchWebhook(MOCK_WEBHOOK, { event: "test" })
    expect(capturedMethod).toBe("POST")

    mockFetch.mockRestore()
  })
})
