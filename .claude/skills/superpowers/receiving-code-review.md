---
name: receiving-code-review
description: "Evaluate review feedback before implementing. Don't accept blindly. Use when receiving review comments."
user_invocable: true
---

# Receiving Code Review — SDA Framework

When you receive code review feedback, EVALUATE before implementing. Don't accept blindly.

## Protocol

### Step 1: Read All Findings
Read every finding completely before acting on any of them.

### Step 2: Classify Each Finding
For each finding, determine:

| Classification | Action |
|---------------|--------|
| **Correct and actionable** | Fix it |
| **Correct but out of scope** | Acknowledge, create follow-up task |
| **Partially correct** | Fix the valid part, explain the rest |
| **Incorrect / stale** | Explain why with evidence |
| **Style preference** | Follow project conventions (bible.md) |

### Step 3: Evaluate Before Acting
Before implementing any fix:
- Does this finding align with the project's conventions?
- Is the suggested fix the right approach for SDA's architecture?
- Could the fix introduce a regression?
- Does the reviewer have context that I'm missing?

### Step 4: Implement Valid Fixes
1. Fix findings in order of severity (Critical → Low)
2. Each fix gets its own test if applicable
3. Run `make test` after each batch of fixes
4. Don't introduce new issues while fixing review findings

### Step 5: Respond to Review
For each finding:
- **Fixed**: "Fixed in [file:line] — [what changed]"
- **Won't fix**: "This is intentional because [reason with evidence]"
- **Deferred**: "Valid finding, created follow-up task for [reason]"

## Anti-patterns
- Accepting every suggestion without thinking
- Rejecting feedback defensively
- Implementing fixes that break other things
- Not running tests after applying fixes
- Ignoring Low severity findings (project rule: fix ALL)
