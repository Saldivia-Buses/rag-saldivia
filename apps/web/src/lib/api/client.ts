/**
 * API client for SDA Framework.
 *
 * - Resolves tenant from hostname (saldivia.sda.app → "saldivia")
 * - Attaches Authorization header with access token
 * - Auto-refreshes expired tokens (401 → POST /v1/auth/refresh)
 * - Single retry on 5xx, no retry on 4xx
 */

import { useAuthStore } from "@/lib/auth/store";

// ---------------------------------------------------------------------------
// Tenant resolution
// ---------------------------------------------------------------------------

/**
 * Extracts the tenant slug from the current hostname.
 * saldivia.sda.app → "saldivia"
 * localhost → reads NEXT_PUBLIC_TENANT_SLUG or defaults to "dev"
 */
export function getTenantSlug(): string {
  if (typeof window === "undefined") return process.env.NEXT_PUBLIC_TENANT_SLUG ?? "dev";

  const host = window.location.hostname;

  // Development: localhost or IP
  if (host === "localhost" || host === "127.0.0.1" || host.startsWith("192.168.")) {
    return process.env.NEXT_PUBLIC_TENANT_SLUG ?? "dev";
  }

  // Production: {slug}.sda.app
  const parts = host.split(".");
  if (parts.length >= 3) return parts[0];

  return "dev";
}

/**
 * Returns the base URL for API requests.
 * In dev: proxied through Next.js or direct to Traefik.
 * In prod: same origin (Cloudflare Tunnel → Traefik).
 */
export function getApiBaseUrl(): string {
  if (typeof window === "undefined") return process.env.NEXT_PUBLIC_API_URL ?? "";
  return process.env.NEXT_PUBLIC_API_URL ?? "";
}

// ---------------------------------------------------------------------------
// Core fetch wrapper
// ---------------------------------------------------------------------------

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
  skipAuth?: boolean;
};

let refreshPromise: Promise<boolean> | null = null;

async function tryRefresh(): Promise<boolean> {
  // Deduplicate concurrent refresh attempts
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    try {
      const base = getApiBaseUrl();
      const res = await fetch(`${base}/v1/auth/refresh`, {
        method: "POST",
        credentials: "include", // send HttpOnly cookie
      });

      if (!res.ok) return false;

      const data = await res.json();
      useAuthStore.getState().setAccessToken(data.access_token);
      return true;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const base = getApiBaseUrl();
  const url = `${base}${path}`;

  const headers: Record<string, string> = {
    ...Object.fromEntries(new Headers(opts.headers).entries()),
  };

  // Attach auth header unless explicitly skipped
  if (!opts.skipAuth) {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
  }

  // Set Content-Type for JSON bodies
  if (opts.body !== undefined && !(opts.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const init: RequestInit = {
    ...opts,
    headers,
    credentials: "include", // always send cookies (for refresh token)
    body:
      opts.body instanceof FormData
        ? (opts.body as BodyInit)
        : opts.body !== undefined
          ? JSON.stringify(opts.body)
          : undefined,
  };

  let res = await fetch(url, init);

  // Auto-refresh on 401 (token expired)
  if (res.status === 401 && !opts.skipAuth) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      // Retry with new token
      const newToken = useAuthStore.getState().accessToken;
      if (newToken) {
        headers["Authorization"] = `Bearer ${newToken}`;
      }
      res = await fetch(url, { ...init, headers });
    } else {
      // Refresh failed — force logout
      useAuthStore.getState().clearAuth();
      throw new ApiError(401, "session expired");
    }
  }

  // Single retry on 5xx
  if (res.status >= 500) {
    res = await fetch(url, init);
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new ApiError(res.status, body.error ?? res.statusText);
  }

  // Handle 204 No Content
  if (res.status === 204) return undefined as T;

  return res.json();
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export const api = {
  get: <T>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "GET" }),

  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "POST", body }),

  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PATCH", body }),

  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PUT", body }),

  delete: <T = void>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "DELETE" }),

  /**
   * SSE streaming for RAG responses.
   * Returns an async iterator that yields parsed SSE data chunks.
   */
  stream: async function* (path: string, body: unknown): AsyncGenerator<string> {
    const base = getApiBaseUrl();
    const token = useAuthStore.getState().accessToken;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    };
    if (token) headers["Authorization"] = `Bearer ${token}`;

    const res = await fetch(`${base}${path}`, {
      method: "POST",
      headers,
      body: JSON.stringify(body),
      credentials: "include",
    });

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new ApiError(res.status, err.error ?? res.statusText);
    }

    const reader = res.body?.getReader();
    if (!reader) return;

    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() ?? "";

      for (const line of lines) {
        if (line.startsWith("data: ")) {
          const data = line.slice(6).trim();
          if (data === "[DONE]") return;
          try {
            const parsed = JSON.parse(data);
            const content = parsed.choices?.[0]?.delta?.content;
            if (content) yield content;
          } catch {
            // skip unparseable SSE lines
          }
        }
      }
    }
  },
};
