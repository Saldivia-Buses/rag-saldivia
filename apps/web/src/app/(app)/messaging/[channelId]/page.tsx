import { notFound } from "next/navigation"
import { requireUser } from "@/lib/auth/current-user"
import { getChannel, getMessages, getChannelMembers, updateLastRead } from "@rag-saldivia/db"
import { ChannelHeader } from "@/components/messaging/ChannelHeader"
import { ChannelView } from "@/components/messaging/ChannelView"

export default async function ChannelPage({
  params,
}: {
  params: Promise<{ channelId: string }>
}) {
  const user = await requireUser()
  const { channelId } = await params

  const [channel, messages, members] = await Promise.all([
    getChannel(channelId),
    getMessages(channelId, { limit: 50 }),
    getChannelMembers(channelId),
  ])

  if (!channel) notFound()

  // Verify membership
  const isMember = channel.channelMembers.some((m) => m.userId === user.id)
  if (!isMember && user.role !== "admin") notFound()

  // Mark as read
  await updateLastRead(channelId, user.id)

  return (
    <>
      <ChannelHeader
        channel={channel}
        memberCount={members.length}
      />
      <ChannelView
        channelId={channelId}
        initialMessages={messages.reverse()}
        currentUserId={user.id}
        members={members.map((m) => ({
          id: m.userId,
          name: m.user?.name ?? "Usuario",
          email: m.user?.email ?? "",
        }))}
      />
    </>
  )
}
