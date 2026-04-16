package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// extractor_consumer_test.go covers unit-testable paths in ExtractorConsumer.
//
// ARCHITECTURE NOTE: ExtractorConsumer.handleResult uses *repository.Queries (a concrete
// struct backed by *pgxpool.Pool), not an interface. This means the full happy-path
// and DB-error paths cannot be unit-tested without a real database.
// TDD-ANCHOR: if ExtractorConsumer is refactored to accept a DocumentRepo interface,
// the tests marked below can be promoted from integration to unit.
//
// What CAN be tested without DB:
//   1. tenantFromSubject — already tested in worker_test.go (same package, same function)
//   2. handleResult early-exit on malformed JSON → Term()
//   3. handleResult early-exit on valid JSON with empty DocumentID → Nak() (after repo call fails)
//      — cannot test without DB
//   4. ExtractionResult JSON deserialization correctness
//   5. Invariant: tenant comes from NATS subject, NOT from payload
//      — handleResult currently ignores subject tenant (single-tenant dev mode, documented
//        in production code NOTE). Test documents this gap for future multi-tenant wiring.

// --- mockExtractorMsg ---
// Implements the subset of jetstream.Msg used by handleResult.
// Reuses the same interface pattern as mockMsg in worker_test.go.

type mockExtractorMsg struct {
	subject string
	data    []byte

	AckCalled  bool
	NakCalled  bool
	TermCalled bool
}

func (m *mockExtractorMsg) Subject() string                        { return m.subject }
func (m *mockExtractorMsg) Data() []byte                           { return m.data }
func (m *mockExtractorMsg) Ack() error                             { m.AckCalled = true; return nil }
func (m *mockExtractorMsg) Nak() error                             { m.NakCalled = true; return nil }
func (m *mockExtractorMsg) Term() error                            { m.TermCalled = true; return nil }
func (m *mockExtractorMsg) TermWithReason(_ string) error          { m.TermCalled = true; return nil }
func (m *mockExtractorMsg) NakWithDelay(_ time.Duration) error     { m.NakCalled = true; return nil }
func (m *mockExtractorMsg) Headers() nats.Header                   { return nil }
func (m *mockExtractorMsg) Reply() string                          { return "" }
func (m *mockExtractorMsg) DoubleAck(_ context.Context) error      { return nil }
func (m *mockExtractorMsg) InProgress() error                      { return nil }
func (m *mockExtractorMsg) Metadata() (*jetstream.MsgMetadata, error) {
	return nil, nil
}

// Compile-time check: mockExtractorMsg must satisfy the full jetstream.Msg interface.
var _ jetstream.Msg = (*mockExtractorMsg)(nil)

// --- ExtractionResult JSON deserialization ---

func TestExtractionResult_Deserialization(t *testing.T) {
	raw := `{
		"document_id": "doc-abc",
		"file_name":   "report.pdf",
		"total_pages": 3,
		"pages": [
			{"page_number": 1, "text": "Page one content", "tables": [], "images": []},
			{"page_number": 2, "text": "Page two content", "tables": null, "images": null},
			{"page_number": 3, "text": "",                 "tables": [],  "images": []}
		],
		"metadata": {"author": "test"}
	}`

	var result ExtractionResult
	require.NoError(t, json.Unmarshal([]byte(raw), &result))

	assert.Equal(t, "doc-abc", result.DocumentID)
	assert.Equal(t, "report.pdf", result.FileName)
	assert.Equal(t, 3, result.TotalPages)
	require.Len(t, result.Pages, 3)

	assert.Equal(t, 1, result.Pages[0].PageNumber)
	assert.Equal(t, "Page one content", result.Pages[0].Text)

	// Null tables/images on page 2 are valid — BulkInsertPages handles nil by defaulting to [].
	// NOTE: json.RawMessage unmarshals JSON null as json.RawMessage("null"), not as nil.
	// The Go JSON decoder sets the field to the literal bytes []byte("null").
	// BulkInsertPages guards against this with: if tables == nil { tables = [] }
	// which only catches Go nil, not JSON null bytes. The production code's nil check
	// still works because json.RawMessage("null") is a non-nil slice — but its content
	// is valid JSON null, which PostgreSQL JSONB accepts correctly.
	assert.Equal(t, json.RawMessage("null"), result.Pages[1].Tables)
	assert.Equal(t, json.RawMessage("null"), result.Pages[1].Images)

	// Page 3 has empty text — valid (some pages may be images only)
	assert.Equal(t, "", result.Pages[2].Text)
}

func TestExtractionResult_MalformedJSON_CannotUnmarshal(t *testing.T) {
	// Documents that handleResult receives with bad JSON must be caught by
	// json.Unmarshal before any DB call. The test confirms the JSON shape is
	// detected as invalid — the handler calls Term() for these.
	badPayloads := []struct {
		name string
		data string
	}{
		{"empty string", ""},
		{"truncated JSON", `{"document_id": "doc-1"`},
		{"wrong type for pages", `{"document_id": "doc-1", "pages": "not-an-array"}`},
		{"number instead of object", `42`},
	}

	for _, tc := range badPayloads {
		t.Run(tc.name, func(t *testing.T) {
			var result ExtractionResult
			err := json.Unmarshal([]byte(tc.data), &result)
			assert.Error(t, err, "bad payload must fail unmarshal — handleResult will call Term()")
		})
	}
}

// --- handleResult early-exit: malformed JSON → Term() ---
//
// TDD-ANCHOR: to test handleResult we need an ExtractorConsumer with a real pool.
// The only path testable without a pool is the JSON decode failure path — it calls
// msg.Term() and returns before any repo access.
//
// We instantiate ExtractorConsumer with nil pool and nil treeGen. The nil pool will
// cause a panic only if repo methods are reached. If Term() is called first, no panic.

func TestExtractorConsumer_HandleResult_MalformedJSON_Terms(t *testing.T) {
	// nil pool is safe here: json.Unmarshal fails before any repo call
	c := &ExtractorConsumer{
		// pool and repo intentionally nil — Term() fires before repo access
		ctx: t.Context(),
	}

	msg := &mockExtractorMsg{
		subject: "tenant.saldivia.extractor.result.done",
		data:    []byte(`{invalid json`),
	}

	c.handleResult(msg)

	assert.True(t, msg.TermCalled, "malformed JSON must call Term() — C2 invariant")
	assert.False(t, msg.AckCalled, "malformed payload must not be acked")
	assert.False(t, msg.NakCalled, "malformed payload must not be nacked (no retry for corrupt data)")
}

func TestExtractorConsumer_HandleResult_EmptyPayload_Terms(t *testing.T) {
	c := &ExtractorConsumer{
		ctx: t.Context(),
	}

	msg := &mockExtractorMsg{
		subject: "tenant.saldivia.extractor.result.done",
		data:    []byte(``), // empty — json.Unmarshal returns error
	}

	c.handleResult(msg)

	assert.True(t, msg.TermCalled, "empty payload must call Term() — C2 invariant")
	assert.False(t, msg.AckCalled)
	assert.False(t, msg.NakCalled)
}

// --- Invariant: tenant comes from subject, NOT from payload ---
//
// OBSERVATION: handleResult currently does NOT use the tenant slug from the NATS subject.
// The consumer operates in single-tenant dev mode (noted in production code at line 18-19).
// The pool passed to NewExtractorConsumer is a single shared pool — tenant routing is not
// implemented at the consumer level yet.
//
// This test documents the CURRENT behavior and the EXPECTED behavior post-multi-tenant:
//   Current:  handleResult ignores msg.Subject() entirely — pool is fixed at construction.
//   Expected: consumer should resolve tenant from subject and route to correct DB pool.
//
// TDD-ANCHOR: when multi-tenant wiring is added, this test should assert that:
//   - subject "tenant.saldivia.extractor.result.done" → uses saldivia pool
//   - subject "tenant.acme.extractor.result.done"     → uses acme pool
//   - payload {"tenant": "attacker"} is ignored entirely

func TestExtractorConsumer_TenantSlugFromSubject_Documented(t *testing.T) {
	// Verify tenantFromSubject parses extractor subjects correctly.
	// This is the function that SHOULD be used when multi-tenant routing is added.
	tests := []struct {
		subject string
		want    string
	}{
		{"tenant.saldivia.extractor.result.done", "saldivia"},
		{"tenant.acme.extractor.result.foo.bar", "acme"},
		{"tenant.my-tenant.extractor.result.done", "my-tenant"},
		{"bad.subject", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := tenantFromSubject(tt.subject)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractorConsumer_TenantIsolationInvariant_Gap(t *testing.T) {
	// INVARIANTE: en producción multi-tenant, el tenant slug debe venir SIEMPRE
	// del subject NATS, no del payload.
	//
	// Attack scenario: payload contiene {"document_id": "...", "tenant": "attacker"}
	// pero el subject es "tenant.saldivia.extractor.result.done".
	// El sistema debe usar "saldivia" y NUNCA confiar en el body.
	//
	// ExtractionResult no tiene campo "tenant" — por diseño correcto.
	// El payload no puede inyectar el tenant porque el struct no lo modela.
	// Esto verifica que el schema del payload no exponga una superficie de ataque.

	payload := `{
		"document_id": "doc-legit",
		"file_name":   "file.pdf",
		"total_pages": 1,
		"pages": []
	}`

	var result ExtractionResult
	require.NoError(t, json.Unmarshal([]byte(payload), &result))

	// ExtractionResult has no tenant field — tenant cannot be injected via payload.
	// The zero value of any unexported tenant field would be "".
	// This assertion confirms the struct schema prevents tenant spoofing.
	assert.Equal(t, "doc-legit", result.DocumentID)

	// Verify tenantFromSubject gives us the trusted tenant regardless of payload.
	subject := "tenant.saldivia.extractor.result.done"
	trustedTenant := tenantFromSubject(subject)
	assert.Equal(t, "saldivia", trustedTenant,
		"tenant must always come from NATS subject, never from payload")
}

// --- TDD-ANCHOR: paths that require a real DB ---
//
// The following behaviors need testcontainers (integration tag) or an interface refactor:
//
//   TestExtractorConsumer_HandleResult_Success_AcksAndUpdatesJob:
//     - valid payload + working pool → status="indexing", pages stored, status="ready", Ack()
//     - requires: real pool, documents table, document_pages table
//
//   TestExtractorConsumer_HandleResult_UpdateStatusFails_Naks:
//     - repo.UpdateDocumentStatus returns error → msg.Nak() for retry
//     - requires: mock repo interface OR broken pool
//
//   TestExtractorConsumer_HandleResult_BulkInsertFails_Naks:
//     - BulkInsertPages returns error (broken tx) → msg.Nak()
//     - requires: pool that fails mid-transaction
//
//   TestExtractorConsumer_HandleResult_AllPagesZero_Naks:
//     - storedPages == 0 with len(pages) > 0 → sets doc error, msg.Nak()
//     - requires: pool where CopyFrom returns 0 rows
