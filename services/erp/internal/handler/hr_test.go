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

type mockHRService struct {
	departments []repository.ErpDepartment
	employees   []repository.ListEmployeeDetailsRow
	employee    repository.GetEmployeeDetailRow
	upserted    repository.UpsertEmployeeDetailRow
	events      []repository.ErpHrEvent
	event       repository.ErpHrEvent
	training    []repository.ErpTraining
	trainingRow repository.ErpTraining
	attendance  []repository.ErpAttendance
	attendRow   repository.ErpAttendance
	err         error
}

func (m *mockHRService) ListDepartments(_ context.Context, _ string) ([]repository.ErpDepartment, error) {
	return m.departments, m.err
}

func (m *mockHRService) CreateDepartment(_ context.Context, _ repository.CreateDepartmentParams, _, _ string) (repository.ErpDepartment, error) {
	if m.err != nil {
		return repository.ErpDepartment{}, m.err
	}
	if len(m.departments) > 0 {
		return m.departments[0], nil
	}
	return repository.ErpDepartment{}, nil
}

func (m *mockHRService) ListEmployees(_ context.Context, _ string, _, _ int) ([]repository.ListEmployeeDetailsRow, error) {
	return m.employees, m.err
}

func (m *mockHRService) GetEmployee(_ context.Context, _ pgtype.UUID, _, _, _ string) (repository.GetEmployeeDetailRow, error) {
	if m.err != nil {
		return repository.GetEmployeeDetailRow{}, m.err
	}
	return m.employee, nil
}

func (m *mockHRService) UpsertEmployee(_ context.Context, _ repository.UpsertEmployeeDetailParams, _, _ string) (repository.UpsertEmployeeDetailRow, error) {
	if m.err != nil {
		return repository.UpsertEmployeeDetailRow{}, m.err
	}
	return m.upserted, nil
}

func (m *mockHRService) ListEvents(_ context.Context, _ string, _ pgtype.UUID, _ string, _, _ int) ([]repository.ErpHrEvent, error) {
	return m.events, m.err
}

func (m *mockHRService) CreateEvent(_ context.Context, _ repository.CreateHREventParams, _ string) (repository.ErpHrEvent, error) {
	if m.err != nil {
		return repository.ErpHrEvent{}, m.err
	}
	return m.event, nil
}

func (m *mockHRService) ListTraining(_ context.Context, _ string, _, _ int) ([]repository.ErpTraining, error) {
	return m.training, m.err
}

func (m *mockHRService) CreateTraining(_ context.Context, _ repository.CreateTrainingParams, _, _ string) (repository.ErpTraining, error) {
	if m.err != nil {
		return repository.ErpTraining{}, m.err
	}
	return m.trainingRow, nil
}

func (m *mockHRService) ListAttendance(_ context.Context, _ string, _ pgtype.UUID, _, _ pgtype.Date, _, _ int) ([]repository.ErpAttendance, error) {
	return m.attendance, m.err
}

func (m *mockHRService) CreateAttendance(_ context.Context, _ repository.CreateAttendanceParams, _, _ string) (repository.ErpAttendance, error) {
	if m.err != nil {
		return repository.ErpAttendance{}, m.err
	}
	return m.attendRow, nil
}

// --- helpers ---

func setupHRRouter(mock *mockHRService) *chi.Mux {
	h := NewHR(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/hr", h.Routes(noopMiddleware))
	return r
}

// --- tests ---

func TestHR_ListEmployees_Success(t *testing.T) {
	mock := &mockHRService{
		employees: []repository.ListEmployeeDetailsRow{
			{Position: "Driver"},
			{Position: "Mechanic"},
		},
	}
	r := setupHRRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHR_ListEmployees_EmptySlice_Returns200(t *testing.T) {
	mock := &mockHRService{employees: []repository.ListEmployeeDetailsRow{}}
	r := setupHRRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHR_ListEmployees_ServiceError_Returns500(t *testing.T) {
	mock := &mockHRService{err: errors.New("db down")}
	r := setupHRRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestHR_GetEmployee_InvalidUUID_Returns400(t *testing.T) {
	r := setupHRRouter(&mockHRService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHR_GetEmployee_Success(t *testing.T) {
	mock := &mockHRService{employee: repository.GetEmployeeDetailRow{Position: "Driver"}}
	r := setupHRRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHR_GetEmployee_NotFound_Returns404(t *testing.T) {
	mock := &mockHRService{err: errors.New("not found")}
	r := setupHRRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHR_UpsertEmployee_InvalidJSON_Returns400(t *testing.T) {
	r := setupHRRouter(&mockHRService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/hr/employees", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHR_UpsertEmployee_Success(t *testing.T) {
	mock := &mockHRService{upserted: repository.UpsertEmployeeDetailRow{Position: "Driver"}}
	r := setupHRRouter(mock)

	body := `{"entity_id":"00000000-0000-0000-0000-000000000001","position":"Driver","schedule_type":"full_time"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/hr/employees", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHR_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupHRRouter(&mockHRService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/hr/employees", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
