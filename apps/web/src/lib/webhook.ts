/**
 * Dispatcher de webhooks salientes con firma HMAC-SHA256.
 * F2.38 — webhooks salientes.
 */

import { createHmac } from "crypto"
import { log } from "@rag-saldivia/logger/backend"
import type { DbWebhook } from "@rag-saldivia/db"

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
      log.warn("webhook.delivery_failed", { webhookId: webhook.id, url: webhook.url, status: res.status })
    } else {
      log.info("webhook.delivered", { webhookId: webhook.id, url: webhook.url })
    }
  } catch (err) {
    // No relanzar — los webhooks no deben interrumpir el flujo principal
    log.warn("webhook.delivery_error", { webhookId: webhook.id, url: webhook.url, error: String(err) })
  }
}

export async function dispatchEvent(eventType: string, payload: Record<string, unknown>) {
  const { listWebhooksByEvent } = await import("@rag-saldivia/db")
  const hooks = await listWebhooksByEvent(eventType)
  await Promise.allSettled(hooks.map((h) => dispatchWebhook(h, { event: eventType, ...payload })))
}
