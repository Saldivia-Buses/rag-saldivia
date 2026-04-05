"use server"

import { z } from "zod"
import { revalidatePath } from "next/cache"
import { authAction, clean } from "@/lib/safe-action"
import {
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
  getDb,
  chatMessages,
  chatSessions,
} from "@rag-saldivia/db"
import { eq as eqDrizzle } from "drizzle-orm"
import { randomUUID } from "crypto"
import { log } from "@rag-saldivia/logger/backend"
import { CitationSchema } from "@rag-saldivia/shared"

export const actionCreateSession = authAction
  .schema(z.object({
    collection: z.string(),
    crossdoc: z.boolean().optional(),
    title: z.string().optional(),
  }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    const session = await createSession(clean({ userId: user.id, ...data }))
    log.info("auth.login", { action: "session_created", collection: data.collection }, { userId: user.id })
    revalidatePath("/chat")
    return session
  })

export const actionRenameSession = authAction
  .schema(z.object({ id: z.string(), title: z.string() }))
  .action(async ({ parsedInput: { id, title }, ctx: { user } }) => {
    const updated = await updateSessionTitle(id, user.id, title)
    revalidatePath("/chat")
    revalidatePath(`/chat/${id}`)
    return updated
  })

export const actionDeleteSession = authAction
  .schema(z.object({ id: z.string() }))
  .action(async ({ parsedInput: { id }, ctx: { user } }) => {
    await deleteSession(id, user.id)
    revalidatePath("/chat")
  })

export const actionAddMessage = authAction
  .schema(z.object({
    sessionId: z.string(),
    role: z.enum(["user", "assistant", "system"]),
    content: z.string(),
    sources: z.array(CitationSchema).optional(),
  }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    const message = await addMessage(clean(data))

    if (data.role === "user") {
      log.info("rag.query", {
        query: data.content.slice(0, 100),
        sessionId: data.sessionId,
      }, { userId: user.id, sessionId: data.sessionId })
    }

    revalidatePath(`/chat/${data.sessionId}`)
    return message
  })

export const actionAddFeedback = authAction
  .schema(z.object({ messageId: z.number(), rating: z.enum(["up", "down"]) }))
  .action(async ({ parsedInput: { messageId, rating }, ctx: { user } }) => {
    await addFeedback(messageId, user.id, rating)
  })

export const actionToggleSaved = authAction
  .schema(z.object({
    messageId: z.number(),
    content: z.string(),
    sessionTitle: z.string(),
    currentlySaved: z.boolean(),
  }))
  .action(async ({ parsedInput: { messageId, content, sessionTitle, currentlySaved }, ctx: { user } }) => {
    if (currentlySaved) {
      await unsaveByMessageId(messageId, user.id)
    } else {
      await saveResponse({ userId: user.id, messageId, content, sessionTitle })
    }
  })

export const actionSaveAnnotation = authAction
  .schema(z.object({
    sessionId: z.string(),
    messageId: z.number().optional(),
    selectedText: z.string(),
    note: z.string().optional(),
  }))
  .action(async ({ parsedInput: data, ctx: { user } }) => {
    await saveAnnotation({
      userId: user.id,
      sessionId: data.sessionId,
      messageId: data.messageId,
      selectedText: data.selectedText,
      note: data.note ?? null,
    })
  })

export const actionCreateSessionForDoc = authAction
  .schema(z.object({ collection: z.string(), docName: z.string() }))
  .action(async ({ parsedInput: { collection, docName }, ctx: { user } }) => {
    const session = await createSession({
      userId: user.id,
      title: `Chat: ${docName}`,
      collection,
      crossdoc: false,
    })
    await addMessage({
      sessionId: session!.id,
      role: "system",
      content: `Only use information from document: ${docName}. Ignore other documents in the collection.`,
    })
    revalidatePath("/chat")
    return session
  })

export const actionAddTag = authAction
  .schema(z.object({ sessionId: z.string(), tag: z.string() }))
  .action(async ({ parsedInput: { sessionId, tag } }) => {
    await addTag(sessionId, tag)
    revalidatePath("/chat")
  })

export const actionRemoveTag = authAction
  .schema(z.object({ sessionId: z.string(), tag: z.string() }))
  .action(async ({ parsedInput: { sessionId, tag } }) => {
    await removeTag(sessionId, tag)
    revalidatePath("/chat")
  })

export const actionForkSession = authAction
  .schema(z.object({ sessionId: z.string(), upToMessageId: z.number() }))
  .action(async ({ parsedInput: { sessionId, upToMessageId }, ctx: { user } }) => {
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

    if (messagesToCopy.length > 0) {
      await db.insert(chatMessages).values(
        messagesToCopy.map((msg) => ({
          sessionId: newId,
          role: msg.role,
          content: msg.content,
          sources: msg.sources,
          timestamp: msg.timestamp,
        }))
      )
    }

    revalidatePath("/chat")
    return newId
  })
