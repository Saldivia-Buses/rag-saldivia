package handler

import (
	"context"
	"encoding/json"
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

type mockAdminService struct {
	communications []repository.ErpCommunication
	communication  repository.ErpCommunication
	events         []repository.ErpCalendarEvent
	event          repository.ErpCalendarEvent
	surveys        []repository.ErpSurvey
	survey         repository.ErpSurvey
	err            error
}

func (m *mockAdminService) ListCommunications(_ context.Context, _ string, _, _ int) ([]repository.ErpCommunication, error) {
	return m.communications, m.err
}

func (m *mockAdminService) CreateCommunication(_ context.Context, _, _, _, _, _, _ string) (repository.ErpCommunication, error) {
	if m.err != nil {
		return repository.ErpCommunication{}, m.err
	}
	return m.communication, nil
}

func (m *mockAdminService) ListCalendarEvents(_ context.Context, _ string, _, _ pgtype.Timestamptz) ([]repository.ErpCalendarEvent, error) {
	return m.events, m.err
}

func (m *mockAdminService) CreateCalendarEvent(_ context.Context, _ repository.CreateCalendarEventParams, _ string) (repository.ErpCalendarEvent, error) {
	if m.err != nil {
		return repository.ErpCalendarEvent{}, m.err
	}
	return m.event, nil
}

func (m *mockAdminService) ListSurveys(_ context.Context, _ string, _, _ int) ([]repository.ErpSurvey, error) {
	return m.surveys, m.err
}

func (m *mockAdminService) CreateSurvey(_ context.Context, _, _, _, _, _ string) (repository.ErpSurvey, error) {
	if m.err != nil {
		return repository.ErpSurvey{}, m.err
	}
	return m.survey, nil
}

func (m *mockAdminService) ListProductSections(_ context.Context, _ string) ([]repository.ErpProductSection, error) {
	return nil, m.err
}
func (m *mockAdminService) ListProducts(_ context.Context, _ string, _, _ int) ([]repository.ErpProduct, error) {
	return nil, m.err
}
func (m *mockAdminService) ListProductAttributes(_ context.Context, _ string, _ bool) ([]repository.ErpProductAttribute, error) {
	return nil, m.err
}
func (m *mockAdminService) ListTools(_ context.Context, _ string, _ int32, _ string, _, _ int) ([]repository.ErpTool, error) {
	return nil, m.err
}

// --- helpers ---

func setupAdminRouter(mock AdminService) *chi.Mux {
	h := NewAdmin(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/admin", h.Routes(noopMiddleware))
	return r
}

func decodeAdminJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests ---

func TestAdmin_ListCommunications_Success(t *testing.T) {
	mock := &mockAdminService{
		communications: []repository.ErpCommunication{
			{Subject: "Aviso de cierre", Priority: "normal"},
		},
	}
	r := setupAdminRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/admin/communications", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeAdminJSON(t, rec, &resp)
	comms, ok := resp["communications"].([]any)
	if !ok || len(comms) != 1 {
		t.Errorf("expected 1 communication in response, got %v", resp["communications"])
	}
}

func TestAdmin_ListCommunications_ServiceError_Returns500(t *testing.T) {
	mock := &mockAdminService{err: errors.New("db error")}
	r := setupAdminRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/admin/communications", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	decodeAdminJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}

func TestAdmin_CreateCommunication_InvalidBody_Returns400(t *testing.T) {
	r := setupAdminRouter(&mockAdminService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/communications", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestAdmin_CreateCommunication_Success(t *testing.T) {
	mock := &mockAdminService{
		communication: repository.ErpCommunication{Subject: "Test", Priority: "normal"},
	}
	r := setupAdminRouter(mock)

	body := `{"Subject":"Test","Body":"Mensaje de prueba","Priority":"normal"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/communications", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdmin_CreateCommunication_ServiceError_Returns500(t *testing.T) {
	mock := &mockAdminService{err: errors.New("db error")}
	r := setupAdminRouter(mock)

	body := `{"Subject":"Test","Body":"Mensaje","Priority":"normal"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/communications", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAdmin_ListCalendarEvents_Success(t *testing.T) {
	mock := &mockAdminService{
		events: []repository.ErpCalendarEvent{
			{Title: "Reunión de equipo"},
		},
	}
	r := setupAdminRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/admin/calendar", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeAdminJSON(t, rec, &resp)
	events, ok := resp["events"].([]any)
	if !ok || len(events) != 1 {
		t.Errorf("expected 1 event in response, got %v", resp["events"])
	}
}

func TestAdmin_CreateCalendarEvent_InvalidBody_Returns400(t *testing.T) {
	r := setupAdminRouter(&mockAdminService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/calendar", strings.NewReader("bad json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestAdmin_CreateCalendarEvent_InvalidStartAt_Returns400(t *testing.T) {
	r := setupAdminRouter(&mockAdminService{})

	// start_at is required and must be a valid timestamp
	body := `{"Title":"Meeting","StartAt":"not-a-date"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/calendar", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid start_at, got %d", rec.Code)
	}
}

func TestAdmin_CreateCalendarEvent_Success(t *testing.T) {
	mock := &mockAdminService{
		event: repository.ErpCalendarEvent{Title: "Reunión"},
	}
	r := setupAdminRouter(mock)

	body := `{"Title":"Reunion","StartAt":"2025-06-01 10:00:00+00"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/calendar", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdmin_ListSurveys_Success(t *testing.T) {
	mock := &mockAdminService{
		surveys: []repository.ErpSurvey{
			{Title: "Encuesta de satisfacción"},
		},
	}
	r := setupAdminRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/admin/surveys", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeAdminJSON(t, rec, &resp)
	surveys, ok := resp["surveys"].([]any)
	if !ok || len(surveys) != 1 {
		t.Errorf("expected 1 survey in response, got %v", resp["surveys"])
	}
}

func TestAdmin_CreateSurvey_InvalidBody_Returns400(t *testing.T) {
	r := setupAdminRouter(&mockAdminService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/surveys", strings.NewReader("bad json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestAdmin_CreateSurvey_Success(t *testing.T) {
	mock := &mockAdminService{
		survey: repository.ErpSurvey{Title: "Encuesta"},
	}
	r := setupAdminRouter(mock)

	body := `{"Title":"Encuesta","Description":"Para medir satisfacción"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/admin/surveys", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdmin_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupAdminRouter(&mockAdminService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/admin/communications", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
