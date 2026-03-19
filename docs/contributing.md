# Contributing

## Development Workflow

This project follows a strict five-phase workflow for non-trivial changes:

1. Research → Brainstorm → Plan → Implement → Review

For details on what counts as "trivial" vs "non-trivial", tool stack per phase, and commit rules, see:

**[Development Workflow Documentation](development-workflow.md)**

Do NOT skip phases. Implementing without a spec and plan leads to poorly designed features and technical debt.

## Code Conventions

### Python (saldivia/, cli/, scripts/)

**Type Hints:**
- Required for all function signatures
- Use `from typing import ...` for generic types
- Example:

```python
def get_collection_stats(collection_name: str) -> dict[str, Any]:
    """Get statistics for a collection."""
    ...
```

**Error Handling:**
- No bare `except:` — always specify exception type
- Use `logging` module for errors, not `print()`
- Example:

```python
try:
    result = risky_operation()
except ValueError as e:
    logger.error(f"Invalid value: {e}")
    raise
```

**Style:**
- Follow existing style in `saldivia/` (Black-compatible)
- Line length: 100 characters (not 80)
- Imports: standard library → third-party → local (grouped with blank lines)
- Use f-strings, not `%` or `.format()`

### TypeScript (services/sda-frontend/)

**Strict Mode:**
- `tsconfig.json` has `strict: true` — no disabling
- No `any` types — use `unknown` and type guards
- Example:

```typescript
// Bad
function process(data: any) { ... }

// Good
function process(data: unknown) {
  if (typeof data === 'object' && data !== null) {
    // Type-safe access
  }
}
```

**Svelte 5 Runes Syntax:**
- Use `$state`, `$derived`, `$effect` (not legacy `let` + `$:`)
- Example:

```typescript
// Bad (Svelte 4 syntax)
let count = 0;
$: doubled = count * 2;

// Good (Svelte 5 runes)
let count = $state(0);
let doubled = $derived(count * 2);
```

**Imports:**
- Use absolute imports from `$lib/` (not relative)
- Example:

```typescript
// Bad
import { formatBytes } from '../../utils/formatting';

// Good
import { formatBytes } from '$lib/utils/formatting';
```

**Naming:**
- Components: PascalCase (`ChatInput.svelte`)
- Utilities: camelCase (`formatBytes.ts`)
- Stores: camelCase with `Store` suffix (`chatStore.ts`)
- Types: PascalCase (`type Message = ...`)

### Zone Boundaries

Do NOT use relative imports that cross zone boundaries. Each zone has its own README and should be self-contained:

**Zones:**
- `saldivia/` — Python SDK
- `cli/` — CLI commands
- `services/sda-frontend/src/lib/` — Frontend library code
- `services/sda-frontend/src/routes/` — Frontend routes
- `scripts/` — Deployment and utility scripts

**Example:**

```typescript
// Bad: importing from another zone via relative path
import { formatBytes } from '../../../saldivia/utils';

// Good: use the zone's public API
import { formatBytes } from '$lib/utils/formatting';
```

## Commit Rules

**Rule:** Commits are ONLY created when explicitly requested by the user.

**DO NOT:**
- Create commits proactively after completing a task
- Assume "done" means "commit now"
- Commit without user approval

**Commit Message Format:**

```
type(scope): description

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

**Types:**
- `feat` — new feature
- `fix` — bug fix
- `docs` — documentation only
- `test` — tests only
- `chore` — maintenance (deps, config, build)
- `refactor` — code change with no behavior change

**Scopes:**
- `deploy` — deployment scripts, profiles, docker-compose
- `crossdoc` — crossdoc pipeline and UI
- `gateway` — auth gateway, RBAC
- `frontend` — SvelteKit frontend
- `cli` — CLI commands
- `ingest` — ingestion pipeline, smart_ingest.py
- `tests` — test code

**Examples:**

```
feat(crossdoc): add decomposition UI with progress tracking
fix(gateway): handle SSE buffer splitting in gatewayGenerateText
docs: add architecture and deployment documentation
test(crossdoc): add E2E tests for 4-phase pipeline
chore(deploy): upgrade to NVIDIA RAG Blueprint v2.5.0
```

## README Maintenance

**Rule:** When modifying code in a zone, updating that zone's README is MANDATORY, not optional.

**Process:**
1. Identify the zone affected by your changes
2. Update the zone's README in the same commit
3. Update sections: new files, new functions, new behavior, examples

**Zone READMEs:**
- `/Users/enzo/rag-saldivia/README.md` — global README
- `/Users/enzo/rag-saldivia/saldivia/README.md` — Python SDK
- `/Users/enzo/rag-saldivia/cli/README.md` — CLI
- `/Users/enzo/rag-saldivia/services/sda-frontend/README.md` — Frontend
- `/Users/enzo/rag-saldivia/scripts/README.md` — Scripts
- `/Users/enzo/rag-saldivia/config/README.md` — Config and profiles

**Example:**

If you add a new API endpoint in `saldivia/gateway.py`, update `saldivia/README.md`:

```diff
## API Endpoints

...

+### POST /crossdoc/decompose
+
+Decompose a query into sub-queries for cross-document retrieval.
+
+**Request:**
+```json
+{ "query": "Compare A and B", "num_subqueries": 3 }
+```
```

**DO NOT commit without updating READMEs.**

## Branch Naming

Use prefixes to indicate the type of change:

- `feat/` — new feature (e.g., `feat/crossdoc-pipeline`)
- `fix/` — bug fix (e.g., `fix/sse-buffer-split`)
- `docs/` — documentation (e.g., `docs/architecture`)
- `test/` — tests (e.g., `test/e2e-crossdoc`)
- `chore/` — maintenance (e.g., `chore/upgrade-blueprint`)

**Naming style:** kebab-case, descriptive (3-5 words max)

**Examples:**

```bash
git checkout -b feat/crossdoc-synthesis
git checkout -b fix/milvus-downtime
git checkout -b docs/testing-guide
```

## PR Process

1. **Create PR against `main`** — no other target branches
2. **All tests must pass** — run `make test` locally before pushing
3. **Self-review first** — read your own diff, check for:
   - Type errors or lint warnings
   - Missing tests (every feature needs tests)
   - Stale READMEs (update if code changed)
   - Commented-out code (remove it)
4. **PR description** — include:
   - Link to spec (docs/superpowers/specs/)
   - Link to plan (docs/superpowers/plans/)
   - Summary of changes (what, why, how)
   - Testing notes (what you tested, what edge cases)
5. **Merge when approved** — squash merge preferred (keeps history clean)

**PR Template:**

```markdown
## Summary

Brief description of what this PR does.

## Links

- Spec: docs/superpowers/specs/2026-03-19-feature-design.md
- Plan: docs/superpowers/plans/2026-03-19-feature-implementation.md

## Changes

- Added X
- Modified Y
- Fixed Z

## Testing

- [ ] Unit tests pass
- [ ] Component tests pass
- [ ] E2E tests pass (if applicable)
- [ ] Manually tested happy path
- [ ] Manually tested error cases

## Screenshots (if UI change)

...
```

## Adding Dependencies

**Rule:** Never add a dependency without discussion.

**Process:**
1. Check if functionality exists in current stack
2. If not, propose the dependency in a GitHub issue or discussion
3. Wait for approval before adding

**Why:** Dependencies add:
- Bundle size (frontend)
- Security risk (supply chain attacks)
- Maintenance burden (upgrades, breaking changes)
- Complexity (learning curve for new contributors)

**Examples of unnecessary dependencies:**

- Don't add `lodash` when native JS has `Array.prototype.map/filter/reduce`
- Don't add `date-fns` when native `Intl.DateTimeFormat` works
- Don't add `axios` when `fetch` is built-in

**Examples of justified dependencies:**

- `@testing-library/svelte` — testing library with better API than manual DOM queries
- `zod` — runtime validation with type inference
- `sveltekit-superforms` — form handling with SSR validation

## Code Review Checklist

Use this checklist for self-review before requesting PR review:

- [ ] All tests pass (`make test`)
- [ ] No type errors (`npm run check` in frontend)
- [ ] No lint warnings (`npm run lint` in frontend)
- [ ] READMEs updated in affected zones
- [ ] Commit message follows format (type(scope): description)
- [ ] No commented-out code
- [ ] No debug print statements (use logging)
- [ ] No hardcoded secrets or production data
- [ ] No relative imports crossing zone boundaries
- [ ] New functions have type hints (Python) or TypeScript types (TS)
- [ ] Edge cases tested (null, undefined, empty, large)
- [ ] Error handling added (no bare except, no unhandled promises)

