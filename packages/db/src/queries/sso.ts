/**
 * SSO provider CRUD queries.
 *
 * Secrets are encrypted at rest with AES-256-GCM via crypto.ts.
 * The login page only sees public fields (id, name, type).
 */

import { eq } from "drizzle-orm"
import { getDb } from "../connection"
import { ssoProviders } from "../schema"
import type { NewSsoProvider } from "../schema"
import { encryptSecret, decryptSecret } from "../crypto"

type SsoProviderInput = Omit<NewSsoProvider, "id" | "createdAt" | "updatedAt" | "clientSecretEncrypted"> & {
  clientSecret: string
}

/** List active providers — public fields only (for login page). */
export async function listPublicSsoProviders() {
  const db = getDb()
  const rows = await db.select({
    id: ssoProviders.id,
    name: ssoProviders.name,
    type: ssoProviders.type,
  }).from(ssoProviders).where(eq(ssoProviders.active, true))
  return rows
}

/** List all providers with full details (for admin). Decrypts secrets. */
export async function listAllSsoProviders() {
  const db = getDb()
  const rows = await db.select().from(ssoProviders)
  return rows.map((r) => ({
    ...r,
    clientSecret: decryptSecret(r.clientSecretEncrypted),
    clientSecretEncrypted: undefined,
  }))
}

/** Get a single provider by type. Decrypts secret. */
export async function getSsoProviderByType(type: "google" | "microsoft" | "github" | "oidc_generic" | "saml") {
  const db = getDb()
  const row = await db.select().from(ssoProviders).where(eq(ssoProviders.type, type)).limit(1)
  const r = row[0]
  if (!r) return null
  return {
    ...r,
    clientSecret: decryptSecret(r.clientSecretEncrypted),
    clientSecretEncrypted: undefined,
  }
}

/** Get a single provider by ID. Decrypts secret. */
export async function getSsoProviderById(id: number) {
  const db = getDb()
  const row = await db.select().from(ssoProviders).where(eq(ssoProviders.id, id)).limit(1)
  const r = row[0]
  if (!r) return null
  return {
    ...r,
    clientSecret: decryptSecret(r.clientSecretEncrypted),
    clientSecretEncrypted: undefined,
  }
}

/** Create a new SSO provider. Encrypts client secret. */
export async function createSsoProvider(data: SsoProviderInput) {
  const db = getDb()
  const now = Date.now()
  const [row] = await db.insert(ssoProviders).values({
    name: data.name,
    type: data.type,
    clientId: data.clientId,
    clientSecretEncrypted: encryptSecret(data.clientSecret),
    tenantId: data.tenantId,
    issuerUrl: data.issuerUrl,
    scopes: data.scopes ?? "openid email profile",
    autoProvision: data.autoProvision ?? true,
    defaultRole: data.defaultRole ?? "user",
    active: data.active ?? true,
    createdAt: now,
    updatedAt: now,
  }).returning()
  return row!
}

/** Update an existing SSO provider. Re-encrypts secret if provided. */
export async function updateSsoProvider(
  id: number,
  data: Partial<Omit<SsoProviderInput, "type">> & { clientSecret?: string | undefined }
) {
  const db = getDb()
  const updates: Record<string, unknown> = { updatedAt: Date.now() }
  if (data.name !== undefined) updates.name = data.name
  if (data.clientId !== undefined) updates.clientId = data.clientId
  if (data.clientSecret !== undefined) updates.clientSecretEncrypted = encryptSecret(data.clientSecret)
  if (data.tenantId !== undefined) updates.tenantId = data.tenantId
  if (data.issuerUrl !== undefined) updates.issuerUrl = data.issuerUrl
  if (data.scopes !== undefined) updates.scopes = data.scopes
  if (data.autoProvision !== undefined) updates.autoProvision = data.autoProvision
  if (data.defaultRole !== undefined) updates.defaultRole = data.defaultRole
  if (data.active !== undefined) updates.active = data.active
  await db.update(ssoProviders).set(updates).where(eq(ssoProviders.id, id))
}

/** Delete an SSO provider. */
export async function deleteSsoProvider(id: number) {
  const db = getDb()
  await db.delete(ssoProviders).where(eq(ssoProviders.id, id))
}
