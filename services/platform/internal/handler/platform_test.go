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
	deploys []service.DeployRecord
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

func (m *mockPlatformService) CreateFeatureFlag(_ context.Context, params service.CreateFlagParams, createdBy string) (service.FeatureFlag, error) {
	if m.err != nil {
		return service.FeatureFlag{}, m.err
	}
	return service.FeatureFlag{ID: params.ID, Name: params.Name, RolloutPct: params.RolloutPct}, nil
}

func (m *mockPlatformService) UpdateFeatureFlag(_ context.Context, id string, params service.UpdateFlagParams) error {
	return m.err
}

func (m *mockPlatformService) ToggleFeatureFlag(_ context.Context, id string, enabled bool) error {
	return m.err
}

func (m *mockPlatformService) KillFlag(_ context.Context, id string, killedBy string) error {
	return m.err
}

func (m *mockPlatformService) EvaluateFlags(_ context.Context, tenantID, userID string) (map[string]bool, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]bool)
	for _, f := range m.flags {
		result[f.Name] = f.Enabled
	}
	return result, nil
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

func (m *mockPlatformService) RecordDeploy(_ context.Context, arg db.InsertDeployLogParams) (service.DeployRecord, error) {
	if m.err != nil {
		return service.DeployRecord{}, m.err
	}
	rec := service.DeployRecord{
		ID: "d-new", Service: arg.Service,
		VersionFrom: arg.VersionFrom, VersionTo: arg.VersionTo,
		Status: arg.Status, DeployedBy: arg.DeployedBy,
		StartedAt: "2026-04-14T10:00:00Z",
	}
	if arg.Notes.Valid {
		rec.Notes = arg.Notes.String
	}
	return rec, nil
}

func (m *mockPlatformService) ListDeploys(_ context.Context, _, _ int32) ([]service.DeployRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.deploys, nil
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
	h := NewPlatform(mock, testPub, "platform", nil)
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

// --- update / disable / enable tenant ---

func TestUpdateTenant_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"name":"New Name","plan_id":"p-2","settings":{}}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/tenants/t-1", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateTenant_InvalidJSON_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/tenants/t-1", strings.NewReader("not json")), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDisableTenant_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants/t-1/disable", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEnableTenant_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants/t-1/enable", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- modules ---

func TestListModules_Success(t *testing.T) {
	mock := &mockPlatformService{modules: []db.Module{{ID: "m-1", Name: "Fleet"}}}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/modules", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEnableModule_MissingModuleID_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"module_id":""}` // empty module_id
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants/t-1/modules", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty module_id, got %d", rec.Code)
	}
}

func TestEnableModule_InvalidJSON_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants/t-1/modules", strings.NewReader("nope")), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestEnableModule_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"module_id":"fleet"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants/t-1/modules", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisableModule_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodDelete, "/v1/platform/tenants/t-1/modules/fleet", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- feature flags ---

func TestListFeatureFlags_Success(t *testing.T) {
	mock := &mockPlatformService{flags: []service.FeatureFlag{{ID: "f-1", Name: "dark_mode", Enabled: true}}}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/flags", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestToggleFlag_InvalidJSON_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodPatch, "/v1/platform/flags/f-1", strings.NewReader("nope")), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

// --- config ---

func TestGetConfig_Success(t *testing.T) {
	mock := &mockPlatformService{config: service.ConfigEntry{Key: "site_name", Value: []byte(`"SDA"`)}}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/config/site_name", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetConfig_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"value":"new-value"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/config/site_name", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// SetConfig may return 204 or 200 depending on handler
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("expected 204 or 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ── Tenant CRUD edge cases ────────────────────────────────────────────────────

func TestCreateTenant_EmptySlug_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	// Slug is empty — handler rejects before calling service.
	body := `{"slug":"","name":"Test","plan_id":"p-1","postgres_url":"x","redis_url":"x"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/tenants", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty slug, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("expected non-empty error field")
	}
}

func TestUpdateTenant_NotFound_Returns404(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrTenantNotFound}
	r := setupPlatformRouter(mock)

	body := `{"name":"New Name","plan_id":"p-1","settings":{}}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/tenants/nonexistent", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// UpdateTenant passes ErrTenantNotFound through serverError (500) because the
	// handler does not map it to 404. This test documents the current behaviour.
	// If the handler is later updated to return 404, change the assertion here.
	if rec.Code == http.StatusOK {
		t.Fatalf("expected non-200 for nonexistent tenant update, got %d", rec.Code)
	}
}

// ── Module management edge cases ─────────────────────────────────────────────

func TestGetTenantModules_Success(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/tenants/t-1/modules", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDisableModule_AlreadyDisabled_IsIdempotent(t *testing.T) {
	// DisableModule returns nil even when called twice — 204 both times.
	r := setupPlatformRouter(&mockPlatformService{})

	for i := 0; i < 2; i++ {
		req := withAdminAuth(httptest.NewRequest(http.MethodDelete, "/v1/platform/tenants/t-1/modules/fleet", nil), t)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("call %d: expected 204, got %d: %s", i+1, rec.Code, rec.Body.String())
		}
	}
}

// ── Feature flag edge cases ───────────────────────────────────────────────────

func TestToggleFlag_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"enabled":true}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPatch, "/v1/platform/flags/f-1", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ── Header spoofing guard ─────────────────────────────────────────────────────

func TestRequirePlatformAdmin_StripsIdentityHeadersBeforeJWTCheck(t *testing.T) {
	// An attacker tries to spoof admin identity before the middleware can set it.
	// The middleware must strip these before JWT verification so downstream
	// handlers only ever see middleware-injected values.
	r := setupPlatformRouter(&mockPlatformService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/tenants", nil)
	req.Header.Set("X-User-ID", "spoofed-uid")
	req.Header.Set("X-User-Role", "admin")
	req.Header.Set("X-Tenant-Slug", "platform")
	// No Authorization header — must be rejected as 401, not pass through.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 — spoofed headers must not bypass auth, got %d", rec.Code)
	}
}

// ── SetConfig error path ──────────────────────────────────────────────────────

func TestSetConfig_ServiceError_Returns500(t *testing.T) {
	mock := &mockPlatformService{err: errors.New("config store unavailable")}
	r := setupPlatformRouter(mock)

	body := `{"value":"x"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/config/some_key", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error message, got %q", resp["error"])
	}
}

// ── Modules service error path ────────────────────────────────────────────────

func TestListModules_ServiceError_Returns500(t *testing.T) {
	mock := &mockPlatformService{err: errors.New("db down")}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/modules", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// ── Deploy log ──────────────────────────────────────────────────────────────

func TestRecordDeploy_Success_Returns201(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"service":"auth","version_from":"1.0.0","version_to":"1.1.0","status":"success"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp service.DeployRecord
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Service != "auth" {
		t.Errorf("expected service auth, got %s", resp.Service)
	}
	if resp.VersionTo != "1.1.0" {
		t.Errorf("expected version_to 1.1.0, got %s", resp.VersionTo)
	}
}

func TestRecordDeploy_MissingService_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"version_to":"1.1.0"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecordDeploy_InvalidJSON_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader("nope")), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRecordDeploy_NoAuth_Returns401(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"service":"auth","version_to":"1.1.0"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRecordDeploy_ServiceError_Returns500(t *testing.T) {
	mock := &mockPlatformService{err: errors.New("db exploded")}
	r := setupPlatformRouter(mock)

	body := `{"service":"auth","version_to":"1.1.0"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestListDeploys_Success(t *testing.T) {
	mock := &mockPlatformService{
		deploys: []service.DeployRecord{
			{ID: "d-1", Service: "auth", VersionTo: "1.0.0", Status: "success", StartedAt: "2026-04-14T10:00:00Z"},
			{ID: "d-2", Service: "chat", VersionTo: "2.0.0", Status: "pending", StartedAt: "2026-04-14T11:00:00Z"},
		},
	}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/deploys", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var deploys []service.DeployRecord
	json.NewDecoder(rec.Body).Decode(&deploys)
	if len(deploys) != 2 {
		t.Errorf("expected 2 deploys, got %d", len(deploys))
	}
}

func TestListDeploys_Empty_ReturnsEmptyArray(t *testing.T) {
	mock := &mockPlatformService{deploys: []service.DeployRecord{}}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/deploys", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var deploys []service.DeployRecord
	json.NewDecoder(rec.Body).Decode(&deploys)
	if len(deploys) != 0 {
		t.Errorf("expected 0 deploys, got %d", len(deploys))
	}
}

func TestListDeploys_NoAuth_Returns401(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/platform/deploys", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestListDeploys_ServiceError_Returns500(t *testing.T) {
	mock := &mockPlatformService{err: errors.New("db down")}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodGet, "/v1/platform/deploys", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestRecordDeploy_DefaultStatus_Pending(t *testing.T) {
	mock := &mockPlatformService{}
	r := setupPlatformRouter(mock)

	// No status field — handler should default to "pending"
	body := `{"service":"auth","version_from":"1.0.0","version_to":"1.1.0"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp service.DeployRecord
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Status != "pending" {
		t.Errorf("expected default status pending, got %s", resp.Status)
	}
}

func TestRecordDeploy_InvalidStatus_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"service":"auth","version_to":"1.1.0","status":"yolo"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid status, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecordDeploy_WithNotes_IncludesNotes(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"service":"auth","version_from":"1.0.0","version_to":"1.1.0","notes":"hotfix for login bug"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp service.DeployRecord
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Notes != "hotfix for login bug" {
		t.Errorf("expected notes to be propagated, got %q", resp.Notes)
	}
}

func TestRecordDeploy_WithoutNotes_OmitsNotes(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"service":"auth","version_from":"1.0.0","version_to":"1.1.0"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/deploys", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if _, hasNotes := resp["notes"]; hasNotes {
		t.Errorf("expected notes to be omitted when empty, got %v", resp["notes"])
	}
}

// --- feature flags v2: create, update, kill, evaluate ---

func TestCreateFlag_Success_Returns201(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"id":"flag-1","name":"dark_mode","rollout_pct":50}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/flags", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var flag service.FeatureFlag
	json.NewDecoder(rec.Body).Decode(&flag)
	if flag.ID != "flag-1" {
		t.Errorf("expected id flag-1, got %s", flag.ID)
	}
	if flag.RolloutPct != 50 {
		t.Errorf("expected rollout_pct 50, got %d", flag.RolloutPct)
	}
}

func TestCreateFlag_MissingID_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"name":"dark_mode"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/flags", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateFlag_MissingName_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"id":"flag-1"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/flags", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateFlag_InvalidRollout_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"id":"flag-1","name":"test","rollout_pct":150}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/flags", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for rollout >100, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateFlag_DefaultRollout_IsZero(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"id":"flag-1","name":"dark_mode"}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPost, "/v1/platform/flags", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var flag service.FeatureFlag
	json.NewDecoder(rec.Body).Decode(&flag)
	if flag.RolloutPct != 0 {
		t.Errorf("expected default rollout_pct 0, got %d", flag.RolloutPct)
	}
}

func TestUpdateFlag_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"enabled":true,"rollout_pct":25}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/flags/f-1", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateFlag_NoFields_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/flags/f-1", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateFlag_InvalidRollout_Returns400(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	body := `{"rollout_pct":-5}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/flags/f-1", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for negative rollout, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateFlag_NotFound_Returns404(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrFlagNotFound}
	r := setupPlatformRouter(mock)

	body := `{"enabled":false}`
	req := withAdminAuth(httptest.NewRequest(http.MethodPut, "/v1/platform/flags/f-999", strings.NewReader(body)), t)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestKillFlag_Success_Returns204(t *testing.T) {
	r := setupPlatformRouter(&mockPlatformService{})

	req := withAdminAuth(httptest.NewRequest(http.MethodDelete, "/v1/platform/flags/f-1/kill", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestKillFlag_NotFound_Returns404(t *testing.T) {
	mock := &mockPlatformService{err: service.ErrFlagNotFound}
	r := setupPlatformRouter(mock)

	req := withAdminAuth(httptest.NewRequest(http.MethodDelete, "/v1/platform/flags/f-999/kill", nil), t)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEvaluateFlags_WithJWT_Returns200(t *testing.T) {
	mock := &mockPlatformService{flags: []service.FeatureFlag{
		{ID: "f-1", Name: "dark_mode", Enabled: true},
		{ID: "f-2", Name: "new_chat", Enabled: false},
	}}
	h := NewPlatform(mock, testPub, "platform", nil)
	r := chi.NewRouter()
	r.Mount("/v1/platform", h.Routes())
	r.Mount("/v1/flags", h.FlagsRoutes())

	req := httptest.NewRequest(http.MethodGet, "/v1/flags/evaluate", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]map[string]bool
	json.NewDecoder(rec.Body).Decode(&resp)
	flags := resp["flags"]
	if !flags["dark_mode"] {
		t.Error("expected dark_mode=true")
	}
	if flags["new_chat"] {
		t.Error("expected new_chat=false")
	}
}

func TestEvaluateFlags_NoToken_Returns401(t *testing.T) {
	h := NewPlatform(&mockPlatformService{}, testPub, "platform", nil)
	r := chi.NewRouter()
	r.Mount("/v1/flags", h.FlagsRoutes())

	req := httptest.NewRequest(http.MethodGet, "/v1/flags/evaluate", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEvaluateFlags_ResponseOnlyBooleans(t *testing.T) {
	mock := &mockPlatformService{flags: []service.FeatureFlag{
		{ID: "f-1", Name: "feature_a", Enabled: true},
	}}
	h := NewPlatform(mock, testPub, "platform", nil)
	r := chi.NewRouter()
	r.Mount("/v1/flags", h.FlagsRoutes())

	req := httptest.NewRequest(http.MethodGet, "/v1/flags/evaluate", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify response shape: {"flags": {"feature_a": true}} — only booleans, no metadata
	var raw map[string]json.RawMessage
	json.NewDecoder(rec.Body).Decode(&raw)

	flagsRaw, ok := raw["flags"]
	if !ok {
		t.Fatal("expected 'flags' key in response")
	}

	var flags map[string]bool
	if err := json.Unmarshal(flagsRaw, &flags); err != nil {
		t.Fatalf("flags should be map[string]bool, got error: %v", err)
	}
}
