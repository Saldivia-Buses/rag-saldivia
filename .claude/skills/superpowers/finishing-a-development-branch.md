---
name: finishing-a-development-branch
description: "Clean up and merge a feature branch. PR creation, final checks, squash merge workflow."
user_invocable: true
---

# Finishing a Development Branch — SDA Framework

Use this skill when implementation + tests are done and the branch is ready to merge.

## Protocol

### Step 1: Pre-merge Checklist
```bash
# All tests pass
make test

# Code compiles
make build

# Lint clean
make lint

# Invariants hold
bash .claude/hooks/check-invariants.sh

# No unintended files
git status
git diff --stat main...HEAD
```

### Step 2: Clean Up
- Remove debug prints / temporary code
- Ensure no `TODO` or `FIXME` left unresolved
- Check for accidentally committed `.env`, credentials, or large files
- Ensure documentation is updated in same PR (bible.md rule)

### Step 3: Create PR
PR against the current working branch (currently `2.0.x`):
- Title: concise, under 70 characters
- Body: Summary bullets + test plan
- Link to plan if applicable

```bash
gh pr create --title "feat(service): description" --body "..."
```

### Step 4: Request Review
Use `requesting-code-review` skill to dispatch appropriate reviewer agents.

### Step 5: Post-Review
1. Fix all findings (including Low severity)
2. Re-run full verification
3. Squash merge when approved

## SDA Branch Conventions
- Feature branches: `feat/`, `fix/`, `refactor/`, `docs/`
- PR target: current working branch (`2.0.x` or as specified)
- Never commit directly to main
- Never force push to shared branches
- Squash merge to keep history clean

## Anti-patterns
- Merging without running full test suite
- Leaving TODO items in the code
- Creating PR without updated documentation
- Skipping the review step
- Pushing to main/upstream without explicit permission
