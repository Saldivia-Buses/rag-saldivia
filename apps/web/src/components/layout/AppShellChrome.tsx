"use client"

import { NavRail } from "./NavRail"
import type { CurrentUser } from "@/lib/auth/current-user"

type Changelog = { version: string; entries: { version: string; content: string }[] }

type Props = {
  user: CurrentUser
  children: React.ReactNode
  changelog: Changelog
}

export function AppShellChrome({
  user,
  children,
  changelog,
}: Props) {
  return (
    <div className="flex h-screen overflow-hidden bg-bg">
      <NavRail user={user} changelog={changelog} />
      <main className="flex-1 overflow-y-auto bg-bg">{children}</main>
    </div>
  )
}
