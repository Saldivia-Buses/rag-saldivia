package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/ingest/internal/service"
)

// --- mock ---

type mockIngestService struct {
	jobs []service.Job
	err  error
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
	part.Write([]byte("fake pdf content"))
	writer.WriteField("collection", collection)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", userID)
	req.Header.Set("X-Tenant-Slug", tenantSlug)
	return req
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
	json.NewDecoder(rec.Body).Decode(&job)
	if job.Status != "pending" {
		t.Errorf("expected status pending, got %q", job.Status)
	}
}

func TestUpload_MissingUserID_Returns401(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "doc.pdf")
	part.Write([]byte("content"))
	writer.WriteField("collection", "test")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// No X-User-ID or X-Tenant-Slug
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
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
	part.Write([]byte("content"))
	// No collection field
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/ingest/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
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
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string][]service.Job
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp["jobs"]) != 1 {
		t.Errorf("expected 1 job for u-1, got %d", len(resp["jobs"]))
	}
}

func TestListJobs_MissingIdentity_Returns401(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/ingest/jobs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
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
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q — service internals may be leaking", resp["error"])
	}
}

func TestDeleteJob_NotFound_Returns404(t *testing.T) {
	r := setupIngestRouter(&mockIngestService{})

	req := httptest.NewRequest(http.MethodDelete, "/v1/ingest/jobs/nonexistent", nil)
	req.Header.Set("X-User-ID", "u-1")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
