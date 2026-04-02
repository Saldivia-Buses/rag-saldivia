import { describe, test, expect, beforeEach, afterEach } from "bun:test"
import { join } from "path"
import { loadConfig, loadRagParams, AppConfigSchema } from "../loader"

// Apunta al directorio config/ real del monorepo
const CONFIG_DIR = join(import.meta.dir, "../../../../../config")

// Guarda y restaura las env vars que se tocan en los tests
const ENV_VARS = [
  "JWT_SECRET", "SYSTEM_API_KEY", "NODE_ENV", "RAG_PROFILE",
  "MOCK_RAG", "LOG_LEVEL", "PORT", "RAG_SERVER_URL",
]

let savedEnv: Record<string, string | undefined> = {}

beforeEach(() => {
  for (const key of ENV_VARS) {
    savedEnv[key] = process.env[key]
  }
  // Env mínima válida para todos los tests
  process.env["NODE_ENV"] = "test"
  process.env["JWT_SECRET"] = "test-secret-32chars-minimum-ok!!"
  process.env["SYSTEM_API_KEY"] = "test-api-key-ok"
})

afterEach(() => {
  for (const key of ENV_VARS) {
    if (savedEnv[key] === undefined) {
      delete process.env[key]
    } else {
      process.env[key] = savedEnv[key]
    }
  }
})

// ── loadConfig ───────────────────────────────────────────────────────────────

describe("loadConfig", () => {
  test("carga sin errores con env mínima válida", () => {
    const config = loadConfig(CONFIG_DIR)
    expect(config).toBeDefined()
    expect(config.jwtSecret).toBe("test-secret-32chars-minimum-ok!!")
    expect(config.systemApiKey).toBe("test-api-key-ok")
  })

  test("defaults correctos cuando no se pasan env vars opcionales", () => {
    delete process.env["LOG_LEVEL"]
    delete process.env["PORT"]
    delete process.env["RAG_SERVER_URL"]
    delete process.env["MOCK_RAG"]

    const config = loadConfig(CONFIG_DIR)

    expect(config.logLevel).toBe("INFO")
    expect(config.port).toBe(3000)
    expect(config.ragServerUrl).toBe("http://localhost:8081")
    expect(config.mockRag).toBe(false)
    expect(config.jwtExpiry).toBe("24h")
  })

  test("no retorna undefined en campos con default definido", () => {
    const config = loadConfig(CONFIG_DIR)

    // Ningún campo con default puede ser undefined
    expect(config.logLevel).not.toBeUndefined()
    expect(config.port).not.toBeUndefined()
    expect(config.ragServerUrl).not.toBeUndefined()
    expect(config.mockRag).not.toBeUndefined()
    expect(config.databasePath).not.toBeUndefined()
    expect(config.jwtExpiry).not.toBeUndefined()
    expect(config.corsOrigins).not.toBeUndefined()
  })

  test("env vars tienen precedencia sobre los defaults", () => {
    process.env["LOG_LEVEL"] = "DEBUG"
    process.env["PORT"] = "4000"
    process.env["MOCK_RAG"] = "true"
    process.env["RAG_SERVER_URL"] = "http://localhost:9999"

    const config = loadConfig(CONFIG_DIR)

    expect(config.logLevel).toBe("DEBUG")
    expect(config.port).toBe(4000)
    expect(config.mockRag).toBe(true)
    expect(config.ragServerUrl).toBe("http://localhost:9999")
  })

  test("MOCK_RAG=false se parsea correctamente como boolean false", () => {
    process.env["MOCK_RAG"] = "false"
    const config = loadConfig(CONFIG_DIR)
    expect(config.mockRag).toBe(false)
  })

  test("carga el perfil workstation-1gpu del YAML sin errores", () => {
    process.env["RAG_PROFILE"] = "workstation-1gpu"
    const config = loadConfig(CONFIG_DIR)
    expect(config.profile).toBe("workstation-1gpu")
  })

  test("perfil inexistente no lanza error — usa config vacía", () => {
    process.env["RAG_PROFILE"] = "perfil-que-no-existe"
    expect(() => loadConfig(CONFIG_DIR)).not.toThrow()
  })

  test("en producción lanza error con JWT_SECRET faltante", () => {
    process.env["NODE_ENV"] = "production"
    delete process.env["JWT_SECRET"]
    delete process.env["SYSTEM_API_KEY"]

    expect(() => loadConfig(CONFIG_DIR)).toThrow("Config inválida")
  })
})

// ── loadRagParams ─────────────────────────────────────────────────────────────

describe("loadRagParams", () => {
  test("retorna defaults correctos sin overrides", () => {
    // Apunta a un dir temporal vacío para no leer admin-overrides.yaml real
    const params = loadRagParams("/tmp/rag-test-config-dir-inexistente")

    expect(params.temperature).toBe(0.2)
    expect(params.top_p).toBe(0.7)
    expect(params.max_tokens).toBe(1024)
    expect(params.vdb_top_k).toBe(10)
    expect(params.reranker_top_k).toBe(5)
    expect(params.use_guardrails).toBe(false)
    expect(params.use_reranker).toBe(true)
    expect(params.chunk_size).toBe(512)
    expect(params.chunk_overlap).toBe(50)
    expect(params.embedding_model).toBe("nvidia/nv-embedqa-e5-v5")
  })

  test("ningún campo de RagParams es undefined", () => {
    const params = loadRagParams("/tmp/rag-test-config-dir-inexistente")

    for (const [key, value] of Object.entries(params)) {
      expect(value, `${key} no debe ser undefined`).not.toBeUndefined()
    }
  })
})

// ── AppConfigSchema ───────────────────────────────────────────────────────────

describe("AppConfigSchema", () => {
  test("valida objeto mínimo correctamente", () => {
    const result = AppConfigSchema.safeParse({
      jwtSecret: "valid-secret-at-least-16",
      systemApiKey: "valid-api-key",
    })
    expect(result.success).toBe(true)
  })

  test("falla con jwtSecret demasiado corto", () => {
    const result = AppConfigSchema.safeParse({
      jwtSecret: "corto",
      systemApiKey: "valid-api-key",
    })
    expect(result.success).toBe(false)
  })

  test("falla con logLevel inválido", () => {
    const result = AppConfigSchema.safeParse({
      jwtSecret: "valid-secret-at-least-16",
      systemApiKey: "valid-api-key",
      logLevel: "VERBOSE",
    })
    expect(result.success).toBe(false)
  })

  test("falla con ragServerUrl que no es URL válida", () => {
    const result = AppConfigSchema.safeParse({
      jwtSecret: "valid-secret-at-least-16",
      systemApiKey: "valid-api-key",
      ragServerUrl: "not-a-url",
    })
    expect(result.success).toBe(false)
  })
})
