---
title: Package: pkg/llm
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ../architecture/llm-sglang.md
  - ./guardrails.md
  - ./traces.md
---

## Purpose

The single OpenAI-compatible HTTP client for the entire SDA Framework. Works
with SGLang (local), OpenAI, Anthropic, or any compatible endpoint. Supports
chat, tool calling, streaming (SSE), and OpenTelemetry trace propagation via
`otelhttp`. See `architecture/llm-sglang.md` for the model-server topology.
Import this whenever a service needs to call an LLM — never roll your own
client.

## Public API

Source: `pkg/llm/client.go:7`

| Symbol | Kind | Description |
|--------|------|-------------|
| `Client` | struct | Endpoint + model + API key + traced HTTP client (120s timeout) |
| `NewClient(endpoint, model, apiKey)` | func | Constructor with `otelhttp.NewTransport` baked in |
| `Message` | struct | OpenAI message: `Role`, `Content`, `ToolCalls`, `ToolCallID`, `Name` |
| `ToolCall` / `FunctionCall` | struct | LLM-issued tool invocation |
| `ToolSchema` / `ToolDefinition` | struct | Tool registered with the LLM |
| `ChatResponse` | struct | `Content`, `ToolCalls`, `InputTokens`, `OutputTokens` |
| `StreamDelta` | struct | One SSE chunk: `Text`, token counts, `Done`, `Err` |
| `ChatClient` | interface | `Chat`, `SimplePrompt`, `StreamChat`, `Model`, `Endpoint` |
| `Client.Chat(ctx, msgs, tools, temp, maxTokens)` | method | Non-streaming chat completion |
| `Client.SimplePrompt(ctx, prompt, temp, maxTokens...)` | method | Single user-message convenience |
| `Client.StreamChat(ctx, msgs, temp, maxTokens)` | method | Returns `<-chan StreamDelta` |
| `Client.Model()` / `Client.Endpoint()` | method | Accessors |

## Usage

```go
c := llm.NewClient("http://sglang:30000", "qwen2.5", "")
resp, err := c.Chat(ctx, []llm.Message{
    {Role: "system", Content: systemPrompt},
    {Role: "user", Content: q},
}, tools, 0.2, 2048)

// Streaming
ch, _ := c.StreamChat(ctx, msgs, 0.7, 4096)
for delta := range ch {
    if delta.Err != nil { return delta.Err }
    fmt.Print(delta.Text)
}
```

## Invariants

- HTTP client is shared and uses `otelhttp.NewTransport` — every request
  participates in the active span (`pkg/llm/client.go:38`).
- `ChatClient` is an interface so handlers and services should depend on it
  rather than `*Client` to enable mocking.
- `StreamChat` uses a separate client without timeout (streams may exceed
  120s) (`pkg/llm/client.go:226`). Caller MUST drain the returned channel.
- Default `maxTokens` for `SimplePrompt` is 4096; for `StreamChat` also 4096
  if 0 is passed.
- API key is sent as `Authorization: Bearer <key>` only when non-empty.

## Importers

`services/agent`, `astro`, `search`, `ingest` (tree generation).
