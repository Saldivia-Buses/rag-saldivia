import { requireUser } from "@/lib/auth/current-user"
import { AppShell } from "@/components/layout/AppShell"

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const user = await requireUser()

  return <AppShell user={user}>{children}</AppShell>
}
