---
name: Plan 23 Phase 3 AI review gates
description: Review of ai-review.yml, claude-assist.yml, and 3 review prompts — JSON extraction broken, concurrency wrong level, HS256/ed25519 mismatch
type: project
---

Plan 23 Phase 3 — AI review gates and @claude assist (2026-04-14).

Result: CAMBIOS REQUERIDOS

**Blockers:**
- B2: JSON extraction with `grep -oP '[\s\S]*'` silently passes when Claude wraps JSON in code fences (the common case) — gate becomes a no-op
- B3: `concurrency:` in claude-assist.yml is at job level, not workflow level — allows multiple parallel runs per issue defeating the rate-limit intent

**Must fix:**
- M1: security.md says HS256 but CLAUDE.md invariant 2 says ed25519 — resolve against pkg/jwt/jwt.go before merge
- M2: quality.md missing header spoofing check, WS JWT pattern, astro mutex checks, correct HTTP status codes
- M3: dependencies.md missing invariant 4 (NATS event on every write) which applies to migration PRs
- M4: grep -oP multiline behavior unreliable on ubuntu-latest — replace with python3
- M5: concurrency group key for pull_request_review_comment may resolve to empty string

**What's correct:**
- DS6 applied correctly in Score findings step (env: var → file → jq, never ${{ }} interpolation of AI output)
- DS1: all jobs on ubuntu-latest
- Anti-injection notice in all 3 prompts (correct wording: reports as critical finding)
- Sonnet for dependency review (cost-appropriate)
- 7 invariants present verbatim in quality.md and security.md matching current CLAUDE.md

**Why:** The JSON extraction bug is the most dangerous — it turns the blocker gate into a soft warning that exits 0, meaning critical findings from Claude would never fail the PR check.
