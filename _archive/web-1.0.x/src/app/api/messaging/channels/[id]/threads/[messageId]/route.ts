/**
 * GET /api/messaging/channels/[id]/threads/[messageId] — Get thread replies.
 */

import { getThreadReplies } from "@rag-saldivia/db"
import { requireAuth, apiOk, apiServerError } from "@/lib/api-utils"

export async function GET(
  request: Request,
  { params }: { params: Promise<{ id: string; messageId: string }> },
) {
  const claims = await requireAuth(request)
  if (claims instanceof Response) return claims

  const { messageId } = await params

  try {
    const replies = await getThreadReplies(messageId)
    return apiOk(replies)
  } catch (err) {
    return apiServerError(err, "GET /api/messaging/threads")
  }
}
