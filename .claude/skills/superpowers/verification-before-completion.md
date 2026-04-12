---
name: verification-before-completion
description: "MANDATORY before declaring any task complete. Run tests, verify output, provide evidence. No claims without proof."
user_invocable: true
---

# Verification Before Completion — SDA Framework

This skill is MANDATORY before declaring ANY task complete. No exceptions.

## Protocol

### Step 1: Compile Check
```bash
make build
```
If this fails, the task is NOT complete.

### Step 2: Test Suite
```bash
# Service-specific tests
go test ./services/{name}/... -count=1 -v

# Full test suite
make test
```
If any test fails, the task is NOT complete.

### Step 3: Lint Check
```bash
make lint
```
If lint fails, fix it before declaring complete.

### Step 4: Structural Verification
```bash
# Run architectural invariants
bash .claude/hooks/check-invariants.sh
```

### Step 5: Evidence Collection
Before claiming completion, provide:

| Claim | Required evidence |
|-------|-------------------|
| "Tests pass" | Show `make test` output |
| "It compiles" | Show `make build` output |
| "No regressions" | Show `make lint` + test results |
| "Migration works" | Show up+down SQL files exist and are paired |
| "sqlc is in sync" | Show `make sqlc` produces no diff |
| "API works" | Show curl/httptest output |
| "Frontend works" | Show browser screenshot or dev server output |

### Step 6: Blast Radius Check
- What files did you change? List them.
- Did you change any `pkg/` package? Check all consumers.
- Did you change a migration? Verify it's reversible.
- Did you change a NATS subject? Check all publishers and subscribers.

## Completion Checklist
```
[ ] make build — compiles
[ ] make test — all tests pass
[ ] make lint — no lint errors
[ ] Invariant checks pass
[ ] No unintended file changes (git diff --stat)
[ ] Evidence provided for every claim
```

## Anti-patterns
- "It should work" without running tests
- "I verified mentally" without actual execution
- Claiming completion with failing tests
- Skipping lint because "it's just style"
- Not checking blast radius of pkg/ changes
