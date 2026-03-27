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
  saveResponse,
  unsaveByMessageId,
  saveAnnotation,
  addTag,
  removeTag,
  listTagsBySession,
  getDb,
  chatMessages,
  chatSessions,
} from "@rag-saldivia/db"
import { eq as eqDrizzle, lte as lteDrizzle } from "drizzle-orm"
import { randomUUID } from "crypto"
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
  sources?: import("@rag-saldivia/shared").Citation[]
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

export async function actionToggleSaved(
  messageId: number,
  content: string,
  sessionTitle: string,
  currentlySaved: boolean
) {
  const user = await requireUser()

  if (currentlySaved) {
    await unsaveByMessageId(messageId, user.id)
  } else {
    await saveResponse({ userId: user.id, messageId, content, sessionTitle })
  }
  revalidatePath("/saved")
}

export async function actionSaveAnnotation(data: {
  sessionId: string
  messageId?: number
  selectedText: string
  note?: string
}) {
  const user = await requireUser()
  await saveAnnotation({
    userId: user.id,
    sessionId: data.sessionId,
    messageId: data.messageId,
    selectedText: data.selectedText,
    note: data.note ?? null,
  })
}

export async function actionCreateSessionForDoc(collection: string, docName: string) {
  const user = await requireUser()
  const session = await createSession({
    userId: user.id,
    title: `Chat: ${docName}`,
    collection,
    crossdoc: false,
  })
  // Agregar system message con contexto del documento
  await addMessage({
    sessionId: session!.id,
    role: "system",
    content: `Only use information from document: ${docName}. Ignore other documents in the collection.`,
  })
  revalidatePath("/chat")
  return session
}

export async function actionAddTag(sessionId: string, tag: string) {
  await requireUser()
  await addTag(sessionId, tag)
  revalidatePath("/chat")
}

export async function actionRemoveTag(sessionId: string, tag: string) {
  await requireUser()
  await removeTag(sessionId, tag)
  revalidatePath("/chat")
}

export async function actionForkSession(sessionId: string, upToMessageId: number) {
  const user = await requireUser()
  const origSession = await getSessionById(sessionId, user.id)
  if (!origSession) return null

  const db = getDb()
  const messages = await db
    .select()
    .from(chatMessages)
    .where(eqDrizzle(chatMessages.sessionId, sessionId))

  const upToIdx = messages.findIndex((m) => m.id === upToMessageId)
  const messagesToCopy = upToIdx >= 0 ? messages.slice(0, upToIdx + 1) : messages

  const newId = randomUUID()
  await db.insert(chatSessions).values({
    id: newId,
    userId: user.id,
    title: `${origSession.title} (bifurcación)`,
    collection: origSession.collection,
    crossdoc: origSession.crossdoc,
    forkedFrom: sessionId,
    createdAt: Date.now(),
    updatedAt: Date.now(),
  })

  for (const msg of messagesToCopy) {
    await db.insert(chatMessages).values({
      sessionId: newId,
      role: msg.role,
      content: msg.content,
      sources: msg.sources,
      timestamp: msg.timestamp,
    })
  }

  revalidatePath("/chat")
  return newId
}
