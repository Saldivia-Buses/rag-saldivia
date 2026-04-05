import { listAreas, listUsers } from "@rag-saldivia/db"
import { AdminAreas } from "@/components/admin/AdminAreas"

export default async function AreasPage() {
  const [areas, users] = await Promise.all([
    listAreas(),
    listUsers(),
  ])

  // listAreas returns areas with areaCollections but not userAreas with user data.
  // Query userAreas with user info for each area.
  const { getDb } = await import("@rag-saldivia/db")
  const db = getDb()
  const allUserAreas = await db.query.userAreas.findMany({
    with: { user: true },
  })

  // Merge userAreas into areas
  const areasWithMembers = areas.map((area) => ({
    ...area,
    userAreas: allUserAreas
      .filter((ua) => ua.areaId === area.id)
      .map((ua) => ({
        userId: ua.userId,
        areaId: ua.areaId,
        user: { id: ua.user.id, name: ua.user.name, email: ua.user.email },
      })),
  }))

  const simpleUsers = users.map((u) => ({ id: u.id, name: u.name, email: u.email }))

  return <AdminAreas areas={areasWithMembers} allUsers={simpleUsers} />
}
