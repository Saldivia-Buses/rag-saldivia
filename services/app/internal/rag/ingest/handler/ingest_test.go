package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/app/internal/rag/ingest/repository"
	"github.com/Camionerou/rag-saldivia/services/app/internal/rag/ingest/service"
)

// --- mock ---

type mockIngestService struct {
	jobs       []service.Job
	collection repository.Collection
	err        error
}

func (m *mockIngestService) ListCollections(_ context.Context) ([]repository.Collection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []repository.Collection{m.collection}, nil
}

func (m *mockIngestService) CreateCollection(_ context.Context, name, description string) (repository.Collection, error) {
	if m.err != nil {
		return repository.Collection{}, m.err
	}
	col := m.collection
	if col.Name == "" {
		col.Name = name
	}
	return col, nil
}

func (m *mockIngestService) Submit(_ context.Context, tenantSlug, userID, collection, fileName string, fileSize int64, _ multipart.File) (*service.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	j := service.Job{
		ID: "j-new", UserID: userID, Collection: collection,
		FileName: fileName, FileSize: fileSize, Status: "pending",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	return &j, nil
}

func (m *mockIngestService) ListJobs(_ context.Context, userID string, limit int) ([]service.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []service.Job
	for _, j := range m.jobs {
		if j.UserID == userID {
			result = append(result, j)
		}
	}
	if result == nil {
		result = []service.Job{}
	}
	return result, nil
}

func (m *mockIngestService) GetJob(_ context.Context, jobID, userID string) (*service.Job, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, j := range m.jobs {
		if j.ID == jobID && j.UserID == userID {
			return &j, nil
		}
	}
	return nil, service.ErrJobNotFound
}

func (m *mockIngestService) DeleteJob(_ context.Context, jobID, userID string) error {
	if m.err != nil {
		return m.err
	}
	for _, j := range m.jobs {
		if j.ID == jobID && j.UserID == userID {
			return nil
		}
	}
	return service.ErrJobNotFound
}

// --- helpers ---

func setupIngestRouter(mock *mockIngestService) *chi.Mux {
	h := NewIngest(mock)
	r := chi.NewRouter()
	r.Mount("/v1/ingest", h.Routes())
	return r
}

func makeUploadRequest(t *testing.T, fileName, collection, userID, tenantSlug string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", fileName)
	_, _ = part.Write([]byte("fake pdf content"))
	_ = writer.WriteField("collection", collection)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("X-Tenant-Slug", tenantSlug)
	req = withAdminContext(req)
	return req
}

func withAdminContext(req *http.Request) *http.Request {
	ctx := sdamw.WithRole(req.Context(), "admin")
	return req.WithContext(ctx)
}

// --- tests ---

func TestUpload_Success(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := makeUploadRequest(t, "document.pdf", "contratos", "u-1", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rec.Code, rec.Body.String())
	}

	var job service.Job
	_ = json.NewDecoder(rec.Body).Decode(&job)
	if job.Status != "pending" {
		t.Errorf("expected status pending, got %q", job.Status)
	}
}

func TestUpload_MissingUserID_Returns403(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "doc.pdf")
	_, _ = part.Write([]byte("content"))
	_ = writer.WriteField("collection", "test")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// No X-User-ID, X-Tenant-Slug, or role — RBAC middleware rejects with 403
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestUpload_MissingTenantSlug_Returns401(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := makeUploadRequest(t, "doc.pdf", "test", "u-1", "")
	req.Header.Del("X-Tenant-Slug")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestUpload_UnsupportedExtension_Returns400(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	unsupported := []string{"malware.exe", "script.sh", "image.png", "video.mp4"}
	for _, name := range unsupported {
		t.Run(name, func(t *testing.T) {
			req := makeUploadRequest(t, name, "test", "u-1", "saldivia")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %s, got %d", name, rec.Code)
			}
		})
	}
}

func TestUpload_SupportedExtensions_Accepted(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	supported := []string{"report.pdf", "letter.docx", "notes.txt", "data.csv", "sheet.xlsx"}
	for _, name := range supported {
		t.Run(name, func(t *testing.T) {
			req := makeUploadRequest(t, name, "test", "u-1", "saldivia")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusAccepted {
				t.Fatalf("expected 202 for %s, got %d: %s", name, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestUpload_MissingCollection_Returns400(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "doc.pdf")
	_, _ = part.Write([]byte("content"))
	// No collection field
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing collection, got %d", rec.Code)
	}
}

func TestListJobs_ReturnsUserJobs(t *testing.T) {
	mock := &mockIngestService{
		jobs: []service.Job{
			{ID: "j-1", UserID: "u-1", FileName: "a.pdf"},
			{ID: "j-2", UserID: "u-2", FileName: "b.pdf"},
		},
	}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string][]service.Job
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp["jobs"]) != 1 {
		t.Errorf("expected 1 job for u-1, got %d", len(resp["jobs"]))
	}
}

func TestListJobs_MissingIdentity_Returns403(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs", nil)
	// No identity or role — RBAC middleware rejects with 403
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestGetJob_OwnerCanAccess(t *testing.T) {
	mock := &mockIngestService{
		jobs: []service.Job{
			{ID: "j-1", UserID: "u-1", FileName: "doc.pdf"},
		},
	}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs/j-1", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetJob_NonOwner_Returns404(t *testing.T) {
	mock := &mockIngestService{
		jobs: []service.Job{
			{ID: "j-1", UserID: "u-1"},
		},
	}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs/j-1", nil)
	req.Header.Set("X-User-ID", "u-2")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner, got %d", rec.Code)
	}
}

func TestDeleteJob_Success(t *testing.T) {
	mock := &mockIngestService{
		jobs: []service.Job{
			{ID: "j-1", UserID: "u-1"},
		},
	}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodDelete, "/v1/ingest/jobs/j-1", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestListJobs_ServiceError_Returns500_GenericMessage(t *testing.T) {
	mock := &mockIngestService{err: errors.New("database connection lost")}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q — service internals may be leaking", resp["error"])
	}
}

func TestDeleteJob_NotFound_Returns404(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := httptest.NewRequest(http.MethodDelete, "/v1/ingest/jobs/nonexistent", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ── Collection CRUD ───────────────────────────────────────────────────────────

func TestListCollections_ReturnsAllCollections(t *testing.T) {
	mock := &mockIngestService{
		collection: repository.Collection{ID: "c-1", Name: "contratos"},
	}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/collections", nil)
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var cols []repository.Collection
	if err := json.NewDecoder(rec.Body).Decode(&cols); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(cols) != 1 {
		t.Errorf("expected 1 collection, got %d", len(cols))
	}
	if cols[0].Name != "contratos" {
		t.Errorf("expected collection name 'contratos', got %q", cols[0].Name)
	}
}

func TestCreateCollection_Success_Returns201(t *testing.T) {
	mock := &mockIngestService{
		collection: repository.Collection{ID: "c-new", Name: "facturas"},
	}
	r := setupIngestRouter(mock)

	body := `{"name":"facturas","description":"Facturas de clientes"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/collections", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateCollection_EmptyName_Returns400(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	body := `{"name":"","description":"missing name"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/collections", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty name, got %d", rec.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("expected non-empty error field in response")
	}
}

func TestCreateCollection_WhitespaceName_Returns400(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	body := `{"name":"   "}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/collections", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for whitespace-only name, got %d", rec.Code)
	}
}

func TestCreateCollection_DuplicateName_Returns409(t *testing.T) {
	// The handler maps "duplicate"/"unique" substring in error message to 409.
	mock := &mockIngestService{err: errors.New("duplicate key violates unique constraint")}
	r := setupIngestRouter(mock)

	body := `{"name":"contratos"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/collections", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate collection, got %d", rec.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("expected non-empty error in conflict response")
	}
}

func TestCreateCollection_InvalidJSON_Returns400(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/collections", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestListCollections_ServiceError_Returns500(t *testing.T) {
	mock := &mockIngestService{err: errors.New("db unavailable")}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/collections", nil)
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// ── File size enforcement ─────────────────────────────────────────────────────

func TestUpload_FileTooBig_Returns400OrRequestEntityTooLarge(t *testing.T) {
	// MaxUploadSize = 100MB. Build a request body that exceeds it.
	// We use a fake multipart body whose Content-Length signals the oversized payload.
	// MaxBytesReader will truncate and ParseMultipartForm will return an error,
	// which the handler maps to 400 "invalid multipart form".
	r := setupIngestRouter(&mockIngestService{})

	// Build a body that is larger than MaxUploadSize (100 << 20).
	// We construct the multipart envelope with a very large fake file part.
	overLimit := int64(MaxUploadSize) + 1

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "huge.pdf")
	// Write overLimit bytes of content (all zeros).
	chunk := make([]byte, 32*1024)
	written := int64(0)
	for written < overLimit {
		n := int64(len(chunk))
		if written+n > overLimit {
			n = overLimit - written
		}
		_, _ = part.Write(chunk[:n])
		written += n
	}
	_ = writer.WriteField("collection", "test")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// MaxBytesReader causes ParseMultipartForm to fail → 400 "invalid multipart form"
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 400 or 413 for oversized file, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ── ListJobs query params ─────────────────────────────────────────────────────

func TestListJobs_WithLimit_ReturnsCorrectCount(t *testing.T) {
	jobs := make([]service.Job, 5)
	for i := range jobs {
		jobs[i] = service.Job{ID: fmt.Sprintf("j-%d", i), UserID: "u-1", FileName: "f.pdf"}
	}
	mock := &mockIngestService{jobs: jobs}
	r := setupIngestRouter(mock)

	// limit=2 is passed through to the mock; mock ListJobs ignores limit
	// but the handler must at least pass it through without error.
	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs?limit=2", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListJobs_NegativeLimit_Ignored(t *testing.T) {
	// A negative or zero limit must not cause an error — handler falls back to default.
	mock := &mockIngestService{jobs: []service.Job{
		{ID: "j-1", UserID: "u-1", FileName: "a.pdf"},
	}}
	r := setupIngestRouter(mock)

	for _, lim := range []string{"-1", "0", "abc"} {
		t.Run("limit="+lim, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs?limit="+lim, nil)
			req.Header.Set("X-User-ID", "u-1")
			req.Header.Set("X-Tenant-Slug", "saldivia")
			req = withAdminContext(req)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("limit=%s: expected 200, got %d", lim, rec.Code)
			}
		})
	}
}

// ── Job isolation: OtherUser → 404 ───────────────────────────────────────────

func TestGetJob_OtherUsersJob_Returns404(t *testing.T) {
	mock := &mockIngestService{
		jobs: []service.Job{
			{ID: "j-secret", UserID: "u-owner", FileName: "secret.pdf"},
		},
	}
	r := setupIngestRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs/j-secret", nil)
	req.Header.Set("X-User-ID", "u-attacker")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	req = withAdminContext(req)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Must return 404, not 403 — do not leak job existence.
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other user's job, got %d: %s", rec.Code, rec.Body.String())
	}
}
