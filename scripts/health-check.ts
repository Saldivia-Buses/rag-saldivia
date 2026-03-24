#!/usr/bin/env bun
/**
 * health-check.ts — reemplaza scripts/health_check.sh
 *
 * Verifica el estado de todos los servicios y retorna exit code 0 si todos
 * los servicios críticos están operativos, 1 si alguno falla.
 *
 * Uso:
 *   bun scripts/health-check.ts
 *   bun scripts/health-check.ts --json    # output en JSON
 */

const services = [
  { name: "Next.js", url: `http://localhost:${process.env["PORT"] ?? 3000}/api/health`, critical: true },
  { name: "RAG Server", url: `${process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"}/health`, critical: true },
  { name: "Milvus", url: "http://localhost:9091/healthz", critical: false },
]

const isJson = process.argv.includes("--json")

type ServiceResult = {
  name: string
  url: string
  ok: boolean
  latency: number
  critical: boolean
  error?: string
}

const results: ServiceResult[] = await Promise.all(
  services.map(async (svc) => {
    const start = Date.now()
    try {
      const res = await fetch(svc.url, { signal: AbortSignal.timeout(3000) })
      return { ...svc, ok: res.ok, latency: Date.now() - start }
    } catch (err) {
      return { ...svc, ok: false, latency: Date.now() - start, error: String(err) }
    }
  })
)

if (isJson) {
  console.log(JSON.stringify(results, null, 2))
} else {
  for (const r of results) {
    const icon = r.ok ? "\x1b[32m✓\x1b[0m" : r.critical ? "\x1b[31m✗\x1b[0m" : "\x1b[33m⚠\x1b[0m"
    const latency = `\x1b[2m(${r.latency}ms)\x1b[0m`
    const tag = r.critical ? "" : " \x1b[2m(opcional)\x1b[0m"
    console.log(`  ${icon} ${r.name.padEnd(15)} ${latency}${tag}`)
    if (!r.ok && r.error) console.log(`      \x1b[33m→ ${r.error}\x1b[0m`)
  }
}

const criticalFailed = results.filter((r) => r.critical && !r.ok)
if (criticalFailed.length > 0) {
  if (!isJson) console.error(`\n  \x1b[31m✗ ${criticalFailed.length} servicio(s) crítico(s) no responden\x1b[0m`)
  process.exit(1)
}

if (!isJson) console.log(`\n  \x1b[32m✓ Todos los servicios críticos operativos\x1b[0m`)
