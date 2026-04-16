---
name: writing-plans
description: "Write implementation plans with bite-sized tasks, TDD anchors, and commit checkpoints. Use when spec/requirements are clear."
user_invocable: true
---

# Writing Plans — SDA Framework

Use this skill when requirements are clear and you need a structured implementation plan.

## Plan Structure

Every plan must have:

### 1. Goal Statement
One sentence: what does this plan deliver?

### 2. Prerequisites
- Which services exist that this depends on?
- Which `pkg/` packages are needed?
- Are there pending migrations?

### 3. Phases (max 5)
Each phase is independently deployable and testable.

```
Phase N: [Name]
├── Migration: [what SQL changes]
├── Backend: [what Go code]
├── Frontend: [what React changes]
├── Events: [what NATS subjects]
├── Tests: [what to test]
└── Checkpoint: make test && make lint && make build
```

### 4. Task Breakdown
Each task must be:
- **Bite-sized**: completable in one focused session
- **TDD-anchored**: test file written BEFORE implementation
- **Independently verifiable**: has a clear "done" signal
- **Commit-worthy**: each task = one commit

### 5. Invariants
What must NEVER break during implementation?
- Tenant isolation
- Existing API contracts
- Auth middleware chain

## SDA Plan Conventions
- Plans go in `docs/plans/` as `2.0.x-planNN-name.md`
- Plans are written in Spanish (per bible.md)
- Each plan must leave reusable infrastructure in `pkg/`
- NATS events follow `sda.{service}.{entity}.{action}` naming
- Migrations have matching `.up.sql` and `.down.sql`

## Anti-patterns
- Phases that can't be tested independently
- Tasks that require more than one service change
- Plans without explicit tenant isolation strategy
- Skipping the TDD anchor on tasks
