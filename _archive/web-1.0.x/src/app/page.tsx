import { redirect } from "next/navigation"
import { headers } from "next/headers"

export default async function HomePage() {
  // El middleware ya verificó el JWT y puso los headers
  // Si llegan acá, están autenticados → redirect al chat
  const headersList = await headers()
  const userId = headersList.get("x-user-id")

  if (!userId) {
    redirect("/login")
  }

  redirect("/chat")
}
