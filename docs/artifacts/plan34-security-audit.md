# Security Audit -- Plan 34 SSO (SAML/OIDC) -- 2026-04-01

## Resumen ejecutivo

SSO implementation is architecturally sound: CSRF via signed state tokens, PKCE for
OIDC, secrets encrypted at rest with AES-256-GCM, admin-only configuration, and no
IdP tokens stored. However, **2 CRITICAL and 3 HIGH findings** need fixing before
production. The most severe is the SAML state token validation bypass and the
`samlCert`/`samlEntryPoint` fields not being persisted to DB -- SAML will fail
silently or use empty certs.

---

## Hallazgos

### CRITICOS (bloquean deploy)

#### C-1. SAML state token validation is optional -- bypass possible

**File:** `apps/web/src/app/api/auth/callback/[provider]/route.ts:121-127`

```typescript
const stateToken = getCookieValue(request, "sso_token")
if (stateToken) {  // <-- NOT validated if cookie is absent
  const tokenPayload = await verifyStateToken(stateToken)
  if (!tokenPayload || ...) {
    return errorRedirect(...)
  }
}
```

**Impact:** In the SAML POST callback, the signed state token (`sso_token` cookie)
is only validated **if the cookie is present**. If an attacker strips the
`sso_token` cookie from the request (trivial -- just don't send it), the signed JWT
verification is completely bypassed. The plain `sso_state` cookie comparison at
line 115 still happens but it is a simple string comparison without cryptographic
binding -- an attacker who can observe the RelayState value (which travels through
the IdP) can forge the `sso_state` cookie.

Compare with the OIDC callback (line 57-63) which correctly treats a missing token
as a hard error:
```typescript
const stateToken = getCookieValue(request, "sso_token")
if (!stateToken) {
  return errorRedirect("Token de estado faltante", "invalid_state")
}
```

**Severity:** CRITICA -- CSRF protection partially bypassed for SAML flow.

**Fix:** Make the `sso_token` check mandatory in the SAML callback, identical to
the OIDC path:
```typescript
const stateToken = getCookieValue(request, "sso_token")
if (!stateToken) {
  return errorRedirect("Token de estado SAML faltante", "invalid_state")
}
const tokenPayload = await verifyStateToken(stateToken)
if (!tokenPayload || tokenPayload.provider !== "saml" || tokenPayload.state !== relayState) {
  return errorRedirect("Token de estado SAML expirado", "expired_state")
}
```

---

#### C-2. `samlCert` and `samlEntryPoint` not persisted to DB -- SAML broken

**File:** `packages/db/src/queries/sso.ts:67-84`

The `createSsoProvider()` function inserts these fields:
```typescript
name, type, clientId, clientSecretEncrypted, tenantId, issuerUrl,
scopes, autoProvision, defaultRole, active, createdAt, updatedAt
```

It does **NOT** insert `samlCert` or `samlEntryPoint`, even though:
- The schema defines them: `packages/db/src/schema/core.ts:256-257`
- The admin action accepts them: `apps/web/src/app/actions/sso.ts:33-34`
- The action **does not pass them** to `createSsoProvider()`: line 38-49
- `loadSamlProvider()` casts to `Record<string, string>` to read them: `sso.ts:197-199`

This means SAML providers created via admin will have `null` for both `samlCert`
and `samlEntryPoint`. The `loadSamlProvider()` function will use:
- `entryPoint: config.issuerUrl ?? ""` (fallback works IF admin sets issuerUrl)
- `idpCert: ""` (empty string -- **NO SIGNATURE VERIFICATION**)

**Impact:** With an empty `idpCert`, node-saml may accept SAML assertions without
verifying the IdP signature, enabling assertion forgery.

**Severity:** CRITICA -- Complete SAML authentication bypass if cert is empty.

**Fix:**
1. Update `createSsoProvider()` to persist `samlCert` and `samlEntryPoint`:
```typescript
samlCert: data.samlCert ?? null,
samlEntryPoint: data.samlEntryPoint ?? null,
```
2. Update `SsoProviderInput` type to include these optional fields.
3. Update `actionCreateSsoProvider` to pass `samlCert` and `samlEntryPoint` to the
   DB function.
4. Update `updateSsoProvider()` similarly.
5. Add a guard in `loadSamlProvider()` to reject providers with empty cert:
```typescript
const cert = config.samlCert ?? ""
if (!cert) return null  // refuse to load without cert
```

---

### ALTOS (corregir antes de produccion)

#### A-1. State comparison not timing-safe

**File:** `apps/web/src/app/api/auth/callback/[provider]/route.ts:51,116`

```typescript
if (!savedState || savedState !== state) {  // line 51 (OIDC)
if (!savedState || savedState !== relayState) {  // line 116 (SAML)
```

The state comparison uses JavaScript `!==` which is not constant-time. This could
theoretically leak information about the state value through timing side-channels.

**Impact:** The practical risk is low because (a) the state is random and
short-lived, (b) the signed JWT token provides a second layer. However, OWASP
recommends constant-time comparison for all security tokens.

**Severity:** ALTA

**Fix:** Use `crypto.timingSafeEqual`:
```typescript
import { timingSafeEqual } from "crypto"
function safeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) return false
  return timingSafeEqual(Buffer.from(a), Buffer.from(b))
}
```

---

#### A-2. Account linking without email verification -- account takeover vector

**File:** `apps/web/src/app/api/auth/callback/[provider]/route.ts:154-162`

```typescript
if (!user) {
  const existingUser = await getUserByEmail(userInfo.email)
  if (existingUser) {
    // ... auto-links SSO to existing account
    await linkSsoToUser(existingUser.id, providerType, userInfo.sub)
```

**Attack scenario:**
1. Victim has account `victim@company.com` (local password auth)
2. Attacker registers `victim@company.com` on a supported IdP (e.g., GitHub with
   any email, or a malicious SAML IdP if configured)
3. Attacker initiates SSO flow with that IdP
4. System finds existing user by email, auto-links SSO
5. Attacker now has full access to victim's account via SSO

The only guard is: `if (existingUser.ssoProvider && existingUser.ssoProvider !== providerType)`
which only prevents re-linking if already linked to a DIFFERENT provider. A user with
`ssoProvider === null` (all password-only users) gets linked without any challenge.

**Impact:** Full account takeover for any user whose email can be registered on a
configured IdP.

**Severity:** ALTA

**Fix:** Do NOT auto-link SSO to existing accounts without verification. Options:
- **Option A (recommended):** Require the user to be logged in (via password) to
  link an SSO provider. If not logged in, show "account exists, login with password
  to link SSO".
- **Option B:** Send email verification before linking.
- **Option C:** Disable auto-linking entirely -- only admin can link accounts.

---

#### A-3. Admin can configure `defaultRole: "admin"` for SSO auto-provisioning

**File:** `apps/web/src/app/actions/sso.ts:30` + `packages/db/src/schema/core.ts:259`

```typescript
defaultRole: z.enum(["admin", "area_manager", "user"]).optional(),
// ...
defaultRole: text("default_role", { enum: ["admin", "area_manager", "user"] })
```

The schema and action allow `defaultRole` to be set to `"admin"`. This means an
admin can configure an SSO provider where every auto-provisioned user gets admin
role. If the IdP is compromised or misconfigured, mass admin creation is possible.

**Impact:** Privilege escalation through IdP compromise + misconfigured
auto-provisioning.

**Severity:** ALTA

**Fix:** Remove `"admin"` from the allowed `defaultRole` values for SSO providers:
```typescript
defaultRole: z.enum(["area_manager", "user"]).optional(),
```
Admin accounts should only be created manually.

---

### MEDIOS (backlog prioritario)

#### M-1. Encryption falls back to plaintext when SYSTEM_API_KEY is missing

**File:** `packages/db/src/crypto.ts:27,42`

```typescript
export function encryptSecret(plaintext: string): string {
  const key = getEncryptionKey()
  if (!key) return plaintext  // <-- stores in plaintext
```

And on decrypt:
```typescript
export function decryptSecret(stored: string): string {
  const key = getEncryptionKey()
  if (!key) return stored  // <-- returns stored value as-is
```

In development mode or if `SYSTEM_API_KEY` is not a valid 32-byte base64 key,
client secrets are stored in plaintext in SQLite. While this is documented as
"dev mode fallback", production deployments could silently store secrets unencrypted
if the env var is malformed.

**Severity:** MEDIA

**Fix:** Add startup validation that `SYSTEM_API_KEY` is properly configured in
production:
```typescript
if (process.env.NODE_ENV === "production" && !getEncryptionKey()) {
  throw new Error("SYSTEM_API_KEY must be a valid 32-byte base64 key in production")
}
```

---

#### M-2. `listAllSsoProviders()` returns decrypted secrets -- serialization risk

**File:** `packages/db/src/queries/sso.ts:30-38`

```typescript
export async function listAllSsoProviders() {
  const rows = await db.select().from(ssoProviders)
  return rows.map((r) => ({
    ...r,
    clientSecret: decryptSecret(r.clientSecretEncrypted),  // decrypted!
    clientSecretEncrypted: undefined,
  }))
}
```

The function returns decrypted client secrets. While:
- The admin page (`admin/sso/page.tsx:9-18`) correctly strips `clientSecret` before
  passing to the component
- The `actionListSsoProviders` returns the full object to the client

The admin action at `actions/sso.ts:17` does `return { providers }` with the full
decrypted secret. This means the client secret travels to the browser in the server
action response.

**Severity:** MEDIA

**Fix:** Strip `clientSecret` in the action before returning:
```typescript
.action(async () => {
  const providers = await listAllSsoProviders()
  return {
    providers: providers.map(({ clientSecret, ...rest }) => rest)
  }
})
```

---

#### M-3. `oidc_generic` type uses Google implementation as placeholder

**File:** `apps/web/src/lib/auth/sso.ts:62-63`

```typescript
case "oidc_generic":
  return new Google(clientId, clientSecret, callbackUrl)
```

The `oidc_generic` provider type creates a Google Arctic instance. This is a
placeholder that will not work for non-Google OIDC providers (different
authorization/token endpoints).

**Severity:** MEDIA (functional, not security -- but misleading for admins)

**Fix:** Document this is not yet implemented, or implement proper generic OIDC using
configurable endpoints. Consider removing `oidc_generic` from the enum until properly
implemented.

---

#### M-4. SSO state cookie SameSite=Lax may be insufficient for SAML POST

**File:** `apps/web/src/app/api/auth/sso/[provider]/route.ts:28`

```typescript
const cookieOpts = `Path=/; HttpOnly; SameSite=Lax; Max-Age=${SSO_STATE_TTL_S}${isProduction ? "; Secure" : ""}`
```

SAML callbacks come as POST requests from the IdP (cross-origin). With
`SameSite=Lax`, cookies are sent on GET navigations but **NOT on cross-origin POST
requests**. This means the `sso_state` and `sso_token` cookies will not be sent by
the browser in the SAML POST callback, causing state validation to always fail.

**Severity:** MEDIA (SAML flow is functionally broken)

**Fix:** For SAML specifically, use `SameSite=None; Secure` for the state cookies.
This requires HTTPS in production (which is already configured via `Secure` flag).
For OIDC (which uses GET callbacks), `SameSite=Lax` is correct.

---

### BAJOS (nice to have)

#### B-1. Error redirect does not sanitize `code` parameter

**File:** `apps/web/src/app/api/auth/callback/[provider]/route.ts:30`

```typescript
const response = NextResponse.redirect(new URL(`/login?sso_error=${code}`, ...))
```

The `code` parameter comes from internal strings (not user input), so this is not
exploitable. However, defensive coding should URL-encode the parameter:
```typescript
`/login?sso_error=${encodeURIComponent(code)}`
```

**Severity:** BAJA

---

#### B-2. CSP allows 'unsafe-inline' and 'unsafe-eval' for scripts

**File:** `apps/web/next.config.ts:101`

```typescript
"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
```

While this is common with Next.js (which uses inline scripts), it weakens XSS
protection from CSP. This is not specific to the SSO implementation.

**Severity:** BAJA

---

#### B-3. SAML configuration type casting is fragile

**File:** `apps/web/src/lib/auth/sso.ts:197-199`

```typescript
entryPoint: (config as unknown as Record<string, string>).samlEntryPoint ?? config.issuerUrl ?? "",
idpCert: (config as unknown as Record<string, string>).samlCert ?? "",
```

Double cast to `unknown` then `Record<string, string>` bypasses type safety. If C-2
is fixed (fields properly persisted), this should use typed access:
```typescript
entryPoint: config.samlEntryPoint ?? config.issuerUrl ?? "",
idpCert: config.samlCert ?? "",
```

**Severity:** BAJA

---

## Positive findings (well done)

1. **CSRF on OIDC flow**: Properly implemented with dual-layer protection (random
   state in cookie + signed JWT state token with expiry). The OIDC callback
   correctly requires both.

2. **PKCE**: Properly implemented for Google and Microsoft. GitHub correctly
   identified as not supporting PKCE (wrapper with `noPkce: true`).

3. **Secrets at rest**: AES-256-GCM encryption for client secrets in DB. Proper
   random IV generation. Auth tag verified on decrypt.

4. **IdP tokens not stored**: The IdP access token is used only in-memory for
   `extractUserInfo()` and discarded. Not saved to DB or cookies.

5. **Cookie security**: All SSO cookies are HttpOnly. Secure flag in production.
   State cookies have proper Max-Age matching `SSO_STATE_TTL_S` (300s).

6. **Admin-only configuration**: All SSO provider CRUD uses `adminAction` with
   Zod validation. The admin layout requires `requireAdmin()`.

7. **Public endpoint is minimal**: `/api/auth/sso/providers` only exposes `id`,
   `name`, `type` -- no secrets, no client IDs.

8. **User deactivation check**: `findOrProvisionUser()` checks `user.active`
   before issuing JWT.

9. **Redirect safety**: All redirects go to hardcoded internal paths (`/chat`,
   `/login`) -- no user-controlled redirect targets.

10. **Session issuance reuses existing JWT pipeline**: `createAccessToken()` +
    `createRefreshToken()` + proper cookie helpers. No new auth code paths.

11. **SAML assertion validation**: `wantAuthnResponseSigned: true` and
    `wantAssertionsSigned: true` are correctly set (assuming cert is present).

12. **Email normalization**: `getUserByEmail()` and `createSsoUser()` both
    normalize to lowercase. However, see note below.

---

## Email normalization note

`getUserByEmail()` normalizes to lowercase (line 27 of users.ts), and
`createSsoUser()` stores lowercase (line 197). However, the SSO callback calls
`getUserByEmail(userInfo.email)` where `userInfo.email` comes directly from the IdP
without normalization. If an IdP returns `User@Example.COM`, the lookup will still
match because `getUserByEmail` lowercases its input. This is correct.

---

## CVEs relevantes

| Package | Version | Known CVEs |
|---|---|---|
| `arctic` | ^3.7.0 | No known CVEs as of 2025-05 |
| `@node-saml/node-saml` | ^5.1.0 | No known CVEs in v5.x as of 2025-05 |
| `jose` | (existing) | No known CVEs for HS256 usage |

Note: `@node-saml/node-saml` v5.x is the actively maintained fork. Earlier versions
(`passport-saml`, `node-saml` v4.x) had signature wrapping vulnerabilities (CVE-2022-39299).
v5.x mitigated these.

---

## Resumen de hallazgos

| Severidad | Cantidad | IDs |
|---|---|---|
| CRITICA | 2 | C-1, C-2 |
| ALTA | 3 | A-1, A-2, A-3 |
| MEDIA | 4 | M-1, M-2, M-3, M-4 |
| BAJA | 3 | B-1, B-2, B-3 |

---

## Veredicto: NO APTO para produccion

**Bloqueado por:** C-1 (SAML state bypass) and C-2 (SAML cert not persisted).

**Condiciones para aprobar:**
1. Fix C-1 and C-2 (mandatory before any deploy)
2. Fix A-1, A-2, A-3 (mandatory before production with real users)
3. Fix M-4 (SAML is broken without this regardless)
4. Remaining MEDIA/BAJA can go to backlog

---

*Audit performed by: security-auditor agent (Opus)*
*Scope: Plan 34 SSO files only (not full codebase re-audit)*
