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

	// Verify connection
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

	// agent_config is seeded with global scope entries
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

func TestResolver_GetString(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	// slot.ocr is seeded as "paddleocr-vl"
	val, err := resolver.GetString(ctx, "", "slot.ocr")
	if err != nil {
		t.Fatalf("expected seeded slot, got: %v", err)
	}
	if val != "paddleocr-vl" {
		t.Fatalf("expected paddleocr-vl, got %s", val)
	}
}

func TestResolver_GetInt(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	val, err := resolver.GetInt(ctx, "", "llm.max_tool_calls_per_turn")
	if err != nil {
		t.Fatalf("expected seeded int config, got: %v", err)
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
		t.Fatalf("expected paddleocr-vl model, got %s", mc.ID)
	}
	if mc.Endpoint != "http://sglang-ocr:8000" {
		t.Fatalf("expected sglang-ocr endpoint, got %s", mc.Endpoint)
	}
}

func TestResolver_ResolveSlot_NotConfigured(t *testing.T) {
	resolver, _ := newTestResolver(t)
	ctx := context.Background()

	// slot.chat is seeded as null
	_, err := resolver.ResolveSlot(ctx, "", "slot.chat")
	if err == nil {
		t.Fatal("expected error for unconfigured slot")
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
		t.Fatal("expected non-empty prompt content")
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
