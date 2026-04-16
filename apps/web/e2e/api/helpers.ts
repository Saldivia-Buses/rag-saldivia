/**
 * Shared helpers for the API smoke suite (bun:test).
 *
 * Run with: TARGET=http://172.22.100.23 bun test apps/web/e2e/api/
 *
 * No browser involved — pure HTTP. The suite logs in once via the
 * test user seeded by db/tenant/migrations/053_e2e_test_user.up.sql
 * and reuses the access token across requests.
 */

export const TARGET = process.env.TARGET ?? "http://localhost";
export const TEST_EMAIL = process.env.TEST_EMAIL ?? "e2e-test@saldivia.local";
export const TEST_PASSWORD = process.env.TEST_PASSWORD ?? "testpassword123";

let cachedToken: string | null = null;

export async function login(): Promise<string> {
  if (cachedToken) return cachedToken;

  const res = await fetch(`${TARGET}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: TEST_EMAIL, password: TEST_PASSWORD }),
  });

  if (!res.ok) {
    const body = await res.text();
    throw new Error(
      `login failed: ${res.status} ${res.statusText} — ${body.slice(0, 200)}`,
    );
  }

  const data = (await res.json()) as { access_token: string };
  if (!data.access_token) {
    throw new Error("login response missing access_token");
  }
  cachedToken = data.access_token;
  return cachedToken;
}

export interface ApiCallResult {
  status: number;
  ok: boolean;
  contentType: string;
  bodyPreview: string;
  parsedJson: unknown | null;
  parseError: string | null;
}

export async function apiGet(
  path: string,
  opts: { token?: string; query?: Record<string, string> } = {},
): Promise<ApiCallResult> {
  const url = new URL(path, TARGET);
  if (opts.query) {
    for (const [k, v] of Object.entries(opts.query)) {
      url.searchParams.set(k, v);
    }
  }

  const headers: Record<string, string> = { Accept: "application/json" };
  if (opts.token) headers.Authorization = `Bearer ${opts.token}`;

  const res = await fetch(url, { method: "GET", headers });
  const contentType = res.headers.get("content-type") ?? "";
  const text = await res.text();

  let parsedJson: unknown = null;
  let parseError: string | null = null;
  if (contentType.includes("json")) {
    try {
      parsedJson = JSON.parse(text);
    } catch (err) {
      parseError = err instanceof Error ? err.message : String(err);
    }
  }

  return {
    status: res.status,
    ok: res.ok,
    contentType,
    bodyPreview: text.slice(0, 300),
    parsedJson,
    parseError,
  };
}

export function resetTokenCache() {
  cachedToken = null;
}
