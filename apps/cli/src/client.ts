/**
 * Cliente HTTP para comunicarse con el servidor Next.js (apps/web).
 * La CLI habla con el mismo servidor que usa el frontend.
 */

const SERVER_URL = process.env["RAG_SERVER_URL_CLI"] ?? process.env["RAG_WEB_URL"] ?? "http://localhost:3000"

type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: string; suggestion?: string }

async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<ApiResult<T>> {
  try {
    const apiKey = process.env["SYSTEM_API_KEY"]
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(apiKey ? { Authorization: `Bearer ${apiKey}` } : {}),
    }

    const res = await fetch(`${SERVER_URL}${path}`, {
      ...options,
      headers: { ...headers, ...(options.headers as Record<string, string> ?? {}) },
      signal: AbortSignal.timeout(15000),
    })

    const body = await res.json().catch(() => null)

    if (!res.ok) {
      return {
        ok: false,
        error: body?.error ?? `HTTP ${res.status}`,
        suggestion: body?.suggestion,
      }
    }

    return { ok: true, data: body?.data ?? body }
  } catch (err) {
    const msg = String(err)
    return {
      ok: false,
      error: msg,
      suggestion: msg.includes("ECONNREFUSED")
        ? `El servidor no está corriendo en ${SERVER_URL}\n→ Inicialo con: bun run dev`
        : undefined,
    }
  }
}

// ── Endpoints ──────────────────────────────────────────────────────────────

export const api = {
  // Health
  health: () => apiFetch<{ status: string }>("/api/health"),

  // Auth
  login: (email: string, password: string) =>
    apiFetch<{ user: { id: number; name: string; role: string } }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  // Users
  users: {
    list: () => apiFetch<unknown[]>("/api/admin/users"),
    create: (data: object) =>
      apiFetch<unknown>("/api/admin/users", { method: "POST", body: JSON.stringify(data) }),
    delete: (id: number) =>
      apiFetch<void>(`/api/admin/users/${id}`, { method: "DELETE" }),
    update: (id: number, data: object) =>
      apiFetch<unknown>(`/api/admin/users/${id}`, { method: "PATCH", body: JSON.stringify(data) }),
  },

  // Areas
  areas: {
    list: () => apiFetch<unknown[]>("/api/admin/areas"),
    create: (name: string, description?: string) =>
      apiFetch<unknown>("/api/admin/areas", {
        method: "POST",
        body: JSON.stringify({ name, description }),
      }),
    delete: (id: number) =>
      apiFetch<void>(`/api/admin/areas/${id}`, { method: "DELETE" }),
  },

  // Collections
  collections: {
    list: () => apiFetch<string[]>("/api/rag/collections"),
    create: (name: string) =>
      apiFetch<unknown>("/api/collections", { method: "POST", body: JSON.stringify({ name }) }),
    delete: (name: string) =>
      apiFetch<void>(`/api/collections/${name}`, { method: "DELETE" }),
  },

  // Ingestion
  ingestion: {
    start: (collection: string, filePath: string) =>
      apiFetch<unknown>("/api/ingestion", {
        method: "POST",
        body: JSON.stringify({ collection, file_path: filePath }),
      }),
    status: () => apiFetch<unknown[]>("/api/ingestion/status"),
    cancel: (jobId: string) =>
      apiFetch<void>(`/api/ingestion/${jobId}`, { method: "DELETE" }),
  },

  // Config
  config: {
    get: () => apiFetch<Record<string, unknown>>("/api/admin/config"),
    set: (key: string, value: unknown) =>
      apiFetch<void>("/api/admin/config", {
        method: "PATCH",
        body: JSON.stringify({ [key]: value }),
      }),
    reset: () =>
      apiFetch<void>("/api/admin/config/reset", { method: "POST" }),
  },

  // Audit / Events
  audit: {
    list: (params?: { limit?: number; level?: string; type?: string }) => {
      const q = new URLSearchParams()
      if (params?.limit) q.set("limit", String(params.limit))
      if (params?.level) q.set("level", params.level)
      if (params?.type) q.set("type", params.type)
      return apiFetch<unknown[]>(`/api/audit?${q}`)
    },
    replay: (fromDate: string) =>
      apiFetch<{ timeline: unknown[]; stats: unknown }>(`/api/audit/replay?from=${fromDate}`),
    export: () => apiFetch<unknown[]>("/api/audit/export"),
  },

  // Sessions
  sessions: {
    list: (userId?: number) =>
      apiFetch<unknown[]>(`/api/admin/sessions${userId ? `?userId=${userId}` : ""}`),
    delete: (id: string) =>
      apiFetch<void>(`/api/chat/sessions/${id}`, { method: "DELETE" }),
  },

  // DB admin
  db: {
    migrate: () => apiFetch<void>("/api/admin/db/migrate", { method: "POST" }),
    seed: () => apiFetch<void>("/api/admin/db/seed", { method: "POST" }),
    reset: () => apiFetch<void>("/api/admin/db/reset", { method: "POST" }),
  },
}

export { SERVER_URL }
