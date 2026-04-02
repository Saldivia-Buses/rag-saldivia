/**
 * @rag-saldivia/config — Config Loader
 *
 * Reemplaza saldivia/config.py (220 líneas Python).
 * Lee los YAMLs existentes en config/, los valida con Zod,
 * aplica overrides de variables de entorno, y exporta config tipada.
 */

import { readFileSync, existsSync } from "fs"
import { join } from "path"
import { parse as parseYaml } from "yaml"
import { z } from "zod"
import { RagParamsSchema } from "@rag-saldivia/shared"

// ── Schemas de configuración ────────────────────────────────────────────────

const ServiceLlmSchema = z.object({
  provider: z.string().default("nvidia-api"),
  model: z.string(),
  parameters: z.object({
    max_tokens: z.number().int().default(4096),
  }).optional(),
})

const ServiceCrossdocSchema = z.object({
  decomposition: z.object({
    provider: z.string().default("openrouter"),
    model: z.string(),
  }).optional(),
})

const IngestionTierConfigSchema = z.object({
  max_pages: z.number().int().nullable().default(null),
  poll_interval: z.number().int().default(10),
  deadlock_threshold: z.number().int().default(60),
  timeout: z.number().int().default(900),
})

const IngestionConfigSchema = z.object({
  parallel_slots_small: z.number().int().default(2),
  parallel_slots_large: z.number().int().default(1),
  client_max_retries: z.number().int().default(3),
  server_max_retries: z.number().int().default(3),
  retry_backoff_base: z.number().int().default(30),
  stall_check_interval: z.number().int().default(60),
  tiers: z.object({
    tiny: IngestionTierConfigSchema.optional(),
    small: IngestionTierConfigSchema.optional(),
    medium: IngestionTierConfigSchema.optional(),
    large: IngestionTierConfigSchema.optional(),
  }).default({}),
})

const ProfileConfigSchema = z.object({
  mode: z.object({
    gpu_memory_gb: z.number().default(96),
    idle_timeout: z.number().int().default(300),
  }).optional(),
  services: z.object({
    llm: ServiceLlmSchema.optional(),
    crossdoc: ServiceCrossdocSchema.optional(),
  }).optional(),
  ingestion: IngestionConfigSchema.optional(),
})

// ── Config completa del sistema ─────────────────────────────────────────────

export const AppConfigSchema = z.object({
  profile: z.string().default("workstation-1gpu"),
  // Servidor
  port: z.number().int().default(3000),
  nodeEnv: z.enum(["development", "production", "test"]).default("development"),
  // Auth
  jwtSecret: z.string().min(16),
  jwtExpiry: z.string().default("24h"),
  systemApiKey: z.string().min(8),
  // DB
  databasePath: z.string().default("./data/app.db"),
  // RAG
  ragServerUrl: z.string().url().default("http://localhost:8081"),
  ragTimeoutMs: z.number().int().default(120000),
  mockRag: z.boolean().default(false),
  // CORS
  corsOrigins: z.array(z.string()).default(["http://localhost:3000"]),
  // Logging
  logLevel: z.enum(["TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"]).default("INFO"),
  // NVIDIA
  ngcApiKey: z.string().optional(),
  nvidiaApiKey: z.string().optional(),
  // Profile-specific
  profileConfig: ProfileConfigSchema.optional(),
  ragParams: RagParamsSchema.optional(),
})

export type AppConfig = z.infer<typeof AppConfigSchema>

// ── Loader ─────────────────────────────────────────────────────────────────

function deepMerge(base: Record<string, unknown>, override: Record<string, unknown>): Record<string, unknown> {
  const result = { ...base }
  for (const [key, value] of Object.entries(override)) {
    if (value !== null && typeof value === "object" && !Array.isArray(value) &&
        key in result && typeof result[key] === "object" && !Array.isArray(result[key])) {
      result[key] = deepMerge(result[key] as Record<string, unknown>, value as Record<string, unknown>)
    } else {
      result[key] = value
    }
  }
  return result
}

function loadYamlFile(path: string): Record<string, unknown> {
  if (!existsSync(path)) return {}
  try {
    const content = readFileSync(path, "utf-8")
    return parseYaml(content) ?? {}
  } catch (e) {
    console.warn(`[config] Failed to parse YAML at ${path}:`, e)
    return {}
  }
}

function parseEnvBoolean(val: string | undefined): boolean | undefined {
  if (val === undefined) return undefined
  return val.toLowerCase() === "true" || val === "1"
}

export function loadConfig(configDir?: string): AppConfig {
  const root = configDir ?? join(process.cwd(), "config")
  const profile = process.env["RAG_PROFILE"] ?? "workstation-1gpu"

  // Cargar YAML del perfil
  const profilePath = join(root, "profiles", `${profile}.yaml`)
  const profileYaml = loadYamlFile(profilePath)

  // Construir config desde env vars (tienen precedencia sobre YAML)
  const corsOrigins = process.env["CORS_ORIGINS"]
    ? process.env["CORS_ORIGINS"].split(",").map((s) => s.trim())
    : ["http://localhost:3000"]

  const raw = {
    profile,
    port: process.env["PORT"] ? parseInt(process.env["PORT"]) : 3000,
    nodeEnv: (process.env["NODE_ENV"] as AppConfig["nodeEnv"]) ?? "development",
    jwtSecret: process.env["JWT_SECRET"] ?? (process.env["NODE_ENV"] === "development" ? "dev-secret-changeme-in-production" : ""),
    jwtExpiry: process.env["JWT_EXPIRY"] ?? "24h",
    systemApiKey: process.env["SYSTEM_API_KEY"] ?? (process.env["NODE_ENV"] === "development" ? "dev-system-key-changeme" : ""),
    databasePath: process.env["DATABASE_PATH"] ?? "./data/app.db",
    ragServerUrl: process.env["RAG_SERVER_URL"] ?? "http://localhost:8081",
    ragTimeoutMs: process.env["RAG_TIMEOUT_MS"] ? parseInt(process.env["RAG_TIMEOUT_MS"]) : 120000,
    mockRag: parseEnvBoolean(process.env["MOCK_RAG"]) ?? false,
    corsOrigins,
    logLevel: (process.env["LOG_LEVEL"] as AppConfig["logLevel"]) ?? "INFO",
    ngcApiKey: process.env["NGC_API_KEY"],
    nvidiaApiKey: process.env["NVIDIA_API_KEY"],
    profileConfig: profileYaml,
  }

  const result = AppConfigSchema.safeParse(raw)

  if (!result.success) {
    const issues = result.error.issues
      .map((i) => `  - ${i.path.join(".")}: ${i.message}`)
      .join("\n")

    if (process.env["NODE_ENV"] === "production") {
      throw new Error(`Config inválida:\n${issues}`)
    } else {
      console.warn(`[config] Advertencias de configuración:\n${issues}`)
      // En dev, retornar config parcial con defaults
      return result.error.flatten().fieldErrors as unknown as AppConfig
    }
  }

  return result.data
}

// ── Admin overrides (RAG params) ────────────────────────────────────────────
// Los admins pueden cambiar parámetros RAG en runtime.
// Se guardan en config/admin-overrides.yaml (excluido del repo por .gitignore).

export function loadRagParams(configDir?: string): z.infer<typeof RagParamsSchema> {
  const root = configDir ?? join(process.cwd(), "config")
  const overridesPath = join(root, "admin-overrides.yaml")
  const overrides = loadYamlFile(overridesPath)

  const defaults = RagParamsSchema.parse({})
  const merged = deepMerge(defaults as Record<string, unknown>, (overrides["rag_params"] as Record<string, unknown>) ?? {})

  return RagParamsSchema.parse(merged)
}

export async function saveRagParams(
  params: Partial<z.infer<typeof RagParamsSchema>>,
  configDir?: string
): Promise<void> {
  const root = configDir ?? join(process.cwd(), "config")
  const overridesPath = join(root, "admin-overrides.yaml")
  const { stringify } = await import("yaml")
  const { writeFileSync } = await import("fs")
  const { mkdirSync } = await import("fs")

  mkdirSync(root, { recursive: true })

  const existing = loadYamlFile(overridesPath)
  const updated = deepMerge(existing, { rag_params: params })

  writeFileSync(overridesPath, stringify(updated), "utf-8")
}
