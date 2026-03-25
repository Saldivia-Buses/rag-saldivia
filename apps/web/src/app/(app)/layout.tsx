import { requireUser } from "@/lib/auth/current-user"
import { AppShell } from "@/components/layout/AppShell"
import { OnboardingTour } from "@/components/onboarding/OnboardingTour"
import { getDb, users } from "@rag-saldivia/db"
import { eq } from "drizzle-orm"

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const user = await requireUser()

  // Leer onboardingCompleted directamente de la DB (el campo puede no estar en CurrentUser)
  const db = getDb()
  const [dbUser] = await db.select({ onboardingCompleted: users.onboardingCompleted }).from(users).where(eq(users.id, user.id)).limit(1)
  const onboardingCompleted = dbUser?.onboardingCompleted ?? true

  return (
    <AppShell user={user}>
      <OnboardingTour completed={onboardingCompleted} />
      {children}
    </AppShell>
  )
}
