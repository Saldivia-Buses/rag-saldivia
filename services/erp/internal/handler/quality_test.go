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
	ncs        []repository.ListNonconformitiesRow
	ncRow      repository.GetNonconformityRow
	ncCreated  repository.CreateNonconformityRow
	cas        []repository.ErpCorrectiveAction
	ca         repository.ErpCorrectiveAction
	audits     []repository.ErpAudit
	audit      repository.ErpAudit
	findings   []repository.ErpAuditFinding
	finding    repository.ErpAuditFinding
	documents  []repository.ErpControlledDocument
	document   repository.ErpControlledDocument
	ncOrigins  []repository.ErpNcOrigin
	ncOrigin   repository.ErpNcOrigin
	plans      []repository.ListActionPlansRow
	plan       repository.ErpQualityActionPlan
	tasks      []repository.ListActionTasksRow
	task       repository.ErpQualityActionTask
	err        error
}

func (m *mockQualityService) ListNC(_ context.Context, _, _, _ string, _, _ int) ([]repository.ListNonconformitiesRow, error) {
	return m.ncs, m.err
}

func (m *mockQualityService) GetNC(_ context.Context, _ pgtype.UUID, _ string) (repository.GetNonconformityRow, error) {
	if m.err != nil {
		return repository.GetNonconformityRow{}, m.err
	}
	return m.ncRow, nil
}

func (m *mockQualityService) CreateNC(_ context.Context, _ repository.CreateNonconformityParams, _ string) (repository.CreateNonconformityRow, error) {
	if m.err != nil {
		return repository.CreateNonconformityRow{}, m.err
	}
	return m.ncCreated, nil
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

func (m *mockQualityService) ListNCOrigins(_ context.Context, _ string) ([]repository.ErpNcOrigin, error) {
	return m.ncOrigins, m.err
}

func (m *mockQualityService) CreateNCOrigin(_ context.Context, _, _, _, _ string) (repository.ErpNcOrigin, error) {
	if m.err != nil {
		return repository.ErpNcOrigin{}, m.err
	}
	return m.ncOrigin, nil
}

func (m *mockQualityService) ListActionPlans(_ context.Context, _, _, _ string, _, _ int) ([]repository.ListActionPlansRow, error) {
	return m.plans, m.err
}

func (m *mockQualityService) GetActionPlan(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpQualityActionPlan, error) {
	if m.err != nil {
		return repository.ErpQualityActionPlan{}, m.err
	}
	return m.plan, nil
}

func (m *mockQualityService) CreateActionPlan(_ context.Context, _ repository.CreateActionPlanParams, _, _ string) (repository.ErpQualityActionPlan, error) {
	if m.err != nil {
		return repository.ErpQualityActionPlan{}, m.err
	}
	return m.plan, nil
}

func (m *mockQualityService) UpdateActionPlanStatus(_ context.Context, _ pgtype.UUID, _, _, _, _ string) error {
	return m.err
}

func (m *mockQualityService) ListActionTasks(_ context.Context, _ string, _ pgtype.UUID) ([]repository.ListActionTasksRow, error) {
	return m.tasks, m.err
}

func (m *mockQualityService) CreateActionTask(_ context.Context, _ repository.CreateActionTaskParams, _, _ string) (repository.ErpQualityActionTask, error) {
	if m.err != nil {
		return repository.ErpQualityActionTask{}, m.err
	}
	return m.task, nil
}

func (m *mockQualityService) ListIndicators(_ context.Context, _, _, _ string) ([]repository.ErpQualityIndicator, error) {
	return nil, m.err
}

func (m *mockQualityService) ListSupplierScorecards(_ context.Context, _ string, _, _ int) ([]repository.ListSupplierScorecardsRow, error) {
	return nil, m.err
}

func (m *mockQualityService) CompleteActionTask(_ context.Context, _ pgtype.UUID, _, _, _ string) error {
	return m.err
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
	mock := &mockQualityService{ncCreated: repository.CreateNonconformityRow{Number: "NC-001"}}
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

// ─── NC Origins ─────────────────────────────────────────────────────────────

func TestQuality_ListNCOrigins_Success(t *testing.T) {
	mock := &mockQualityService{
		ncOrigins: []repository.ErpNcOrigin{{Name: "Internal"}},
	}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/nc-origins", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CreateNCOrigin_Success(t *testing.T) {
	mock := &mockQualityService{ncOrigin: repository.ErpNcOrigin{Name: "Supplier"}}
	r := setupQualityRouter(mock)

	body := `{"name":"Supplier"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/nc-origins", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CreateNCOrigin_MissingName_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"name":""}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/nc-origins", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ─── Action Plans ────────────────────────────────────────────────────────────

func TestQuality_ListActionPlans_Success(t *testing.T) {
	mock := &mockQualityService{
		plans: []repository.ListActionPlansRow{{Description: "Fix welding"}},
	}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/action-plans", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_GetActionPlan_InvalidUUID_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/action-plans/bad-id", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestQuality_GetActionPlan_NotFound_Returns404(t *testing.T) {
	mock := &mockQualityService{err: errors.New("not found")}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestQuality_CreateActionPlan_Success(t *testing.T) {
	mock := &mockQualityService{plan: repository.ErpQualityActionPlan{Description: "Fix welding"}}
	r := setupQualityRouter(mock)

	body := `{"description":"Fix welding","planned_start":"2025-01-01","target_date":"2025-03-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/action-plans", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CreateActionPlan_MissingDescription_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"planned_start":"2025-01-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/action-plans", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestQuality_UpdateActionPlanStatus_Success(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"status":"active"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_UpdateActionPlanStatus_InvalidStatus_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"status":"unknown"}`
	req := withAdmin(httptest.NewRequest(http.MethodPatch, "/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/status", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ─── Action Tasks ─────────────────────────────────────────────────────────────

func TestQuality_ListActionTasks_Success(t *testing.T) {
	mock := &mockQualityService{
		tasks: []repository.ListActionTasksRow{{Description: "Inspect welds"}},
	}
	r := setupQualityRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/tasks", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CreateActionTask_Success(t *testing.T) {
	mock := &mockQualityService{task: repository.ErpQualityActionTask{Description: "Inspect welds"}}
	r := setupQualityRouter(mock)

	body := `{"description":"Inspect welds","target_date":"2025-03-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/tasks", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CreateActionTask_MissingDescription_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	body := `{"target_date":"2025-03-01"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/tasks", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestQuality_CompleteActionTask_Success(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	req := withAdmin(httptest.NewRequest(http.MethodPatch,
		"/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/tasks/00000000-0000-0000-0000-000000000002/complete", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuality_CompleteActionTask_InvalidTaskUUID_Returns400(t *testing.T) {
	r := setupQualityRouter(&mockQualityService{})

	req := withAdmin(httptest.NewRequest(http.MethodPatch,
		"/v1/erp/quality/action-plans/00000000-0000-0000-0000-000000000001/tasks/bad-task-id/complete", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
