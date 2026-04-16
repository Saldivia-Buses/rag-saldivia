---
name: requesting-code-review
description: "Request formal code review using specialized agents. Use after completing features or before merge."
user_invocable: true
---

# Requesting Code Review — SDA Framework

Use specialized review agents for formal code review after completing features.

## Protocol

### Step 1: Determine Review Type
| Changed files | Agent to use |
|---------------|-------------|
| `services/*/internal/`, `pkg/` | `gateway-reviewer` |
| `apps/web/`, `apps/login/` | `frontend-reviewer` |
| Auth, JWT, tenant, secrets | `security-auditor` |
| Both backend + frontend | Both reviewers in parallel |
| Pre-release | All three |

### Step 2: Prepare Review Context
Before dispatching the reviewer:
- Ensure all tests pass (`make test`)
- Ensure code compiles (`make build`)
- Have a clear summary of what changed and why

### Step 3: Dispatch Review Agent
```
Agent({
  subagent_type: "gateway-reviewer",
  description: "Review [feature name]",
  prompt: "Review the following changes for [feature]. 
    Files changed: [list]. 
    Purpose: [what and why].
    Pay special attention to: [specific concerns]."
})
```

### Step 4: Process Findings
Reviewer will report findings by severity:
- **Critical**: Must fix before merge
- **High**: Should fix before merge
- **Medium**: Fix in this PR if feasible
- **Low**: Fix in follow-up (per project convention: fix ALL including Low)

### Step 5: Fix All Findings
Per project convention, fix EVERYTHING down to Low severity. No open fronts.

## Review Checklist (what reviewers check)

**Gateway reviewer:**
- [ ] Tenant isolation in every handler
- [ ] Auth middleware on all protected routes
- [ ] RBAC permission checks
- [ ] Error wrapping with context
- [ ] sqlc query safety
- [ ] NATS event naming convention
- [ ] slog structured logging

**Frontend reviewer:**
- [ ] Real-time via WebSocket (no polling)
- [ ] Auth token handling
- [ ] Module guard for tenant-specific features
- [ ] Spanish UI text
- [ ] shadcn/ui component usage
- [ ] TanStack Query patterns

**Security auditor:**
- [ ] JWT validation
- [ ] SQL injection prevention
- [ ] Tenant boundary enforcement
- [ ] Secret exposure
- [ ] RBAC bypass vectors
