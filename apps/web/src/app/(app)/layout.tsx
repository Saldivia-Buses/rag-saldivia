import { requireUser } from "@/lib/auth/current-user"
import { AppShell } from "@/components/layout/AppShell"
import { OnboardingTour } from "@/components/onboarding/OnboardingTour"
import { getDb, users, listSessionsByUser, listProjects } from "@rag-saldivia/db"
import { eq } from "drizzle-orm"
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

  const db = getDb()
  const [dbUser, sessions, projects] = await Promise.all([
    db.select({ onboardingCompleted: users.onboardingCompleted }).from(users).where(eq(users.id, user.id)).limit(1),
    listSessionsByUser(user.id),
    listProjects(user.id),
  ])

  const onboardingCompleted = dbUser[0]?.onboardingCompleted ?? true
  const changelog = loadChangelog()

  return (
    <AppShell
      user={user}
      initialSessions={sessions}
      initialProjects={projects}
      changelog={changelog}
    >
      <OnboardingTour completed={onboardingCompleted} />
      {/* Densidad spacious por defecto — los layouts de admin sobrescriben con compact */}
      <div data-density="spacious" className="h-full">
        {children}
      </div>
    </AppShell>
  )
}
