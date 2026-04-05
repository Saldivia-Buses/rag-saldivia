/**
 * Connector registry — maps provider names to IConnector implementations.
 *
 * Connectors register themselves at import time. The external-sync worker
 * looks up the connector by provider name to process sync jobs.
 */

import type { IConnector } from "./types"

const connectors = new Map<string, IConnector>()

export function registerConnector(connector: IConnector): void {
  connectors.set(connector.provider, connector)
}

export function getConnector(provider: string): IConnector {
  const c = connectors.get(provider)
  if (!c) throw new Error(`Connector not registered: ${provider}`)
  return c
}

export function listRegisteredConnectors(): string[] {
  return Array.from(connectors.keys())
}
