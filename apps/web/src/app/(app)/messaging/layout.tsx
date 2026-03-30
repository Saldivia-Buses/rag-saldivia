import { requireUser } from "@/lib/auth/current-user"
import { getChannelsByUser, getUnreadCounts, listUsers } from "@rag-saldivia/db"
import { ChannelList } from "@/components/messaging/ChannelList"

export default async function MessagingLayout({ children }: { children: React.ReactNode }) {
  const user = await requireUser()
  const [channels, unreadCounts, users] = await Promise.all([
    getChannelsByUser(user.id),
    getUnreadCounts(user.id),
    listUsers().catch(() => []),
  ])

  return (
    <div className="flex h-full">
      <ChannelList
        channels={channels}
        unreadCounts={unreadCounts}
        userId={user.id}
        allUsers={users.map((u) => ({ id: u.id, name: u.name, email: u.email }))}
      />
      <div className="flex-1 flex flex-col min-w-0">
        {children}
      </div>
    </div>
  )
}
