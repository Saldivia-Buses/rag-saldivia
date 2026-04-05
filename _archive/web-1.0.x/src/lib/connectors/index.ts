/**
 * Connector auto-registration — import this to register all connectors.
 *
 * Used by: external-sync worker on startup.
 */

import { registerConnector } from "./registry"
import { GoogleDriveConnector } from "./google-drive"
import { SharePointConnector } from "./sharepoint"
import { ConfluenceConnector } from "./confluence"
import { WebCrawlerConnector } from "./web-crawler"

registerConnector(new GoogleDriveConnector())
registerConnector(new SharePointConnector())
registerConnector(new ConfluenceConnector())
registerConnector(new WebCrawlerConnector())

export { getConnector, listRegisteredConnectors } from "./registry"
export type { IConnector, DocumentMeta, DocumentContent, ChangedDocument, ListOptions, ListResult } from "./types"
