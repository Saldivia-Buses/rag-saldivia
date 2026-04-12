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

type mockStockService struct {
	articles   []repository.ErpArticle
	article    repository.ErpArticle
	warehouses []repository.ErpWarehouse
	levels     []repository.GetStockLevelsRow
	movements  []repository.ListStockMovementsRow
	movement   repository.ErpStockMovement
	bom        []repository.ListBOMRow
	err        error
}

func (m *mockStockService) ListArticles(_ context.Context, _, _, _ string, _ bool, _, _ int) ([]repository.ErpArticle, error) {
	return m.articles, m.err
}

func (m *mockStockService) GetArticle(_ context.Context, _ pgtype.UUID, _ string) (repository.ErpArticle, error) {
	if m.err != nil {
		return repository.ErpArticle{}, m.err
	}
	return m.article, nil
}

func (m *mockStockService) CreateArticle(_ context.Context, _ service.CreateArticleRequest) (repository.ErpArticle, error) {
	if m.err != nil {
		return repository.ErpArticle{}, m.err
	}
	return m.article, nil
}

func (m *mockStockService) ListWarehouses(_ context.Context, _ string, _ bool) ([]repository.ErpWarehouse, error) {
	return m.warehouses, m.err
}

func (m *mockStockService) CreateWarehouse(_ context.Context, _, _, _, _, _, _ string) (repository.ErpWarehouse, error) {
	if m.err != nil {
		return repository.ErpWarehouse{}, m.err
	}
	if len(m.warehouses) > 0 {
		return m.warehouses[0], nil
	}
	return repository.ErpWarehouse{}, nil
}

func (m *mockStockService) GetStockLevels(_ context.Context, _ string, _, _ pgtype.UUID) ([]repository.GetStockLevelsRow, error) {
	return m.levels, m.err
}

func (m *mockStockService) ListMovements(_ context.Context, _ string, _ pgtype.UUID, _, _ int) ([]repository.ListStockMovementsRow, error) {
	return m.movements, m.err
}

func (m *mockStockService) CreateMovement(_ context.Context, _ service.CreateMovementRequest) (repository.ErpStockMovement, error) {
	if m.err != nil {
		return repository.ErpStockMovement{}, m.err
	}
	return m.movement, nil
}

func (m *mockStockService) ListBOM(_ context.Context, _ string, _ pgtype.UUID) ([]repository.ListBOMRow, error) {
	return m.bom, m.err
}

// --- helpers ---

func setupStockRouter(mock StockService) *chi.Mux {
	h := NewStock(mock)
	r := chi.NewRouter()
	noopMiddleware := func(next http.Handler) http.Handler { return next }
	r.Mount("/v1/erp/stock", h.Routes(noopMiddleware))
	return r
}

func decodeStockJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- tests ---

func TestStock_ListArticles_Success(t *testing.T) {
	mock := &mockStockService{
		articles: []repository.ErpArticle{
			{Code: "ART-001", Name: "Tornillo"},
			{Code: "ART-002", Name: "Tuerca"},
		},
	}
	r := setupStockRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/stock/articles", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	decodeStockJSON(t, rec, &resp)
	articles, ok := resp["articles"].([]any)
	if !ok || len(articles) != 2 {
		t.Errorf("expected 2 articles in response, got %v", resp["articles"])
	}
}

func TestStock_ListArticles_ServiceError_Returns500(t *testing.T) {
	mock := &mockStockService{err: errors.New("db error")}
	r := setupStockRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/stock/articles", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	decodeStockJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error message, got %q", resp["error"])
	}
}

func TestStock_GetArticle_InvalidUUID_Returns400(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/stock/articles/not-a-uuid", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", rec.Code)
	}
}

func TestStock_GetArticle_NotFound_Returns404(t *testing.T) {
	mock := &mockStockService{err: errors.New("not found")}
	r := setupStockRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/stock/articles/00000000-0000-0000-0000-000000000001", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestStock_CreateArticle_InvalidBody_Returns400(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/articles", strings.NewReader("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestStock_CreateArticle_Success(t *testing.T) {
	mock := &mockStockService{article: repository.ErpArticle{Code: "ART-001", Name: "Tornillo"}}
	r := setupStockRouter(mock)

	body := `{"code":"ART-001","name":"Tornillo","article_type":"product","min_stock":"10","max_stock":"100","reorder_point":"20"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/articles", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStock_CreateArticle_ServiceError_Returns500(t *testing.T) {
	mock := &mockStockService{err: errors.New("duplicate code")}
	r := setupStockRouter(mock)

	body := `{"code":"ART-001","name":"Tornillo","article_type":"product"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/articles", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestStock_CreateMovement_InvalidBody_Returns400(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/movements", strings.NewReader("bad json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", rec.Code)
	}
}

func TestStock_CreateMovement_InvalidArticleID_Returns400(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	body := `{"article_id":"not-a-uuid","warehouse_id":"00000000-0000-0000-0000-000000000001","movement_type":"in","quantity":"5"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/movements", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid article_id, got %d", rec.Code)
	}
}

func TestStock_CreateMovement_InvalidWarehouseID_Returns400(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	body := `{"article_id":"00000000-0000-0000-0000-000000000001","warehouse_id":"not-a-uuid","movement_type":"in","quantity":"5"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/movements", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid warehouse_id, got %d", rec.Code)
	}
}

func TestStock_CreateMovement_Success(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	body := `{"article_id":"00000000-0000-0000-0000-000000000001","warehouse_id":"00000000-0000-0000-0000-000000000002","movement_type":"in","quantity":"10","unit_cost":"5.50"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/movements", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStock_CreateMovement_ServiceError_Returns500(t *testing.T) {
	mock := &mockStockService{err: errors.New("invalid movement type")}
	r := setupStockRouter(mock)

	body := `{"article_id":"00000000-0000-0000-0000-000000000001","warehouse_id":"00000000-0000-0000-0000-000000000002","movement_type":"bad","quantity":"10"}`
	req := withAdmin(httptest.NewRequest(http.MethodPost, "/v1/erp/stock/movements", strings.NewReader(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestStock_ListMovements_Success(t *testing.T) {
	mock := &mockStockService{
		movements: []repository.ListStockMovementsRow{
			{MovementType: "in"},
		},
	}
	r := setupStockRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/stock/movements", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestStock_ListMovements_ServiceError_Returns500(t *testing.T) {
	mock := &mockStockService{err: errors.New("db error")}
	r := setupStockRouter(mock)

	req := withAdmin(httptest.NewRequest(http.MethodGet, "/v1/erp/stock/movements", nil))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestStock_RequirePermission_WithoutRole_Returns403(t *testing.T) {
	r := setupStockRouter(&mockStockService{})

	// No role injected → RequirePermission blocks with 403
	req := httptest.NewRequest(http.MethodGet, "/v1/erp/stock/articles", nil)
	req.Header.Set("X-Tenant-Slug", "test-tenant")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without auth, got %d", rec.Code)
	}
}
