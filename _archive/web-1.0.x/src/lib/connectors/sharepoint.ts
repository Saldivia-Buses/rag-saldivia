/**
 * SharePoint/OneDrive connector — Microsoft Graph API.
 *
 * Supports OneDrive personal (me/drive) and SharePoint sites (sites/{id}/drive).
 * Delta queries for incremental sync. Respects Retry-After on 429.
 *
 * Credentials: { accessToken, refreshToken, clientId, clientSecret, tenantId, siteId?, driveId? }
 */

import type { IConnector, DocumentMeta, DocumentContent, ChangedDocument, ListOptions, ListResult } from "./types"

const GRAPH_BASE = "https://graph.microsoft.com/v1.0"
const TOKEN_URL_BASE = "https://login.microsoftonline.com"
const RATE_LIMIT_MS = 100

type SPCredentials = {
  accessToken: string
  refreshToken: string
  clientId: string
  clientSecret: string
  tenantId: string
  siteId?: string | undefined
  driveId?: string | undefined
  deltaLink?: string | undefined
}

function parseCreds(raw: Record<string, unknown>): SPCredentials {
  return {
    accessToken: String(raw.accessToken ?? ""),
    refreshToken: String(raw.refreshToken ?? ""),
    clientId: String(raw.clientId ?? process.env["AZURE_CLIENT_ID"] ?? ""),
    clientSecret: String(raw.clientSecret ?? process.env["AZURE_CLIENT_SECRET"] ?? ""),
    tenantId: String(raw.tenantId ?? process.env["AZURE_TENANT_ID"] ?? "common"),
    siteId: raw.siteId ? String(raw.siteId) : undefined,
    driveId: raw.driveId ? String(raw.driveId) : undefined,
    deltaLink: raw.deltaLink ? String(raw.deltaLink) : undefined,
  }
}

async function refreshToken(creds: SPCredentials): Promise<string> {
  const res = await fetch(`${TOKEN_URL_BASE}/${creds.tenantId}/oauth2/v2.0/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: creds.clientId,
      client_secret: creds.clientSecret,
      refresh_token: creds.refreshToken,
      grant_type: "refresh_token",
      scope: "Files.Read.All Sites.Read.All offline_access",
    }),
  })
  if (!res.ok) throw new Error(`Microsoft token refresh failed: ${res.status}`)
  const data = (await res.json()) as { access_token: string }
  return data.access_token
}

async function graphRequest(url: string, token: string): Promise<Response> {
  await new Promise((r) => setTimeout(r, RATE_LIMIT_MS))
  const res = await fetch(url, { headers: { Authorization: `Bearer ${token}` } })
  if (res.status === 429) {
    const retryAfter = parseInt(res.headers.get("Retry-After") ?? "10") * 1000
    await new Promise((r) => setTimeout(r, retryAfter))
    return fetch(url, { headers: { Authorization: `Bearer ${token}` } })
  }
  return res
}

function drivePath(creds: SPCredentials): string {
  if (creds.driveId) return `/drives/${creds.driveId}`
  if (creds.siteId) return `/sites/${creds.siteId}/drive`
  return "/me/drive"
}

export class SharePointConnector implements IConnector {
  readonly provider = "sharepoint"

  async authenticate(credentials: Record<string, unknown>): Promise<void> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshToken(creds)
    const res = await graphRequest(`${GRAPH_BASE}/me`, token)
    if (!res.ok) throw new Error(`SharePoint auth failed: ${res.status}`)
  }

  async listDocuments(credentials: Record<string, unknown>, options?: ListOptions): Promise<ListResult> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshToken(creds)

    const path = drivePath(creds)
    let url = `${GRAPH_BASE}${path}/root/children?$top=${options?.maxResults ?? 100}&$select=id,name,file,size,lastModifiedDateTime`
    if (options?.pageToken) url = options.pageToken // Graph uses full URL as next link

    const res = await graphRequest(url, token)
    if (!res.ok) throw new Error(`SharePoint list failed: ${res.status}`)

    const data = (await res.json()) as {
      value: Array<{ id: string; name: string; file?: { mimeType: string }; size?: number; lastModifiedDateTime?: string }>
      "@odata.nextLink"?: string
    }

    const documents: DocumentMeta[] = data.value
      .filter((item) => item.file) // Only files, not folders
      .map((item) => ({
        externalId: item.id,
        title: item.name,
        mimeType: item.file?.mimeType ?? "application/octet-stream",
        sizeBytes: item.size,
        lastModified: item.lastModifiedDateTime ? new Date(item.lastModifiedDateTime).getTime() : undefined,
      }))

    const result: ListResult = { documents }
    if (data["@odata.nextLink"]) result.nextPageToken = data["@odata.nextLink"]
    return result
  }

  async fetchDocument(credentials: Record<string, unknown>, externalId: string): Promise<DocumentContent> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshToken(creds)
    const path = drivePath(creds)

    // Get metadata
    const metaRes = await graphRequest(`${GRAPH_BASE}${path}/items/${externalId}?$select=name,file,size`, token)
    if (!metaRes.ok) throw new Error(`SharePoint metadata failed: ${metaRes.status}`)
    const meta = (await metaRes.json()) as { name: string; file?: { mimeType: string }; size?: number }

    // Download content
    const contentRes = await graphRequest(`${GRAPH_BASE}${path}/items/${externalId}/content`, token)
    if (!contentRes.ok) throw new Error(`SharePoint download failed: ${contentRes.status}`)

    const buffer = Buffer.from(await contentRes.arrayBuffer())
    return {
      externalId,
      title: meta.name,
      content: buffer,
      mimeType: meta.file?.mimeType ?? "application/octet-stream",
      sizeBytes: buffer.length,
    }
  }

  async detectChanges(credentials: Record<string, unknown>, _since: number): Promise<ChangedDocument[]> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshToken(creds)
    const path = drivePath(creds)

    let url = creds.deltaLink ?? `${GRAPH_BASE}${path}/root/delta?$select=id,name,file,deleted,lastModifiedDateTime`
    const changes: ChangedDocument[] = []

    while (url) {
      const res = await graphRequest(url, token)
      if (!res.ok) break

      const data = (await res.json()) as {
        value: Array<{
          id: string; name: string
          file?: { mimeType: string }
          deleted?: { state: string }
          lastModifiedDateTime?: string
        }>
        "@odata.nextLink"?: string
        "@odata.deltaLink"?: string
      }

      for (const item of data.value) {
        if (item.deleted) {
          changes.push({ externalId: item.id, changeType: "deleted", title: item.name ?? "unknown" })
        } else if (item.file) {
          changes.push({
            externalId: item.id,
            changeType: "modified",
            title: item.name,
            lastModified: item.lastModifiedDateTime ? new Date(item.lastModifiedDateTime).getTime() : undefined,
          })
        }
      }

      url = data["@odata.nextLink"] ?? ""
    }

    return changes
  }
}
