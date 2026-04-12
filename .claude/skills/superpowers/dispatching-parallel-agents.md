---
name: dispatching-parallel-agents
description: "Run 2+ independent tasks simultaneously via parallel agents. Use when plan has tasks with no dependencies between them."
user_invocable: true
---

# Dispatching Parallel Agents — SDA Framework

When a plan has 2+ tasks that are independent (different services, different files), run them in parallel.

## Protocol

### Step 1: Identify Parallelizable Tasks
Tasks are parallelizable when:
- They modify different services
- They modify different files within the same service
- They have no data dependency (output of one is not input of another)
- They don't share database migrations

Tasks are NOT parallelizable when:
- One task creates a `pkg/` package that another task imports
- One task writes a migration that another task's sqlc queries depend on
- Both tasks modify the same handler or service file

### Step 2: Dispatch in Single Message
All parallel agents MUST be dispatched in a single message block:

```
Agent({ description: "Task A — [service]", prompt: "..." })
Agent({ description: "Task B — [service]", prompt: "..." })
Agent({ description: "Task C — [service]", prompt: "..." })
```

### Step 3: Each Agent Gets
- Self-contained context (no reference to "the conversation")
- Specific file paths and patterns to follow
- TDD requirement
- Verification command for their scope

### Step 4: Collect and Integrate
When all agents return:
1. Review each agent's output
2. Check for unintended overlaps
3. Run `make test && make lint && make build`
4. Resolve any conflicts
5. Single integration commit or per-agent commits

## SDA Parallel Patterns

| Pattern | Agent A | Agent B |
|---------|---------|---------|
| Service + Frontend | Go handler/service | React component/hook |
| Multi-service | auth changes | chat changes |
| Backend + Tests | Implementation | Test writer |
| Code + Docs | Implementation | doc-writer agent |

## Anti-patterns
- Dispatching agents sequentially when they could be parallel
- Dispatching agents that will edit the same file
- Not reviewing agent outputs before integrating
- Missing the integration test after all agents complete
