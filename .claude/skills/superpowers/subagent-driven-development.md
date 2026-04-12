---
name: subagent-driven-development
description: "Dispatch fresh subagent per task with TDD + two-stage review. Use when executing plan tasks."
user_invocable: true
---

# Subagent-Driven Development — SDA Framework

When executing plan tasks, dispatch a fresh subagent for each independent task.

## Protocol

### Step 1: Task Briefing
Each subagent gets a self-contained prompt with:
- **Goal**: What to implement (specific files, functions, endpoints)
- **Context**: What exists, what patterns to follow, which `pkg/` to use
- **Constraints**: Tenant isolation, auth middleware, NATS naming, sqlc conventions
- **TDD requirement**: Write test first, implement second
- **Verification**: What commands to run (`make test`, `make lint`, `go vet`)

### Step 2: Dispatch
```
Agent({
  description: "Implement [task name]",
  prompt: "[self-contained briefing with file paths, patterns, constraints]",
  subagent_type: "general-purpose"  // or specialized agent type
})
```

### Step 3: Two-Stage Review
When the subagent returns:

**Stage 1 — Spec Review:**
- Did the implementation match the spec/plan?
- Are all requirements addressed?
- Is tenant isolation respected?

**Stage 2 — Code Quality Review:**
- Does it follow Go conventions (chi, sqlc, slog)?
- Are error paths handled with wrapped errors?
- Is context propagated correctly?
- Are there tests with table-driven patterns?

### Step 4: Integration
After all subagents complete:
1. Run `make test` (full suite)
2. Run `make lint`
3. Run `make build`
4. Check for conflicts between subagent outputs
5. Verify NATS event flow across services

## Available Specialized Agents

| Agent | Use for |
|-------|---------|
| `gateway-reviewer` | Review Go handlers, middleware, auth, NATS |
| `frontend-reviewer` | Review React components, hooks, auth |
| `security-auditor` | Security audit before release |
| `test-writer` | Write Go tests + frontend tests |
| `debugger` | Diagnose failures |

## Anti-patterns
- Dispatching without self-contained context (agent has no memory of this conversation)
- Skipping the two-stage review
- Having subagents modify the same files (conflicts)
- Not running integration tests after all subagents complete
