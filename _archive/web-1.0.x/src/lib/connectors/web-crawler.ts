/**
 * Web Crawler connector — fetches and indexes public web pages.
 *
 * Respects robots.txt, configurable max depth and allowed domains.
 * Rate limiting: 1 req/sec per domain by default.
 *
 * Credentials: { urls: string[], maxDepth?, allowedDomains?, headers? }
 */

import type { IConnector, DocumentMeta, DocumentContent, ListOptions, ListResult } from "./types"
import { stripHtml } from "./html-utils"

const DEFAULT_MAX_DEPTH = 2
const RATE_LIMIT_MS = 1000 // 1 req/s per domain
const MAX_CONTENT_SIZE = 10 * 1024 * 1024 // 10MB
const USER_AGENT = "SaldiviaRAG/1.0 (+https://github.com/Camionerou/rag-saldivia)"
const FETCH_TIMEOUT_MS = 30_000

type CrawlerCredentials = {
  urls: string[]
  maxDepth: number
  allowedDomains: string[]
  headers: Record<string, string>
}

function parseCreds(raw: Record<string, unknown>): CrawlerCredentials {
  const urls = Array.isArray(raw.urls) ? raw.urls.map(String) : [String(raw.urls ?? "")]
  const allowedDomains = Array.isArray(raw.allowedDomains) ? raw.allowedDomains.map(String) : []

  // If no allowed domains specified, extract from seed URLs
  if (allowedDomains.length === 0) {
    for (const url of urls) {
      try {
        allowedDomains.push(new URL(url).hostname)
      } catch { /* skip invalid URLs */ }
    }
  }

  return {
    urls: urls.filter((u) => u.startsWith("http")),
    maxDepth: Number(raw.maxDepth ?? DEFAULT_MAX_DEPTH),
    allowedDomains,
    headers: (raw.headers as Record<string, string>) ?? {},
  }
}

// Simple robots.txt parser
const robotsCache = new Map<string, Set<string>>()

async function isAllowedByRobots(url: string): Promise<boolean> {
  try {
    const parsed = new URL(url)
    const origin = parsed.origin
    if (!robotsCache.has(origin)) {
      const res = await fetch(`${origin}/robots.txt`, {
        headers: { "User-Agent": USER_AGENT },
        signal: AbortSignal.timeout(5000),
      }).catch(() => null)

      const disallowed = new Set<string>()
      if (res?.ok) {
        const text = await res.text()
        let relevant = false
        for (const line of text.split("\n")) {
          const trimmed = line.trim().toLowerCase()
          if (trimmed.startsWith("user-agent:")) {
            const agent = trimmed.slice(11).trim()
            relevant = agent === "*" || agent === "saldiviarag"
          } else if (relevant && trimmed.startsWith("disallow:")) {
            const path = line.trim().slice(9).trim()
            if (path) disallowed.add(path)
          }
        }
      }
      robotsCache.set(origin, disallowed)
    }

    const disallowed = robotsCache.get(origin)!
    for (const path of disallowed) {
      if (parsed.pathname.startsWith(path)) return false
    }
    return true
  } catch {
    return true
  }
}

function extractLinks(html: string, baseUrl: string, allowedDomains: string[]): string[] {
  const links: string[] = []
  const linkRegex = /href=["']([^"']+)["']/gi
  let match
  while ((match = linkRegex.exec(html)) !== null) {
    try {
      const resolved = new URL(match[1]!, baseUrl)
      if (
        (resolved.protocol === "http:" || resolved.protocol === "https:") &&
        allowedDomains.includes(resolved.hostname) &&
        !resolved.pathname.match(/\.(png|jpg|jpeg|gif|svg|css|js|ico|woff|woff2|ttf|eot)$/i)
      ) {
        links.push(resolved.toString())
      }
    } catch { /* skip invalid URLs */ }
  }
  return [...new Set(links)]
}

export class WebCrawlerConnector implements IConnector {
  readonly provider = "web_crawler"

  async authenticate(credentials: Record<string, unknown>): Promise<void> {
    const creds = parseCreds(credentials)
    if (creds.urls.length === 0) throw new Error("No URLs provided")

    // Verify at least the first URL is reachable
    const res = await fetch(creds.urls[0]!, {
      method: "HEAD",
      headers: { "User-Agent": USER_AGENT },
      signal: AbortSignal.timeout(10_000),
      redirect: "follow",
    })
    if (!res.ok && res.status !== 405) {
      throw new Error(`URL not reachable: ${creds.urls[0]} (${res.status})`)
    }
  }

  async listDocuments(credentials: Record<string, unknown>, _options?: ListOptions): Promise<ListResult> {
    const creds = parseCreds(credentials)
    const visited = new Set<string>()
    const documents: DocumentMeta[] = []

    // BFS crawl
    let queue = creds.urls.filter((u) => u.startsWith("http"))
    let depth = 0

    while (queue.length > 0 && depth <= creds.maxDepth) {
      const nextQueue: string[] = []

      for (const url of queue) {
        if (visited.has(url)) continue
        visited.add(url)

        if (!(await isAllowedByRobots(url))) continue

        try {
          await new Promise((r) => setTimeout(r, RATE_LIMIT_MS))
          const res = await fetch(url, {
            headers: { "User-Agent": USER_AGENT, ...creds.headers },
            signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
            redirect: "follow",
          })

          if (!res.ok) continue

          const contentType = res.headers.get("content-type") ?? ""
          const contentLength = parseInt(res.headers.get("content-length") ?? "0")

          if (contentLength > MAX_CONTENT_SIZE) continue

          documents.push({
            externalId: url,
            title: url,
            mimeType: contentType.split(";")[0]?.trim() ?? "text/html",
            sizeBytes: contentLength || undefined,
          })

          // Extract links for next depth level
          if (contentType.includes("text/html") && depth < creds.maxDepth) {
            const html = await res.text()
            const links = extractLinks(html, url, creds.allowedDomains)
            nextQueue.push(...links.filter((l) => !visited.has(l)))
          }
        } catch { /* skip failed URLs */ }
      }

      queue = nextQueue
      depth++
    }

    return { documents }
  }

  async fetchDocument(credentials: Record<string, unknown>, externalId: string): Promise<DocumentContent> {
    const creds = parseCreds(credentials)
    const url = externalId // externalId IS the URL

    const res = await fetch(url, {
      headers: { "User-Agent": USER_AGENT, ...creds.headers },
      signal: AbortSignal.timeout(FETCH_TIMEOUT_MS),
      redirect: "follow",
    })

    if (!res.ok) throw new Error(`Fetch failed: ${url} (${res.status})`)

    const contentType = res.headers.get("content-type") ?? "text/html"

    if (contentType.includes("text/html")) {
      const html = await res.text()
      const text = stripHtml(html)
      return {
        externalId: url,
        title: url,
        content: text,
        mimeType: "text/plain",
        sizeBytes: Buffer.byteLength(text, "utf8"),
      }
    }

    // Non-HTML: return raw bytes
    const buffer = Buffer.from(await res.arrayBuffer())
    return {
      externalId: url,
      title: url,
      content: buffer,
      mimeType: contentType.split(";")[0]?.trim() ?? "application/octet-stream",
      sizeBytes: buffer.length,
    }
  }

  // Web crawler doesn't support change detection — relies on content hash comparison
}
