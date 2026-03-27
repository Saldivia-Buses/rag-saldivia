/**
 * Tests del cliente Redis singleton.
 * Usa ioredis-mock (activado via bunfig.toml preload) — no requiere Redis real.
 */

import { describe, test, expect, afterEach } from "bun:test"
import { getRedisClient, _resetRedisForTesting } from "../redis"

afterEach(() => {
  _resetRedisForTesting()
})

describe("getRedisClient", () => {
  test("lanza error claro si REDIS_URL no está configurado", () => {
    const orig = process.env["REDIS_URL"]
    delete process.env["REDIS_URL"]
    expect(() => getRedisClient()).toThrow("REDIS_URL no configurado")
    process.env["REDIS_URL"] = orig ?? "redis://localhost:6379"
  })

  test("retorna instancia Redis cuando REDIS_URL está configurado", () => {
    process.env["REDIS_URL"] = "redis://localhost:6379"
    const client = getRedisClient()
    expect(client).toBeDefined()
    expect(typeof client.get).toBe("function")
  })

  test("retorna el mismo singleton en llamadas sucesivas", () => {
    process.env["REDIS_URL"] = "redis://localhost:6379"
    const c1 = getRedisClient()
    const c2 = getRedisClient()
    expect(c1).toBe(c2)
  })

  test("_resetRedisForTesting permite crear una nueva instancia", () => {
    process.env["REDIS_URL"] = "redis://localhost:6379"
    const c1 = getRedisClient()
    _resetRedisForTesting()
    const c2 = getRedisClient()
    expect(c1).not.toBe(c2)
  })
})
