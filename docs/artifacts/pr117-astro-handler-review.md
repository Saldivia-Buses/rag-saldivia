# Gateway Review -- PR #117 Astro Handler Wiring (Phase 13)

**Fecha:** 2026-04-06
**Archivo:** `services/astro/internal/handler/astro.go`
**Resultado:** CAMBIOS REQUERIDOS

---

## Bloqueantes

### B1. `tenantAndUser` silently ignores missing tenant [SECURITY] -- line 41

```go
info, _ := tenant.FromContext(r.Context())
```

The error from `FromContext` is discarded. If the middleware chain is misconfigured or the context is empty, `info.ID` is `""`, `tid.Scan("")` produces an invalid/zero `pgtype.UUID`, and all queries silently execute with a zero-UUID tenant filter. This does not leak cross-tenant data (zero UUID matches nothing), but it silently masks a broken auth chain -- the user gets "contact not found" instead of a clear 401/500.

**Fix:** Return an error and surface it as 401/500 in every caller:

```go
func tenantAndUser(r *http.Request) (pgtype.UUID, pgtype.UUID, error) {
    info, err := tenant.FromContext(r.Context())
    if err != nil {
        return pgtype.UUID{}, pgtype.UUID{}, fmt.Errorf("tenant: %w", err)
    }
    uid := sdamw.UserIDFromContext(r.Context())
    if uid == "" {
        return pgtype.UUID{}, pgtype.UUID{}, fmt.Errorf("missing user id in context")
    }
    var tid, uidPG pgtype.UUID
    if err := tid.Scan(info.ID); err != nil {
        return pgtype.UUID{}, pgtype.UUID{}, fmt.Errorf("invalid tenant id: %w", err)
    }
    if err := uidPG.Scan(uid); err != nil {
        return pgtype.UUID{}, pgtype.UUID{}, fmt.Errorf("invalid user id: %w", err)
    }
    return tid, uidPG, nil
}
```

### B2. `CreateContact` -- Content-Type header lost on 201 -- lines 325-326

```go
w.WriteHeader(http.StatusCreated)
jsonOK(w, contact)
```

`jsonOK` sets `Content-Type` AFTER `WriteHeader(201)` has already been called. Once `WriteHeader` is called, subsequent `w.Header().Set()` calls have no effect on the sent response. The response body is JSON but the Content-Type header is not set (defaults to `text/plain` via Go's sniffing, or potentially `application/octet-stream`).

**Fix:** Set header before WriteHeader, or add a status param to `jsonOK`:

```go
func jsonResponse(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}
```

### B3. No `http.MaxBytesReader` on any endpoint -- all handlers

None of the handlers limit request body size. A client can send a multi-GB body to any POST endpoint and the server will read it all into memory via `json.NewDecoder(r.Body).Decode()`.

**Fix:** Add at the top of `parseRequest` and `Query`, or as a middleware:

```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
```

---

## Debe corregirse

### D1. SSE Query leaks raw `err.Error()` to client -- lines 236, 269

```go
sseError(w, flusher, "invalid request: "+err.Error())  // line 236
sseError(w, flusher, "LLM error: "+err.Error())        // line 269
```

`err.Error()` from JSON decode can contain internal details (field names, types). The LLM error can contain the upstream URL, API keys in error messages, or internal status codes. Log the full error server-side, send a generic message to the client.

**Fix:**
```go
slog.Error("query decode failed", "error", err)
sseError(w, flusher, "invalid request body")
```

### D2. No validation on `contact_name` -- `parseRequest` line 84-92

`ContactName` is never checked for empty string. An empty `contact_name` will hit the DB with `lower("") = lower(name)` which could match a contact whose name is empty, or more likely return a confusing "contact not found" error. All 7 technique endpoints plus Query are affected.

**Fix:** Add to `parseRequest`:
```go
req.ContactName = strings.TrimSpace(req.ContactName)
if req.ContactName == "" {
    return nil, fmt.Errorf("contact_name is required")
}
```

### D3. Year validation missing -- `parseRequest`

Year 0, negative years, and absurdly far-future years (e.g., 99999) are accepted. Ephemeris calculations may panic or return garbage for years outside the Swiss Ephemeris data range (typically ~5000 BCE to ~5000 CE).

**Fix:**
```go
if req.Year < -5000 || req.Year > 5000 {
    return nil, fmt.Errorf("year out of supported range")
}
```

### D4. `resolveContact` does not distinguish "not found" from DB error -- line 49-58

```go
c, err := h.q.GetContactByName(r.Context(), ...)
if err != nil {
    return nil, err
}
```

All callers treat any error as 404. A connection timeout, a permission error, or a query syntax error all produce "contact not found" (404). The actual DB error is swallowed.

**Fix:** Use `pgx.ErrNoRows` to distinguish:
```go
if errors.Is(err, pgx.ErrNoRows) {
    return nil, ErrContactNotFound
}
return nil, fmt.Errorf("resolve contact: %w", err)
```

Then in handlers: check `errors.Is(err, ErrContactNotFound)` for 404, else 500.

### D5. `CreateContact` leaks raw DB error to client -- line 322

```go
jsonError(w, "create failed: "+err.Error(), http.StatusInternalServerError)
```

`err.Error()` from pgx contains table names, constraint names, and potentially SQL fragments. Log it, return a generic message.

### D6. SSE headers set before body decode -- Query handler lines 225-228 vs 235

The SSE headers (`text/event-stream`, etc.) are written before the request body is parsed. If the JSON decode fails, the error is sent as an SSE event. This means the client receives `Content-Type: text/event-stream` for what is effectively a 400 error. Clients that check Content-Type before opening an EventSource reader may be confused.

**Fix:** Parse and validate the request body first, then set SSE headers.

### D7. `ListContacts` hardcodes limit=100, no pagination params -- line 297

Limit and offset are hardcoded. The client cannot paginate. For a user with many contacts this returns only the first 100 with no way to get the rest.

**Fix:** Accept `?limit=` and `?offset=` query params with sane defaults and a max cap.

---

## Sugerencias

1. **SSE "fake streaming"**: The Query handler calls `SimplePrompt` which waits for the full LLM response, then chops it into 50-char chunks. This is cosmetic streaming -- the latency is the same as a non-streaming call. Consider using a real streaming chat completion method when available.

2. **`h.q == nil` guard**: Only `ListContacts` and `CreateContact` check for `h.q == nil`. The technique endpoints (Natal, SolarArc, Profections, Firdaria, Brief, Query) all call `resolveContact` which calls `h.q.GetContactByName` -- this will nil-pointer panic if `db` was not configured. Either add the guard everywhere or make `db` required at startup.

3. **SSE `Connection: keep-alive` header**: This is a hop-by-hop header that proxies (including Traefik) strip. It has no effect in HTTP/1.1 (where keep-alive is default) or HTTP/2. Harmless but unnecessary.

4. **Stub endpoints**: `Transits`, `Directions`, `Progressions`, `Returns`, `FixedStars` return 501. Fine for a WIP phase, but ensure they are tracked so they do not ship to prod as stubs.

5. **`sseEvent` ignores marshal error**: Line 338 `b, _ := json.Marshal(data)` -- if data contains a channel or func type, marshal silently fails and sends `null`. Low risk given the current data types but worth a defensive check.

---

## Lo que esta bien

- **Tenant isolation in SQL**: All sqlc queries (GetContactByName, ListContacts, CreateContact, etc.) filter by both `tenant_id` AND `user_id`. No query is missing the tenant filter. Contacts are per-user-per-tenant.
- **Auth middleware properly wired**: `cmd/main.go` applies `authMw` (with blacklist) to all `/v1/astro/*` routes. Health endpoint is outside the auth group.
- **SSE group correctly excludes chi Timeout**: The SSE endpoint is in a separate router group without `middleware.Timeout`, preventing the 5-minute deadline from killing long SSE connections. The `http.Server.WriteTimeout` of 5 minutes acts as the safety net.
- **Header spoofing protection**: The auth middleware strips `X-User-ID` et al before JWT verification. The handler reads identity from context (not headers directly).
- **Rate limiting**: Both route groups apply `rateMw` (10 req/min per user).
- **pgtype.Time handling**: The microseconds-to-hours conversion in `contactToChart` (line 65) is correct: `float64(us) / (3600 * 1e6)`.
- **Server timeouts**: `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout`, and `IdleTimeout` are all set with reasonable values.
- **JSON handler for slog**: Configured correctly in `cmd/main.go` line 29.
- **OTel instrumentation**: Both HTTP server (`otelhttp.NewHandler`) and LLM client (`otelhttp.NewTransport`) are instrumented for trace propagation.
