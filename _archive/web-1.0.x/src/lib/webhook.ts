/**
 * Outbound webhook dispatcher with HMAC-SHA256 signing.
 *
 * When system events occur (e.g. new document ingested, collection created),
 * this module delivers POST payloads to registered webhook URLs.
 *
 * Data flow:
 *   System event -> dispatchEvent(eventType, payload)
 *     -> listWebhooksByEvent (from DB) -> dispatchWebhook for each
 *       -> POST to webhook.url with JSON body + X-Signature header
 *
 * Security: each webhook has a unique secret; the body is signed with
 * HMAC-SHA256 and sent as `X-Signature: sha256=<hex>` for verification.
 *
 * Error handling: webhook failures are logged but never thrown — they must
 * not interrupt the main application flow (fire-and-forget pattern).
 *
 * Used by: server actions, API routes that emit domain events
 * Depends on: @rag-saldivia/db (webhook registry), @rag-saldivia/logger
 */

import { createHmac } from "crypto"
import { log } from "@rag-saldivia/logger/backend"
import type { DbWebhook } from "@rag-saldivia/db"

/**
 * Send a signed POST request to a single webhook endpoint.
 *
 * The payload is JSON-serialized with an injected `timestamp` field.
 * The request has a 5-second timeout to prevent slow endpoints from
 * blocking the caller. Failures are logged as warnings, never thrown.
 */
export async function dispatchWebhook(webhook: DbWebhook, payload: Record<string, unknown>) {
  const body = JSON.stringify({ ...payload, timestamp: Date.now() })
  const signature = createHmac("sha256", webhook.secret).update(body).digest("hex")

  try {
    const res = await fetch(webhook.url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Signature": `sha256=${signature}`,
        "X-Webhook-Id": webhook.id,
      },
      body,
      signal: AbortSignal.timeout(5000),
    })

    if (!res.ok) {
      log.warn("system.warning", { webhookId: webhook.id, url: webhook.url, status: res.status })
    } else {
      log.info("system.warning", { webhookId: webhook.id, url: webhook.url })
    }
  } catch (err) {
    // No relanzar — los webhooks no deben interrumpir el flujo principal
    log.warn("system.warning", { webhookId: webhook.id, url: webhook.url, error: String(err) })
  }
}

/**
 * Fan-out an event to all webhooks registered for that event type.
 *
 * Uses `Promise.allSettled` so one failing webhook doesn't block others.
 * The DB import is dynamic to avoid circular dependencies in edge bundles.
 */
export async function dispatchEvent(eventType: string, payload: Record<string, unknown>) {
  const { listWebhooksByEvent } = await import("@rag-saldivia/db")
  const hooks = await listWebhooksByEvent(eventType)
  await Promise.allSettled(hooks.map((h) => dispatchWebhook(h, { event: eventType, ...payload })))
}
