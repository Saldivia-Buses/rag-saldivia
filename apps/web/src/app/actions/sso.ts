"use server"

import { z } from "zod"
import { adminAction } from "@/lib/safe-action"
import {
  listAllSsoProviders,
  createSsoProvider,
  updateSsoProvider,
  deleteSsoProvider,
} from "@rag-saldivia/db"
import { SsoProviderTypeSchema } from "@rag-saldivia/shared"
import { revalidatePath } from "next/cache"

export const actionListSsoProviders = adminAction
  .schema(z.object({}))
  .action(async () => {
    const providers = await listAllSsoProviders()
    return { providers }
  })

export const actionCreateSsoProvider = adminAction
  .schema(
    z.object({
      name: z.string().min(1).max(100),
      type: SsoProviderTypeSchema,
      clientId: z.string().min(1),
      clientSecret: z.string().min(1),
      tenantId: z.string().optional(),
      issuerUrl: z.string().optional(),
      scopes: z.string().optional(),
      autoProvision: z.boolean().optional(),
      defaultRole: z.enum(["admin", "area_manager", "user"]).optional(),
      samlCert: z.string().optional(),
      samlEntryPoint: z.string().optional(),
    })
  )
  .action(async ({ parsedInput: data }) => {
    await createSsoProvider({
      name: data.name,
      type: data.type,
      clientId: data.clientId,
      clientSecret: data.clientSecret,
      tenantId: data.tenantId ?? null,
      issuerUrl: data.issuerUrl ?? null,
      scopes: data.scopes,
      autoProvision: data.autoProvision,
      defaultRole: data.defaultRole,
      active: true,
    })
    revalidatePath("/admin/sso")
    return { ok: true }
  })

export const actionUpdateSsoProvider = adminAction
  .schema(
    z.object({
      id: z.number().int().positive(),
      name: z.string().min(1).max(100).optional(),
      clientId: z.string().min(1).optional(),
      clientSecret: z.string().min(1).optional(),
      tenantId: z.string().optional(),
      issuerUrl: z.string().optional(),
      scopes: z.string().optional(),
      autoProvision: z.boolean().optional(),
      defaultRole: z.enum(["admin", "area_manager", "user"]).optional(),
      active: z.boolean().optional(),
    })
  )
  .action(async ({ parsedInput: { id, ...data } }) => {
    // Build update object manually to satisfy exactOptionalPropertyTypes
    const updates: Record<string, unknown> = {}
    if (data.name !== undefined) updates.name = data.name
    if (data.clientId !== undefined) updates.clientId = data.clientId
    if (data.clientSecret !== undefined) updates.clientSecret = data.clientSecret
    if (data.tenantId !== undefined) updates.tenantId = data.tenantId
    if (data.issuerUrl !== undefined) updates.issuerUrl = data.issuerUrl
    if (data.scopes !== undefined) updates.scopes = data.scopes
    if (data.autoProvision !== undefined) updates.autoProvision = data.autoProvision
    if (data.defaultRole !== undefined) updates.defaultRole = data.defaultRole
    if (data.active !== undefined) updates.active = data.active
    await updateSsoProvider(id, updates as Parameters<typeof updateSsoProvider>[1])
    revalidatePath("/admin/sso")
    return { ok: true }
  })

export const actionDeleteSsoProvider = adminAction
  .schema(z.object({ id: z.number().int().positive() }))
  .action(async ({ parsedInput: { id } }) => {
    await deleteSsoProvider(id)
    revalidatePath("/admin/sso")
    return { ok: true }
  })
