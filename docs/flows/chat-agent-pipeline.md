---
title: Flow: Chat + Agent Pipeline
audience: ai
last_reviewed: 2026-04-15
related:
  - ../services/chat.md
  - ../services/agent.md
  - ./websocket-realtime.md
---

## Purpose

Sequence of a single user turn: HTTP message lands in `chat`, agent runs
the LLM/tool loop, and the assistant reply flows back to the user via REST
+ NATS-driven WebSocket. Read this before changing the agent loop, the
chat persistence layer, the tool executor, or the events broadcast to the
hub. Architecture (RAG strategy, model contracts) is in
`services/agent.md` — this file owns the runtime sequence.

## Steps

1. Authenticated client posts the user message via
   `POST /v1/chat/sessions/{sessionID}/messages`; handler in
   `services/chat/internal/handler/chat.go` requires `chat.write` and
   delegates to `Chat.AddMessage`.
2. `Chat.AddMessage` (`services/chat/internal/service/chat.go:163`)
   creates the row in `chat.messages`, touches `chat.sessions.updated_at`,
   and returns the persisted `Message`.
3. For `role=user` the service emits a `chat.new_message` notification
   through the injected `EventPublisher`; the publisher writes a NATS
   message on `tenant.{slug}.notification.*` so the WS hub picks it up.
4. The frontend issues `POST /v1/agent/query` with the same prompt plus
   recent history; handler in `services/agent/internal/handler/agent.go`
   enforces `chat.read` and forwards the bearer JWT for tool calls.
5. `Agent.Query` (`services/agent/internal/service/agent.go:82`) reads the
   tenant from context, calls `tracePublisher.TraceStart`, runs input
   guardrails, and filters history to user/assistant roles only.
6. The service builds `messages = [system_prompt, ...history, user]` and
   enters the loop bounded by `MaxLoopIterations`; each iteration calls
   `pkg/llm/client.go` `Client.Chat` with `toolSchemas`.
7. If the LLM returns no tool calls, `guardrails.ValidateOutput` strips
   prompt-leak fragments and `publishTraceEnd` emits the final trace; the
   handler responds with `QueryResult{Response, ToolCalls, tokens}`.
8. Otherwise each `ToolCall` is validated against its schema and dispatched
   through `tools.Executor.Execute` at
   `services/agent/internal/tools/executor.go:70`; tools flagged
   `RequiresConfirmation` short-circuit with `pending_confirmation`.
9. Non-confirming tools execute via `executeHTTP` (or `grpcSearch` for
   `search_documents`) using the user's JWT, and the JSON result is
   appended back to `messages` with role `tool` for the next loop pass.
10. The frontend persists the assistant reply by posting it back to
    `Chat.AddMessage` (role `assistant`), which writes the row and lets
    the NATS bridge push the update to every WS subscriber on
    `tenant.{slug}.chat.messages:{session_id}`.

## Invariants

- The agent loop stops at `MaxLoopIterations` and at `MaxToolCallsPerTurn`
  (defaults 10 and 25); these MUST stay enforced before any LLM call.
- Tool calls always run with the original user JWT — the executor MUST NOT
  swap in a service token, otherwise tenant isolation breaks.
- All LLM/tool spans go through `TracePublisher` (start + end) so
  `services/traces` can compute cost and latency per turn.
- Output guardrails (`ValidateOutput`) MUST run before the response
  leaves the agent — they redact system-prompt leakage detected via
  `SystemPromptFragments`.
- Any tool with `RequiresConfirmation=true` MUST go through
  `/v1/agent/confirm`; `ExecuteConfirmed` rejects tools that lack the
  flag (`executor.go:105`).
- Every chat message persisted by `Chat.AddMessage` MUST happen before the
  NATS notification publishes — order matters for the WS hub consumers.

## Failure modes

- `400 invalid request body` / `400 message is required` — payload too
  large or empty; handler caps body at 256KB
  (`services/agent/internal/handler/agent.go:42`).
- `guardrails blocked` — input violated `guardrails.InputConfig`; check
  the rule that fired in `pkg/guardrails/input.go`.
- `loop detected, breaking` (slog warn) — `DetectLoop` saw repeated tool
  calls; break out and inspect `loopHistory` for the offending tool.
- `llm call: ...` errors — SGLang/OpenAI endpoint is down; the client has
  a 120s timeout and OTel span propagation, see `pkg/llm/client.go:38`.
- `unknown tool: ...` — manifest in `modules/{module}/tools.yaml` not
  loaded; verify `tools.Loader` registered the definition at startup.
- `pending_confirmation` returned but never resumed — the user dropped the
  approval; the agent does not auto-cancel, the frontend must drive
  `/v1/agent/confirm` or discard.
- WS clients miss messages — `chat.AddMessage` succeeded but the NATS
  publish failed silently (slog warn); see
  `services/chat/internal/service/chat.go:200` and confirm subject prefix
  `tenant.{slug}.notification`.
- Tool HTTP 4xx/5xx — `executeHTTP` records `{"error":"tool execution
  failed"}` in the loop; downstream service logs hold the real cause.
