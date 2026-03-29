import chalk from "chalk"
import { out, statusBadge, makeTable } from "../output.js"
import { SERVER_URL } from "../client.js"

const SERVICES = [
  { name: "Next.js server", url: `${SERVER_URL}/api/health`, critical: true },
  { name: "RAG Server", url: `${process.env["RAG_SERVER_URL"] ?? "http://localhost:8081"}/health`, critical: true },
  { name: "Milvus", url: `${process.env["MILVUS_URL"] ?? "http://localhost:9091"}/healthz`, critical: false },
  { name: "Mode Manager", url: `${process.env["MODE_MANAGER_URL"] ?? "http://localhost:8082"}/health`, critical: false },
  { name: "OpenRouter Proxy", url: `${process.env["OPENROUTER_PROXY_URL"] ?? "http://localhost:8083"}/health`, critical: false },
]

export async function statusCommand() {
  out.section("Estado del sistema")

  const results = await Promise.allSettled(
    SERVICES.map(async (svc) => {
      const start = Date.now()
      try {
        const res = await fetch(svc.url, { signal: AbortSignal.timeout(3000) })
        return { ...svc, ok: res.ok, latency: Date.now() - start }
      } catch {
        return { ...svc, ok: false, latency: Date.now() - start }
      }
    })
  )

  const rows: string[][] = []
  let criticalOk = 0
  let criticalTotal = 0

  for (const result of results) {
    const svc = result.status === "fulfilled" ? result.value : { ...SERVICES[0]!, ok: false, latency: 0 }

    if (svc.critical) {
      criticalTotal++
      if (svc.ok) criticalOk++
    }

    rows.push([
      chalk.bold(svc.name),
      statusBadge(svc.ok, svc.latency),
      svc.critical ? "" : chalk.dim("opcional"),
    ])
  }

  console.log(makeTable(["Servicio", "Estado", ""], rows))
  console.log("")
  console.log(
    `  ${criticalOk === criticalTotal ? chalk.green("✓") : chalk.yellow("⚠")} ${chalk.bold(`${criticalOk}/${criticalTotal}`)} servicios críticos operativos`
  )
  console.log("")
}
