package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// --- mock ---

type mockQualityService struct {
	ncs       []repository.ListNonconformitiesRow
	nc        repository.ErpNonconformity
	cas       []repository.ErpCorrectiveAction
	ca        repository.ErpCorrectiveAction
	audits    []repository.ErpAudit
	audit     repository.ErpAudit
	findings  []repository.ErpAuditFinding
	finding   repository.ErpAuditFinding
	documents []repository.ErpControlledDocument
	document  repository.ErpControlledDocument
	err       error
}

func (m *mockQualityService) ListNC(_ context.Context, _, _, _ string, _, _ int) ([]repository.ListNonconformitiesRow, error) {
	return m.ncs, m.err
}

func (m *mockQualityService) GetNC(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpNonconformity, error) {
	if m.err != nil {
		return repository.ErpNonconformity{}, m.err
	}
	return m.nc, nil
}

func (m *mockQualityService) CreateNC(_ context.Context, _ repository.CreateNonconformityParams, _ string) (repository.ErpNonconformity, error) {
	if m.err != nil {
		return repository.ErpNonconformity{}, m.err
	}
	return m.nc, nil
}

func (m *mockQualityService) UpdateNCStatus(_ context.Context, _ pgtype.UUID, _, _, _, _ string) error {
	return m.err
}

func (m *mockQualityService) ListCA(_ context.Context, _ pgtype.UUID, _ string) ([]repository.ErpCorrectiveAction, error) {
	return m.cas, m.err
}

func (m *mockQualityService) CreateCA(_ context.Context, _ repository.CreateCorrectiveActionParams, _, _ string) (repository.ErpCorrectiveAction, error) {
	if m.err != nil {
		return repository.ErpCorrectiveAction{}, m.err
	}
	return m.ca, nil
}

func (m *mockQualityService) ListAudits(_ context.Context, _ string, _, _ int) ([]repository.ErpAudit, error) {
	return m.audits, m.err
}

func (m *mockQualityService) CreateAudit(_ context.Context, _ repository.CreateAuditParams, _, _ string) (repository.ErpAudit, error) {
	if m.err != nil {
		return repository.ErpAudit{}, m.err
	}
	return m.audit, nil
}

func (m *mockQualityService) ListAuditFindings(_ context.Context, _ pgtype.UUID, _ string) ([]repository.ErpAuditFinding, error) {
	return m.findings, m.err
}

func (m *mockQualityService) CreateAuditFinding(_ context.Context, _ repository.CreateAuditFindingParams, _, _ string) (repository.ErpAuditFinding, error) {
	if m.err != nil {
		return repository.ErpAuditFinding{}, m.err
	}
	return m.finding, nil
}

func (m *mockQualityService) ListDocuments(_ context.Context, _, _ string, _, _ int) ([]repository.ErpControlledDocument, error) {
	return m.documents, m.err
}

func (m *mockQualityService) CreateDocument(_ context.Context, _ repository.CreateControlledDocumentParams, _, _ string) (repository.ErpControlledDocument, error) {
	if m.err != nil {
		return repository.ErpControlledDocument{}, m.err
	}
	return m.document, nil
}

// --- helpers ---

func setupQualityRouter(mock *mockQualityService) *chi.Mux {
	h := NewQuality(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/quality", h.Routes(noopMiddleware))
	return r
}

// --- tests ---

func TestQuality_ListNC_Success(t *testing.T) {
	mock := &mockQualityService{
		ncs: []repository.ListNonconformitiesRow{
			{Number: "NC-001"},
		},
	}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/nc", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_ListNC_ServiceError_Returns500(t *testing.T) {
	mock := &mockQualityService{err: errors.New("db error")}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/nc", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestQuality_CreateNC_InvalidBody_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/nc", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestQuality_CreateNC_Success(t *testing.T) {
	mock := &mockQualityService{nc: repository.ErpNonconformity{Number: "NC-001"}}
	r := setupQualityRouter(mock)

	body := `{"number":"NC-001","description":"Defective weld","severity":"minor","date":"2025-01-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/nc", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CreateNC_InvalidSeverity_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"number":"NC-001","description":"Defective weld","severity":"bad_value","date":"2025-01-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/nc", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid severity, got %d", rec.Code)
	}
}

func TestQuality_UpdateNCStatus_InvalidUUID_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"status":"closed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/quality/nc/bad-uuid/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestQuality_UpdateNCStatus_Success(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"status":"closed"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/quality/nc/00000000-0000-0000-0000-000000000001/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_ListAudits_Success(t *testing.T) {
	mock := &mockQualityService{
		audits: []repository.ErpAudit{{Number: "AUD-001"}},
	}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/audits", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_ListDocuments_WithStatus_Returns200(t *testing.T) {
	mock := &mockQualityService{
		documents: []repository.ErpControlledDocument{{Code: "DOC-001"}},
	}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/documents?status=active", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/quality/nc", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
