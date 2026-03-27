import { test, expect } from "@playwright/test"

test.describe("Redis smoke tests", () => {
  test("GET /api/health retorna ok cuando Redis está up", async ({ request }) => {
    const res = await request.get("/api/health")
    expect(res.status()).toBe(200)
    const body = (await res.json()) as { ok?: boolean; status?: string }
    expect(body.ok).toBe(true)
    expect(body.status).toBe("healthy")
  })

  test("token JWT revocado después del logout", async ({ request }) => {
    const loginRes = await request.post("/api/auth/login", {
      data: { email: "admin@localhost", password: "changeme" },
    })
    expect(loginRes.ok()).toBeTruthy()

    const setCookie = loginRes.headers()["set-cookie"]
    expect(setCookie).toBeTruthy()
    const cookieHeader = Array.isArray(setCookie) ? setCookie.join(";") : setCookie
    const match = cookieHeader.match(/auth_token=([^;]+)/)
    expect(match?.[1]).toBeTruthy()
    const cookieValue = `auth_token=${match![1]}`

    const logoutRes = await request.delete("/api/auth/logout", {
      headers: { cookie: cookieValue },
    })
    expect(logoutRes.ok()).toBeTruthy()

    const protectedRes = await request.get("/api/admin/users", {
      headers: { cookie: cookieValue },
    })
    expect(protectedRes.status()).toBe(401)
  })

  test("GET /api/health documenta contrato redis/ts", async ({ request }) => {
    const res = await request.get("/api/health")
    const body = (await res.json()) as Record<string, unknown>
    expect(body).toHaveProperty("ok")
    expect(body).toHaveProperty("ts")
  })
})
