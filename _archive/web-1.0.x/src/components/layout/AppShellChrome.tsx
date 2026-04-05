/**
 * Client-side app chrome — wraps all authenticated page content.
 *
 * Provides three things:
 *   1. SidebarProvider — context for sidebar open/closed state across the app
 *   2. NavRail — the 64px vertical navigation strip on the left
 *   3. Main content area — flex-1 scrollable region for page content
 *
 * Layout: `flex h-screen` → NavRail (fixed 64px) + main (flex-1, scrollable)
 *
 * This is a Client Component ("use client") because SidebarProvider uses
 * useState + localStorage + keyboard listeners. The parent AppShell is a
 * Server Component that fetches user data before passing it here.
 *
 * Rendered by: AppShell (Server Component)
 * Depends on: NavRail, SidebarProvider (ChatLayout)
 */
"use client"

import { NavRail } from "./NavRail"
import { SidebarProvider } from "@/components/chat/ChatLayout"
import type { CurrentUser } from "@/lib/auth/current-user"

type Changelog = { version: string; entries: { version: string; content: string }[] }

type Props = {
  user: CurrentUser
  children: React.ReactNode
  changelog: Changelog
  defaultCollection?: string
}

export function AppShellChrome({
  user,
  children,
  changelog,
  defaultCollection,
}: Props) {
  return (
    <SidebarProvider>
      <div className="flex h-screen overflow-hidden bg-bg">
        <NavRail user={user} changelog={changelog} {...(defaultCollection ? { defaultCollection } : {})} />
        <main className="flex-1 overflow-y-auto bg-bg">{children}</main>
      </div>
    </SidebarProvider>
  )
}
