# Fase 5.1 + 5.2 — Documentation & Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Document the entire project with navigable READMEs at every level, and establish a full test pyramid (unit → component → E2E) covering all critical paths.

**Architecture:**
- Phase 5.1: Rewrite global README + 5 thematic docs in `docs/` + ~22 zone/subfolder READMEs in English, following an adaptable base template.
- Phase 5.2: Expand existing Vitest unit tests (≥80% coverage, threshold enforced), add `@testing-library/svelte` component tests for 6 critical UI components, add Playwright E2E with Page Object Model for 5 critical flows using `page.route()` mocks for CI-safe runs.
- All test data is illustrative — use generic, non-production values for credentials, queries, fixtures, and file contents.

**Tech Stack:** SvelteKit 5, TypeScript, Vitest 3.x, `@testing-library/svelte`, `@testing-library/user-event`, `@playwright/test`, `@vitest/coverage-v8`, Tailwind v4, Python/pytest (backend already covered).

**Branch:** `feat/fase-51-52` (created in Task 1)

**Commit strategy:** One commit per logical block (zone or test layer), not per file.

**TDD mode:**
- Existing code → write test, verify it passes
- New code → TDD strict: write failing test → implement → verify green

**README language:** English throughout.

**README template (adaptable base):**
```markdown
# <Folder Name>

Short description of the folder's purpose (1-3 lines).

## Files

| File | What it does | Key dependencies |
|------|-------------|-----------------|
| `file.ts` | ... | ... |

## Design notes (optional)

Only include if there is non-obvious design rationale worth explaining.
```

> **Note:** The items listed in tables and steps throughout this plan (test case names, component states, file contents) are **illustrative examples**. The implementor should read the actual source files and adapt, expand, or replace cases based on what they find.

---

## File Map

### Phase 5.1 — Created / Modified

| Action | Path |
|--------|------|
| Modify | `README.md` |
| Create | `docs/architecture.md` |
| Create | `docs/development-workflow.md` |
| Create | `docs/testing.md` |
| Create | `docs/deployment.md` |
| Create | `docs/contributing.md` |
| Create | `services/sda-frontend/README.md` |
| Create | `services/sda-frontend/src/lib/components/README.md` |
| Create | `services/sda-frontend/src/lib/components/chat/README.md` |
| Create | `services/sda-frontend/src/lib/components/ui/README.md` |
| Create | `services/sda-frontend/src/lib/components/sidebar/README.md` |
| Create | `services/sda-frontend/src/lib/components/layout/README.md` |
| Create | `services/sda-frontend/src/lib/stores/README.md` |
| Create | `services/sda-frontend/src/lib/utils/README.md` |
| Create | `services/sda-frontend/src/lib/crossdoc/README.md` |
| Create | `services/sda-frontend/src/lib/server/README.md` |
| Create | `services/sda-frontend/src/lib/actions/README.md` |
| Create | `services/sda-frontend/src/routes/README.md` |
| Create | `services/sda-frontend/src/routes/(app)/README.md` |
| Create | `services/sda-frontend/src/routes/(auth)/README.md` |
| Create | `services/sda-frontend/src/routes/api/README.md` |
| Create | `services/sda-frontend/src/routes/api/crossdoc/README.md` |
| Create | `saldivia/README.md` |
| Create | `saldivia/auth/README.md` |
| Create | `saldivia/tests/README.md` |
| Create | `config/README.md` |
| Create | `config/profiles/README.md` |
| Create | `patches/README.md` |
| Create | `patches/frontend/README.md` |
| Create | `scripts/README.md` |
| Create | `cli/README.md` |

### Phase 5.2 — Created / Modified

| Action | Path |
|--------|------|
| Modify | `services/sda-frontend/vitest.config.ts` |
| Modify | `services/sda-frontend/package.json` |
| Modify | `services/sda-frontend/src/lib/crossdoc/pipeline.test.ts` |
| Modify | `services/sda-frontend/src/lib/stores/chat.svelte.test.ts` |
| Modify | `services/sda-frontend/src/lib/stores/collections.svelte.test.ts` |
| Modify | `services/sda-frontend/src/lib/server/auth.test.ts` |
| Modify | `services/sda-frontend/src/lib/utils/markdown.test.ts` |
| Modify | `services/sda-frontend/src/lib/utils/scroll.test.ts` |
| Modify | `services/sda-frontend/src/routes/api/upload/upload.test.ts` |
| Modify | `services/sda-frontend/src/routes/api/crossdoc/decompose/decompose.test.ts` |
| Create | `services/sda-frontend/src/lib/actions/clickOutside.test.ts` |
| Create | `services/sda-frontend/src/routes/api/chat/stream/stream.test.ts` |
| Create | `services/sda-frontend/src/lib/components/chat/ChatInput.component.test.ts` |
| Create | `services/sda-frontend/src/lib/components/chat/MessageList.component.test.ts` |
| Create | `services/sda-frontend/src/lib/components/chat/CrossdocProgress.component.test.ts` |
| Create | `services/sda-frontend/src/lib/components/ui/Toast.component.test.ts` |
| Create | `services/sda-frontend/src/lib/components/ui/Modal.component.test.ts` |
| Create | `services/sda-frontend/src/routes/(app)/collections/_components/CollectionCard.component.test.ts` |
| Create | `services/sda-frontend/playwright.config.ts` |
| Create | `services/sda-frontend/tests/e2e/pages/LoginPage.ts` |
| Create | `services/sda-frontend/tests/e2e/pages/ChatPage.ts` |
| Create | `services/sda-frontend/tests/e2e/pages/CollectionsPage.ts` |
| Create | `services/sda-frontend/tests/e2e/pages/UploadPage.ts` |
| Create | `services/sda-frontend/tests/e2e/pages/AdminPage.ts` |
| Create | `services/sda-frontend/tests/e2e/fixtures/auth.ts` |
| Create | `services/sda-frontend/tests/e2e/fixtures/sample.pdf` |
| Create | `services/sda-frontend/tests/e2e/flows/auth.spec.ts` |
| Create | `services/sda-frontend/tests/e2e/flows/chat.spec.ts` |
| Create | `services/sda-frontend/tests/e2e/flows/collections.spec.ts` |
| Create | `services/sda-frontend/tests/e2e/flows/upload.spec.ts` |
| Create | `services/sda-frontend/tests/e2e/flows/crossdoc.spec.ts` |
| Modify | `Makefile` |

---

## Phase 5.1 — Documentation

---

### Task 1: Create branch

**Files:**
- No file changes — branch setup only

- [ ] **Step 1: Create and switch to branch**

```bash
cd /Users/enzo/rag-saldivia
git checkout -b feat/fase-51-52
git status
```

Expected: on branch `feat/fase-51-52`, working tree clean.

---

### Task 2: Write `docs/architecture.md`

**Files:**
- Create: `docs/architecture.md`

Read before writing: `CLAUDE.md` (project architecture section), `config/profiles/brev-2gpu.yaml`, `saldivia/gateway.py` (first 50 lines for port/service info).

- [ ] **Step 1: Write the file**

The file must cover:
1. **What it is** — overlay on NVIDIA RAG Blueprint v2.5.0, what it adds (auth, RBAC, multi-collection, SvelteKit 5 frontend, CLI, SDK, deployment profiles)
2. **Service map** — ports and roles:
   - Port 3000: SDA Frontend (SvelteKit 5 BFF)
   - Port 8081: RAG Server (NVIDIA Blueprint)
   - Port 8082: NV-Ingest
   - Port 9000: Auth Gateway (Saldivia)
   - Milvus: vector DB (no public port)
   - NIMs: embed, rerank, OCR (internal)
3. **Request flow diagram** (ASCII): `User → SDA Frontend (JWT cookie) → Auth Gateway (Bearer + RBAC) → RAG Server → Milvus + NIMs → LLM`
4. **Deployment profiles table**: brev-2gpu / workstation-1gpu / full-cloud — GPUs, services, use case
5. **1-GPU mode explanation**: QUERY vs INGEST mode, VRAM table, mode manager logic
6. **Key architectural decisions** (with rationale):
   - Client-side crossdoc decomposition (avoids reranker diversity loss)
   - HNSW on CPU + hybrid search (avoids GPU_CAGRA VRAM pressure)
   - Milvus GPU memory pool disabled (saves 3.7 GB VRAM)
   - Smart ingest with serial batching (avoids NV-Ingest deadlocks)

- [ ] **Step 2: Verify file exists and is non-empty**

```bash
wc -l docs/architecture.md
```

Expected: at least 60 lines.

---

### Task 3: Write `docs/development-workflow.md`

**Files:**
- Create: `docs/development-workflow.md`

This is the most important doc. Read before writing: `docs/superpowers/specs/` (any 2 spec files to understand the format), `docs/superpowers/plans/` (any 1 plan file).

- [ ] **Step 1: Write the file**

The file must cover:
1. **The fundamental rule**: no non-trivial change without `Research → Brainstorm → Plan → Implement → Review`
2. **Trivial vs non-trivial** with concrete examples:
   - Trivial: typo fix, 1-3 line obvious change, updating a README
   - Non-trivial: new feature, bug fix that touches >3 files, refactor, any new dependency
3. **Tool stack for each phase**:
   - Research: `mcp__CodeGraphContext__find_code`, `mcp__repomix__pack_codebase`, `firecrawl scrape`
   - Brainstorm: `superpowers:brainstorming` skill (mandatory before any implementation)
   - Plan: `superpowers:writing-plans` skill
   - Implement: `superpowers:subagent-driven-development` skill
   - Review: `superpowers:requesting-code-review` skill
4. **Phase lifecycle** — how a spec becomes a plan becomes code: spec saved to `docs/superpowers/specs/`, plan saved to `docs/superpowers/plans/`, commits only when Enzo asks explicitly
5. **How to read `docs/superpowers/`** — difference between `specs/` (what to build) and `plans/` (how to build it step by step), naming convention `YYYY-MM-DD-<feature>-design.md`
6. **README maintenance rule** — when modifying code in a zone, updating that zone's README is mandatory, not optional

- [ ] **Step 2: Verify**

```bash
wc -l docs/development-workflow.md
```

Expected: at least 80 lines.

---

### Task 4: Write `docs/testing.md`

**Files:**
- Create: `docs/testing.md`

Read before writing: `services/sda-frontend/vitest.config.ts`, `services/sda-frontend/package.json`.

- [ ] **Step 1: Write the file**

The file must cover:
1. **Test pyramid** — three layers with their purpose, tools, and speed
2. **Running tests**:
   ```bash
   make test           # full pyramid
   make test-unit      # vitest only (<5s)
   make test-e2e       # playwright only (needs app running)
   make test-coverage  # vitest + coverage report
   ```
3. **Vitest unit tests** — where they live (`src/**/*.test.ts`), how to run a single file, what `environment: node` vs `jsdom` means and when each applies
4. **Component tests** — where they live (`src/**/*.component.test.ts`), what `@testing-library/svelte` provides, example of rendering a component and querying by role
5. **Playwright E2E** — where they live (`tests/e2e/`), Page Object Model explanation, how `page.route()` mocks work, difference between CI-safe flows and `@slow` flows that need Brev
6. **Preview server setup for E2E**:
   ```bash
   cd services/sda-frontend
   npm run build
   npm run preview &   # starts on port 4173
   # Then in another terminal:
   npx playwright test
   ```
7. **Coverage** — threshold is 80% on `.ts` and `.svelte.ts` files in `src/lib/` and `src/routes/api/`; how to read the HTML report (`coverage/index.html`)
8. **Test data convention** — all data in tests is illustrative; no production credentials, no real vault data

- [ ] **Step 2: Verify**

```bash
wc -l docs/testing.md
```

Expected: at least 80 lines.

---

### Task 5: Write `docs/deployment.md`

**Files:**
- Create: `docs/deployment.md`

Read before writing: `config/profiles/brev-2gpu.yaml`, `config/profiles/workstation-1gpu.yaml`, `config/.env.saldivia`, `Makefile`.

- [ ] **Step 1: Write the file**

The file must cover:
1. **Deployment profiles table** — brev-2gpu, workstation-1gpu, full-cloud: what hardware, what services run locally vs via API
2. **Deploying to Brev** — step by step: SSH, pull, `make deploy PROFILE=brev-2gpu`, verify with `make status`
3. **Environment variables** — table of all vars in `.env.saldivia` with purpose and example value (no real secrets)
4. **Makefile commands reference** — all `make` targets with description
5. **Port table** — 3000, 8081, 8082, 9000 and what runs on each
6. **Known gotchas** (from CLAUDE.md patterns):
   - `docker network connect --alias` fails silently if container already on network → disconnect first
   - `PYTHONPATH` not defined + `set -u` = crash → use `${PYTHONPATH:-}`
   - Milvus `detect_types=PARSE_DECLTYPES` crashes with date-only timestamps
7. **1-GPU mode switching** — how the mode manager works, QUERY vs INGEST, VRAM table

- [ ] **Step 2: Verify**

```bash
wc -l docs/deployment.md
```

Expected: at least 70 lines.

---

### Task 6: Write `docs/contributing.md`

**Files:**
- Create: `docs/contributing.md`

- [ ] **Step 1: Write the file**

The file must cover:
1. **Development workflow** — reference to `docs/development-workflow.md`, don't repeat it
2. **Code conventions**:
   - Python: follow existing style in `saldivia/`, type hints required, no bare `except`
   - TypeScript: strict mode, no `any`, Svelte 5 runes syntax (`$state`, `$derived`, `$effect`)
   - Imports: absolute from `$lib/` in frontend, no relative imports crossing zone boundaries
3. **Commit rules** — only when Enzo explicitly asks; message format: `type(scope): description` (feat, fix, docs, test, chore, refactor)
4. **README maintenance rule** — mandatory: when modifying code in a zone, update that zone's README in the same commit
5. **Branch naming** — `feat/`, `fix/`, `docs/`, `test/` prefixes + kebab-case description
6. **PR process** — create PR against `main`, all tests must pass, self-review before requesting review
7. **Adding dependencies** — never add a dependency without discussing first; check if the functionality exists in the existing stack

- [ ] **Step 2: Verify**

```bash
wc -l docs/contributing.md
```

Expected: at least 60 lines.

---

### Task 7: Rewrite `README.md` global

**Files:**
- Modify: `README.md`

Read the current file first: `README.md`.

- [ ] **Step 1: Rewrite the file**

Structure (in order):
1. `# RAG Saldivia` + one-line description
2. Badges: CI status (placeholder), coverage (placeholder)
3. **What it is** — 3 lines max: overlay on NVIDIA RAG Blueprint v2.5.0, adds auth/RBAC/multi-collection/SvelteKit 5 frontend/CLI
4. **Architecture** — ASCII diagram of the service stack
5. **Quick Start** — exactly 5 commands to go from zero to running
6. **Documentation** — table of contents linking to `docs/`:
   | Doc | Description |
   |-----|-------------|
   | [Architecture](docs/architecture.md) | Service map, request flow, design decisions |
   | [Development Workflow](docs/development-workflow.md) | How to contribute and build features |
   | [Testing](docs/testing.md) | How to run and write tests |
   | [Deployment](docs/deployment.md) | Profiles, Brev, environment variables |
   | [Contributing](docs/contributing.md) | Code conventions, commits, PRs |
7. **Roadmap** — table of phases (1-5 completed, 6+ planned) — see existing docs for phase names
8. **Quick links** — `make test`, `make status`, `make deploy`

Keep it under 150 lines.

- [ ] **Step 2: Verify length**

```bash
wc -l README.md
```

Expected: ≤150 lines.

- [ ] **Step 3: Commit Phase 5.1 docs/ block**

```bash
cd /Users/enzo/rag-saldivia
git add README.md docs/architecture.md docs/development-workflow.md docs/testing.md docs/deployment.md docs/contributing.md
git commit -m "docs: add thematic docs and rewrite global README (5.1)"
```

---

### Task 8: READMEs — backend zones (saldivia/, config/, patches/, scripts/, cli/)

**Files:**
- Create: `saldivia/README.md`, `saldivia/auth/README.md`, `saldivia/tests/README.md`
- Create: `config/README.md`, `config/profiles/README.md`
- Create: `patches/frontend/README.md`
- Create: `scripts/README.md`
- Create: `cli/README.md`

Read each folder's files before writing its README.

- [ ] **Step 1: Write `saldivia/README.md`**

Files to document (read each briefly before writing):
`gateway.py`, `auth/`, `cache.py`, `collections.py`, `config.py`, `ingestion_queue.py`, `ingestion_worker.py`, `mcp_server.py`, `mode_manager.py`, `providers.py`, `watch.py`

Include a design note about the `_ts()` helper pattern (SQLite timestamp workaround).

- [ ] **Step 2: Write `saldivia/auth/README.md`**

Files: `database.py`, `models.py`. Include design note about why `detect_types=PARSE_DECLTYPES` was avoided.

- [ ] **Step 3: Write `saldivia/tests/README.md`**

Document each test file and what module it covers. Include how to run:
```bash
cd /Users/enzo/rag-saldivia
uv run pytest saldivia/tests/ -v
uv run pytest saldivia/tests/test_auth.py -v   # single file
```

- [ ] **Step 4: Write `config/README.md`**

Document each YAML and ENV file. Include a note about the compose file layering strategy.

- [ ] **Step 5: Write `config/profiles/README.md`**

Document each profile (brev-2gpu, workstation-1gpu, full-cloud) with a brief description of hardware requirements and use case.

- [ ] **Step 6: Write `patches/README.md`**

Overview of the `patches/` directory: what it is (overlay patch system for the NVIDIA Blueprint), what the two subdirectories do (`frontend/new/` for added files, `frontend/patches/` for `.patch` files).

- [ ] **Step 7: Write `patches/frontend/README.md`**

Document `new/` (files added to the blueprint: `SaldiviaSection.tsx`, `useCrossdocDecompose.ts`, `useCrossdocStream.ts`) and `patches/` (`.patch` files that modify existing blueprint files). Explain how to re-apply patches after a blueprint upgrade.

- [ ] **Step 8: Write `scripts/README.md`**

Files: `smart_ingest.py`, `crossdoc_client.py`, `stress_test.py`, `deploy.sh`, `health_check.sh`, `setup.sh`. For `smart_ingest.py` specifically, explain the tier system (tiny/small/medium/large by page count) and deadlock detection.

- [ ] **Step 9: Write `cli/README.md`**

Files: `main.py`, `areas.py`, `audit.py`, `collections.py`, `ingest.py`, `users.py`. Include a usage examples section with the most common commands.

- [ ] **Step 10: Commit**

```bash
git add saldivia/README.md saldivia/auth/README.md saldivia/tests/README.md \
        config/README.md config/profiles/README.md \
        patches/README.md patches/frontend/README.md \
        scripts/README.md cli/README.md
git commit -m "docs: add READMEs for backend zones (5.1)"
```

---

### Task 9: READMEs — frontend root + components

**Files:**
- Create: `services/sda-frontend/README.md`
- Create: `services/sda-frontend/src/lib/components/README.md`
- Create: `services/sda-frontend/src/lib/components/chat/README.md`
- Create: `services/sda-frontend/src/lib/components/ui/README.md`
- Create: `services/sda-frontend/src/lib/components/sidebar/README.md`
- Create: `services/sda-frontend/src/lib/components/layout/README.md`

Read each component's `.svelte` file briefly before writing its entry.

- [ ] **Step 1: Write `services/sda-frontend/README.md`**

Cover: what it is (SvelteKit 5 BFF), how to run (`npm run dev`, `npm run build`, `npm run preview`), how to run tests (`npm run test`), environment variables it needs (gateway URL, JWT secret), tech stack (SvelteKit 5, TypeScript, Tailwind v4, Vitest, Playwright).

- [ ] **Step 2: Write `src/lib/components/README.md`**

Overview of the three component categories: `chat/` (chat-specific), `ui/` (generic primitives), `sidebar/` and `layout/` (structural).

- [ ] **Step 3: Write `src/lib/components/chat/README.md`**

Components to document:
- `ChatInput.svelte` — message input with send button, crossdoc toggle, file attach
- `CrossdocProgress.svelte` — 4-phase progress indicator (decompose/search/synthesize/done)
- `CrossdocSettingsPopover.svelte` — settings panel for crossdoc parameters
- `DecompositionView.svelte` — shows sub-query decomposition results
- `HistoryPanel.svelte` — chat session history sidebar
- `MarkdownRenderer.svelte` — renders markdown with syntax highlighting
- `MessageList.svelte` — renders conversation messages with streaming support
- `SourcesPanel.svelte` — shows document sources for a response

- [ ] **Step 4: Write `src/lib/components/ui/README.md`**

Components: `Badge`, `Button`, `Card`, `Input`, `Modal`, `Skeleton`, `Toast`, `ToastContainer`.

- [ ] **Step 5: Write `src/lib/components/sidebar/README.md`**

Components: `Sidebar`, `SidebarItem`.

- [ ] **Step 6: Write `src/lib/components/layout/README.md`**

Components: `Sidebar` (layout variant — note the naming distinction from sidebar/).

- [ ] **Step 7: Commit**

```bash
git add services/sda-frontend/README.md \
        services/sda-frontend/src/lib/components/
git commit -m "docs: add READMEs for frontend root and components (5.1)"
```

---

### Task 10: READMEs — lib/ subfolders (stores, utils, crossdoc, server, actions)

**Files:**
- Create: `src/lib/stores/README.md`, `src/lib/utils/README.md`
- Create: `src/lib/crossdoc/README.md`, `src/lib/server/README.md`, `src/lib/actions/README.md`

All paths relative to `services/sda-frontend/`.

- [ ] **Step 1: Write `src/lib/stores/README.md`**

Files: `chat.svelte.ts`, `collections.svelte.ts`, `crossdoc.svelte.ts`, `toast.svelte.ts`. Explain the Svelte 5 `$state` runes pattern used. Note that stores are classes, not the old Svelte 4 writable() pattern.

- [ ] **Step 2: Write `src/lib/utils/README.md`**

Files: `chat-utils.ts`, `markdown.ts`, `scroll.ts`. For `markdown.ts` note the DOMPurify XSS sanitization step.

- [ ] **Step 3: Write `src/lib/crossdoc/README.md`**

Files: `pipeline.ts`, `types.ts`. Explain the 4-phase pipeline (decompose → parallel subquery → rerank/dedup → synthesize), the Jaccard dedup algorithm, and the `hasUsefulData` gate. This is the most complex module in `lib/` — spend extra detail here.

- [ ] **Step 4: Write `src/lib/server/README.md`**

Files: `auth.ts`, `gateway.ts`. Note that these are server-only files (BFF layer) — they import `$env/static/private` and must never be imported from client-side code.

- [ ] **Step 5: Write `src/lib/actions/README.md`**

Files: `clickOutside.ts`. Include a usage example:
```svelte
<div use:clickOutside on:clickoutside={handleClose}>...</div>
```

- [ ] **Step 6: Commit**

```bash
git add services/sda-frontend/src/lib/stores/README.md \
        services/sda-frontend/src/lib/utils/README.md \
        services/sda-frontend/src/lib/crossdoc/README.md \
        services/sda-frontend/src/lib/server/README.md \
        services/sda-frontend/src/lib/actions/README.md
git commit -m "docs: add READMEs for lib/ subfolders (5.1)"
```

---

### Task 11: READMEs — routes

**Files:**
- Create: `src/routes/README.md`, `src/routes/(app)/README.md`
- Create: `src/routes/(auth)/README.md`, `src/routes/api/README.md`
- Create: `src/routes/api/crossdoc/README.md`

All paths relative to `services/sda-frontend/`.

- [ ] **Step 1: Write `src/routes/README.md`**

Explain SvelteKit route groups: `(app)/` is auth-guarded (layout checks JWT), `(auth)/` is public (login), `api/` is BFF endpoints. Include the redirect logic: `/` → `/chat` if authenticated.

- [ ] **Step 2: Write `src/routes/(app)/README.md`**

List all routes: `chat/`, `chat/[id]/`, `collections/`, `collections/[name]/`, `admin/users/`, `admin/areas/`, `admin/permissions/`, `admin/rag-config/`, `admin/system/`, `audit/`, `settings/`, `upload/`. One line per route explaining what it does.

- [ ] **Step 3: Write `src/routes/(auth)/README.md`**

Only `login/`. Explain the form, what it posts to, and how the JWT cookie is set.

- [ ] **Step 4: Write `src/routes/api/README.md`**

Document all BFF endpoints:
- `auth/session/` — validate/refresh JWT session
- `chat/sessions/` — CRUD for chat sessions
- `chat/sessions/[id]/` — single session operations
- `chat/stream/[id]/` — SSE proxy to gateway
- `collections/` — list/create collections
- `collections/[name]/` — get/delete single collection
- `crossdoc/decompose/` — query decomposition
- `crossdoc/subquery/` — single subquery execution
- `crossdoc/synthesize/` — final synthesis
- `upload/` — file upload to NV-Ingest
- `dev-login/` — development-only login bypass

- [ ] **Step 5: Write `src/routes/api/crossdoc/README.md`**

Explain the 3 crossdoc endpoints as a sequence: client calls decompose → for each subquery calls subquery → calls synthesize. Include sequence diagram in ASCII.

- [ ] **Step 6: Commit**

```bash
git add services/sda-frontend/src/routes/
git commit -m "docs: add READMEs for routes (5.1)"
```

---

## Phase 5.2 — Tests

---

### Task 12: Setup — install deps + update vitest.config.ts

**Files:**
- Modify: `services/sda-frontend/package.json`
- Modify: `services/sda-frontend/vitest.config.ts`

- [ ] **Step 1: Install dependencies**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm install --save-dev @testing-library/svelte @testing-library/user-event @vitest/coverage-v8
```

Verify installation:
```bash
npm ls @testing-library/svelte @testing-library/user-event @vitest/coverage-v8
```

Expected: three packages listed without errors.

- [ ] **Step 2: Update `vitest.config.ts`**

Replace current content with:

```ts
import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
    plugins: [sveltekit()],
    test: {
        environment: 'node',
        environmentMatchGlobs: [
            ['src/**/*.svelte.test.ts', 'jsdom'],
            ['src/**/components/**/*.test.ts', 'jsdom'],
            ['src/**/components/**/*.component.test.ts', 'jsdom'],
            ['src/**/_components/**/*.test.ts', 'jsdom'],
            ['src/routes/**/*.component.test.ts', 'jsdom'],
        ],
        coverage: {
            provider: 'v8',
            include: ['src/lib/**/*.ts', 'src/lib/**/*.svelte.ts', 'src/routes/api/**/*.ts'],
            exclude: ['src/**/*.test.ts', 'src/**/*.spec.ts'],
            thresholds: {
                lines: 80,
                functions: 80,
                branches: 80,
                statements: 80,
            },
            reporter: ['text', 'html', 'lcov'],
        },
    },
});
```

- [ ] **Step 3: Run existing tests to verify nothing broke**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm run test
```

Expected: all existing tests pass. If any fail, investigate before continuing.

- [ ] **Step 4: Update package.json scripts**

Add to the `scripts` section:
```json
"test:coverage": "vitest run --coverage",
"test:watch": "vitest"
```

- [ ] **Step 5: Commit**

```bash
cd /Users/enzo/rag-saldivia
git add services/sda-frontend/package.json services/sda-frontend/package-lock.json services/sda-frontend/vitest.config.ts
git commit -m "test: install @testing-library/svelte + coverage-v8, configure vitest (5.2 setup)"
```

---

### Task 13: Expand unit tests — crossdoc pipeline + stores

**Files:**
- Modify: `src/lib/crossdoc/pipeline.test.ts`
- Modify: `src/lib/stores/chat.svelte.test.ts`
- Modify: `src/lib/stores/collections.svelte.test.ts`

Read the current test file and the source file before adding cases. All paths relative to `services/sda-frontend/`.

- [ ] **Step 1: Read current state of each file**

```bash
cat src/lib/crossdoc/pipeline.test.ts
cat src/lib/stores/chat.svelte.test.ts
cat src/lib/stores/collections.svelte.test.ts
```

- [ ] **Step 2: Expand `pipeline.test.ts`**

Add cases for (adapt based on what's already covered):
- `jaccard('')` with empty strings
- `dedup` with threshold at exactly 0.65
- `parseSubQueries` with malformed input (no numbered list, empty string, extra whitespace)
- `hasUsefulData` with all-empty chunks, with null/undefined fields

- [ ] **Step 3: Expand `chat.svelte.test.ts`**

Add cases for:
- Multi-turn: append multiple messages, verify order preserved
- `reset()`: verify all state returns to initial values
- `appendToken` with markdown characters (`**bold**`, `\n`, code blocks)

- [ ] **Step 4: Expand `collections.svelte.test.ts`**

Add cases for:
- `select(name)` → selected collection changes
- `clearSelected()` → selection removed
- Error state after failed fetch: store reflects error, not stale data

- [ ] **Step 5: Run tests**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm run test -- src/lib/crossdoc src/lib/stores
```

Expected: all pass.

---

### Task 14: Expand unit tests — server/auth + utils

**Files:**
- Modify: `src/lib/server/auth.test.ts`
- Modify: `src/lib/utils/markdown.test.ts`
- Modify: `src/lib/utils/scroll.test.ts`

All paths relative to `services/sda-frontend/`.

- [ ] **Step 1: Read current test files and sources**

```bash
cat src/lib/server/auth.test.ts
cat src/lib/utils/markdown.test.ts
cat src/lib/utils/scroll.test.ts
```

- [ ] **Step 2: Expand `auth.test.ts`**

Add cases for (adapt based on what exists):
- Expired JWT: token with `exp` in the past → function returns null or throws
- Malformed cookie: non-JWT string → handled gracefully, no crash
- Missing required claim (`name`, `sub`) → handled

- [ ] **Step 3: Expand `markdown.test.ts`**

Add cases for:
- XSS: `<script>alert('xss')</script>` in input → not present in output
- Table markdown → renders `<table>` with correct rows
- Fenced code block with language → `<code class="language-python">` present

- [ ] **Step 4: Expand `scroll.test.ts`**

Add cases for:
- Already at bottom → `scrollToBottom` is a no-op (no error)
- Element doesn't exist → handled gracefully

- [ ] **Step 5: Run tests**

```bash
npm run test -- src/lib/server src/lib/utils
```

Expected: all pass.

---

### Task 15: Expand unit tests — API routes

**Files:**
- Modify: `src/routes/api/upload/upload.test.ts`
- Modify: `src/routes/api/crossdoc/decompose/decompose.test.ts`

All paths relative to `services/sda-frontend/`.

- [ ] **Step 1: Read current files**

```bash
cat src/routes/api/upload/upload.test.ts
cat src/routes/api/crossdoc/decompose/decompose.test.ts
```

- [ ] **Step 2: Expand `upload.test.ts`**

Add cases for:
- File type not PDF → returns 400 with error message
- Gateway returns 503 → propagates error to client
- File size exceeds limit → returns 413 (if limit exists) or passes through

- [ ] **Step 3: Expand `decompose.test.ts`**

Add cases for:
- Gateway timeout → returns 504 or error response
- Malformed JSON from gateway → returns 500
- Empty query string → returns 400

- [ ] **Step 4: Run tests**

```bash
npm run test -- src/routes/api
```

Expected: all pass.

---

### Task 16: New unit test — `actions/clickOutside.ts` (TDD)

**Files:**
- Create: `src/lib/actions/clickOutside.test.ts`

Read source first: `cat src/lib/actions/clickOutside.ts`

- [ ] **Step 1: Read the source BEFORE writing the test**

```bash
cat src/lib/actions/clickOutside.ts
```

Understand the actual API: does it receive a callback, dispatch a custom event, or both? The tests below are **illustrative** — DO NOT copy them without adapting to the actual implementation.

⚠️ **Known gotcha:** The current `clickOutside.ts` uses a callback parameter `(node, callback) => void`, NOT a custom event. The test must use the callback API. If the implementation uses `callback()`, the test looks like:

```ts
// src/lib/actions/clickOutside.test.ts
import { describe, it, expect, vi } from 'vitest';
import { clickOutside } from './clickOutside.js';

describe('clickOutside action', () => {
    it('calls callback when clicking outside the element', () => {
        const node = document.createElement('div');
        document.body.appendChild(node);
        const handler = vi.fn();

        clickOutside(node, handler);
        document.dispatchEvent(new MouseEvent('click', { bubbles: true }));

        expect(handler).toHaveBeenCalledOnce();
        document.body.removeChild(node);
    });

    it('does not call callback when clicking inside the element', () => {
        const node = document.createElement('div');
        document.body.appendChild(node);
        const handler = vi.fn();

        clickOutside(node, handler);
        node.dispatchEvent(new MouseEvent('click', { bubbles: true }));

        expect(handler).not.toHaveBeenCalled();
        document.body.removeChild(node);
    });

    it('cleans up listener on destroy', () => {
        const node = document.createElement('div');
        document.body.appendChild(node);
        const handler = vi.fn();

        const action = clickOutside(node, handler);
        action?.destroy?.();
        document.dispatchEvent(new MouseEvent('click', { bubbles: true }));

        expect(handler).not.toHaveBeenCalled();
        document.body.removeChild(node);
    });
});
```

Adapt the test if the actual implementation differs from the above.

- [ ] **Step 2: Run and verify it fails** *(TDD: red)*

```bash
npm run test -- src/lib/actions/clickOutside.test.ts
```

Expected: FAIL — the test file doesn't exist yet, so Vitest reports no tests or import error.

- [ ] **Step 3: Verify/fix the implementation**

If the source already has all three behaviors (call on outside click, no-op on inside click, cleanup on destroy), the tests should pass as-is. Implement only what's missing.

- [ ] **Step 4: Run and verify it passes**

```bash
npm run test -- src/lib/actions/clickOutside.test.ts
```

Expected: 3 tests pass.

---

### Task 17: New unit test — chat stream API (TDD)

**Files:**
- Create: `src/routes/api/chat/stream/stream.test.ts`

Read source first: `cat src/routes/api/chat/stream/[id]/+server.ts`

- [ ] **Step 1: Read the source BEFORE writing the test**

```bash
cat src/routes/api/chat/stream/[id]/+server.ts
```

Understand: what HTTP method does it export (`GET` or `POST`)? How does it authenticate (via `locals.user`, `cookies`, or `request.headers`)? What does it return?

⚠️ **Known gotcha:** The handler exports `POST`, not `GET`. Authentication uses `locals.user`, not `cookies.get()`. The test below matches this pattern — adapt if the actual source differs:

```ts
// src/routes/api/chat/stream/stream.test.ts
import { describe, it, expect, vi } from 'vitest';

vi.mock('$lib/server/gateway.js', () => ({
    streamChat: vi.fn(),
}));

import { POST } from '../[id]/+server.js';
import { streamChat } from '$lib/server/gateway.js';

describe('POST /api/chat/stream/[id]', () => {
    it('returns 401 when user not in locals', async () => {
        const response = await POST({
            request: new Request('http://localhost/api/chat/stream/test-id', {
                method: 'POST',
                body: JSON.stringify({ query: 'example question', collection_names: ['example'] }),
                headers: { 'Content-Type': 'application/json' },
            }),
            params: { id: 'test-id' },
            locals: { user: null },
        } as any);
        expect(response.status).toBe(401);
    });

    it('proxies SSE stream from gateway when authenticated', async () => {
        const mockStream = new ReadableStream({
            start(controller) {
                controller.enqueue(new TextEncoder().encode('data: {"type":"token","content":"example"}\n\n'));
                controller.close();
            },
        });
        vi.mocked(streamChat).mockResolvedValue(new Response(mockStream, {
            headers: { 'Content-Type': 'text/event-stream' },
        }));

        const response = await POST({
            request: new Request('http://localhost/api/chat/stream/test-id', {
                method: 'POST',
                body: JSON.stringify({ query: 'example question', collection_names: ['example'] }),
                headers: { 'Content-Type': 'application/json' },
            }),
            params: { id: 'test-id' },
            locals: { user: { name: 'Test User', email: 'test@example.com', role: 'user' } },
        } as any);

        expect(response.status).toBe(200);
        expect(response.headers.get('content-type')).toContain('text/event-stream');
    });
});
```

- [ ] **Step 2: Run and verify it fails** *(TDD: red)*

```bash
npm run test -- src/routes/api/chat/stream/stream.test.ts
```

Expected: FAIL — file doesn't exist yet.

- [ ] **Step 3: Adjust test to match actual implementation**

Read `+server.ts` again carefully if Step 2 fails for reasons other than "file not found". Adapt the method name (`GET`/`POST`), request body shape, and auth mechanism to match reality. Never modify the handler itself.

- [ ] **Step 4: Run and verify it passes**

```bash
npm run test -- src/routes/api/chat/stream/stream.test.ts
```

Expected: pass.

- [ ] **Step 5: Run coverage to check threshold**

```bash
npm run test:coverage
```

Expected: ≥80% on all metrics, or close to it. Note any files below threshold.

- [ ] **Step 6: Commit unit tests block**

```bash
cd /Users/enzo/rag-saldivia
git add services/sda-frontend/src/
git commit -m "test: expand unit tests + add clickOutside and stream tests (5.2 capa 1)"
```

---

### Task 18: Component tests — ChatInput + MessageList

**Files:**
- Create: `src/lib/components/chat/ChatInput.component.test.ts`
- Create: `src/lib/components/chat/MessageList.component.test.ts`

All paths relative to `services/sda-frontend/`. Read each `.svelte` file before writing its test.

- [ ] **Step 1: Read the components**

```bash
cat src/lib/components/chat/ChatInput.svelte
cat src/lib/components/chat/MessageList.svelte
```

- [ ] **Step 2: Write `ChatInput.component.test.ts`**

⚠️ **Critical:** Read `ChatInput.svelte` first and identify the exact props interface. This is Svelte 5 — `component.$on()` does NOT exist. Events are passed as props (e.g., `onsubmit`, `onstop`). The template below matches the known API (`onsubmit: (query: string) => void`, `onstop: () => void`, `oncrossdoctoggle: () => void`, `streaming: boolean`). Adapt if it has changed.

```ts
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import ChatInput from './ChatInput.svelte';

const defaultProps = {
    streaming: false,
    crossdoc: false,
    onsubmit: vi.fn(),
    onstop: vi.fn(),
    oncrossdoctoggle: vi.fn(),
};

describe('ChatInput', () => {
    it('renders a text input', () => {
        render(ChatInput, { props: defaultProps });
        expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    it('shows stop button when streaming=true', () => {
        render(ChatInput, { props: { ...defaultProps, streaming: true } });
        expect(screen.getByRole('button', { name: /stop/i })).toBeInTheDocument();
    });

    it('calls onsubmit with message text when send is clicked', async () => {
        const user = userEvent.setup();
        const onsubmit = vi.fn();
        render(ChatInput, { props: { ...defaultProps, onsubmit } });

        await user.type(screen.getByRole('textbox'), 'example question');
        await user.click(screen.getByRole('button', { name: /send/i }));

        expect(onsubmit).toHaveBeenCalledWith('example question');
    });

    it('does not call onsubmit when input is empty', async () => {
        const user = userEvent.setup();
        const onsubmit = vi.fn();
        render(ChatInput, { props: { ...defaultProps, onsubmit } });

        await user.click(screen.getByRole('button', { name: /send/i }));

        expect(onsubmit).not.toHaveBeenCalled();
    });

    it('calls onstop when stop button clicked during streaming', async () => {
        const user = userEvent.setup();
        const onstop = vi.fn();
        render(ChatInput, { props: { ...defaultProps, streaming: true, onstop } });

        await user.click(screen.getByRole('button', { name: /stop/i }));

        expect(onstop).toHaveBeenCalledOnce();
    });
});
```

- [ ] **Step 3: Write `MessageList.component.test.ts`**

Cover states:
- Empty list → renders empty state or nothing (no errors)
- Single user message → text visible in DOM
- Single assistant message → text visible in DOM
- Streaming message (if component accepts a `streaming` prop) → loading indicator visible

- [ ] **Step 4: Run tests**

```bash
npm run test -- src/lib/components/chat/
```

Expected: all pass. If jsdom errors appear, verify the `environmentMatchGlobs` config from Task 12 is correct.

---

### Task 19: Component tests — CrossdocProgress + Toast + Modal + CollectionCard

**Files:**
- Create: `src/lib/components/chat/CrossdocProgress.component.test.ts`
- Create: `src/lib/components/ui/Toast.component.test.ts`
- Create: `src/lib/components/ui/Modal.component.test.ts`
- Create: `src/routes/(app)/collections/_components/CollectionCard.component.test.ts`

Read each `.svelte` file before writing its test.

- [ ] **Step 1: Read the components**

```bash
cat src/lib/components/chat/CrossdocProgress.svelte
cat src/lib/components/ui/Toast.svelte
cat src/lib/components/ui/Modal.svelte
cat src/routes/\(app\)/collections/_components/CollectionCard.svelte
```

- [ ] **Step 2: Write `CrossdocProgress.component.test.ts`**

Cover each phase of the crossdoc pipeline. The component likely receives a `phase` or `status` prop. Test that the correct label or indicator appears for each phase value.

- [ ] **Step 3: Write `Toast.component.test.ts`**

Cover: success variant shows correct icon/color class, error variant, warning variant. If auto-dismiss is implemented via timeout, test with `vi.useFakeTimers()`.

- [ ] **Step 4: Write `Modal.component.test.ts`**

Cover: `open=true` → content visible; `open=false` → content hidden or not rendered; close button click triggers close event.

- [ ] **Step 5: Write `CollectionCard.component.test.ts`**

Cover: renders collection name, renders document count, selected state has distinct visual indicator, loading skeleton when no data.

- [ ] **Step 6: Run all component tests**

```bash
npm run test -- src/lib/components src/routes/\(app\)/collections/_components
```

Expected: all pass.

- [ ] **Step 7: Commit component tests**

```bash
cd /Users/enzo/rag-saldivia
git add services/sda-frontend/src/
git commit -m "test: add component tests with @testing-library/svelte (5.2 capa 2)"
```

---

### Task 20: Install Playwright + browsers

**Files:**
- No code changes — installation only

- [ ] **Step 1: Install Playwright**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm install --save-dev @playwright/test
```

- [ ] **Step 2: Install browser binaries + OS dependencies**

```bash
npx playwright install chromium --with-deps
```

On Debian/Ubuntu (Brev instance), if `--with-deps` fails or is insufficient:
```bash
npx playwright install-deps chromium
npx playwright install chromium
```

- [ ] **Step 3: Verify installation**

```bash
npx playwright --version
```

Expected: prints version (e.g., `Version 1.x.x`).

- [ ] **Step 4: Verify browser launches**

```bash
npx playwright test --list
```

Expected: no crashes (even if no tests exist yet, the command should not error on browser init).

---

### Task 21: Configure Playwright + setup preview server

**Files:**
- Create: `services/sda-frontend/playwright.config.ts`

- [ ] **Step 1: Create `playwright.config.ts`**

```ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
    testDir: './tests/e2e/flows',
    fullyParallel: false,
    forbidOnly: !!process.env.CI,
    retries: process.env.CI ? 1 : 0,
    workers: 1,
    reporter: 'html',
    use: {
        baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:4173',
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
    },
    projects: [
        {
            name: 'chromium',
            use: { ...devices['Desktop Chrome'] },
        },
    ],
});
```

- [ ] **Step 2: Add E2E script to `package.json`**

Add to `scripts`:
```json
"test:e2e": "playwright test",
"test:e2e:ui": "playwright test --ui",
"test:e2e:headed": "playwright test --headed"
```

- [ ] **Step 3: Document preview server startup in a comment at top of `playwright.config.ts`**

Add as a comment block before the import:
```ts
/**
 * E2E tests run against the built app in preview mode.
 *
 * Before running tests:
 *   npm run build && npm run preview
 *   # App starts on http://localhost:4173
 *
 * Or set PLAYWRIGHT_BASE_URL to point to Brev:
 *   PLAYWRIGHT_BASE_URL=https://brev-instance-url npx playwright test
 *
 * Flows tagged @slow only run when PLAYWRIGHT_BASE_URL is set (Brev only).
 */
```

- [ ] **Step 4: Build and start preview server to verify it works**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm run build
npm run preview &
PREVIEW_PID=$!
sleep 3
curl -s http://localhost:4173 | grep -q "html" && echo "Preview server OK" || echo "Preview server FAILED"
kill $PREVIEW_PID
```

Expected: `Preview server OK`.

---

### Task 22: Write Page Objects

**Files:**
- Create: `tests/e2e/pages/LoginPage.ts`
- Create: `tests/e2e/pages/ChatPage.ts`
- Create: `tests/e2e/pages/CollectionsPage.ts`
- Create: `tests/e2e/pages/UploadPage.ts`
- Create: `tests/e2e/pages/AdminPage.ts`

All paths relative to `services/sda-frontend/`. Page Objects encapsulate selectors — flows import Page Objects and call methods, never use raw selectors in spec files.

- [ ] **Step 1: Create `tests/e2e/pages/LoginPage.ts`**

```ts
import type { Page } from '@playwright/test';

export class LoginPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/login');
    }

    async login(email: string, password: string) {
        await this.page.getByLabel(/email/i).fill(email);
        await this.page.getByLabel(/password/i).fill(password);
        await this.page.getByRole('button', { name: /login/i }).click();
    }

    async getErrorMessage() {
        return this.page.getByRole('alert').textContent();
    }
}
```

- [ ] **Step 2: Create `tests/e2e/pages/ChatPage.ts`**

```ts
import type { Page } from '@playwright/test';

export class ChatPage {
    constructor(private page: Page) {}

    async goto() {
        await this.page.goto('/chat');
    }

    async sendMessage(text: string) {
        await this.page.getByRole('textbox').fill(text);
        await this.page.getByRole('button', { name: /send/i }).click();
    }

    async waitForResponse() {
        await this.page.waitForSelector('[data-testid="assistant-message"]');
    }

    async enableCrossdoc() {
        await this.page.getByRole('button', { name: /crossdoc/i }).click();
    }

    async getLastMessage() {
        const messages = this.page.locator('[data-testid="message"]');
        return messages.last().textContent();
    }
}
```

- [ ] **Step 3: Create `tests/e2e/pages/CollectionsPage.ts`**

Methods: `goto()`, `createCollection(name)`, `deleteCollection(name)`, `getCollectionNames(): Promise<string[]>`, `openCollection(name)`.

- [ ] **Step 4: Create `tests/e2e/pages/UploadPage.ts`**

Methods: `goto()`, `uploadFile(filePath: string)`, `waitForUploadComplete()`, `getUploadStatus(): Promise<string>`.

- [ ] **Step 5: Create `tests/e2e/pages/AdminPage.ts`**

Methods: `goto()`, `getUsers(): Promise<string[]>`, `createUser(email, role)`, `deleteUser(email)`.

**Note:** Selectors in Page Objects (`getByLabel`, `getByRole`, `[data-testid]`) must match actual HTML in the components. Read the Svelte component source to find the right selectors. Add `data-testid` attributes to components if needed — that's a valid part of this task.

---

### Task 23: Write auth fixture

**Files:**
- Create: `tests/e2e/fixtures/auth.ts`

All paths relative to `services/sda-frontend/`.

- [ ] **Step 1: Create `tests/e2e/fixtures/auth.ts`** *(TDD: fixture first)*

```ts
import { test as base } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage.js';

export const test = base.extend<{ loginPage: LoginPage }>({
    loginPage: async ({ page }, use) => {
        const loginPage = new LoginPage(page);
        await loginPage.goto();
        // Use example credentials — must match a test user in the system
        // or be intercepted via page.route() before being called
        await loginPage.login(
            process.env.TEST_USER_EMAIL ?? 'test@example.com',
            process.env.TEST_USER_PASSWORD ?? 'test-password-example'
        );
        // Wait for redirect to /chat
        await page.waitForURL('**/chat');
        await use(loginPage);
    },
});

export { expect } from '@playwright/test';
```

- [ ] **Step 2: Verify the fixture file compiles**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npx tsc --noEmit tests/e2e/fixtures/auth.ts 2>&1 || true
```

Expected: no errors (or only path resolution warnings that don't affect runtime).

---

### Task 24: E2E flow — auth.spec.ts

**Files:**
- Create: `tests/e2e/flows/auth.spec.ts`

All paths relative to `services/sda-frontend/`. This flow runs against the real gateway (via preview server or Brev). Uses `page.route()` to mock the gateway auth endpoint for CI safety.

- [ ] **Step 1: Start preview server**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm run build && npm run preview &
PREVIEW_PID=$!
sleep 3
```

- [ ] **Step 2: Create `tests/e2e/flows/auth.spec.ts`**

```ts
import { test, expect } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage.js';

test.describe('Auth flow', () => {
    test('valid login redirects to /chat', async ({ page }) => {
        // Mock the BFF auth endpoint
        await page.route('**/api/auth/session', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    token: 'example-jwt-token',
                    user: { name: 'Test User', email: 'test@example.com', role: 'user' },
                }),
            });
        });

        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login('test@example.com', 'password123');
        await expect(page).toHaveURL(/\/chat/);
    });

    test('invalid credentials shows error message', async ({ page }) => {
        await page.route('**/api/auth/session', async (route) => {
            await route.fulfill({
                status: 401,
                contentType: 'application/json',
                body: JSON.stringify({ error: 'Invalid credentials' }),
            });
        });

        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login('wrong@example.com', 'wrongpassword');
        const error = await loginPage.getErrorMessage();
        expect(error).toBeTruthy();
    });

    test('unauthenticated access to /chat redirects to /login', async ({ page }) => {
        await page.goto('/chat');
        await expect(page).toHaveURL(/\/login/);
    });

    test('logout clears session and redirects to /login', async ({ page }) => {
        // Setup authenticated state via storage (or via login flow with mock)
        await page.route('**/api/auth/session', async (route) => {
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    token: 'example-jwt-token',
                    user: { name: 'Test User', email: 'test@example.com', role: 'user' },
                }),
            });
        });

        const loginPage = new LoginPage(page);
        await loginPage.goto();
        await loginPage.login('test@example.com', 'password123');
        await page.waitForURL(/\/chat/);

        // Click logout (find the actual logout button selector in the layout)
        await page.getByRole('button', { name: /logout/i }).click();
        await expect(page).toHaveURL(/\/login/);
    });
});
```

- [ ] **Step 3: Run the flow**

```bash
npx playwright test tests/e2e/flows/auth.spec.ts --headed
```

Expected: all tests pass. If selectors don't match, read the actual Svelte components and update selectors in LoginPage.ts and the spec.

---

### Task 25: E2E flow — chat.spec.ts

**Files:**
- Create: `tests/e2e/flows/chat.spec.ts`

Uses `page.route()` to mock SSE streaming — CI safe.

- [ ] **Step 1: Create `tests/e2e/flows/chat.spec.ts`**

```ts
import { test, expect } from '@playwright/test';
import { ChatPage } from '../pages/ChatPage.js';

// Mock auth for all tests in this file
test.beforeEach(async ({ page }) => {
    await page.route('**/api/auth/session', async (route) => {
        await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
                token: 'example-jwt-token',
                user: { name: 'Test User', email: 'test@example.com', role: 'user' },
            }),
        });
    });

    // Mock SSE stream endpoint
    await page.route('**/api/chat/stream/**', async (route) => {
        const stream = new ReadableStream({
            start(controller) {
                const encoder = new TextEncoder();
                controller.enqueue(encoder.encode('data: {"type":"token","content":"This"}\n\n'));
                controller.enqueue(encoder.encode('data: {"type":"token","content":" is"}\n\n'));
                controller.enqueue(encoder.encode('data: {"type":"token","content":" an"}\n\n'));
                controller.enqueue(encoder.encode('data: {"type":"token","content":" example"}\n\n'));
                controller.enqueue(encoder.encode('data: {"type":"done"}\n\n'));
                controller.close();
            },
        });

        await route.fulfill({
            status: 200,
            headers: { 'Content-Type': 'text/event-stream' },
            body: stream,
        });
    });
});

test.describe('Chat flow', () => {
    test('sends message and receives streamed response', async ({ page }) => {
        const chatPage = new ChatPage(page);
        await chatPage.goto();
        await chatPage.sendMessage('What is an example question?');
        await chatPage.waitForResponse();
        const lastMessage = await chatPage.getLastMessage();
        expect(lastMessage).toContain('example');
    });

    test('input clears after sending', async ({ page }) => {
        const chatPage = new ChatPage(page);
        await chatPage.goto();
        await chatPage.sendMessage('Test message');
        const input = page.getByRole('textbox');
        await expect(input).toHaveValue('');
    });
});
```

- [ ] **Step 2: Run the flow**

```bash
npx playwright test tests/e2e/flows/chat.spec.ts
```

Expected: pass. Adapt SSE format to match the actual format from `src/routes/api/chat/stream/[id]/+server.ts`.

---

### Task 26: E2E flows — collections + upload + crossdoc

**Files:**
- Create: `tests/e2e/flows/collections.spec.ts`
- Create: `tests/e2e/flows/upload.spec.ts`
- Create: `tests/e2e/flows/crossdoc.spec.ts`
- Create: `tests/e2e/fixtures/sample.pdf` (tiny valid PDF)

- [ ] **Step 1: Create a minimal valid PDF fixture**

```bash
cd /Users/enzo/rag-saldivia/services/sda-frontend/tests/e2e/fixtures
# Create a minimal valid PDF (1 page, plain text)
printf '%%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/MediaBox[0 0 612 792]/Parent 2 0 R>>endobj\nxref\n0 4\n0000000000 65535 f\n0000000009 00000 n\n0000000058 00000 n\n0000000115 00000 n\ntrailer<</Size 4/Root 1 0 R>>\nstartxref\n190\n%%%%EOF' > sample.pdf
```

- [ ] **Step 2: Create `collections.spec.ts`**

Mock `**/api/collections*` with `page.route()`. Cover:
- Listing collections (mock returns array of example collections)
- Creating a collection (mock POST returns 201)
- Deleting a collection (mock DELETE returns 204)

- [ ] **Step 3: Create `upload.spec.ts`**

Tag with `@slow` annotation (Playwright `test.slow()` or custom tag). This test only makes sense against a real backend. Mock the upload endpoint minimally for CI:

```ts
import { fileURLToPath } from 'url';
import path from 'path';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SAMPLE_PDF = path.join(__dirname, '../fixtures/sample.pdf');

test('upload PDF shows success state', async ({ page }) => {
    await page.route('**/api/upload', async (route) => {
        await route.fulfill({ status: 200, body: JSON.stringify({ status: 'queued' }) });
    });
    const uploadPage = new UploadPage(page);
    await uploadPage.goto();
    await uploadPage.uploadFile(SAMPLE_PDF); // absolute path — safe regardless of CWD
    await uploadPage.waitForUploadComplete();
    const status = await uploadPage.getUploadStatus();
    expect(status).toMatch(/queued|success/i);
});
```

- [ ] **Step 4: Create `crossdoc.spec.ts`**

Mock all three crossdoc endpoints with `page.route()`:
- `/api/crossdoc/decompose` → returns example sub-queries array
- `/api/crossdoc/subquery` → returns example chunks
- `/api/crossdoc/synthesize` → returns example SSE stream

Cover: activating crossdoc mode, sending a query, seeing the decomposition panel appear, seeing the final response.

- [ ] **Step 5: Run all E2E flows**

```bash
npx playwright test
```

Expected: all non-`@slow` tests pass.

- [ ] **Step 6: Kill preview server**

```bash
kill $PREVIEW_PID 2>/dev/null || true
```

- [ ] **Step 7: Commit Playwright block**

```bash
cd /Users/enzo/rag-saldivia
git add services/sda-frontend/playwright.config.ts \
        services/sda-frontend/tests/ \
        services/sda-frontend/package.json \
        services/sda-frontend/package-lock.json
git commit -m "test: add Playwright E2E tests with POM (5.2 capa 3)"
```

---

### Task 27: Integrate test targets in Makefile

**Files:**
- Modify: `Makefile`

Read the current Makefile first.

- [ ] **Step 1: Read current Makefile**

```bash
cat /Users/enzo/rag-saldivia/Makefile | grep -A2 "^test\|^\.PHONY"
```

- [ ] **Step 2: Add test targets**

Add the following targets to the Makefile (adapt to match existing style):

```makefile
## Testing
test: test-unit test-e2e  ## Run full test pyramid (unit + E2E)

test-unit:  ## Run Vitest unit + component tests
	cd services/sda-frontend && npm run test

test-coverage:  ## Run Vitest with coverage report (fails if <80%)
	cd services/sda-frontend && npm run test:coverage
	@echo "Coverage report: services/sda-frontend/coverage/index.html"

test-e2e:  ## Run Playwright E2E tests (requires app running on port 4173)
	cd services/sda-frontend && npm run build && npm run preview & \
	sleep 5 && npx playwright test; \
	kill %1

test-e2e-brev:  ## Run E2E tests against Brev instance
	cd services/sda-frontend && PLAYWRIGHT_BASE_URL=$(BREV_URL) npx playwright test

test-backend:  ## Run Python pytest tests
	uv run pytest saldivia/tests/ -v
```

- [ ] **Step 3: Verify `make test-unit` works**

```bash
cd /Users/enzo/rag-saldivia
make test-unit
```

Expected: all Vitest tests pass.

- [ ] **Step 4: Verify `make test-coverage` works**

```bash
make test-coverage
```

Expected: coverage report generated, threshold met (≥80%). If threshold not met, go back to Tasks 13-17 and add more test cases.

- [ ] **Step 5: Final commit**

```bash
git add Makefile
git commit -m "chore: add test pyramid make targets (5.2 complete)"
```

---

## Done

After Task 27:

1. Run `git log --oneline` — verify commit history is clean and semantic
2. Run `make test-unit` — all unit + component tests pass
3. Run `make test-coverage` — coverage ≥80%
4. Run `make test-e2e` — E2E tests pass (non-@slow)
5. Open `services/sda-frontend/coverage/index.html` in browser — review coverage gaps

Push branch and open PR against `main` when Enzo asks.
