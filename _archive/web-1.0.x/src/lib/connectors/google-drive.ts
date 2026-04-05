/**
 * Google Drive connector — implements IConnector for Google Drive API v3.
 *
 * Uses fetch nativo (no googleapis SDK). OAuth2 refresh handled internally.
 * Supports folder filtering, Google Docs export to PDF, and change detection
 * via the Changes API with startPageToken.
 *
 * Credentials shape: { accessToken, refreshToken, clientId, clientSecret, folderId? }
 */

import type { IConnector, DocumentMeta, DocumentContent, ChangedDocument, ListOptions, ListResult } from "./types"

const API_BASE = "https://www.googleapis.com/drive/v3"
const TOKEN_URL = "https://oauth2.googleapis.com/token"
const RATE_LIMIT_MS = 100 // 10 req/s max

type DriveCredentials = {
  accessToken: string
  refreshToken: string
  clientId: string
  clientSecret: string
  folderId?: string | undefined
  startPageToken?: string | undefined
}

function parseCreds(raw: Record<string, unknown>): DriveCredentials {
  return {
    accessToken: String(raw.accessToken ?? ""),
    refreshToken: String(raw.refreshToken ?? ""),
    clientId: String(raw.clientId ?? process.env["GOOGLE_CLIENT_ID"] ?? ""),
    clientSecret: String(raw.clientSecret ?? process.env["GOOGLE_CLIENT_SECRET"] ?? ""),
    folderId: raw.folderId ? String(raw.folderId) : undefined,
    startPageToken: raw.startPageToken ? String(raw.startPageToken) : undefined,
  }
}

async function refreshAccessToken(creds: DriveCredentials): Promise<string> {
  const res = await fetch(TOKEN_URL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: creds.clientId,
      client_secret: creds.clientSecret,
      refresh_token: creds.refreshToken,
      grant_type: "refresh_token",
    }),
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(`Token refresh failed (${res.status}): ${text.slice(0, 200)}`)
  }
  const data = (await res.json()) as { access_token: string }
  return data.access_token
}

async function driveRequest(url: string, accessToken: string, init?: RequestInit): Promise<Response> {
  await new Promise((r) => setTimeout(r, RATE_LIMIT_MS))
  const res = await fetch(url, {
    ...init,
    headers: { Authorization: `Bearer ${accessToken}`, ...init?.headers },
  })
  if (res.status === 429) {
    // Rate limited — wait and retry once
    const retryAfter = parseInt(res.headers.get("Retry-After") ?? "5") * 1000
    await new Promise((r) => setTimeout(r, retryAfter))
    return fetch(url, {
      ...init,
      headers: { Authorization: `Bearer ${accessToken}`, ...init?.headers },
    })
  }
  return res
}

// Google Docs native types that need export (can't download directly)
const EXPORT_MIME_MAP: Record<string, string> = {
  "application/vnd.google-apps.document": "application/pdf",
  "application/vnd.google-apps.spreadsheet": "text/csv",
  "application/vnd.google-apps.presentation": "application/pdf",
  "application/vnd.google-apps.drawing": "image/png",
}

export class GoogleDriveConnector implements IConnector {
  readonly provider = "google_drive"

  async authenticate(credentials: Record<string, unknown>): Promise<void> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshAccessToken(creds)
    const res = await driveRequest(`${API_BASE}/about?fields=user`, token)
    if (!res.ok) throw new Error(`Google Drive auth failed: ${res.status}`)
  }

  async listDocuments(credentials: Record<string, unknown>, options?: ListOptions): Promise<ListResult> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshAccessToken(creds)

    const params = new URLSearchParams({
      fields: "nextPageToken,files(id,name,mimeType,size,modifiedTime,parents)",
      pageSize: String(options?.maxResults ?? 100),
    })

    // Folder filter
    const folderId = creds.folderId ?? (options?.folderPath ? options.folderPath : null)
    let q = "trashed = false"
    if (folderId) q += ` and '${folderId}' in parents`
    params.set("q", q)

    if (options?.pageToken) params.set("pageToken", options.pageToken)

    const res = await driveRequest(`${API_BASE}/files?${params}`, token)
    if (!res.ok) throw new Error(`Google Drive list failed: ${res.status}`)

    const data = (await res.json()) as {
      files: Array<{ id: string; name: string; mimeType: string; size?: string; modifiedTime?: string }>
      nextPageToken?: string
    }

    const documents: DocumentMeta[] = data.files.map((f) => ({
      externalId: f.id,
      title: f.name,
      mimeType: f.mimeType,
      sizeBytes: f.size ? parseInt(f.size) : undefined,
      lastModified: f.modifiedTime ? new Date(f.modifiedTime).getTime() : undefined,
    }))

    const result: ListResult = { documents }
    if (data.nextPageToken) result.nextPageToken = data.nextPageToken
    return result
  }

  async fetchDocument(credentials: Record<string, unknown>, externalId: string): Promise<DocumentContent> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshAccessToken(creds)

    // First get metadata to check if it's a Google native type
    const metaRes = await driveRequest(`${API_BASE}/files/${externalId}?fields=name,mimeType,size`, token)
    if (!metaRes.ok) throw new Error(`Google Drive metadata failed: ${metaRes.status}`)
    const meta = (await metaRes.json()) as { name: string; mimeType: string; size?: string }

    const exportMime = EXPORT_MIME_MAP[meta.mimeType]

    let contentRes: Response
    if (exportMime) {
      // Google native type — export
      contentRes = await driveRequest(
        `${API_BASE}/files/${externalId}/export?mimeType=${encodeURIComponent(exportMime)}`,
        token,
      )
    } else {
      // Regular file — download
      contentRes = await driveRequest(`${API_BASE}/files/${externalId}?alt=media`, token)
    }

    if (!contentRes.ok) throw new Error(`Google Drive fetch failed: ${contentRes.status}`)

    const buffer = Buffer.from(await contentRes.arrayBuffer())
    return {
      externalId,
      title: meta.name,
      content: buffer,
      mimeType: exportMime ?? meta.mimeType,
      sizeBytes: buffer.length,
    }
  }

  async detectChanges(credentials: Record<string, unknown>, _since: number): Promise<ChangedDocument[]> {
    const creds = parseCreds(credentials)
    const token = creds.accessToken || await refreshAccessToken(creds)

    // Get or initialize startPageToken
    let pageToken = creds.startPageToken
    if (!pageToken) {
      const res = await driveRequest(`${API_BASE}/changes/startPageToken`, token)
      if (!res.ok) throw new Error(`Google Drive startPageToken failed: ${res.status}`)
      const data = (await res.json()) as { startPageToken: string }
      pageToken = data.startPageToken
    }

    const changes: ChangedDocument[] = []
    let nextPageToken: string | undefined = pageToken

    while (nextPageToken) {
      const params = new URLSearchParams({
        pageToken: nextPageToken,
        fields: "nextPageToken,newStartPageToken,changes(fileId,removed,file(name,mimeType,modifiedTime))",
        pageSize: "100",
      })

      const res = await driveRequest(`${API_BASE}/changes?${params}`, token)
      if (!res.ok) break

      const data = (await res.json()) as {
        changes: Array<{
          fileId: string
          removed: boolean
          file?: { name: string; mimeType: string; modifiedTime?: string }
        }>
        nextPageToken?: string
        newStartPageToken?: string
      }

      for (const change of data.changes) {
        if (change.removed) {
          changes.push({
            externalId: change.fileId,
            changeType: "deleted",
            title: change.file?.name ?? "unknown",
          })
        } else if (change.file) {
          changes.push({
            externalId: change.fileId,
            changeType: "modified",
            title: change.file.name,
            lastModified: change.file.modifiedTime ? new Date(change.file.modifiedTime).getTime() : undefined,
          })
        }
      }

      nextPageToken = data.nextPageToken
    }

    return changes
  }
}
