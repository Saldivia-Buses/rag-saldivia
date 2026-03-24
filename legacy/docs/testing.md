# Testing

## Test Pyramid

RAG Saldivia uses a three-layer test pyramid for the SDA Frontend (SvelteKit 5):

```
           /\
          /  \  E2E (Playwright)        — Slow, high confidence, full stack
         /    \                         — ~10-20 tests, run in CI + manually
        /------\
       / Comp.  \  Component Tests      — Medium speed, UI behavior
      /  Tests   \  (@testing-library)  — ~30-50 tests, render + query
     /------------\
    /    Unit      \  Unit Tests        — Fast, pure logic
   /     Tests      \  (Vitest)         — ~100+ tests, <5s total
  /------------------\
```

**Rationale:**
- Unit tests are fast and catch logic bugs early
- Component tests validate UI behavior without browser overhead
- E2E tests verify the full stack works (auth, SSE streaming, file uploads)

## Running Tests

### Backend Tests (Python)

The backend test suite is already installed and working. From the repo root:

```bash
# All backend tests
uv run pytest saldivia/tests/ -v

# Specific test file
uv run pytest saldivia/tests/test_gateway.py -v

# With coverage
uv run pytest saldivia/tests/ --cov=saldivia -v
```

**22 tests** covering: gateway, auth, config, mode manager, providers, collections.

### Frontend Tests (Vitest + Playwright)

> **Status:** Planned — not yet installed. See Phase 5.2 implementation.

Once implemented, the commands will be:

```bash
# From repo root
make test           # Full pyramid (unit + component + E2E)
make test-unit      # Vitest only (<5s)
make test-e2e       # Playwright only (requires app running)
make test-coverage  # Vitest with coverage

# From services/sda-frontend/
npm run test:unit      # Vitest unit + component tests
npm run test:e2e       # Playwright E2E tests
npm run test:coverage  # Coverage report
npm run test:watch     # Watch mode
```

**Note:** E2E tests require a running preview server:

```bash
cd services/sda-frontend
npm run build && npm run preview
# App starts on http://localhost:4173
```

## Vitest Unit Tests

**Location:** `services/sda-frontend/src/**/*.test.ts`

> **Status:** Vitest is already installed. `environmentMatchGlobs` and coverage thresholds are planned (Phase 5.2, not yet configured).

**Environment:**
- Default: `node` (for pure logic, stores, utilities)
- Component tests: `jsdom` (for DOM-dependent code)
- Configured via `environmentMatchGlobs` in `vitest.config.ts`

**Examples:**

```typescript
// src/lib/utils/formatting.test.ts
import { describe, it, expect } from 'vitest';
import { formatBytes, formatDate } from './formatting';

describe('formatBytes', () => {
  it('formats bytes to human-readable string', () => {
    expect(formatBytes(0)).toBe('0 B');
    expect(formatBytes(1024)).toBe('1.0 KB');
    expect(formatBytes(1048576)).toBe('1.0 MB');
  });
});
```

**What to test:**
- Pure functions (utilities, helpers, validators)
- Store logic (state transitions, derived values)
- API clients (mock fetch with vi.fn())
- Edge cases (null, undefined, empty arrays, large numbers)

**What NOT to test:**
- Svelte internals ($state, $derived, $effect) — trust the framework
- Browser APIs directly — use component tests with jsdom
- External libraries — trust their tests

## Component Tests

**Location:** `services/sda-frontend/src/**/*.component.test.ts`

> **Status:** Planned — `@testing-library/svelte` is not yet installed (Phase 5.2).

**Tools:**
- `@testing-library/svelte` — render components, query by role, fire events
- `jsdom` — simulated DOM environment (faster than real browser)

**Pattern:**

```typescript
import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import MyComponent from './MyComponent.svelte';

describe('MyComponent', () => {
  it('renders button and handles click', async () => {
    const { component } = render(MyComponent, { props: { label: 'Click me' } });

    const button = screen.getByRole('button', { name: 'Click me' });
    expect(button).toBeTruthy();

    await fireEvent.click(button);
    // Assert state change or callback invocation
  });
});
```

**What to test:**
- Component renders with correct props
- User interactions (click, input, submit)
- Conditional rendering ({#if}, {:else})
- Accessibility (role, aria-label, keyboard navigation)

**What NOT to test:**
- CSS styles (use visual regression testing tools instead)
- SSE streaming (use E2E tests with real server)
- Authentication flows (use E2E tests with mocked gateway)

## Playwright E2E Tests

**Location:** `services/sda-frontend/tests/e2e/`

> **Status:** Planned — Playwright is not yet installed (Phase 5.2).

**Tools:**
- Playwright (Chromium, Firefox, WebKit)
- Page Object Model pattern (tests/e2e/pages/)
- `page.route()` for mocking API responses (CI-safe)

**Pattern:**

```typescript
import { test, expect } from '@playwright/test';
import { LoginPage } from './pages/LoginPage';

test.describe('Authentication', () => {
  test('login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page);
    
    await loginPage.goto();
    await loginPage.login('test@example.com', 'password123');
    
    await expect(page).toHaveURL(/\/chat/);
  });
});
```

**Mocking for CI:**

E2E tests run in CI where the gateway is not available. Use `page.route()` to mock API responses:

```typescript
test('chat with mocked gateway', async ({ page }) => {
  await page.route('**/api/chat/generate', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'text/event-stream',
      body: 'data: {"type":"chunk","content":"Hello"}\n\ndata: {"type":"done"}\n\n'
    });
  });

  // Test chat interaction
});
```

**Tagging:**

Tests that require Brev infrastructure (Milvus, NIMs, LLM) are tagged `@slow`:

```typescript
test('@slow ingest document and verify in chat', async ({ page }) => {
  // Requires full stack running on Brev
});
```

Run slow tests only on Brev:

```bash
npm run test:e2e -- --grep @slow
```

## Preview Server Setup for E2E

E2E tests run against the production build (not dev server) to match deployment:

```bash
cd services/sda-frontend

# 1. Build the app
npm run build

# 2. Start preview server (port 4173)
npm run preview

# 3. In another terminal, run E2E tests
npm run test:e2e
```

**Why preview instead of dev?**
- Matches production environment (SSR + prerendering)
- Faster startup than dev server (no Vite HMR overhead)
- Catches build-time errors (missing imports, type errors)

## Coverage

**Target:** 80% coverage on .ts and .svelte.ts files in `src/lib/` and `src/routes/api/`

> **Status:** Planned — `@vitest/coverage-v8` is not yet installed (Phase 5.2).

**Exclusions:**
- `src/routes/` (Svelte files) — covered by E2E tests
- `src/lib/components/` (Svelte files) — covered by component + E2E tests
- `src/app.html` — static HTML, no logic
- `src/hooks.server.ts` — integration point, covered by E2E tests

**Generate report:**

```bash
# Frontend coverage (Phase 5.2, not yet available)
# cd services/sda-frontend && npm run test:coverage

# Backend coverage (available now)
uv run pytest saldivia/tests/ --cov=saldivia -v
# HTML report at coverage/index.html
```

**View in browser:**

```bash
open coverage/index.html
```

**CI enforcement:**

The CI pipeline fails if coverage drops below 80%. Check coverage locally before pushing:

```bash
npm run test:coverage
```

## Test Data Convention

**Rule:** All test data is illustrative. No production credentials, no real vault data.

**Examples:**

```typescript
// Good: fake data
const testUser = { email: 'test@example.com', name: 'Test User' };
const testCollection = { name: 'test-collection', description: 'Test data' };

// Bad: real data
const testUser = { email: 'enzosaldivia@gmail.com', name: 'Enzo Saldivia' };
const testCollection = { name: 'tecpia', description: 'Tecpia knowledge base' };
```

**Rationale:**
- Test data appears in CI logs, coverage reports, and screenshots
- Real data leaks are a security risk
- Illustrative data makes tests easier to understand

**Mocked API responses:**

Use realistic but fake data in mocked responses:

```typescript
await page.route('**/api/collections', (route) => {
  route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify([
      { id: '1', name: 'test-collection-1', description: 'Test data 1' },
      { id: '2', name: 'test-collection-2', description: 'Test data 2' }
    ])
  });
});
```

