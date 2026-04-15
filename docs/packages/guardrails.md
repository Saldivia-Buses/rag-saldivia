---
title: Package: pkg/guardrails
audience: ai
last_reviewed: 2026-04-15
related:
  - ../README.md
  - ./llm.md
---

## Purpose

Two-layer input/output validation for LLM interactions. Layer 1 is
deterministic (microseconds, zero cost): pattern match, length cap, JSON
schema check, loop detection. Layer 2 is semantic: an injected `LLMClassifier`
catches prompt injection attempts that pattern matching misses. Import this
in any service that ships user input to an LLM or returns LLM output.

## Public API

Source: `pkg/guardrails/guardrails.go:9`

| Symbol | Kind | Description |
|--------|------|-------------|
| `DefaultBlockPatterns` | var | Baseline prompt injection patterns (ES + EN) |
| `InputConfig` | struct | `MaxLength` (runes), `BlockPatterns` |
| `DefaultInputConfig(maxLength)` | func | InputConfig with `DefaultBlockPatterns` |
| `OutputConfig` | struct | `SystemPromptFragments` to redact if leaked |
| `LoopConfig` | struct | `MaxIterations`, `MaxIdenticalToolCalls` |
| `ToolCallRecord` | struct | `Tool`, `Params` (serialized) |
| `LLMClassifier` | interface | `Classify(ctx, prompt) (safe, reason, err)` |
| `ValidationError` | struct | `Layer` (1 or 2), `Reason` |
| `ValidateInput(ctx, input, cfg, llm)` | func | Truncate → patterns → classifier (if non-nil) |
| `ValidateOutput(output, cfg)` | func | Case-insensitive redaction of leaked fragments |
| `ValidateToolParams(params, schema)` | func | JSON validity, required fields, type check |
| `DetectLoop(history, cfg)` | func | Iteration cap + identical-call streak |

## Usage

```go
cfg := guardrails.DefaultInputConfig(8000)
clean, err := guardrails.ValidateInput(ctx, userInput, cfg, classifier)
if err != nil {
    httperr.WriteError(w, r, httperr.InvalidInput(err.Error()))
    return
}
```

## Invariants

- `MaxLength` truncates by RUNES, not bytes (`pkg/guardrails/guardrails.go:86`)
  to avoid breaking multi-byte characters.
- Layer 2 fails OPEN when the classifier returns an error
  (`pkg/guardrails/guardrails.go:105`) — the classifier being down doesn't
  block legitimate traffic.
- Pattern matching is case-insensitive (`pkg/guardrails/guardrails.go:93`).
- `ValidateOutput` does NOT call the LLM — output sanitization is Layer 1 only.
- `LLMClassifier` is an interface owned by the caller; this package does not
  pick a model.

## Importers

`services/agent/internal/service/agent.go` (input + tool params + loop),
`services/agent/cmd/main.go` (config wiring),
`services/chat/internal/handler/chat.go`,
`services/search/internal/service/search.go`.
