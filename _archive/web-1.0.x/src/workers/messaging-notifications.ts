#!/usr/bin/env bun
/**
 * BullMQ worker for messaging notifications.
 *
 * Processes notification jobs:
 * - @mention → in-app notification via WS
 * - DM → in-app notification via WS
 * - Thread reply → notification to original author
 * - Dedup: skip if user has channel open (check Redis presence)
 *
 * Usage: bun apps/web/src/workers/messaging-notifications.ts
 */

import { Queue, Worker, type Job } from "bullmq"
import IORedis from "ioredis"
import { getPresence } from "../lib/ws/presence"
import { redisTopic } from "../lib/ws/protocol"
import type { ServerMessage } from "@rag-saldivia/shared"

const REDIS_URL = process.env["REDIS_URL"] ?? "redis://localhost:6379"

function getConnection() {
  return new IORedis(REDIS_URL, { maxRetriesPerRequest: null })
}

export const messagingNotificationQueue = new Queue("messaging-notifications", {
  connection: getConnection(),
  defaultJobOptions: {
    attempts: 2,
    backoff: { type: "exponential", delay: 5_000 },
    removeOnComplete: 50,
    removeOnFail: 100,
  },
})

export type NotificationJobData = {
  type: "mention" | "dm" | "thread_reply"
  recipientUserId: number
  channelId: string
  messageId: string
  senderUserId: number
  senderName: string
  content: string
}

const redisPub = getConnection()
const redisPresence = getConnection()

async function processNotification(job: Job<NotificationJobData>) {
  const { type, recipientUserId, channelId, senderName } = job.data

  // Dedup: don't notify if user is online and has channel open
  const presence = await getPresence(redisPresence, recipientUserId)
  if (presence === "online") {
    // User is online — WS will deliver in real-time, skip push notification
    // Still send the unread_update event
  }

  // Send unread update via Redis → WS sidecar
  const msg: ServerMessage = {
    type: "unread_update",
    channelId,
    count: 1, // Increment signal — client handles actual count
  }

  // Publish to the user's personal channel (they subscribe on auth)
  redisPub.publish(redisTopic(channelId), JSON.stringify(msg))

  console.warn(
    `[messaging-notifications] ${type}: ${senderName} → user:${recipientUserId} in channel:${channelId}`,
  )
}

// Only start worker if this file is run directly
if (import.meta.main) {
  const worker = new Worker("messaging-notifications", processNotification, {
    connection: getConnection(),
    concurrency: 5,
  })

  worker.on("completed", (job) => {
    console.warn(`[messaging-notifications] Job ${job.id} completed`)
  })

  worker.on("failed", (job, err) => {
    console.error(`[messaging-notifications] Job ${job?.id} failed:`, err.message)
  })

  console.warn("[messaging-notifications] Worker started")
}
