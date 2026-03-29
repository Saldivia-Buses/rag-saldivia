/**
 * Top-level app shell — Server Component boundary.
 *
 * This is the outermost layout wrapper for all authenticated pages.
 * It exists as a thin Server Component that passes pre-fetched data
 * (user, changelog) down to AppShellChrome, which is a Client Component.
 *
 * The SC/CC split here is intentional: user data is fetched server-side
 * (no client waterfall), while AppShellChrome needs "use client" for
 * the SidebarProvider context and NavRail interactivity.
 *
 * Rendered by: /app/(app)/layout.tsx (the authenticated layout)
 * Depends on: AppShellChrome (Client Component child)
 */
import type { CurrentUser } from "@/lib/auth/current-user"
import { AppShellChrome } from "./AppShellChrome"

type Changelog = { version: string; entries: { version: string; content: string }[] }
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
