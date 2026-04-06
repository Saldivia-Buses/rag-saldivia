package handler

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	sdajwt "github.com/Camionerou/rag-saldivia/pkg/jwt"
	"github.com/Camionerou/rag-saldivia/services/platform/db"
	"github.com/Camionerou/rag-saldivia/services/platform/internal/service"
)

var testPub ed25519.PublicKey
var testPriv ed25519.PrivateKey

func init() {
	testPub, testPriv, _ = ed25519.GenerateKey(nil)
}

// --- mock ---

type mockPlatformService struct {
	tenants []db.ListTenantsRow
	tenant  service.TenantDetail
	modules []db.Module
	flags   []service.FeatureFlag
	config  service.ConfigEntry
	err     error
}

func (m *mockPlatformService) ListTenants(_ context.Context, _, _ int32) ([]db.ListTenantsRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tenants, nil
}

func (m *mockPlatformService) GetTenant(_ context.Context, slug string) (service.TenantDetail, error) {
	if m.err != nil {
		return service.TenantDetail{}, m.err
	}
	return m.tenant, nil
}

func (m *mockPlatformService) CreateTenant(_ context.Context, arg db.CreateTenantParams) (service.TenantDetail, error) {
	if m.err != nil {
		return service.TenantDetail{}, m.err
	}
	return service.TenantDetail{ID: "t-new", Slug: arg.Slug, Name: arg.Name}, nil
}

func (m *mockPlatformService) UpdateTenant(_ context.Context, arg db.UpdateTenantParams) error {
	return m.err
}

func (m *mockPlatformService) DisableTenant(_ context.Context, id string) error {
	return m.err
}

func (m *mockPlatformService) EnableTenant(_ context.Context, id string) error {
	return m.err
}

func (m *mockPlatformService) ListModules(_ context.Context) ([]db.Module, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.modules, nil
}

func (m *mockPlatformService) GetTenantModules(_ context.Context, tenantID string) ([]db.GetEnabledModulesForTenantRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func (m *mockPlatformService) EnableModule(_ context.Context, arg db.EnableModuleForTenantParams) error {
	return m.err
}

func (m *mockPlatformService) DisableModule(_ context.Context, tenantID, moduleID string) error {
	return m.err
}

func (m *mockPlatformService) ListFeatureFlags(_ context.Context) ([]service.FeatureFlag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.flags, nil
}

func (m *mockPlatformService) ToggleFeatureFlag(_ context.Context, id string, enabled bool) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockPlatformService) GetConfig(_ context.Context, key string) (service.ConfigEntry, error) {
	if m.err != nil {
		return service.ConfigEntry{}, m.err
	}
	return m.config, nil
}

func (m *mockPlatformService) SetConfig(_ context.Context, key string, value []byte, updatedBy string) error {
	return m.err
}

// --- helpers ---

func adminToken(t *testing.T) string {
	t.Helper()
	cfg := sdajwt.DefaultConfig(testPriv, testPub)
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID: "u-admin", Email: "admin@sda.app", TenantID: "platform",
		Slug: "platform", Role: "admin",
	})
	if err != nil {
		t.Fatalf("create admin token: %v", err)
	}
	return token
}

func userToken(t *testing.T) string {
	t.Helper()
	cfg := sdajwt.DefaultConfig(testPriv, testPub)
	token, err := sdajwt.CreateAccess(cfg, sdajwt.Claims{
		UserID: "u-user", Email: "user@tenant.com", TenantID: "t-1",
		Slug: "saldivia", Role: "user",
	})
	if err != nil {
		t.Fatalf("create user token: %v", err)
	}
	return token
}

func setupPlatformRouter(mock *mockPlatformService) *chi.Mux {
	h := NewPlatform(mock, testPub, "platform")
	r := chi.NewRouter()
	r.Mount("/v1/platform", h.Routes())
	return r
}

func withAdminAuth(req *http.Request, t *testing.T) *http.Request {
	t.Helper()
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	return req
}

// --- auth middleware tests ---

func TestPlatformAdmin_NoToken_Returns401(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/tenants", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPlatformAdmin_InvalidToken_Returns401(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/tenants", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.here")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPlatformAdmin_UserRole_Returns403(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", rec.Code)
	}
}

func TestPlatformAdmin_AdminRole_Allowed(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/tenants", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- tenant CRUD tests ---

func TestCreateTenant_Success(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"slug":"acme","name":"Acme Corp","plan_id":"p-1","postgres_url":"postgres://...","redis_url":"redis://..."}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTenant_MissingFields_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"slug":"acme"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing fields, got %d", rec.Code)
	}
}

func TestCreateTenant_InvalidSlug_Returns400(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrInvalidSlug}
	r := setupPlatformRouter(mock)

	body := `{"slug":"BAD SLUG!","name":"Test","plan_id":"p-1","postgres_url":"x","redis_url":"x"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid slug, got %d", rec.Code)
	}
}

func TestCreateTenant_DuplicateSlug_Returns409(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrSlugTaken}
	r := setupPlatformRouter(mock)

	body := `{"slug":"saldivia","name":"Dup","plan_id":"p-1","postgres_url":"x","redis_url":"x"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate slug, got %d", rec.Code)
	}
}

func TestGetTenant_NotFound_Returns404(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrTenantNotFound}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/tenants/nonexistent", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// --- feature flags ---

func TestToggleFlag_NotFound_Returns404(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrFlagNotFound}
	r := setupPlatformRouter(mock)

	body := `{"enabled":true}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPatch, "/v1/platform/flags/nonexistent", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// --- config ---

func TestGetConfig_NotFound_Returns404(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrConfigNotFound}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/config/missing-key", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// --- 500 error ---

func TestListTenants_ServiceError_Returns500(t *testing.T) {
	mock := &mockPlatformService{err: errors.New("database exploded")}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/tenants", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}
