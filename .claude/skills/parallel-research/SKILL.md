---
name: parallel-research
description: Use when a task needs investigation across two or more independent fronts — e.g. "audit all call sites of X AND check the tests AND look at the events", or "compare how service A and service B handle Y". Dispatches one-shot Agent() subagents in parallel (Explore for searches, general-purpose for multi-step research), then the main Claude synthesizes. No subagents live on disk; everything is spun up ad-hoc.
---

# parallel-research

Scope: any task where the work splits into investigation fronts that **don't depend
on each other**.

## When this skill applies

Activate when the task prompt contains 2+ **independent** questions, or when you
are about to do the same kind of search in 2+ different parts of the repo. Signals:

- "audit X across services A, B, and C"
- "how do we handle Y in backend vs frontend"
- "find every caller of Z and also look at the tests"
- "compare implementation in services/foo and services/bar"
- "before touching this, see how each consumer currently uses it"

If the fronts depend on each other ("first find X, then based on X find Y"),
**don't parallelize** — do it sequentially in the main context.

## The pattern

1. **Decompose** the task into N independent fronts. Each front has a clear
   question and a bounded scope (one directory, one concern).
2. **Choose agent type per front:**
   - **Explore** — fast lookup, grep/glob/read only. Use for "find all X", "list files
     matching Y", "who imports Z".
   - **general-purpose** — multi-step, may need to read several files and reason
     across them. Use for "how does flow X work end-to-end", "compare implementations".
3. **Dispatch in a single tool turn** — all Agent() calls in the same response.
   They run concurrently.
4. **Brief each agent self-contained** — they don't see this conversation. Give them:
   - What you are trying to accomplish (one sentence).
   - The specific question for their front.
   - The relevant files/directories to look at.
   - The format of the answer you want ("report as a bulleted list, ≤ 150 words").
5. **Synthesize yourself.** Agents report back; the main Claude combines them.
   Do not dispatch a 4th agent to "combine the other three" — that is your job.

## Example dispatch

Task: "Before I change the signature of `pkg/jwt.Parse`, tell me what breaks."

```
Agent(
  description: "Find jwt.Parse callers",
  subagent_type: "Explore",
  thoroughness: "medium",
  prompt: "List every call site of jwt.Parse in services/ and pkg/.
           For each, report file:line and whether it handles the returned error.
           Format: bullet per call, under 200 words."
)

Agent(
  description: "jwt.Parse test coverage",
  subagent_type: "Explore",
  thoroughness: "quick",
  prompt: "Find tests that exercise jwt.Parse. List file:line and what scenario each
           covers. Under 150 words."
)

Agent(
  description: "jwt.Parse in middleware",
  subagent_type: "general-purpose",
  prompt: "Read pkg/middleware/jwtmw/*.go. Explain exactly how jwtmw calls
           jwt.Parse, what context it expects, what it does on error. Under 200 words."
)
```

Three parallel reports → one synthesis → proceed.

## Rules

- **Never parallelize dependent work.** If front B needs the output of front A, run
  them in sequence.
- **Budget each agent.** Say "under 200 words" or "bullet list only". Unbounded
  agents return walls of text that pollute your context.
- **Don't delegate synthesis.** The main Claude is the one with the task goal —
  subagents don't know why the question matters.
- **Don't dispatch when a single `Grep` would do.** If the task is "find the file
  that defines X", that's one tool call, not an agent.
- **Cap at ~4 concurrent agents.** Beyond that, returns diminish.

## Anti-patterns

- "Spin up 5 agents to do the full feature in parallel." — Parallel **research**,
  not parallel **implementation**. Implementation is sequential in the main context
  (so edits don't conflict).
- "Dispatch an agent to decide what to do next." — Decisions stay with the main
  Claude. Agents gather facts.
- "Agent, then tell me what you found and I'll read your report." — The report is
  the only output you get; read it, extract what matters, synthesize. Don't ask
  for a second pass.

## Adversarial review pattern (doer + judge)

For high-risk fixes — anything touching auth, tenant boundaries, migrations with
backfill, changes to hot paths — a single-pass review is not enough. Dispatch
two agents **in the same turn** with opposing mandates:

- **Doer** — `general-purpose`, prompted to "propose the fix for <bug> in <file>.
  Include the code change, the reasoning, and the test that proves it works."
- **Judge** — `general-purpose`, prompted to "assume the proposed fix (paste the
  doer's intended change) is wrong. Find the failure modes it does not handle.
  Look for: race conditions under concurrency, edge cases with empty/nil,
  backwards-compat, silent data loss, cross-tenant leakage. Report under 300 words."

Then the main Claude synthesizes: accept, revise, or reject. Don't dispatch a
third agent to resolve the disagreement — that's the human/main-loop job.

Example:

```
# Two independent agents, one turn:

Agent(general-purpose,
  description: "Propose JWT clock-skew fix",
  prompt: "pkg/jwt/verify.go:62 rejects tokens issued ≤1s in the future because
           of clock skew between auth service and downstream. Propose a fix
           with leeway of 5s. Include code change, reasoning, and a test case
           matrix. Under 300 words.")

Agent(general-purpose,
  description: "Critique JWT leeway",
  prompt: "A change is about to add ±5s leeway to pkg/jwt.Verify to tolerate
           clock skew. Assume the fix is broken. Find the security/correctness
           holes: replay windows, revocation gaps, NTP failure modes, symmetric
           vs asymmetric risk. Report under 300 words. Be hostile.")
```

When to use it:

- Security-sensitive fixes.
- Migrations that touch production data.
- Concurrent / race-suspect bugs.
- Any fix where a wrong answer means an outage.

When **not** to use it:

- Typos, style nits, pure refactors with tests unchanged.
- Fixes where the failure mode is observable and cheap to revert.

Parallel research + adversarial review can chain: first do a 2–3 agent research
pass to gather facts, then a doer/judge pair on the proposed fix. The main
Claude stays the conductor.
