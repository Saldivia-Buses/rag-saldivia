/**
 * Confluence connector — Atlassian REST API with API tokens.
 *
 * Authenticates via Basic auth (email:apiToken). Supports space filtering,
 * CQL-based change detection, and HTML-to-text conversion for page bodies.
 *
 * Credentials: { baseUrl, email, apiToken, spaceKey? }
 */

import type { IConnector, DocumentMeta, DocumentContent, ChangedDocument, ListOptions, ListResult } from "./types"
import { stripHtml } from "./html-utils"

const RATE_LIMIT_MS = 100 // 10 req/s

type ConfluenceCredentials = {
  baseUrl: string
  email: string
  apiToken: string
  spaceKey?: string | undefined
}

function parseCreds(raw: Record<string, unknown>): ConfluenceCredentials {
  return {
    baseUrl: String(raw.baseUrl ?? "").replace(/\/$/, ""),
    email: String(raw.email ?? ""),
    apiToken: String(raw.apiToken ?? ""),
    spaceKey: raw.spaceKey ? String(raw.spaceKey) : undefined,
  }
}

function authHeader(creds: ConfluenceCredentials): string {
  return `Basic ${Buffer.from(`${creds.email}:${creds.apiToken}`).toString("base64")}`
}

async function confluenceRequest(url: string, auth: string): Promise<Response> {
  await new Promise((r) => setTimeout(r, RATE_LIMIT_MS))
  return fetch(url, { headers: { Authorization: auth, Accept: "application/json" } })
}

export class ConfluenceConnector implements IConnector {
  readonly provider = "confluence"

  async authenticate(credentials: Record<string, unknown>): Promise<void> {
    const creds = parseCreds(credentials)
    const res = await confluenceRequest(`${creds.baseUrl}/wiki/rest/api/user/current`, authHeader(creds))
    if (!res.ok) throw new Error(`Confluence auth failed: ${res.status}`)
  }

  async listDocuments(credentials: Record<string, unknown>, options?: ListOptions): Promise<ListResult> {
    const creds = parseCreds(credentials)
    const auth = authHeader(creds)

    const params = new URLSearchParams({
      type: "page",
      limit: String(options?.maxResults ?? 50),
      expand: "version",
    })
    if (creds.spaceKey) params.set("spaceKey", creds.spaceKey)
    if (options?.pageToken) params.set("start", options.pageToken)

    const res = await confluenceRequest(`${creds.baseUrl}/wiki/rest/api/content?${params}`, auth)
    if (!res.ok) throw new Error(`Confluence list failed: ${res.status}`)

    const data = (await res.json()) as {
      results: Array<{
        id: string; title: string; type: string
        version?: { when?: string }
      }>
      _links?: { next?: string }
      size: number
      start: number
    }

    const documents: DocumentMeta[] = data.results.map((page) => ({
      externalId: page.id,
      title: page.title,
      mimeType: "text/html",
      lastModified: page.version?.when ? new Date(page.version.when).getTime() : undefined,
    }))

    const result: ListResult = { documents }
    if (data._links?.next) {
      result.nextPageToken = String(data.start + data.size)
    }
    return result
  }

  async fetchDocument(credentials: Record<string, unknown>, externalId: string): Promise<DocumentContent> {
    const creds = parseCreds(credentials)
    const auth = authHeader(creds)

    const res = await confluenceRequest(
      `${creds.baseUrl}/wiki/rest/api/content/${externalId}?expand=body.export_view`,
      auth,
    )
    if (!res.ok) throw new Error(`Confluence fetch failed: ${res.status}`)

    const data = (await res.json()) as {
      title: string
      body?: { export_view?: { value?: string } }
    }

    const html = data.body?.export_view?.value ?? ""
    const text = stripHtml(html)

    return {
      externalId,
      title: data.title,
      content: text,
      mimeType: "text/plain",
      sizeBytes: Buffer.byteLength(text, "utf8"),
    }
  }

  async detectChanges(credentials: Record<string, unknown>, since: number): Promise<ChangedDocument[]> {
    const creds = parseCreds(credentials)
    const auth = authHeader(creds)

    const sinceDate = new Date(since).toISOString().split("T")[0] // yyyy-MM-dd
    let cql = `type=page AND lastModified>="${sinceDate}"`
    if (creds.spaceKey) cql += ` AND space="${creds.spaceKey}"`

    const params = new URLSearchParams({
      cql,
      limit: "100",
      expand: "version",
    })

    const res = await confluenceRequest(`${creds.baseUrl}/wiki/rest/api/content/search?${params}`, auth)
    if (!res.ok) throw new Error(`Confluence search failed: ${res.status}`)

    const data = (await res.json()) as {
      results: Array<{
        id: string; title: string
        version?: { when?: string }
      }>
    }

    return data.results.map((page) => ({
      externalId: page.id,
      changeType: "modified" as const,
      title: page.title,
      lastModified: page.version?.when ? new Date(page.version.when).getTime() : undefined,
    }))
  }
}
