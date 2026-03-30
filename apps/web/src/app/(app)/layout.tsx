import { requireUser } from "@/lib/auth/current-user"
import { getUserById } from "@rag-saldivia/db"
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
  const fullUser = await getUserById(user.id)
  const prefs = fullUser?.preferences as Record<string, unknown> | undefined
  const defaultCollection = (prefs?.defaultCollection as string) ?? undefined

  return (
    <AppShell user={user} changelog={changelog} defaultCollection={defaultCollection}>
      <div data-density="spacious" className="h-full">
        {children}
      </div>
    </AppShell>
  )
}
