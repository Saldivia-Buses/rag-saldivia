---
name: test-driven-development
description: "Write tests FIRST, then implement. Go table-driven tests with testify. Never write implementation without a failing test."
user_invocable: true
---

# Test-Driven Development — SDA Framework

Write the test FIRST. Then make it pass. Then refactor. No exceptions.

## Protocol

### Step 1: Write the Test
Before touching any implementation file:
1. Create or open `*_test.go` next to the file you'll implement
2. Write a table-driven test with testify assertions
3. Include happy path + at least one error case
4. Run the test — it MUST fail (red)

```go
func TestCreateInvoice(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateInvoiceParams
        wantErr bool
    }{
        {
            name:  "valid invoice",
            input: CreateInvoiceParams{TenantID: "t1", Amount: 100},
        },
        {
            name:    "missing tenant",
            input:   CreateInvoiceParams{Amount: 100},
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // arrange, act, assert
        })
    }
}
```

### Step 2: Implement (Green)
Write the minimum code to make the test pass.
- Don't add features the test doesn't cover
- Don't optimize yet
- Run `go test ./path/to/package/ -count=1`

### Step 3: Refactor
Now clean up, but only if tests still pass.
- Extract common patterns to `pkg/`
- Improve naming
- Run full `make test` to verify no regressions

## Test Types by Layer

| Layer | Test approach | Tools |
|-------|--------------|-------|
| Handler | HTTP test with `httptest.NewRecorder()` | `net/http/httptest`, chi |
| Service | Unit test with mocked repository | testify/mock |
| Repository | Integration test with testcontainers | testcontainers-go, pgx |
| Middleware | HTTP test with test handler chain | `httptest` |
| pkg/ | Unit test, no mocks needed | testify/assert |

## SDA-Specific Patterns
- **Tenant isolation**: every test must set tenant context via `tenant.WithContext(ctx, tenantID)`
- **Auth context**: use `jwt.WithClaims(ctx, claims)` for authenticated tests
- **NATS events**: use `nats.NewTestConn()` or mock the publisher interface
- **sqlc**: test against real DB with testcontainers, not mocks (per project convention)

## Anti-patterns
- Writing implementation first and tests after
- Testing only happy paths
- Mocking the database (use testcontainers)
- Tests that depend on external services
- Tests without tenant context
