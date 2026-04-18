---
name: agent-tools
description: Use when adding a new agent tool, changing an existing tool's behavior, or wiring tool permissions. Owns the Phase 0 "every tool declares a capability + perm check" gate and the Phase 2 "Chat ↔ UI capability parity" goal (the agent is the user's representative — it can do anything the user can do in the UI, never more).
---

# agent-tools

Scope: every function the agent can call on the user's behalf.
Today these live in `services/app/internal/rag/agent/tools/` (Go) and
`modules/erp/tools.yaml` (declarative ERP tools). The contract
extends into `services/app/internal/core/auth/` for the RBAC check
at dispatch time.

## The north-star contract

**The agent is the user's representative.** It can do anything the
user can do by clicking in the UI, and nothing more. Every tool call
is subject to the same permission check as the corresponding HTTP
route the UI uses.

Put concretely:

- When the user says "cargame esta factura", the agent's `create_invoice`
  tool runs the same write path `POST /v1/erp/invoicing/invoices` uses,
  with the same auth header, the same validations, the same events.
- If the user doesn't have `erp.invoice.create` in their role, the
  tool refuses — same as the UI would reject the button.
- The audit log records "agent acting on behalf of user X" — not
  "system wrote this".

## Tool anatomy

A tool is five things. Missing any of them = not shippable.

```go
type Tool struct {
    Name        string              // "create_invoice"
    Description string              // for the LLM; must be operational, not marketing
    InputSchema *jsonschema.Schema  // validated before dispatch
    Capability  string              // "erp.invoice.create" — checked vs user perms
    Handler     ToolHandler         // executes the work
}

type ToolHandler func(ctx context.Context, user UserContext, input json.RawMessage) (ToolResult, error)
```

Rules:

1. **Capability is canonical**: dot-separated, `<area>.<entity>.<verb>`.
   `erp.invoice.create`, `erp.stock.transfer`, `rag.tree.query`. Register
   once, use everywhere. No free-form strings.
2. **Input is validated** against the schema before the handler runs.
   Schema errors → tool returns a structured error the LLM can read.
3. **Handler receives `UserContext`**, not a raw user ID. The context
   carries the JWT claims so the tool shares auth with every HTTP
   handler.
4. **Permission check is at dispatch** (framework concern), not inside
   the handler. A tool author cannot skip it by omission.
5. **Handler should prefer the existing HTTP handler's service layer**
   — avoid duplicating validation/business logic. The agent dispatcher
   and the HTTP handler funnel through the same service calls.

## Registration + discovery

Tools live in one of two places:

- **Go tools** — registered in `services/app/internal/rag/agent/tools/registry.go`.
  Each tool is a struct literal in an `init()` or a bootstrap list. One
  file per domain (`tools_erp.go`, `tools_rag.go`, `tools_calendar.go`).
- **Declarative tools** — in `modules/erp/tools.yaml` (or similar per
  module). The loader generates Go wrappers at startup. Useful for
  read-only tools whose handler is "run this SQL and format the rows".

The agent requests tool definitions from the registry at conversation
start, filtered by the user's capabilities. The LLM never sees tools
it can't call.

## Permission enforcement

At dispatch time, the framework does (pseudocode):

```go
if !user.HasCapability(tool.Capability) {
    return ToolResult{
        Error: "permission_denied",
        Detail: fmt.Sprintf("User lacks capability %s", tool.Capability),
    }, nil
}
// audit: "agent dispatching <tool> on behalf of <user>"
result, err := tool.Handler(ctx, user, input)
// audit: result summary
```

**Never** bypass this. If a tool needs elevated privilege (e.g., a
memory curator running as "system"), it runs in a **background
agent context** (see `background-agents` skill) — not as a user tool.

## Testing (non-negotiable)

Every tool ships with:

1. **A permission test per role** — `TestTool_Denies_WhenMissingCapability`
   for at least two roles (allowed + denied).
2. **A happy-path test** — handler produces the expected output for
   a canonical input.
3. **An input-validation test** — malformed input is rejected before
   the handler runs.
4. **An audit-log test** — the tool call was recorded with the user
   context.

Tests sit next to the tool file (`tools_erp_test.go`) and use
testcontainers for any tool that touches the DB.

## Chat ↔ UI parity checklist

For every tool you ship, answer:

- Does the UI have a corresponding button/form? (Yes → link the route in
  the tool description.)
- Does the tool call the same service-layer function the UI uses? (No →
  you're duplicating logic.)
- Does the permission string match the one the UI's route middleware
  checks? (No → bug.)
- Does a user without the permission get the same error message shape
  (UI vs chat)? (Consistency matters — the user shouldn't learn two
  systems of rejection.)

## Batch tools + confirm flow

Write tools that the user might want to preview before committing (bulk
stock transfer, price list update) should:

- Be split into `<tool>_preview` (read-only dry-run) + `<tool>_execute`
  (the actual write). The LLM is instructed to call preview, show the
  user, then confirm + execute.
- Return a `dry_run: true` field in the preview result so the UI can
  render a diff.
- Write an idempotency key so accidental double-confirmation doesn't
  double-execute.

The chat UI supports this pattern today — see
`apps/web/src/components/chat/ToolCall.tsx` for the confirm widget.

## What agent tools should NEVER do

- **Never** ignore the permission check (no "internal tool" exemption
  for user-agent tools — elevate via background agent if needed).
- **Never** bypass business logic — no `db.Exec` in a tool handler;
  always go through the service layer.
- **Never** take free-form input that expands to a SQL WHERE clause.
  Input is schemaed; free-form goes through the search tool which has
  its own safety.
- **Never** return raw DB errors to the LLM. Wrap in user-friendly
  messages that let the LLM explain.

## Integration with other skills

- **auth-security** owns the RBAC model — tool capabilities draw from
  that role map.
- **prompt-layers** defines how the tool registry is injected into
  the system context.
- **rag-pipeline** hosts the tool dispatcher; read it for the
  streaming / cancellation semantics.
- **htx-parity** owns the UI capability surface that tools mirror.

## Integration with ADR 027

This skill owns the Phase 0 "every tool declares a capability + perm
check" gate and the Phase 2 "Chat coverage: top-20 ERP write actions"
item. Every new tool either ticks one of those or extends the coverage
list in the ADR.

## Don't

- Don't ship a tool without its permission tests.
- Don't describe a tool in marketing language — the LLM uses the
  description to decide when to call it; be operational.
- Don't create a god-tool ("do_erp_thing" with a string verb). One
  tool per capability.
- Don't skip the audit log — the agent acting on a user's behalf must
  be traceable, per ADR 026.
