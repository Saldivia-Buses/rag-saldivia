---
name: using-superpowers
description: "Meta-skill: establishes the skills framework at session start. Maps which skill to invoke for each type of task."
user_invocable: true
---

# Using Superpowers — SDA Framework

You have access to a set of specialized skills (superpowers) that dramatically improve your work quality. This skill establishes the framework.

## Skill Map

| Situation | Skill to invoke |
|-----------|----------------|
| New feature, service, or module | `brainstorming` first, then `writing-plans` |
| Clear spec/requirements ready | `writing-plans` |
| About to write implementation code | `test-driven-development` |
| Bug, test failure, unexpected behavior | `systematic-debugging` |
| Plan with 2+ independent tasks | `dispatching-parallel-agents` |
| Executing a multi-step plan | `executing-plans` |
| Each task dispatched to an agent | `subagent-driven-development` |
| About to declare work complete | `verification-before-completion` |
| Feature ready for review | `requesting-code-review` |
| Received review feedback | `receiving-code-review` |
| Implementation + tests done, ready to merge | `finishing-a-development-branch` |
| Need isolation for risky changes | `using-git-worktrees` |
| Creating or editing a skill | `writing-skills` |
| After implementing, review for quality | `simplify` |

## Priority Rules

When multiple skills could apply:
1. **Safety first**: `systematic-debugging` > `test-driven-development` > everything else
2. **Think before act**: `brainstorming` > `writing-plans` > implementation skills
3. **Verify before claim**: `verification-before-completion` is MANDATORY before declaring any task complete

## Rationalization Alerts

Stop and re-evaluate if you catch yourself thinking:
- "I'll skip the test since the change is small" → NO. Use `test-driven-development`.
- "I know what the bug is, let me just fix it" → NO. Use `systematic-debugging`.
- "This is working, I'll skip verification" → NO. Use `verification-before-completion`.
- "I'll refactor this while I'm here" → NO. Stay on task. Only fix what was asked.
- "I don't need to check blast radius for this" → NO. Always check.

## SDA-Specific Context

- **10 specialized agents** available: gateway-reviewer, frontend-reviewer, security-auditor, test-writer, debugger, deploy, status, doc-writer, plan-writer, ingest
- **MCP tools**: Repowise (get_overview, get_context, get_risk), CodeGraphContext (find_code, analyze_code_relationships)
- **Build system**: `make test`, `make lint`, `make build` — always verify with these
- **Multi-tenant**: Every change must respect tenant isolation
- **Go conventions**: chi handlers, sqlc queries, slog logging, testify tests
