package config_test

import (
	"context"
	"os"
	"testing"

	"github.com/Camionerou/rag-saldivia/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestResolver(t *testing.T) (*config.Resolver, *pgxpool.Pool) {
	t.Helper()

	dbURL := os.Getenv("POSTGRES_PLATFORM_URL")
	if dbURL == "" {
		dbURL = "postgres://sda:sda_dev@localhost:5432/sda_platform?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Platform DB not available: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Platform DB not reachable: %v", err)
	}

	t.Cleanup(func() { pool.Close() })
	return config.NewResolver(pool, nil), pool
}

func TestResolver_Get_GlobalScope(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	val, err := resolver.Get(ctx, "", "guardrails.input_max_length")
	if err != nil {
		t.Fatalf("expected seeded config, got: %v", err)
	}
	if string(val) != "10000" {
		t.Fatalf("expected 10000, got %s", string(val))
	}
}

func TestResolver_Get_NotFound(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	_, err := resolver.Get(ctx, "", "nonexistent.key.12345")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestResolver_Get_PlanScope(t *testing.T) {
	resolver, pool := newTestResolver(t)
	ctx := context.Background()

	// rate_limits.queries_per_minute is seeded as 10 for plan:starter, 60 for plan:business
	// Create a temp tenant on starter plan to test cascade
	pool.Exec(ctx, `INSERT INTO tenants (id, slug, name, plan_id, postgres_url, redis_url)
		VALUES ('test-cascade-tenant', 'test-cascade', 'Test', 'starter',
			'postgres://localhost/test', 'redis://localhost/0')
		ON CONFLICT (id) DO NOTHING`)
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM tenants WHERE id = 'test-cascade-tenant'`)
	})

	val, err := resolver.GetInt(ctx, "test-cascade-tenant", "rate_limits.queries_per_minute")
	if err != nil {
		t.Fatalf("expected plan-scoped config, got: %v", err)
	}
	if val != 10 {
		t.Fatalf("expected 10 (starter plan), got %d", val)
	}
}

func TestResolver_Get_TenantOverride(t *testing.T) {
	resolver, pool := newTestResolver(t)
	ctx := context.Background()

	// Create tenant + override
	pool.Exec(ctx, `INSERT INTO tenants (id, slug, name, plan_id, postgres_url, redis_url)
		VALUES ('test-override-tenant', 'test-override', 'Test Override', 'starter',
			'postgres://localhost/test', 'redis://localhost/0')
		ON CONFLICT (id) DO NOTHING`)
	pool.Exec(ctx, `INSERT INTO agent_config (scope, key, value, updated_by)
		VALUES ('tenant:test-override-tenant', 'guardrails.input_max_length', '5000', 'test')
		ON CONFLICT (scope, key) DO UPDATE SET value = '5000'`)
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM agent_config WHERE scope = 'tenant:test-override-tenant'`)
		pool.Exec(ctx, `DELETE FROM tenants WHERE id = 'test-override-tenant'`)
	})

	val, err := resolver.GetInt(ctx, "test-override-tenant", "guardrails.input_max_length")
	if err != nil {
		t.Fatalf("expected tenant override, got: %v", err)
	}
	if val != 5000 {
		t.Fatalf("expected 5000 (tenant override), got %d", val)
	}
}

func TestResolver_GetString(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	val, err := resolver.GetString(ctx, "", "slot.ocr")
	if err != nil {
		t.Fatalf("expected seeded slot, got: %v", err)
	}
	if val != "paddleocr-vl" {
		t.Fatalf("expected paddleocr-vl, got %s", val)
	}
}

func TestResolver_GetString_TypeMismatch(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	// guardrails.input_max_length is 10000 (number, not string)
	_, err := resolver.GetString(ctx, "", "guardrails.input_max_length")
	if err == nil {
		t.Fatal("expected error for type mismatch (number as string)")
	}
}

func TestResolver_GetInt(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	val, err := resolver.GetInt(ctx, "", "llm.max_tool_calls_per_turn")
	if err != nil {
		t.Fatalf("expected seeded int, got: %v", err)
	}
	if val != 25 {
		t.Fatalf("expected 25, got %d", val)
	}
}

func TestResolver_ResolveSlot(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	mc, err := resolver.ResolveSlot(ctx, "", "slot.ocr")
	if err != nil {
		t.Fatalf("expected resolved slot, got: %v", err)
	}
	if mc.ID != "paddleocr-vl" {
		t.Fatalf("expected paddleocr-vl, got %s", mc.ID)
	}
	if mc.Endpoint != "http://sglang-ocr:8000" {
		t.Fatalf("expected sglang-ocr endpoint, got %s", mc.Endpoint)
	}
}

func TestResolver_ResolveSlot_NotConfigured(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	_, err := resolver.ResolveSlot(ctx, "", "slot.chat")
	if err == nil {
		t.Fatal("expected error for unconfigured slot (null)")
	}
}

func TestResolver_GetActivePrompt(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	content, err := resolver.GetActivePrompt(ctx, "tree_search")
	if err != nil {
		t.Fatalf("expected seeded prompt, got: %v", err)
	}
	if content == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestResolver_GetActivePrompt_NotFound(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	_, err := resolver.GetActivePrompt(ctx, "nonexistent_prompt")
	if err == nil {
		t.Fatal("expected error for nonexistent prompt")
	}
}
