# Gateway Review — Plan 32 Fase 1 & 5 (Self-Healing Error UX)

**Fecha:** 2026-04-01
**Tipo:** review
**Intensity:** standard

## Resultado
**CAMBIOS REQUERIDOS** (1 bloqueante, 2 debe corregirse, 5 sugerencias)

## Archivos revisados

| Archivo | Cambio |
|---|---|
| `apps/web/src/app/api/rag/generate/route.ts` | Rate limit details + error code in ragGenerateStream errors |
| `apps/web/src/lib/rag/client.ts` | `RATE_LIMITED` added to RagError code union |
| `apps/web/src/lib/error-recovery.ts` | New: getErrorRecovery + parseUseChatError |
| `apps/web/src/lib/api-utils.ts` | NOT modified (verified) |
| `apps/web/src/components/ui/error-recovery.tsx` | New: ErrorRecovery + ErrorRecoveryFromError components |
| `apps/web/src/components/chat/ChatInterface.tsx` | ErrorRecovery integration in chat |

## Hallazgos

### Bloqueantes

1. **[route.ts:113] 403 response leaks collection name from user input without sanitization**

   ```typescript
   return apiError(`Sin acceso a la colección '${col}'`, 403)
   ```

   The collection name `col` comes from user input (`body.collection_names`). While it IS validated by `CollectionNameSchema` (Zod), the 403 error echoes the requested collection name back to the client. This is an information disclosure vector: an attacker can enumerate collection names by trying names and seeing which get 403 (exists but no access) vs which would get a different error if the collection doesn't exist at all. The same issue exists on line 138 with `collectionName`.

   **This is a pre-existing issue (not introduced by Plan 32)**, but Plan 32 now surfaces these messages more prominently in the UI via `getErrorRecovery`. The `error-recovery.ts` FORBIDDEN handler (line 81) checks for "coleccion" in the message to decide the sub-variant. If the error message format changes, the recovery mapping breaks silently.

   **Fix:** Change both 403 messages to a generic `"Sin acceso a la colección solicitada"` (without echoing the name). The user already knows which collection they requested — they don't need it repeated. Alternatively, if you want to keep it for multi-collection requests where the user needs to know which one failed, at minimum ensure `error-recovery.ts` line 81 matching is robust (it currently is, since the API always includes "coleccion" in Spanish).

   **Severity:** Bloqueante because the bible says "La seguridad no es un tradeoff. Es una restriccion."

### Debe corregirse

2. **[route.ts:91-96] Rate limit response exposes `currentCount` — unnecessary internal detail**

   ```typescript
   return apiError(`Límite de ${maxQph} queries/hora alcanzado.`, 429, {
     code: "RATE_LIMITED",
     retryAfterMs: 3600_000,
     currentCount: count,
     maxCount: maxQph,
   })
   ```

   `retryAfterMs` and `maxCount` are useful for the UI and harmless (the user sees "10 queries/hora" in their own settings). But `currentCount` serves no UI purpose — `error-recovery.ts` never reads it. It tells an attacker exactly how many queries have been made, which is internal state they shouldn't need. Remove `currentCount` from the response. If needed for debugging, log it server-side (which line 180 already does).

   **Fix:** Remove `currentCount: count` from the details object.

3. **[error-recovery.ts:197] `parseUseChatError` status extraction regex is too broad — matches false positives**

   ```typescript
   const statusMatch = message.match(/\b(4\d{2}|5\d{2})\b/)
   ```

   This regex matches ANY 3-digit number starting with 4 or 5 in the error message. Examples of false positives:
   - `"Timeout después de 45000ms"` → matches `450` → `status: 450`
   - `"Error processing 500 documents"` → matches `500` → `status: 500`
   - `"Response contained 404 bytes"` → matches `404` → `status: 404`

   In practice, most error messages from `apiError` follow the pattern `"Límite de N queries/hora"` so collisions are unlikely today, but this is brittle. The regex should be anchored to actual HTTP error patterns.

   **Fix:** Either anchor the regex to typical HTTP error patterns like `/\bHTTP (\d{3})\b/` or `/ (\d{3})[:;]/`, or better yet: rely solely on `error.status` (which useChat provides when the response is non-2xx) and drop the regex fallback entirely. The AI SDK's useChat already sets `error.status` from the Response status code — the regex is a redundant heuristic.

### Sugerencias

4. **[route.ts:93] `retryAfterMs: 3600_000` is always worst-case — consider using standard `Retry-After` header**

   The 429 response hardcodes `retryAfterMs: 3600_000` (60 minutes). The plan explicitly chose this as a "conservative estimate" to avoid a new DB query, which is a reasonable YAGNI decision. However, the HTTP standard for rate limiting is the `Retry-After` header (seconds). Adding it costs one line:

   ```typescript
   return NextResponse.json(
     { ok: false, error: "...", details: { ... } },
     { status: 429, headers: { "Retry-After": "3600" } }
   )
   ```

   This would make the API more standards-compliant and help any future CLI/API consumers. The frontend can continue using `retryAfterMs` from the body.

5. **[error-recovery.ts:63-77] Rate limit recovery always shows "Respuestas guardadas" action — but the user might not have any saved responses**

   When rate-limited, the only action is `{ label: "Respuestas guardadas", type: "navigate", href: "/chat" }`. This navigates to the chat list. If the user has no previous sessions, this leads to an empty state. Consider adding a "dismiss" action as secondary, or changing the label to "Ver conversaciones" which sets a more accurate expectation.

6. **[error-recovery.ts:170-173] Generic fallback passes `message` directly as `description`**

   ```typescript
   description: message || "Ocurrió un error.",
   ```

   If the error comes from an unhandled server error that somehow bypasses `apiServerError`, the raw message could contain stack traces or internal details. The `apiServerError` function in `api-utils.ts` already sanitizes to "Error interno del servidor" for 500s, so in practice this path only fires for non-API errors (client-side exceptions). Still, defensively, consider capping the message length or stripping anything that looks like a file path.

7. **[error-recovery.ts:202] `parseUseChatError` heuristic for "no disponible" is locale-fragile**

   The code checks `message.includes("no disponible")` to detect UNAVAILABLE. This works because the API returns Spanish messages. But if any upstream error message happens to contain this substring in a different context, it would misclassify. Since the API now returns `code` in details, it would be better to attempt parsing the JSON error body to extract the code directly.

   Currently `parseUseChatError` receives a plain `Error` from useChat, not the parsed response body. The AI SDK's `useChat` in recent versions exposes the response body on error. If you upgrade or change how errors are caught, this heuristic becomes unnecessary. For now it's fine — just a note for future.

8. **[client.ts:33] `RATE_LIMITED` added to RagError code but never emitted by `createRagError`**

   The `RATE_LIMITED` code is in the RagError type union but `createRagError()` is never called with it — the rate limit is checked in the route handler before calling `ragGenerateStream`. This means the code in the union is purely for type completeness / documentation. That's acceptable, but it could be confusing to someone reading `client.ts` who expects every code to have a producer. Consider adding a brief comment on line 33 noting that `RATE_LIMITED` is only produced by the route handler, not by `ragGenerateStream`.

### Lo que esta bien

- **Auth pipeline is correct.** `requireAuth` validates JWT (including Redis revocation check via `extractClaims`). The middleware propagates `x-user-jti` correctly. No auth bypasses found.

- **Status code mapping is correct.** TIMEOUT -> 504, UNAVAILABLE -> 503, everything else -> 502. These match HTTP semantics exactly.

- **`apiError` correctly handles the new `details` format.** The existing signature `apiError(error: string, status = 400, details?: unknown)` handles any shape of details via spread. No changes needed in `api-utils.ts` — good decision in the plan.

- **`apiServerError` never leaks internals.** Line 43 always returns "Error interno del servidor" for 500s while logging the real error server-side. The Plan 32 changes don't bypass this.

- **Error recovery types are clean.** `ErrorInput`, `UserErrorRecovery`, and `ErrorAction` have well-defined shapes. All fields use `| undefined` instead of optional for `exactOptionalPropertyTypes` compatibility.

- **Test coverage is solid.** 17 tests for `getErrorRecovery` covering all codes, statuses, message patterns, fallback, and the invariant that every recovery has at least one action. 6 tests for `parseUseChatError` covering the main heuristics.

- **Priority ordering in `getErrorRecovery` is correct.** Code-based matching runs first (most specific), then status-based, then message pattern, then generic fallback. This prevents ambiguity where a code and status could conflict.

- **No `console.log` with sensitive data.** All logging goes through the structured logger. Error responses contain only user-facing messages.

- **Drizzle queries only.** `getRateLimit`, `countQueriesLastHour`, `canAccessCollection`, `getUserCollections` all use Drizzle ORM. No raw SQL.

- **SSE pipeline untouched.** `ai-stream.ts` was correctly left alone — the error handling is pre-stream in the route handler, and the stream adapter remains simple.

- **Redis patterns respected.** No new Redis imports in edge/middleware. `getRedisClient()` usage is confined to `jwt.ts` which runs in Node.js route handlers.

## Summary

The Plan 32 changes to the gateway layer are well-structured and follow the established patterns. The main concerns are:

1. **(Bloqueante)** Collection name echo in 403 errors — minor info disclosure that becomes more visible now that the UI renders these errors prominently.
2. **(Debe corregirse)** `currentCount` in rate limit response is unnecessary internal state.
3. **(Debe corregirse)** `parseUseChatError` status regex can false-positive on numbers in error messages.

Everything else is clean. The separation between server-side suggestions (`getSuggestion` in logger) and user-facing recovery (`getErrorRecovery` in lib) is a good architectural call.
