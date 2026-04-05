# Agent Runtime

Chat interface — orchestrates LLM + tools. Replaces `services/rag/`.

Receives user messages, runs guardrails, calls LLM with available tools,
executes tool calls against other services, and returns responses with citations.

## Architecture

```
User message → guardrails → LLM (via slot.chat) → tool calls?
  YES → execute tools (Search, Ingest, etc.) → feed results back → LLM → ...
  NO  → output guardrails → response with citations
```

The loop runs up to `max_loop_iterations` times or until the LLM returns text.

## Endpoints

| Path | Method | Description |
|---|---|---|
| `/health` | GET | Health check |
| `/v1/agent/query` | POST | Run a query through the agent |

### POST /v1/agent/query

```json
{
  "message": "medida del disco de freno delantero del 9.20 LE",
  "history": []
}
```

## Environment

| Variable | Required | Description |
|---|---|---|
| `AGENT_PORT` | No | Default: `8004` |
| `JWT_PUBLIC_KEY` | Yes | Ed25519 public key (base64) |
| `SGLANG_LLM_URL` | No | Default: `http://localhost:8102` |
| `SGLANG_LLM_MODEL` | No | Model ID for chat slot |
| `LLM_API_KEY` | No | API key for cloud models |
| `SEARCH_SERVICE_URL` | No | Default: `http://localhost:8010` |
| `SYSTEM_PROMPT` | No | Override default system prompt |
