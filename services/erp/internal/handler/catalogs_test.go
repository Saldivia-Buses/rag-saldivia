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
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// --- mock ---

type mockCatalogsService struct {
	catalogs []repository.ErpCatalog
	catalog  repository.ErpCatalog
	types    []string
	err      error
}

func (m *mockCatalogsService) List(_ context.Context, _, _ string, _ bool) ([]repository.ErpCatalog, error) {
	return m.catalogs, m.err
}

func (m *mockCatalogsService) ListTypes(_ context.Context, _ string) ([]string, error) {
	return m.types, m.err
}

func (m *mockCatalogsService) Get(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpCatalog, error) {
	if m.err != nil {
		return repository.ErpCatalog{}, m.err
	}
	return m.catalog, nil
}

func (m *mockCatalogsService) Create(_ context.Context, _ service.CreateCatalogRequest) (repository.ErpCatalog, error) {
	if m.err != nil {
		return repository.ErpCatalog{}, m.err
	}
	return m.catalog, nil
}

func (m *mockCatalogsService) Update(_ context.Context, _ service.UpdateCatalogRequest) (repository.ErpCatalog, error) {
	if m.err != nil {
		return repository.ErpCatalog{}, m.err
	}
	return m.catalog, nil
}

func (m *mockCatalogsService) Delete(_ context.Context, _ pgtype.UUID, _, _, _ string) error {
	return m.err
}

// --- helpers ---

func setupCatalogsRouter(mock CatalogsService) *chi.Mux {
	h := NewCatalogs(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/catalogs", h.Routes(noopMiddleware))
	return r
}

func decodeCatalogsJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests ---

func TestCatalogs_List_MissingType_Returns400(t *testing.T) {
	r := setupCatalogsRouter(&mockCatalogsService{})

	// ?type is required — omitting it returns 400
	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 without type param, got %d", rec.Code)
	}
}

func TestCatalogs_List_Success(t *testing.T) {
	mock := &mockCatalogsService{
		catalogs: []repository.ErpCatalog{
			{Type: "unit", Code: "KG", Name: "Kilogramo"},
			{Type: "unit", Code: "L", Name: "Litro"},
		},
	}
	r := setupCatalogsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/?type=unit", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeCatalogsJSON(t, rec, &resp)
	catalogs, ok := resp["catalogs"].([]any)
	if !ok || len(catalogs) != 2 {
		t.Errorf("expected 2 catalogs in response, got %v", resp["catalogs"])
	}
}

func TestCatalogs_List_ServiceError_Returns500(t *testing.T) {
	mock := &mockCatalogsService{err: errors.New("db error")}
	r := setupCatalogsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/?type=unit", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	decodeCatalogsJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}

func TestCatalogs_ListTypes_Success(t *testing.T) {
	mock := &mockCatalogsService{types: []string{"unit", "currency", "payment_term"}}
	r := setupCatalogsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/types", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeCatalogsJSON(t, rec, &resp)
	types, ok := resp["types"].([]any)
	if !ok || len(types) != 3 {
		t.Errorf("expected 3 types in response, got %v", resp["types"])
	}
}

func TestCatalogs_Get_InvalidUUID_Returns400(t *testing.T) {
	r := setupCatalogsRouter(&mockCatalogsService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestCatalogs_Get_NotFound_Returns404(t *testing.T) {
	mock := &mockCatalogsService{err: errors.New("not found")}
	r := setupCatalogsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCatalogs_Get_Success(t *testing.T) {
	mock := &mockCatalogsService{
		catalog: repository.ErpCatalog{Type: "unit", Code: "KG", Name: "Kilogramo"},
	}
	r := setupCatalogsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCatalogs_Create_InvalidBody_Returns400(t *testing.T) {
	r := setupCatalogsRouter(&mockCatalogsService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/catalogs/", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestCatalogs_Create_Success(t *testing.T) {
	mock := &mockCatalogsService{
		catalog: repository.ErpCatalog{Type: "unit", Code: "KG", Name: "Kilogramo"},
	}
	r := setupCatalogsRouter(mock)

	body := `{"type":"unit","code":"KG","name":"Kilogramo","sort_order":1}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/catalogs/", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCatalogs_Create_ServiceError_Returns500(t *testing.T) {
	mock := &mockCatalogsService{err: errors.New("duplicate code")}
	r := setupCatalogsRouter(mock)

	body := `{"type":"unit","code":"KG","name":"Kilogramo"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/catalogs/", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestCatalogs_Delete_InvalidUUID_Returns400(t *testing.T) {
	r := setupCatalogsRouter(&mockCatalogsService{})

	req := withAdmin(httptest.NewRequest(http.MethodDelete, "/v1/erp/catalogs/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestCatalogs_Delete_Success(t *testing.T) {
	r := setupCatalogsRouter(&mockCatalogsService{})

	req := withAdmin(httptest.NewRequest(http.MethodDelete, "/v1/erp/catalogs/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCatalogs_Delete_NotFound_Returns404(t *testing.T) {
	mock := &mockCatalogsService{err: errors.New("catalog not found")}
	r := setupCatalogsRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodDelete, "/v1/erp/catalogs/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCatalogs_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupCatalogsRouter(&mockCatalogsService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/erp/catalogs/?type=unit", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
