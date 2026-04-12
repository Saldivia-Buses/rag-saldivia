---
name: using-git-worktrees
description: "Use git worktrees for isolated feature work. Prevents breaking the main working tree."
user_invocable: true
---

# Using Git Worktrees — SDA Framework

Use worktrees when you need isolation for risky changes or parallel feature work.

## When to Use
- Risky refactors that might break the build
- Exploring an approach you might abandon
- Working on a feature while keeping main tree clean
- Agent-dispatched work that needs isolation

## Protocol

### Step 1: Create Worktree
Use the Agent tool with `isolation: "worktree"`:
```
Agent({
  description: "Implement [feature]",
  prompt: "...",
  isolation: "worktree"
})
```

Or manually:
```bash
git worktree add ../rag-saldivia-feat-name feat/name
cd ../rag-saldivia-feat-name
```

### Step 2: Work in Isolation
- Make changes freely without affecting the main tree
- Run service-specific tests: `go test ./services/{name}/... -count=1`
- Commit frequently

### Step 3: Verify Before Merging Back
```bash
make test
make lint
make build
```

### Step 4: Merge or Discard
- **Success**: merge worktree changes back to feature branch
- **Failure**: discard worktree, no damage to main tree

### Cleanup
Worktrees from Agent tool are auto-cleaned if no changes.
Manual worktrees:
```bash
git worktree remove ../rag-saldivia-feat-name
```

## Anti-patterns
- Creating worktrees for trivial changes
- Forgetting to clean up abandoned worktrees
- Working in worktree without running tests before merge
