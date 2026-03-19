# Development Workflow

## The Fundamental Rule

No non-trivial change is implemented without following this five-phase workflow:

1. **Research** — Understand the codebase and gather context
2. **Brainstorm** — Explore the problem space and design space
3. **Plan** — Create a step-by-step implementation plan
4. **Implement** — Execute the plan with subagent-driven development
5. **Review** — Self-review and test before commit

This is not optional. Skipping phases leads to poorly designed features, broken tests, and technical debt.

## Trivial vs Non-trivial

### Trivial Changes (immediate implementation)

- Fixing a typo in code or documentation
- Updating a README with new information (1-2 sections)
- Changing a default value in a config file
- Adding a log statement for debugging
- Renaming a variable for clarity (1-3 occurrences)
- Formatting code (no logic change)

**Rule:** If the change touches ≤3 lines and the fix is obvious, implement immediately.

### Non-trivial Changes (mandatory workflow)

- New feature (new API endpoint, new UI component, new CLI command)
- Bug fix that touches >3 files or changes logic in >1 function
- Refactoring (extracting a function, splitting a file, changing data structures)
- Adding a new dependency (npm package, Python library, Docker service)
- Performance optimization (caching, batching, parallelization)
- Changing behavior of existing features (even if 1 line)

**Rule:** If you need to think about how to implement it, use the workflow.

## Tool Stack per Phase

### Phase 1: Research

**Objective:** Understand the existing codebase and gather external context.

**Tools:**

- **CodeGraphContext MCP** (mcp__CodeGraphContext__*)
  - `find_code`: Search for symbols, functions, classes by name or pattern
  - `analyze_code_relationships`: Find callers, callees, imports, exports
  - Use case: "Where is `gatewayGenerateStream` called?" → analyze_code_relationships
  - Use case: "Find all components that use `$state` runes" → find_code

- **repomix MCP** (mcp__repomix__pack_codebase)
  - Pack entire codebase or subdirectory into a single file for AI analysis
  - Use case: "Understand the structure of services/sda-frontend/src/lib/stores/"

- **firecrawl CLI** (never use WebSearch or WebFetch directly)
  - `firecrawl search "query"` — Search the web for documentation or solutions
  - `firecrawl scrape "https://url.com"` — Scrape a specific page (docs, blog, GitHub issue)
  - Use case: "How does Svelte 5's $effect handle cleanup?" → firecrawl search

**Output:** A clear understanding of:
- What code already exists and how it works
- What external libraries or patterns are relevant
- What constraints or gotchas exist (from docs/problems-and-solutions.md, CLAUDE.md)

### Phase 2: Brainstorm

**Objective:** Explore the problem space and design space before committing to a solution.

**Tools:**

- **superpowers:brainstorming skill** (MANDATORY for all non-trivial changes)
  - Generates a spec in markdown format
  - Saved to `docs/superpowers/specs/YYYY-MM-DD-feature-design.md`
  - Template includes: Problem, Goals, Non-goals, Exploration (3+ options), Decision, Risks

**Process:**
1. Invoke the skill with a clear problem statement
2. Review the generated spec
3. Iterate if needed (add more options, refine constraints)
4. Approve the spec before moving to planning

**DO NOT skip this step.** Implementing without a spec leads to:
- Solving the wrong problem
- Overengineering or underengineering
- Conflicts with existing architecture

### Phase 3: Plan

**Objective:** Create a step-by-step implementation plan.

**Tools:**

- **superpowers:writing-plans skill**
  - Generates a detailed plan in markdown format
  - Saved to `docs/superpowers/plans/YYYY-MM-DD-feature-implementation.md`
  - Template includes: Tasks (numbered, with file paths), Dependencies, Testing strategy, Rollout

**Process:**
1. Pass the approved spec to the planning skill
2. Review the generated plan for completeness
3. Check that all dependencies are handled (new libraries, config changes, database migrations)
4. Verify the testing strategy covers unit, component, and E2E tests

**Output:** A plan with 5-20 numbered tasks, each with:
- Clear description of what to do
- File paths to modify or create
- Success criteria

### Phase 4: Implement

**Objective:** Execute the plan with minimal deviation.

**Tools:**

- **superpowers:subagent-driven-development skill**
  - Executes the plan task-by-task
  - Handles file creation, modification, testing
  - Stops if tests fail or unexpected errors occur

**Process:**
1. Invoke the skill with the plan
2. Monitor progress (task 1/N, 2/N, ...)
3. If blocked → investigate, update plan, resume
4. If tests fail → fix tests or implementation, do not skip

**Rules:**
- Do NOT deviate from the plan without updating the plan file first
- Do NOT skip tests to "make it work"
- Do NOT commit until Phase 5 (review) is complete

### Phase 5: Review

**Objective:** Self-review and test before commit.

**Tools:**

- **superpowers:requesting-code-review skill**
  - Generates a review checklist
  - Runs all tests (unit, component, E2E)
  - Checks for common issues (type errors, missing tests, README not updated)

**Process:**
1. Invoke the skill with the list of changed files
2. Review the generated checklist
3. Run tests: `make test` (full pyramid) or `make test-unit` (fast feedback)
4. Fix any issues found
5. Update README files in affected zones (see README Maintenance Rule below)
6. Only after all checks pass → commit

**DO NOT commit until:**
- All tests pass
- README files in affected zones are updated
- No type errors or lint warnings
- You have manually tested the happy path

## Phase Lifecycle

### Specs and Plans Storage

- **Specs** (from brainstorming) → `docs/superpowers/specs/`
- **Plans** (from writing-plans) → `docs/superpowers/plans/`
- **Naming convention:** `YYYY-MM-DD-feature-description.md`
  - Example: `2026-03-19-crossdoc-pipeline-design.md`
  - Example: `2026-03-19-crossdoc-pipeline-implementation.md`

### Commits

**Rule:** Commits are ONLY created when explicitly requested by the user.

**DO NOT:**
- Create commits proactively after completing a task
- Assume "done" means "commit"
- Commit without user approval

**Reason:** The user may want to test, review, or modify the implementation before committing.

### How to Read docs/superpowers/

- **specs/** — "What to build" — problem, goals, options explored, final decision
- **plans/** — "How to build it" — step-by-step tasks with file paths and success criteria
- **Naming** — YYYY-MM-DD prefix for chronological sorting, kebab-case description

When starting a new feature:
1. Read relevant specs to understand past decisions
2. Read relevant plans to see how similar features were implemented
3. Follow the same structure for new specs and plans

## README Maintenance Rule

**Rule:** When modifying code in a zone (directory or feature), updating that zone's README is MANDATORY, not optional.

**Rationale:** READMEs are the primary documentation for developers. Stale READMEs lead to confusion, wasted time, and incorrect assumptions.

**Zones:**
- `saldivia/` → `saldivia/README.md`
- `cli/` → `cli/README.md`
- `services/sda-frontend/` → `services/sda-frontend/README.md`
- `scripts/` → `scripts/README.md`
- `config/` → `config/README.md`
- Root → `README.md`

**Process:**
1. Identify the zone affected by your changes
2. Update the zone's README in the same commit
3. Update sections: new files, new functions, new behavior, examples
4. If the change affects multiple zones, update all relevant READMEs

**Example:**
- Adding a new API endpoint in `saldivia/gateway.py` → update `saldivia/README.md` with endpoint description
- Adding a new CLI command in `cli/main.py` → update `cli/README.md` with command usage
- Adding a new route in `services/sda-frontend/src/routes/` → update `services/sda-frontend/README.md` with route description

**DO NOT commit without updating READMEs.**

