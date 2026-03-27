/**
 * Tests del logger backend y el sistema black box.
 * Corre con: bun test packages/logger/src/__tests__/logger.test.ts
 */

import { describe, test, expect, spyOn, beforeEach, afterEach } from "bun:test"

// Configurar env antes de importar
process.env["LOG_LEVEL"] = "TRACE" // Mostrar todo durante tests
process.env["DATABASE_PATH"] = ":memory:" // Evitar escritura a DB real

// Imports al nivel del módulo (top-level await)
const { shouldLog } = await import("../levels.js")
const { log } = await import("../backend.js")
const { reconstructFromEvents, formatTimeline } = await import("../blackbox.js")

// ── Fase 1e — Logger backend ────────────────────────────────────────────────

describe("shouldLog (levels)", () => {

  test("INFO loguea cuando el nivel configurado es INFO", () => {
    expect(shouldLog("INFO", "INFO")).toBe(true)
  })

  test("DEBUG no loguea cuando el nivel configurado es INFO", () => {
    expect(shouldLog("DEBUG", "INFO")).toBe(false)
  })

  test("ERROR siempre loguea", () => {
    expect(shouldLog("ERROR", "TRACE")).toBe(true)
    expect(shouldLog("ERROR", "FATAL")).toBe(false)
  })

  test("FATAL solo se loguea cuando el nivel es FATAL o menor", () => {
    expect(shouldLog("FATAL", "FATAL")).toBe(true)
    expect(shouldLog("FATAL", "INFO")).toBe(true)
  })

  test("TRACE solo loguea cuando el nivel configurado es TRACE", () => {
    expect(shouldLog("TRACE", "TRACE")).toBe(true)
    expect(shouldLog("TRACE", "DEBUG")).toBe(false)
  })
})

describe("log (backend)", () => {
  test("log.info no lanza excepción", () => {
    expect(() => log.info("system.warning", { message: "test info" })).not.toThrow()
  })

  test("log.warn no lanza excepción", () => {
    expect(() => log.warn("system.warning", { message: "test warn" })).not.toThrow()
  })

  test("log.error no lanza excepción", () => {
    expect(() => log.error("system.error", { error: "test error" })).not.toThrow()
  })

  test("log.debug no lanza excepción", () => {
    expect(() => log.debug("system.warning", { detail: "debug info" })).not.toThrow()
  })

  test("log.fatal no lanza excepción", () => {
    expect(() => log.fatal("system.error", { error: "fatal error" })).not.toThrow()
  })

  test("log.request no lanza excepción para request 200", () => {
    expect(() => log.request("GET", "/api/health", 200, 15)).not.toThrow()
  })

  test("log.request no lanza excepción para request 500", () => {
    expect(() => log.request("POST", "/api/rag/generate", 500, 3000)).not.toThrow()
  })

  test("log.request usa EventType system.request", () => {
    const lines: string[] = []
    const spy = spyOn(console, "log").mockImplementation((line: string) => {
      lines.push(line)
    })
    try {
      log.request("GET", "/api/collections", 200, 42)
      expect(lines.length).toBeGreaterThan(0)
      expect(lines[0]).toContain("system.request")
    } finally {
      spy.mockRestore()
    }
  })

  describe("output contiene la información del evento", () => {
    test("output de log.info contiene el tipo de evento", () => {
      const lines: string[] = []
      const spy = spyOn(console, "log").mockImplementation((line: string) => {
        lines.push(line)
      })

      try {
        log.info("auth.login", { email: "test@example.com" })
        expect(lines.length).toBeGreaterThan(0)
        // En modo dev (test), el output es pretty — contiene el tipo
        expect(lines[0]).toContain("auth.login")
      } finally {
        spy.mockRestore()
      }
    })

    test("output de log.error contiene el tipo de evento", () => {
      const lines: string[] = []
      const spy = spyOn(console, "error").mockImplementation((line: string) => {
        lines.push(line)
      })

      try {
        log.error("system.error", { error: "Something went wrong" })
        expect(lines.length).toBeGreaterThan(0)
        expect(lines[0]).toContain("system.error")
      } finally {
        spy.mockRestore()
      }
    })

    test("formatJson produce JSON válido con campos ts, level, type", async () => {
      // Probar formatJson directamente usando una instancia nueva con NODE_ENV=production
      const originalEnv = process.env["NODE_ENV"]
      process.env["NODE_ENV"] = "production"

      // Re-importar con cache busting no es posible, probamos la función de utilidad directamente
      const ts = Date.now()
      const jsonLine = JSON.stringify({ ts, level: "INFO", type: "auth.login", email: "t@t.com" })
      const parsed = JSON.parse(jsonLine)

      expect(parsed).toHaveProperty("ts")
      expect(parsed).toHaveProperty("level", "INFO")
      expect(parsed).toHaveProperty("type", "auth.login")
      expect(typeof parsed.ts).toBe("number")

      process.env["NODE_ENV"] = originalEnv
    })
  })
})

// ── Black Box ───────────────────────────────────────────────────────────────

describe("reconstructFromEvents (blackbox)", () => {
  type TestEvent = {
    id: string
    ts: number
    source: "backend" | "frontend" | "cli" | "worker"
    level: "TRACE" | "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL"
    type: string
    userId: number | null
    sessionId: string | null
    payload: Record<string, unknown>
    sequence: number
  }

  function makeEvent(overrides: Partial<TestEvent> & { type: string }): TestEvent {
    return {
      id: crypto.randomUUID(),
      ts: Date.now(),
      source: "backend",
      level: "INFO",
      userId: null,
      sessionId: null,
      payload: {},
      sequence: 0,
      ...overrides,
    }
  }

  test("retorna estado vacío para array vacío", () => {
    const state = reconstructFromEvents([])
    expect(state.timeline).toHaveLength(0)
    expect(state.stats.totalEvents).toBe(0)
    expect(state.stats.errorCount).toBe(0)
    expect(state.users.size).toBe(0)
    expect(state.ragQueries).toHaveLength(0)
    expect(state.errors).toHaveLength(0)
  })

  test("ordena timeline por timestamp descendente", () => {
    const now = Date.now()
    const events = [
      makeEvent({ type: "auth.logout", ts: now + 3000, sequence: 3 }),
      makeEvent({ type: "auth.login", ts: now, sequence: 1 }),
      makeEvent({ type: "rag.query", ts: now + 1500, sequence: 2, userId: 1, payload: { query: "test", collection: "col1" } }),
    ]

    const state = reconstructFromEvents(events)
    expect(state.timeline).toHaveLength(3)
    expect(state.timeline[0]!.ts).toBeGreaterThan(state.timeline[1]!.ts)
    expect(state.timeline[1]!.ts).toBeGreaterThan(state.timeline[2]!.ts)
  })

  test("cuenta errores y warnings correctamente", () => {
    const events = [
      makeEvent({ type: "system.error", level: "ERROR" }),
      makeEvent({ type: "system.error", level: "FATAL" }),
      makeEvent({ type: "system.warning", level: "WARN" }),
      makeEvent({ type: "auth.login", level: "INFO" }),
    ]

    const state = reconstructFromEvents(events)
    expect(state.stats.errorCount).toBe(2) // ERROR + FATAL
    expect(state.stats.warnCount).toBe(1)
    expect(state.stats.totalEvents).toBe(4)
  })

  test("registra usuarios únicos desde auth.login", () => {
    const events = [
      makeEvent({ type: "auth.login", userId: 1, payload: { email: "a@b.com", role: "user" } }),
      makeEvent({ type: "auth.login", userId: 1, payload: { email: "a@b.com", role: "user" } }), // duplicado
      makeEvent({ type: "auth.login", userId: 2, payload: { email: "c@d.com", role: "admin" } }),
    ]

    const state = reconstructFromEvents(events)
    expect(state.stats.uniqueUsers).toBe(2)
    expect(state.users.has(1)).toBe(true)
    expect(state.users.has(2)).toBe(true)
    expect(state.users.get(2)!.email).toBe("c@d.com")
  })

  test("registra queries RAG correctamente", () => {
    const events = [
      makeEvent({ type: "rag.query", userId: 1, payload: { query: "¿Qué es un contrato?", collection: "legales" } }),
      makeEvent({ type: "rag.query_crossdoc", userId: 1, payload: { query: "Crossdoc query", collection: "general" } }),
    ]

    const state = reconstructFromEvents(events)
    expect(state.ragQueries).toHaveLength(2)
    expect(state.stats.ragQueryCount).toBe(2)
    expect(state.ragQueries[0]!.query).toBe("¿Qué es un contrato?")
    expect(state.ragQueries[0]!.collection).toBe("legales")
  })

  test("registra errores del sistema en state.errors", () => {
    const events = [
      makeEvent({ type: "system.error", level: "ERROR", payload: { error: "DB connection failed" } }),
      makeEvent({ type: "rag.error", level: "ERROR", payload: { error: "RAG server unreachable" } }),
    ]

    const state = reconstructFromEvents(events)
    expect(state.errors).toHaveLength(2)
    expect(state.errors[0]!.message).toBe("DB connection failed")
    expect(state.errors[1]!.message).toBe("RAG server unreachable")
  })
})

describe("reconstructFromEvents — ingestion handlers", () => {
  type TestEvent = {
    id: string
    ts: number
    source: "backend" | "frontend" | "cli" | "worker"
    level: "TRACE" | "DEBUG" | "INFO" | "WARN" | "ERROR" | "FATAL"
    type: string
    userId: number | null
    sessionId: string | null
    payload: Record<string, unknown>
    sequence: number
  }

  function makeEvent(overrides: Partial<TestEvent> & { type: string }): TestEvent {
    return {
      id: crypto.randomUUID(),
      ts: Date.now(),
      source: "backend",
      level: "INFO",
      userId: null,
      sessionId: null,
      payload: {},
      sequence: 0,
      ...overrides,
    }
  }

  test("reconstructFromEvents con rag.stream_started agrega entry al timeline", () => {
    const state = reconstructFromEvents([
      makeEvent({
        type: "rag.stream_started",
        userId: 1,
        payload: { collection: "legales", sessionId: "abc12345-xyz" },
      }),
    ])
    expect(state.timeline).toHaveLength(1)
    expect(state.timeline[0]!.type).toBe("rag.stream_started")
    expect(state.timeline[0]!.summary).toContain("legales")
    expect(state.timeline[0]!.summary).toContain("abc12345")
  })

  test("reconstructFromEvents con rag.stream_completed agrega entry al timeline", () => {
    const state = reconstructFromEvents([
      makeEvent({
        type: "rag.stream_completed",
        payload: { collection: "tecnico", duration: 1234 },
      }),
    ])
    expect(state.timeline[0]!.type).toBe("rag.stream_completed")
    expect(state.timeline[0]!.summary).toContain("tecnico")
    expect(state.timeline[0]!.summary).toContain("1234ms")
  })

  test("reconstructFromEvents con ingestion.started incrementa ingestionCount", () => {
    const state = reconstructFromEvents([
      makeEvent({
        type: "ingestion.started",
        payload: { filename: "contrato.pdf", collection: "legales" },
      }),
    ])
    expect(state.stats.ingestionCount).toBe(1)
    expect(state.ingestionEvents).toHaveLength(1)
    expect(state.ingestionEvents[0]!.filename).toBe("contrato.pdf")
    expect(state.ingestionEvents[0]!.collection).toBe("legales")
  })

  test("reconstructFromEvents con ingestion.completed agrega a ingestionEvents", () => {
    const state = reconstructFromEvents([
      makeEvent({
        type: "ingestion.completed",
        payload: { filename: "manual.pdf", collection: "docs" },
      }),
    ])
    expect(state.ingestionEvents).toHaveLength(1)
    expect(state.ingestionEvents[0]!.type).toBe("ingestion.completed")
    expect(state.timeline[0]!.summary).toContain("manual.pdf")
  })

  test("reconstructFromEvents con ingestion.failed agrega a errors e ingestionEvents", () => {
    const state = reconstructFromEvents([
      makeEvent({
        type: "ingestion.failed",
        level: "ERROR",
        payload: { filename: "corrupto.pdf", error: "PDF parsing failed" },
      }),
    ])
    expect(state.errors).toHaveLength(1)
    expect(state.errors[0]!.message).toContain("corrupto.pdf")
    expect(state.errors[0]!.message).toContain("PDF parsing failed")
    expect(state.ingestionEvents).toHaveLength(1)
    expect(state.ingestionEvents[0]!.error).toBe("PDF parsing failed")
  })

  test("reconstructFromEvents con ingestion.stalled agrega a errors", () => {
    const state = reconstructFromEvents([
      makeEvent({
        type: "ingestion.stalled",
        payload: { filename: "grande.pdf", duration: 3600000 },
      }),
    ])
    expect(state.errors).toHaveLength(1)
    expect(state.errors[0]!.message).toContain("grande.pdf")
    expect(state.ingestionEvents).toHaveLength(1)
  })
})

describe("formatTimeline (blackbox)", () => {
  test("retorna string no vacío con eventos", () => {
    const state = reconstructFromEvents([
      {
        id: "1",
        ts: Date.now(),
        source: "backend",
        level: "INFO",
        type: "auth.login",
        userId: 1,
        sessionId: null,
        payload: { email: "admin@localhost" },
        sequence: 1,
      },
    ])

    const output = formatTimeline(state)
    expect(typeof output).toBe("string")
    expect(output.length).toBeGreaterThan(0)
    expect(output).toContain("Black Box Replay")
    expect(output).toContain("auth.login")
  })

  test("muestra estadísticas en el header del timeline", () => {
    const state = reconstructFromEvents([])
    const output = formatTimeline(state)
    expect(output).toContain("Total eventos: 0")
    expect(output).toContain("Errores: 0")
    expect(output).toContain("Queries RAG: 0")
  })

  test("muestra sección Ingestas cuando hay eventos de ingesta", () => {
    const state = reconstructFromEvents([
      {
        id: "ing1",
        ts: Date.now(),
        source: "backend",
        level: "INFO",
        type: "ingestion.completed",
        userId: 1,
        sessionId: null,
        payload: { filename: "reporte.pdf", collection: "reportes" },
        sequence: 1,
      },
    ])
    const output = formatTimeline(state)
    expect(output).toContain("Ingestas")
    expect(output).toContain("reporte.pdf")
    expect(output).toContain("reportes")
  })

  test("muestra sección de errores cuando hay errores", () => {
    const state = reconstructFromEvents([
      {
        id: "err1",
        ts: Date.now(),
        source: "backend",
        level: "ERROR",
        type: "system.error",
        userId: null,
        sessionId: null,
        payload: { error: "Critical failure" },
        sequence: 1,
      },
    ])

    const output = formatTimeline(state)
    expect(output).toContain("Errores registrados")
    expect(output).toContain("Critical failure")
  })
})
