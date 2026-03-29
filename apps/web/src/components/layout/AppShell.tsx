import type { CurrentUser } from "@/lib/auth/current-user"
import { AppShellChrome } from "./AppShellChrome"

type Changelog = { version: string; entries: { version: string; content: string }[] }

/**
 * Server Component — pasa datos pre-fetched a AppShellChrome.
 */
export function AppShell({
  user,
  children,
  changelog,
}: {
  user: CurrentUser
  children: React.ReactNode
  changelog: Changelog
}) {
  return (
    <AppShellChrome
      user={user}
      changelog={changelog}
    >
      {children}
    </AppShellChrome>
  )
}
