"use server"
/**
 * Server Actions — Chat (sesiones y mensajes)
 */

import { revalidatePath } from "next/cache"
import { requireUser } from "@/lib/auth/current-user"
import {
  listSessionsByUser,
  getSessionById,
  createSession,
  updateSessionTitle,
  deleteSession,
  addMessage,
  addFeedback,
} from "@rag-saldivia/db"
import { log } from "@rag-saldivia/logger/backend"

export async function actionListSessions() {
  const user = await requireUser()
  return listSessionsByUser(user.id)
}

export async function actionGetSession(id: string) {
  const user = await requireUser()
  return getSessionById(id, user.id)
}

export async function actionCreateSession(data: {
  collection: string
  crossdoc?: boolean
  title?: string
}) {
  const user = await requireUser()
  const session = await createSession({ userId: user.id, ...data })
  log.info("auth.login", { action: "session_created", collection: data.collection }, { userId: user.id })
  revalidatePath("/chat")
  return session
}

export async function actionRenameSession(id: string, title: string) {
  const user = await requireUser()
  const updated = await updateSessionTitle(id, user.id, title)
  revalidatePath("/chat")
  revalidatePath(`/chat/${id}`)
  return updated
}

export async function actionDeleteSession(id: string) {
  const user = await requireUser()
  await deleteSession(id, user.id)
  revalidatePath("/chat")
}

export async function actionAddMessage(data: {
  sessionId: string
  role: "user" | "assistant" | "system"
  content: string
  sources?: unknown[]
}) {
  const user = await requireUser()
  const message = await addMessage(data)

  if (data.role === "user") {
    log.info("rag.query", {
      query: data.content.slice(0, 100),
      sessionId: data.sessionId,
    }, { userId: user.id, sessionId: data.sessionId })
  }

  revalidatePath(`/chat/${data.sessionId}`)
  return message
}

export async function actionAddFeedback(
  messageId: number,
  rating: "up" | "down"
) {
  const user = await requireUser()
  await addFeedback(messageId, user.id, rating)
}
