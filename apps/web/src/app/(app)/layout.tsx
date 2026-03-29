import { requireUser } from "@/lib/auth/current-user"
import { AppShell } from "@/components/layout/AppShell"
import { parseChangelog } from "@/lib/changelog"
import { readFileSync } from "fs"
import { join } from "path"

function loadChangelog() {
  try {
    const changelogPath = join(process.cwd(), "../../CHANGELOG.md")
    const pkgPath = join(process.cwd(), "../../package.json")
    const raw = readFileSync(changelogPath, "utf-8")
    const entries = parseChangelog(raw)
    const pkg = JSON.parse(readFileSync(pkgPath, "utf-8")) as { version?: string }
    return { version: pkg.version ?? "0.1.0", entries }
  } catch {
    return { version: "0.1.0", entries: [] }
  }
}

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const user = await requireUser()
  const changelog = loadChangelog()

  return (
    <AppShell user={user} changelog={changelog}>
      <div data-density="spacious" className="h-full">
        {children}
      </div>
    </AppShell>
  )
}
