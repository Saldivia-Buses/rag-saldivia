/**
 * Connector interface — all external source connectors implement this.
 *
 * Each connector handles authentication, listing, fetching, and optionally
 * change detection for a specific external system (Google Drive, SharePoint, etc.).
 */

export interface DocumentMeta {
  externalId: string
  title: string
  mimeType: string
  sizeBytes?: number | undefined
  lastModified?: number | undefined // epoch ms
  parentPath?: string | undefined // "/folder/subfolder"
}

export interface DocumentContent {
  externalId: string
  title: string
  content: Buffer | string // raw file bytes or extracted text
  mimeType: string
  sizeBytes: number
}

export interface ChangedDocument {
  externalId: string
  changeType: "created" | "modified" | "deleted"
  title: string
  lastModified?: number | undefined
}

export interface ListOptions {
  pageToken?: string | undefined
  maxResults?: number | undefined
  folderPath?: string | undefined
}

export interface ListResult {
  documents: DocumentMeta[]
  nextPageToken?: string | undefined
}

export interface IConnector {
  readonly provider: string

  /** Test credentials / connection. Throws on failure. */
  authenticate(credentials: Record<string, unknown>): Promise<void>

  /** List documents, paginated. */
  listDocuments(
    credentials: Record<string, unknown>,
    options?: ListOptions | undefined
  ): Promise<ListResult>

  /** Fetch a single document's content. */
  fetchDocument(
    credentials: Record<string, unknown>,
    externalId: string
  ): Promise<DocumentContent>

  /** Detect changes since a given timestamp. Not all connectors support this. */
  detectChanges?(
    credentials: Record<string, unknown>,
    since: number // epoch ms
  ): Promise<ChangedDocument[]>
}
