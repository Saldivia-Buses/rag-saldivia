/**
 * Configuración de NextAuth v5 para SSO.
 * F3.47 — SSO Google / Azure AD.
 *
 * Modo mixto: usuarios SSO y usuarios con password coexisten.
 * Al autenticarse via SSO se emite el mismo JWT que el flujo de password
 * para compatibilidad con el middleware RBAC existente.
 */

import NextAuth from "next-auth"
import Google from "next-auth/providers/google"
import MicrosoftEntraID from "next-auth/providers/microsoft-entra-id"
import { getDb, users } from "@rag-saldivia/db"
import { and, eq, sql } from "drizzle-orm"
import { createJwt, makeAuthCookie } from "@/lib/auth/jwt"

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [
    Google({
      clientId: process.env["GOOGLE_CLIENT_ID"] ?? "",
      clientSecret: process.env["GOOGLE_CLIENT_SECRET"] ?? "",
    }),
    MicrosoftEntraID({
      clientId: process.env["AZURE_AD_CLIENT_ID"] ?? "",
      clientSecret: process.env["AZURE_AD_CLIENT_SECRET"] ?? "",
      tenantId: process.env["AZURE_AD_TENANT_ID"] ?? "common",
    }),
  ],
  callbacks: {
    async signIn({ user, account }) {
      if (!account || !user.email) return false

      const db = getDb()
      const provider = account.provider  // "google" | "microsoft-entra-id"
      const subject = account.providerAccountId

      // Buscar usuario existente por SSO provider+subject
      let existingUser = await db
        .select()
        .from(users)
        .where(and(eq(users.ssoProvider, provider), eq(users.ssoSubject, subject)))
        .limit(1)
        .then((r) => r[0] ?? null)

      if (!existingUser) {
        // Buscar por email (puede que ya tenga cuenta con password)
        existingUser = await db
          .select()
          .from(users)
          .where(eq(users.email, user.email.toLowerCase()))
          .limit(1)
          .then((r) => r[0] ?? null)

        if (existingUser) {
          // Vincular SSO a la cuenta existente
          await db.update(users)
            .set({ ssoProvider: provider, ssoSubject: subject })
            .where(eq(users.id, existingUser.id))
        } else {
          // Crear nuevo usuario SSO
          const [newUser] = await db.insert(users).values({
            email: user.email.toLowerCase(),
            name: user.name ?? user.email,
            role: "user",
            apiKeyHash: `sso-${subject}`,
            ssoProvider: provider,
            ssoSubject: subject,
            preferences: {},
            active: true,
            onboardingCompleted: false,
            createdAt: Date.now(),
          }).returning()
          existingUser = newUser ?? null
        }
      }

      if (!existingUser || !existingUser.active) return false

      // Guardar el userId en el token para extraerlo en jwt callback
      ;(user as Record<string, unknown>)._systemUserId = existingUser.id
      ;(user as Record<string, unknown>)._systemRole = existingUser.role
      return true
    },

    async jwt({ token, user }) {
      if (user) {
        token.systemUserId = (user as Record<string, unknown>)._systemUserId
        token.systemRole = (user as Record<string, unknown>)._systemRole
      }
      return token
    },

    async session({ session, token }) {
      ;(session as Record<string, unknown>).systemUserId = token.systemUserId
      ;(session as Record<string, unknown>).systemRole = token.systemRole
      return session
    },
  },
  pages: {
    signIn: "/login",
  },
})
