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
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// --- mock ---

type mockSuggestionsService struct {
	suggestions []repository.ListSuggestionsRow
	suggestion  repository.ErpSuggestion
	responses   []repository.ErpSuggestionResponse
	response    repository.ErpSuggestionResponse
	unread      int32
	err         error
}

func (m *mockSuggestionsService) List(_ context.Context, _ string, _, _ int) ([]repository.ListSuggestionsRow, error) {
	return m.suggestions, m.err
}

func (m *mockSuggestionsService) Get(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpSuggestion, []repository.ErpSuggestionResponse, error) {
	if m.err != nil {
		return repository.ErpSuggestion{}, nil, m.err
	}
	return m.suggestion, m.responses, nil
}

func (m *mockSuggestionsService) Create(_ context.Context, _ service.CreateRequest) (repository.ErpSuggestion, error) {
	if m.err != nil {
		return repository.ErpSuggestion{}, m.err
	}
	return m.suggestion, nil
}

func (m *mockSuggestionsService) Respond(_ context.Context, _ service.RespondRequest) (repository.ErpSuggestionResponse, error) {
	if m.err != nil {
		return repository.ErpSuggestionResponse{}, m.err
	}
	return m.response, nil
}

func (m *mockSuggestionsService) MarkRead(_ context.Context, _ pgtype.UUID, _ string) error {
	return m.err
}

func (m *mockSuggestionsService) CountUnread(_ context.Context, _ string) (int32, error) {
	return m.unread, m.err
}

// --- helpers ---

func setupSuggestionsRouter(mock *mockSuggestionsService) *chi.Mux {
	h := NewSuggestions(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/suggestions", h.Routes(noopMiddleware))
	return r
}

// --- tests ---

func TestSuggestions_List_Success(t *testing.T) {
	mock := &mockSuggestionsService{
		suggestions: []repository.ListSuggestionsRow{
			{Body: "Improve shift scheduling"},
		},
	}
	r := setupSuggestionsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSuggestions_List_EmptySlice_Returns200(t *testing.T) {
	mock := &mockSuggestionsService{suggestions: []repository.ListSuggestionsRow{}}
	r := setupSuggestionsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSuggestions_List_ServiceError_Returns500(t *testing.T) {
	mock := &mockSuggestionsService{err: errors.New("db down")}
	r := setupSuggestionsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSuggestions_Create_InvalidJSON_Returns400(t *testing.T) {
	r := setupSuggestionsRouter(&mockSuggestionsService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/suggestions/", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSuggestions_Create_Success(t *testing.T) {
	mock := &mockSuggestionsService{
		suggestion: repository.ErpSuggestion{Body: "Better scheduling"},
	}
	r := setupSuggestionsRouter(mock)

	body := `{"origin":"web","body":"Better scheduling"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/suggestions/", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSuggestions_Create_ServiceError_Returns500(t *testing.T) {
	mock := &mockSuggestionsService{err: errors.New("insert failed")}
	r := setupSuggestionsRouter(mock)

	body := `{"origin":"web","body":"Better scheduling"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/suggestions/", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestSuggestions_Get_InvalidUUID_Returns400(t *testing.T) {
	r := setupSuggestionsRouter(&mockSuggestionsService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestSuggestions_Get_NotFound_Returns404(t *testing.T) {
	mock := &mockSuggestionsService{err: errors.New("not found")}
	r := setupSuggestionsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestSuggestions_Get_Success(t *testing.T) {
	mock := &mockSuggestionsService{
		suggestion: repository.ErpSuggestion{Body: "Better scheduling"},
		responses:  []repository.ErpSuggestionResponse{},
	}
	r := setupSuggestionsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSuggestions_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupSuggestionsRouter(&mockSuggestionsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/suggestions/", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
