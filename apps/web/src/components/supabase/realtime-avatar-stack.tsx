'use client'

import { useMemo } from 'react'

import { AvatarStack } from '@/components/supabase/avatar-stack'
import { useRealtimePresenceRoom } from '@/components/supabase/use-realtime-presence-room'

export const RealtimeAvatarStack = ({ roomName }: { roomName: string }) => {
  const { users: usersMap } = useRealtimePresenceRoom(roomName)
  const avatars = useMemo(() => {
    return Object.values(usersMap).map((user) => ({
      name: user.name,
      image: user.image,
    }))
  }, [usersMap])

  return <AvatarStack avatars={avatars} />
}
