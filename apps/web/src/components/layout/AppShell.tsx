import type { CurrentUser } from "@/lib/auth/current-user"
import { AppShellChrome } from "./AppShellChrome"

/**
 * Server Component — solo pasa el user a AppShellChrome.
 * Todo el estado de UI (zen mode, notificaciones, hotkeys) vive en AppShellChrome.
 */
export function AppShell({
  user,
  children,
}: {
  user: CurrentUser
  children: React.ReactNode
}) {
  return <AppShellChrome user={user}>{children}</AppShellChrome>
}
