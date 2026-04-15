---
title: Convention: Testing Strategy
audience: ai
last_reviewed: 2026-04-15
related:
  - ./go.md
  - ./frontend.md
  - ./sqlc.md
  - ../architecture/overview.md
---

If the code has no test, it isn't done. CI enforces this — the Test gate runs `make test` and refuses merges on failure.

## Layers

| Layer | Tooling | Where | When required |
|---|---|---|---|
| Unit (Go) | `go test` + testify | `*_test.go` next to source | Pure logic, helpers, encoders, validators |
| Integration (Go) | testify + testcontainers (Postgres, NATS, Redis) | `*_integration_test.go` next to source | Repositories, NATS publishers, anything touching infra |
| Contract (Go) | testify HTTP roundtrips | `services/{name}/internal/handler/*_test.go` | Every HTTP handler |
| Unit (frontend) | bun:test + happy-dom | `apps/web/src/**/__tests__/` | Pure logic, hooks (after extraction to `lib/`), components |
| E2E (frontend) | Playwright | `apps/web/tests/e2e/` | Critical user flows: login, chat, admin |
| Visual regression | Playwright screenshots vs Storybook | `apps/web/tests/visual/` | Every primitive in the design system |
| A11y | axe-playwright | `apps/web/tests/a11y/` | Top-level pages |

## Go: table-driven tests

DO structure tests as a slice of cases iterated with `t.Run(case.name, ...)`:

```
cases := []struct {
    name    string
    in      string
    want    string
    wantErr error
}{...}
for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

DO use testify (`github.com/stretchr/testify/assert` and `require`) — `require` for fatal assertions that prevent the rest of the test from making sense, `assert` for non-fatal checks.

DO name test functions `TestXxx` mirroring the function under test (`TestLogin`, `TestVerifyRefreshToken`).

DON'T share state across `t.Run` cases. Each case constructs its own fixtures.

## Go: integration tests with testcontainers

DO spin up real Postgres / NATS / Redis containers in integration tests. Use `testcontainers-go` to start them per test package (not per test) — `TestMain` or `t.Helper` setup.

DO point at the testcontainer URL in test config; the production resolver is bypassed.

DO clean up via `t.Cleanup(func() { container.Terminate(ctx) })`. Tests must leave no containers behind.

DON'T mock infrastructure that has cheap real implementations. Integration confidence beats mock fidelity for DB and NATS.

## Go: handler contract tests

DO build a `httptest.NewRecorder` and `httptest.NewRequest`, invoke the handler, and assert on status code, body JSON shape, and the events published to a NATS test bus.

DO use a fake or interface-mocked service layer for handler tests; the integration tests cover the service layer separately.

## Frontend: unit and component tests

DO put `afterEach(cleanup)` at the top of every component test file — without it, renders accumulate and queries return cross-test elements.

DO use scoped queries from the render result, never the global `screen`:
- `const { getByRole } = render(<Button />)` — correct
- `screen.getByRole("button")` — incorrect (multi-test contamination)

DO use `fireEvent` from `@testing-library/react`. `userEvent` has incompatibilities with happy-dom in multi-file runs.

DO mock server actions with `mock.module("@/app/actions/...", () => ({ actionXxx: mock(...) }))`.

DON'T test through the real backend in unit tests — mock at the network boundary.

## Frontend: E2E

DO write Playwright specs for the critical flows: login, send a chat message, navigate admin, list collections.

DO run E2E with `MOCK_RAG=true` so the LLM is stubbed and runs are deterministic.

DON'T add E2E for behaviour adequately covered by component tests. E2E is expensive — reserve it for full-stack flows.

## Coverage targets

| Layer | Target | Enforcement |
|---|---|---|
| `pkg/*` (shared Go libs) | 95% | CI fails below |
| `services/*/internal/service/` (business logic) | 90% | CI fails below |
| `services/*/internal/handler/` (HTTP) | 80% | CI fails below |
| Frontend `lib/` | 95% | CI fails below |
| Frontend `hooks/` | 80% | CI fails below |
| React components | not enforced | reviewer judgment |

DON'T game coverage with tests that exercise lines without asserting behaviour. The reviewer flags shape-only tests.

## TDD

DO write the failing test first for new behaviour. Run it, see it fail for the right reason, then implement minimally to pass. Then refactor.

If you wrote implementation before the test, delete it and start over. This is the iron rule — the `test-driven-development` skill enforces it.

## Test data

DO seed minimum data per test in `t.Helper()` constructors — not in shared fixtures. Tests read clearer when their setup is in the same scope as the assertion.

DON'T rely on test execution order. `go test` may parallelise.

## Speed

DO mark expensive tests `t.Parallel()` where they are isolated. Integration tests with shared containers cannot parallelise.

DON'T leave `time.Sleep` calls in tests as a synchronisation primitive — use channels, `WaitGroup`, or explicit polling with timeout.
