---
title: LLM via SGLang
audience: ai
last_reviewed: 2026-04-15
related:
  - ../packages/llm.md
  - ../services/agent.md
  - ../services/extractor.md
  - rag-tree-search.md
---

This document describes how SDA talks to large language models. Read it
before swapping models, adding a new provider, changing the LLM client
interface, or moving inference between local SGLang and a cloud endpoint —
this is the single seam through which every service consumes an LLM.

## Why SGLang

Inference runs on the inhouse workstation (RTX PRO 6000, 96 GB VRAM). SDA
hosts each model behind its own SGLang server, exposing an
**OpenAI-compatible** HTTP API (`/v1/chat/completions`). Model swap is a
config change — services depend on the API shape, not the model.

## Endpoints

| Env var              | Default                  | Purpose                  |
|----------------------|--------------------------|--------------------------|
| `SGLANG_LLM_URL`     | `http://localhost:8102`  | General-purpose chat / tool calling |
| `SGLANG_LLM_MODEL`   | (empty — server default) | Model id to send in payloads |
| `SGLANG_OCR_URL`     | `http://localhost:8100`  | PaddleOCR-VL document OCR |
| `SGLANG_VISION_URL`  | `http://localhost:8101`  | Qwen3.5-9B image description |
| `LLM_API_KEY`        | (empty)                  | Bearer auth for cloud providers |

Defaults are observed in `services/agent/cmd/main.go:48`,
`services/search/cmd/main.go`, `services/astro/cmd/main.go`,
`services/extractor/main.py:84`. In production each service reaches the
host's SGLang via `host.docker.internal`
(`deploy/docker-compose.prod.yml:773`).

## One process per model (slot per pipeline step)

Each pipeline step that needs a different model runs against a separate
SGLang process so the GPU loads each weight set once. As of today the
deployed slots are:

- LLM chat — `:8102`, used by `agent`, `search`, `astro`, `ingest` (tree
  generation and summaries).
- OCR — `:8100`, used by `extractor` for text/table extraction.
- Vision — `:8101`, used by `extractor` for image captioning.

Adding a new model class (e.g. a re-ranker) means standing up another
SGLang process on its own port and threading a new `*_URL` env through the
consumers. Do not multiplex two models behind one URL.

## The `llm` client

`pkg/llm/client.go:25` is the only HTTP client that should call an LLM.
Highlights:

- `NewClient(endpoint, model, apiKey)` returns a `*Client` whose
  transport is `otelhttp.NewTransport` so every call becomes a span
  (`client.go:33`). 120s default timeout.
- `Chat(ctx, messages, tools, temperature, maxTokens)` posts to
  `endpoint + "/v1/chat/completions"` and returns content, tool calls, and
  token counts.
- `SimplePrompt` is a one-shot convenience used by `search` and `ingest`
  prompts (`client.go:181`).
- `StreamChat` consumes the OpenAI SSE stream and emits `StreamDelta`
  values on a channel until `[DONE]` (`client.go:198`).
- The exported `ChatClient` interface is what services should depend on so
  tests can substitute a fake.

## Tool calling

The agent passes a `[]ToolSchema` to `Chat`. Each schema has a `name`,
`description`, and JSON-schema `parameters`
(`pkg/llm/client.go:67`). Tool definitions for the agent come from
`services/agent/internal/tools` and the YAML manifests in `modules/`
(`services/agent/cmd/main.go:62`). The LLM responds with `tool_calls` the
agent dispatches to the right downstream service.

## Tracing & guardrails

Because the client wraps `http.DefaultTransport` with `otelhttp`, every
LLM call appears in Tempo with parent context propagated (see
observability.md). The search service additionally validates user input
with `pkg/guardrails` before the LLM ever sees it
(`services/search/internal/service/search.go:73`).

## What you must never do

- Build a second LLM HTTP client — extend `pkg/llm` instead.
- Hardcode a model name in service code; read it from `SGLANG_LLM_MODEL`.
- Bypass guardrails on free-form user input that reaches the LLM.
- Run two SGLang processes on the same port.
