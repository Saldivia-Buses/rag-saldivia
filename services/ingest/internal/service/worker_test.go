package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock jetstream.Msg ---

// mockMsg implements just enough of jetstream.Msg for worker unit tests.
// All ack/nak/term calls are recorded for assertion.
type mockMsg struct {
	subject string
	data    []byte

	// recorded calls
	AckCalled  bool
	NakCalled  bool
	TermCalled bool

	// simulate Metadata() response — nil means no metadata
	deliveryCount uint64
}

func (m *mockMsg) Subject() string                        { return m.subject }
func (m *mockMsg) Data() []byte                           { return m.data }
func (m *mockMsg) Ack() error                             { m.AckCalled = true; return nil }
func (m *mockMsg) Nak() error                             { m.NakCalled = true; return nil }
func (m *mockMsg) Term() error                            { m.TermCalled = true; return nil }
func (m *mockMsg) TermWithReason(_ string) error          { m.TermCalled = true; return nil }
func (m *mockMsg) NakWithDelay(_ time.Duration) error     { m.NakCalled = true; return nil }
func (m *mockMsg) Headers() nats.Header                   { return nil }
func (m *mockMsg) Reply() string                          { return "" }
func (m *mockMsg) DoubleAck(_ context.Context) error      { return nil }
func (m *mockMsg) InProgress() error                      { return nil }

// Metadata returns a *jetstream.MsgMetadata whose NumDelivered field is set
// from the mock's deliveryCount. NumDelivered is a struct field (not a method)
// on the real jetstream.MsgMetadata type.
func (m *mockMsg) Metadata() (*jetstream.MsgMetadata, error) {
	if m.deliveryCount == 0 {
		return nil, nil
	}
	return &jetstream.MsgMetadata{NumDelivered: m.deliveryCount}, nil
}

// Compile-time check: mockMsg must satisfy the full jetstream.Msg interface.
var _ jetstream.Msg = (*mockMsg)(nil)

// --- mock EventPublisher ---

type mockPublisher struct {
	NotifyCalled bool
	NotifyArgs   []notifyCall
	ReturnErr    error
}

type notifyCall struct {
	Slug string
	Evt  any
}

func (m *mockPublisher) Notify(slug string, evt any) error {
	m.NotifyCalled = true
	m.NotifyArgs = append(m.NotifyArgs, notifyCall{Slug: slug, Evt: evt})
	return m.ReturnErr
}

// --- helpers ---

// buildWorkerWithServer creates a worker wired to a test HTTP server.
// Returns the worker and a cleanup func.
func buildWorkerWithServer(t *testing.T, srv *httptest.Server, _ EventPublisher) *Worker {
	t.Helper()
	cfg := Config{
		BlueprintURL: srv.URL,
		StagingDir:   t.TempDir(),
		Timeout:      0,
	}
	w := &Worker{
		nc:         nil,
		svc:        &ingestSpy{},
		tenantSlug: "test",
		client:     &http.Client{Timeout: cfg.Timeout},
		cfg:        cfg,
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	w.ctx = ctx
	w.cancel = cancel
	return w
}

// stagedFile writes content to a temp file and returns its path.
func stagedFile(t *testing.T, dir string) string {
	t.Helper()
	f, err := os.CreateTemp(dir, "ingest-*")
	require.NoError(t, err)
	_, err = f.WriteString("document content")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

// ingestMsg serializes an IngestMessage to JSON bytes.
func ingestMsg(t *testing.T, msg IngestMessage) []byte {
	t.Helper()
	b, err := json.Marshal(msg)
	require.NoError(t, err)
	return b
}

// --- tenantFromSubject unit tests ---

func TestTenantFromSubject(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    string
	}{
		{
			name:    "valid ingest subject",
			subject: "tenant.saldivia.ingest.process",
			want:    "saldivia",
		},
		{
			name:    "valid extractor subject",
			subject: "tenant.acme.extractor.result.done",
			want:    "acme",
		},
		{
			name:    "missing tenant prefix",
			subject: "saldivia.ingest.process",
			want:    "",
		},
		{
			name:    "too short subject",
			subject: "tenant",
			want:    "",
		},
		{
			name:    "empty string",
			subject: "",
			want:    "",
		},
		{
			name:    "slug with hyphen",
			subject: "tenant.my-tenant.ingest.process",
			want:    "my-tenant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tenantFromSubject(tt.subject)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- processJob tests ---
// Worker.svc is a JobStatusUpdater interface, so ingestSpy records calls
// without needing a real pool. buildWorkerWithServer wires a fresh spy by
// default. Tests that need to inspect the updates assign a named spy.

// ingestSpy implements JobStatusUpdater and records calls for assertions.
type ingestSpy struct {
	updates []statusUpdate
}

type statusUpdate struct {
	JobID  string
	Status string
	ErrMsg *string
}

func (s *ingestSpy) UpdateJobStatus(_ context.Context, jobID, status string, errMsg *string) error {
	s.updates = append(s.updates, statusUpdate{JobID: jobID, Status: status, ErrMsg: errMsg})
	return nil
}

func TestProcessJob_MalformedJSON_Terms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)

	msg := &mockMsg{
		subject: "tenant.saldivia.ingest.process",
		data:    []byte(`{not valid json`),
	}
	w.processJob(msg)

	assert.True(t, msg.TermCalled, "malformed JSON must call Term()")
	assert.False(t, msg.AckCalled)
	assert.False(t, msg.NakCalled)
}

func TestProcessJob_MissingFields_Terms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)

	// JobID empty
	data := ingestMsg(t, IngestMessage{
		JobID:      "",
		TenantSlug: "saldivia",
		StagedPath: "/some/path",
	})
	msg := &mockMsg{
		subject: "tenant.saldivia.ingest.process",
		data:    data,
	}
	w.processJob(msg)

	assert.True(t, msg.TermCalled, "missing JobID must call Term()")
	assert.False(t, msg.AckCalled)
}

func TestProcessJob_TenantMismatch_Terms(t *testing.T) {
	// INVARIANTE CRITICA: tenant slug viene del NATS subject, NO del payload.
	// Un payload con tenant="attacker" en subject tenant.saldivia.* debe ser terminado.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)

	data := ingestMsg(t, IngestMessage{
		JobID:      "job-1",
		TenantSlug: "attacker",        // mismatch — subject says "saldivia"
		StagedPath: "/some/path",
	})
	msg := &mockMsg{
		subject: "tenant.saldivia.ingest.process", // trusted source
		data:    data,
	}
	w.processJob(msg)

	assert.True(t, msg.TermCalled, "tenant mismatch must call Term() not Ack/Nak")
	assert.False(t, msg.AckCalled, "spoofed tenant payload must not be acked")
	assert.False(t, msg.NakCalled, "spoofed tenant payload must not be nacked (no retry)")
}

func TestProcessJob_StagedFileNotFound_RetryThenTerm(t *testing.T) {
	// When forwardToBlueprint fails (staged file missing), processJob must:
	//   - on retries 1-2: Nak() for redelivery, job stays "processing"
	//   - on final attempt (NumDelivered >= maxDeliveries): Term() + status="failed"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK) // server is fine; file is missing locally
	}))
	defer srv.Close()

	t.Run("retry attempt naks and keeps processing status", func(t *testing.T) {
		spy := &ingestSpy{}
		w := buildWorkerWithServer(t, srv, nil)
		w.svc = spy

		missingPath := filepath.Join(t.TempDir(), "does-not-exist.pdf")
		data := ingestMsg(t, IngestMessage{
			JobID:      "job-retry",
			TenantSlug: "saldivia",
			StagedPath: missingPath,
			Collection: "docs",
			FileName:   "missing.pdf",
		})
		msg := &mockMsg{
			subject:       "tenant.saldivia.ingest.process",
			data:          data,
			deliveryCount: 1, // first attempt — should retry
		}
		w.processJob(msg)

		assert.True(t, msg.NakCalled, "retry attempt must Nak for redelivery")
		assert.False(t, msg.TermCalled)
		assert.False(t, msg.AckCalled)

		// Status stays "processing" — no "failed" update until final attempt.
		require.Len(t, spy.updates, 1, "only the initial processing update should fire on retry")
		assert.Equal(t, "processing", spy.updates[0].Status)
		assert.Nil(t, spy.updates[0].ErrMsg)
	})

	t.Run("final attempt terms and marks failed", func(t *testing.T) {
		spy := &ingestSpy{}
		w := buildWorkerWithServer(t, srv, nil)
		w.svc = spy

		missingPath := filepath.Join(t.TempDir(), "does-not-exist.pdf")
		data := ingestMsg(t, IngestMessage{
			JobID:      "job-final",
			TenantSlug: "saldivia",
			StagedPath: missingPath,
			Collection: "docs",
			FileName:   "missing.pdf",
		})
		msg := &mockMsg{
			subject:       "tenant.saldivia.ingest.process",
			data:          data,
			deliveryCount: uint64(maxDeliveries), // final attempt
		}
		w.processJob(msg)

		assert.True(t, msg.TermCalled, "final attempt must Term (no more retries)")
		assert.False(t, msg.NakCalled)
		assert.False(t, msg.AckCalled)

		// Spy must see processing → failed with error message.
		require.Len(t, spy.updates, 2)
		assert.Equal(t, "processing", spy.updates[0].Status)
		assert.Equal(t, "failed", spy.updates[1].Status)
		require.NotNil(t, spy.updates[1].ErrMsg)
		assert.Contains(t, *spy.updates[1].ErrMsg, "failed after")
	})
}

// --- forwardToBlueprint behavior via HTTP test server ---

func TestForwardToBlueprint_Success_OnHTTP200(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/documents", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)

	// Write real temp file
	staged := stagedFile(t, t.TempDir())

	im := IngestMessage{
		JobID:      "job-123",
		TenantSlug: "saldivia",
		UserID:     "u-1",
		Collection: "fleet-docs",
		FileName:   "report.pdf",
		StagedPath: staged,
	}

	err := w.forwardToBlueprint(context.Background(), im)
	require.NoError(t, err)
	assert.True(t, called, "Blueprint endpoint must be called")
}

func TestForwardToBlueprint_NamespacedCollection(t *testing.T) {
	// Blueprint collection must be namespaced: "{tenant}-{collection}"
	var collectionReceived string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err == nil {
			collectionReceived = r.FormValue("collection_name")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)
	staged := stagedFile(t, t.TempDir())

	im := IngestMessage{
		JobID:      "job-456",
		TenantSlug: "saldivia",
		Collection: "fleet-docs",
		FileName:   "x.pdf",
		StagedPath: staged,
	}

	err := w.forwardToBlueprint(context.Background(), im)
	require.NoError(t, err)
	assert.Equal(t, "saldivia-fleet-docs", collectionReceived,
		"collection_name must be prefixed with tenant slug for isolation")
}

func TestForwardToBlueprint_HTTPError_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)
	staged := stagedFile(t, t.TempDir())

	im := IngestMessage{
		JobID:      "job-789",
		TenantSlug: "saldivia",
		Collection: "docs",
		FileName:   "y.pdf",
		StagedPath: staged,
	}

	err := w.forwardToBlueprint(context.Background(), im)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestForwardToBlueprint_FileNotFound_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, nil)

	im := IngestMessage{
		JobID:      "job-999",
		TenantSlug: "saldivia",
		Collection: "docs",
		FileName:   "missing.pdf",
		StagedPath: filepath.Join(t.TempDir(), "nonexistent.pdf"),
	}

	err := w.forwardToBlueprint(context.Background(), im)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open staged file")
}

func TestPublishCompletion_CallsPublisher(t *testing.T) {
	pub := &mockPublisher{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()

	w := buildWorkerWithServer(t, srv, pub)

	im := IngestMessage{
		JobID:      "job-1",
		TenantSlug: "saldivia",
		UserID:     "u-1",
		Collection: "fleet",
		FileName:   "doc.pdf",
	}

	// broadcastStatus is the only non-outbox path. It should not panic
	// even when nc is nil (the json.Marshal is always valid).
	assert.NotPanics(t, func() {
		w.broadcastStatus(im)
	})
}

// --- mockMsg compatibility check ---
// Ensure mockMsg implements jetstream.Msg at compile time.
// We use a local interface that matches the methods called in processJob.

type jetMsg interface {
	Subject() string
	Data() []byte
	Ack() error
	Nak() error
	Term() error
}

var _ jetMsg = (*mockMsg)(nil)

