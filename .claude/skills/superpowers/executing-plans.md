---
name: executing-plans
description: "Batch execution of multi-step plans with checkpoints and progress tracking. Use when running plans in session."
user_invocable: true
---

# Executing Plans — SDA Framework

Use this skill when executing a plan that spans multiple phases and tasks.

## Protocol

### Step 1: Load the Plan
- Read the plan from `docs/plans/`
- Identify which phases/tasks are in scope for this session
- Create TaskCreate entries for each task

### Step 2: Pre-flight Check
Before starting any implementation:
- `git status` — clean working tree?
- `make build` — does it compile?
- `make test` — are existing tests passing?
- Check branch: are we on the right feature branch?

### Step 3: Execute Phase by Phase
For each phase:
1. Mark phase tasks as `in_progress`
2. Execute tasks in dependency order
3. After each task: `go test ./services/{service}/... -count=1`
4. Phase checkpoint: `make test && make lint && make build`
5. Commit with descriptive message: `feat({service}): plan N phase M — {description}`
6. Mark phase tasks as `completed`

### Step 4: Cross-Phase Verification
After all phases:
1. Full test suite: `make test`
2. Full lint: `make lint`
3. Full build: `make build`
4. Check migrations: are all `.up.sql` paired with `.down.sql`?
5. Check sqlc: `make sqlc` — regenerate and verify no diff
6. Use `verification-before-completion` skill

### Step 5: Document Progress
- Update the plan doc with completion status
- Note any deviations from the original plan
- List follow-up items for future sessions

## Checkpoint Commands
```bash
# Per-service check
go test ./services/{name}/... -count=1
go vet ./services/{name}/...

# Full platform check
make test
make lint
make build

# Migration check
ls db/tenant/migrations/*.up.sql | sed 's/.up.sql//' | while read f; do
  [ -f "${f}.down.sql" ] || echo "MISSING: ${f}.down.sql"
done
```

## Anti-patterns
- Skipping checkpoints between phases
- Continuing after a failing checkpoint
- Not tracking progress with tasks
- Implementing phases out of order without justification
