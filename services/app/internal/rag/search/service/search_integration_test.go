//go:build integration

// Integration tests for the Search service.
// Requires Docker (testcontainers-go spins up PostgreSQL automatically).
// Run: go test -tags=integration -v -timeout 90s ./internal/service/

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// intelligenceSchema is the subset of migration 007 needed by the search service.
// Mirrors db/tenant/migrations/007_intelligence_schema.up.sql.
const intelligenceSchema = `
CREATE TABLE IF NOT EXISTS documents (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name        TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    file_type   TEXT NOT NULL,
    file_hash   TEXT NOT NULL,
    size_bytes  BIGINT NOT NULL,
    total_pages INT,
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'extracting', 'indexing', 'ready', 'error')),
    metadata    JSONB DEFAULT '{}',
    uploaded_by TEXT NOT NULL,
    error       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS document_pages (
    id            TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    document_id   TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    page_number   INT NOT NULL,
    text          TEXT NOT NULL DEFAULT '',
    tables        JSONB DEFAULT '[]',
    images        JSONB DEFAULT '[]',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(document_id, page_number)
);

CREATE TABLE IF NOT EXISTS document_trees (
    id              TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    document_id     TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tree            JSONB NOT NULL,
    doc_description TEXT,
    tree_version    INT NOT NULL DEFAULT 1,
    model_used      TEXT NOT NULL,
    node_count      INT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS collections (
    id          TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS collection_documents (
    collection_id TEXT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    document_id   TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, document_id)
);
`

// setupTestDB spins up a postgres container and applies the intelligence schema.
// The caller is responsible for terminating the container via t.Cleanup.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := t.Context()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("sda_search_test"),
		postgres.WithUsername("sda"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "start postgres container")
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "get connection string")

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "create pool")
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, intelligenceSchema)
	require.NoError(t, err, "apply schema")

	return pool
}

// setupTestDBWithConnStr spins up a postgres container and returns the pool AND
// the raw connection string — used for creating secondary pools to the same server.
func setupTestDBWithConnStr(t *testing.T, dbName string) (*pgxpool.Pool, string) {
	t.Helper()
	ctx := t.Context()

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername("sda"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "start postgres container for %s", dbName)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("terminate container %s: %v", dbName, err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "get connection string for %s", dbName)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err, "create pool for %s", dbName)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, intelligenceSchema)
	require.NoError(t, err, "apply schema for %s", dbName)

	return pool, connStr
}

// mockLLMServer returns an httptest.Server that implements the OpenAI chat
// completions API. The handler calls nodeSelectorFn with the prompt and returns
// whatever node IDs it provides as a comma-separated string.
//
// This replaces the real SGLang/OpenAI endpoint so that integration tests do
// not require GPU infrastructure.
func mockLLMServer(t *testing.T, nodeIDs ...string) *httptest.Server {
	t.Helper()
	response := strings.Join(nodeIDs, ", ")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Minimal OpenAI-compatible response
		resp := map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": response,
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 5,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// seedDocument inserts a document + document_tree into the given pool.
// Returns the document ID.
func seedDocument(t *testing.T, pool *pgxpool.Pool, name, nodeID, hash string) string {
	t.Helper()
	ctx := t.Context()

	// Insert document
	var docID string
	err := pool.QueryRow(ctx,
		`INSERT INTO documents (name, storage_key, file_type, file_hash, size_bytes, uploaded_by, status)
		 VALUES ($1, $2, 'pdf', $3, 1024, 'user-1', 'ready')
		 RETURNING id`,
		name, "storage/"+hash, hash,
	).Scan(&docID)
	require.NoError(t, err, "seed document %s", name)

	// Build a minimal tree
	tree := []TreeNode{
		{
			NodeID:     nodeID,
			Title:      fmt.Sprintf("%s — Chapter 1", name),
			Summary:    fmt.Sprintf("Main content of %s", name),
			StartIndex: 1,
			EndIndex:   3,
		},
	}
	treeJSON, err := json.Marshal(tree)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO document_trees (document_id, tree, doc_description, model_used, node_count)
		 VALUES ($1, $2, $3, 'test-model', 1)`,
		docID, treeJSON, fmt.Sprintf("Description of %s", name),
	)
	require.NoError(t, err, "seed document_tree for %s", name)

	// Insert a page so extractPages can fetch content
	_, err = pool.Exec(ctx,
		`INSERT INTO document_pages (document_id, page_number, text, tables, images)
		 VALUES ($1, 1, $2, '[]', '[]')`,
		docID, fmt.Sprintf("Page content of %s", name),
	)
	require.NoError(t, err, "seed document_page for %s", name)

	return docID
}

// TestSearch_NoDocuments_ReturnsEmpty_Integration verifies that when the
// tenant DB has no document_trees, SearchDocuments returns an empty
// Selections slice (not nil, not an error).
func TestSearch_NoDocuments_ReturnsEmpty_Integration(t *testing.T) {
	pool := setupTestDB(t)

	// LLM mock returns a node ID that won't match anything (no docs in DB)
	llmSrv := mockLLMServer(t, "node-xyz")
	svc := New(pool, llmSrv.URL, "test-model")

	result, err := svc.SearchDocuments(t.Context(), "what is the revenue?", "", 5)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "what is the revenue?", result.Query)
	assert.Empty(t, result.Selections,
		"empty DB must yield empty Selections, not nil or error")
}

// TestSearch_WithDocumentTree_FindsRelevantNodes_Integration verifies the happy
// path: a seeded document with a tree produces non-empty Selections when the
// mock LLM returns a valid node ID.
func TestSearch_WithDocumentTree_FindsRelevantNodes_Integration(t *testing.T) {
	pool := setupTestDB(t)

	const nodeID = "ch1-revenue"
	seedDocument(t, pool, "Annual Report 2025", nodeID, "hash-ar2025")

	// LLM mock returns the seeded node ID
	llmSrv := mockLLMServer(t, nodeID)
	svc := New(pool, llmSrv.URL, "test-model")

	result, err := svc.SearchDocuments(t.Context(), "what is the revenue?", "", 5)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Selections,
		"expected at least one selection when LLM returns a valid node ID")

	sel := result.Selections[0]
	assert.Equal(t, "Annual Report 2025", sel.Document,
		"selection must reference the seeded document")
	assert.Contains(t, sel.NodeIDs, nodeID,
		"selection NodeIDs must include the LLM-selected node")
	assert.Contains(t, sel.Text, "Page content of Annual Report 2025",
		"selection Text must include extracted page content")
}

// TestSearch_TenantIsolation_OnlyOwnDocuments_Integration is the CRITICAL invariant
// test for search tenant isolation.
//
// SDA architecture: each tenant has its own PostgreSQL database. The search service
// receives a *pgxpool.Pool that is already scoped to the tenant's DB — it never
// queries across tenant boundaries.
//
// This test verifies that invariant: when Search is given poolA (tenant A's DB)
// it sees ONLY tenant A's documents, even when tenant B's DB has documents with
// the same node IDs.
//
// If this test FAILS it means the service is cross-contaminating tenant data
// (e.g., by using a shared pool or hardcoded connection string).
func TestSearch_TenantIsolation_OnlyOwnDocuments_Integration(t *testing.T) {
	// Spin up two separate postgres containers — one per tenant DB.
	poolA, _ := setupTestDBWithConnStr(t, "tenant_a_db")
	poolB, _ := setupTestDBWithConnStr(t, "tenant_b_db")

	// Seed tenant A's DB with a document
	const nodeA = "sec-financials-a"
	seedDocument(t, poolA, "Tenant A Report", nodeA, "hash-a")

	// Seed tenant B's DB with a different document using the SAME node ID.
	// If the service leaks across pools, it would return B's docs when A queries.
	const nodeB = "sec-financials-a" // same node ID — intentional to expose any leakage
	seedDocument(t, poolB, "Tenant B Report", nodeB, "hash-b")

	// LLM mock returns nodeA for both services (simulates same query result)
	llmSrv := mockLLMServer(t, nodeA)

	// Service A uses pool A
	svcA := New(poolA, llmSrv.URL, "test-model")
	resultA, err := svcA.SearchDocuments(t.Context(), "financials", "", 5)
	require.NoError(t, err, "svcA.SearchDocuments must not error")

	// Service B uses pool B
	svcB := New(poolB, llmSrv.URL, "test-model")
	resultB, err := svcB.SearchDocuments(t.Context(), "financials", "", 5)
	require.NoError(t, err, "svcB.SearchDocuments must not error")

	// CRITICAL: resultA must only contain Tenant A's document
	require.NotEmpty(t, resultA.Selections,
		"tenant A search must return its own documents")
	for _, sel := range resultA.Selections {
		assert.Equal(t, "Tenant A Report", sel.Document,
			"TENANT ISOLATION VIOLATION: tenant A search returned document %q (expected only 'Tenant A Report')",
			sel.Document)
	}

	// CRITICAL: resultB must only contain Tenant B's document
	require.NotEmpty(t, resultB.Selections,
		"tenant B search must return its own documents")
	for _, sel := range resultB.Selections {
		assert.Equal(t, "Tenant B Report", sel.Document,
			"TENANT ISOLATION VIOLATION: tenant B search returned document %q (expected only 'Tenant B Report')",
			sel.Document)
	}

	// Cross-check: no doc name from A appears in B's results and vice versa
	aDocs := make(map[string]bool)
	for _, sel := range resultA.Selections {
		aDocs[sel.Document] = true
	}
	for _, sel := range resultB.Selections {
		assert.False(t, aDocs[sel.Document],
			"TENANT ISOLATION VIOLATION: document %q appeared in both tenant A and B results",
			sel.Document)
	}

	t.Logf("Tenant isolation verified: A saw %d doc(s), B saw %d doc(s)",
		len(resultA.Selections), len(resultB.Selections))
}
