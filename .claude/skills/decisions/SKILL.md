---
name: decisions
description: Use when creating, updating, or deprecating architectural decisions (ADRs) in docs/decisions/. Covers the ADR file format, numbering, status lifecycle, when a change actually deserves an ADR vs a code comment, and how to deprecate a superseded decision cleanly.
---

# decisions

Scope: `docs/decisions/NNN-slug.md`.

## What an ADR is — and isn't

An ADR records a decision whose **consequences outlive its author**: why we picked
one option over others, under what assumptions, and what becomes invalid if those
assumptions change.

Write an ADR when:

- The decision constrains code that future contributors will write (e.g. "ed25519
  for JWT, not HS256").
- The decision rules out an obvious alternative (e.g. "tree-RAG, not vector kNN").
- Undoing it would require a migration, not just a refactor.

Don't write an ADR when:

- It's a local style choice (put that in the skill, not an ADR).
- It's a bug fix (that's a commit message).
- The decision is already obvious from the code (don't narrate).

## File format

```markdown
# ADR NNN — Short title

**Status:** proposed | accepted | superseded-by-NNN | deprecated
**Date:** YYYY-MM-DD
**Deciders:** <names or roles>

## Context

What question forced a decision. What constraints existed.

## Decision

The decision, in one sentence, then supporting detail.

## Consequences

Positive, negative, neutral. What becomes easier, what becomes harder.

## Alternatives considered

Each alternative, and why it was rejected.
```

No fluff, no history narration, no "we might revisit this later". If the decision
is not stable enough to commit to, it is not ready for an ADR.

## Numbering

- Sequential, zero-padded, three digits: `013`, `014`, `015`, …
- Next number: `ls docs/decisions/ | tail -1`, add 1.
- Never renumber. Never reuse a retired number.

## Lifecycle

1. **proposed** — draft, under discussion. Rare in this repo; most land as "accepted".
2. **accepted** — in force. Code must obey it.
3. **superseded-by-NNN** — a later ADR replaced it. Keep the file; update the header.
4. **deprecated** — no longer applies (e.g. the component it described was removed).
   Keep the file; update the header.

Never delete an ADR. The file is the audit trail.

## Superseding an ADR

When ADR `014` is replaced by ADR `042`:

1. Create `042-<slug>.md` with `Status: accepted` and a `## Supersedes` section
   pointing to `014`.
2. Edit `014`: change `Status:` to `superseded-by-042`. Leave the body intact.
3. Commit both in one change.

## After any architectural change

Before closing a session that changed how the system works, `ls docs/decisions/`
and ask: does an existing ADR now describe the world inaccurately? If yes — update
it, supersede it, or deprecate it in the same PR. Stale ADRs are worse than no ADRs.

## Checklist before merging a new ADR

- [ ] Number is sequential.
- [ ] Status is `accepted` (not `proposed`).
- [ ] Context names the specific problem, not a general theme.
- [ ] Decision is one sentence at the top of that section.
- [ ] Alternatives explains **why** each was rejected, not just that it was.
- [ ] No code sample longer than ~15 lines (put it in the repo, not the ADR).
- [ ] No other ADR is silently invalidated (if it is, supersede it explicitly).
