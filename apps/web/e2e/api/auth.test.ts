import { test, expect, beforeAll } from "bun:test";
import { TARGET, TEST_EMAIL, TEST_PASSWORD, login, apiGet, resetTokenCache } from "./helpers";

beforeAll(() => {
  resetTokenCache();
});

test("POST /v1/auth/login with valid creds returns access_token", async () => {
  const res = await fetch(`${TARGET}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: TEST_EMAIL, password: TEST_PASSWORD }),
  });
  expect(res.status).toBe(200);
  const data = (await res.json()) as { access_token: string; refresh_token?: string };
  expect(data.access_token).toBeTruthy();
  expect(typeof data.access_token).toBe("string");
});

test("POST /v1/auth/login with bad password returns 401", async () => {
  const res = await fetch(`${TARGET}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: TEST_EMAIL, password: "wrong-password" }),
  });
  expect(res.status).toBe(401);
});

test("POST /v1/auth/login with unknown email returns 401", async () => {
  const res = await fetch(`${TARGET}/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email: "no-such-user@nowhere.test", password: "x" }),
  });
  expect(res.status).toBe(401);
});

test("GET /v1/auth/me with token returns user object", async () => {
  const token = await login();
  const res = await apiGet("/v1/auth/me", { token });
  expect(res.status).toBe(200);
  expect(res.parsedJson).toBeTruthy();
  const user = res.parsedJson as { email: string; tenant_slug: string };
  expect(user.email).toBe(TEST_EMAIL);
  expect(user.tenant_slug).toBeTruthy();
});

test("GET /v1/auth/me without token returns 401", async () => {
  const res = await apiGet("/v1/auth/me");
  expect(res.status).toBe(401);
});

test("GET /v1/modules/enabled with token returns array", async () => {
  const token = await login();
  const res = await apiGet("/v1/modules/enabled", { token });
  expect(res.status).toBe(200);
  expect(Array.isArray(res.parsedJson)).toBe(true);
});
