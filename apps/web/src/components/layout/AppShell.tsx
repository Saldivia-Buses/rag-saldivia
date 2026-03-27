import type { CurrentUser } from "@/lib/auth/current-user"
import type { DbChatSession, DbProject } from "@rag-saldivia/db"
import { AppShellChrome } from "./AppShellChrome"

type Changelog = { version: string; entries: { version: string; content: string }[] }

/**
 * Server Component — pasa datos pre-fetched a AppShellChrome.
 * Todo el estado de UI vive en AppShellChrome.
 */
export function AppShell({
  user,
  children,
  initialSessions,
  initialProjects,
  changelog,
}: {
  user: CurrentUser
  children: React.ReactNode
  initialSessions: Pick<DbChatSession, "id" | "title" | "collection">[]
  initialProjects: DbProject[]
  changelog: Changelog
}) {
  return (
    <AppShellChrome
      user={user}
      initialSessions={initialSessions}
      initialProjects={initialProjects}
      changelog={changelog}
    >
      {children}
    </AppShellChrome>
  )
}
