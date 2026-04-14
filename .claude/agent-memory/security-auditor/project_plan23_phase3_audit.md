---
name: Plan 23 Phase 3 AI gates audit
description: Security audit of ai-review.yml, claude-assist.yml, and review prompts for Plan 23 Phase 3
type: project
---

Audit of the AI review gate workflows for Plan 23 Phase 3 (2026-04-14).

**Verdict:** APTO with conditions — 0 critical, 2 high, 4 medium, 2 low.

**Key findings:**

HIGH 1: Scoring silently bypassed when model produces code-fenced output. The prompt examples use markdown code fences but instruct JSON-only output — the model follows the example format. jq fails, grep multiline extraction fails (no -z flag), file is empty, exit 0. The security gate looks green but never actually ran. Fix: remove fences from prompt examples, fix grep to use -z, fail-closed on jq parse error.

HIGH 2: anthropics/claude-code-action@v1 and actions/checkout@v4 are tag-pinned (mutable), not SHA-pinned. Compromise of upstream tag would affect all runs and could exfiltrate ANTHROPIC_API_KEY.

DS1 (runner isolation): PASS — all jobs on ubuntu-latest.
DS6 (anti-injection): PARTIAL — good heredoc delimiters, env: for exec_file, anti-injection in prompts. Scoring bypass undermines the guarantee.

**Why:** Model behavior with code fences is predictable and common. The bypass is not theoretical — it is the default behavior when the prompt contains fenced examples.

**How to apply:** When reviewing future AI-in-CI workflows, always trace the full path from model output to blocking decision. Silent exit 0 fallbacks are a red flag.
