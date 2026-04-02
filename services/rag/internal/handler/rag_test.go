package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/rag/internal/service"
)

// --- mock ---

type mockRAGService struct {
	collections    []string
	streamBody     string
	generateErr    error
	collectionsErr error
	healthErr      error
}

func (m *mockRAGService) GenerateStream(_ context.Context, tenantSlug string, req service.GenerateRequest) (io.ReadCloser, string, error) {
	if m.generateErr != nil {
		return nil, "", m.generateErr
	}
	body := m.streamBody
	if body == "" {
		body = `data: {"choices":[{"delta":{"content":"hello"}}]}` + "\ndata: [DONE]\n"
	}
	return io.NopCloser(strings.NewReader(body)), "text/event-stream", nil
}

func (m *mockRAGService) ListCollections(_ context.Context, tenantSlug string) ([]string, error) {
	if m.collectionsErr != nil {
		return nil, m.collectionsErr
	}
	return m.collections, nil
}

func (m *mockRAGService) Health(_ context.Context) error {
	return m.healthErr
}

// --- helpers ---

func setupRAGRouter(mock *mockRAGService) *chi.Mux {
	h := NewRAG(mock)
	r := chi.NewRouter()
	r.Get("/health", h.Health)
	r.Mount("/v1/rag", h.Routes())
	return r
}

// --- generate tests ---

func TestGenerate_Success_StreamsSSE(t *testing.T) {
	mock := &mockRAGService{
		streamBody: "data: {\"choices\":[{\"delta\":{\"content\":\"hola\"}}]}\ndata: [DONE]\n",
	}
	r := setupRAGRouter(mock)

	body := `{"messages":[{"role":"user","content":"test"}],"collection_name":"docs"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rag/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %q", ct)
	}
	if !strings.Contains(rec.Body.String(), "hola") {
		t.Error("expected streamed content in response body")
	}
}

func TestGenerate_EmptyMessages_Returns400(t *testing.T) {
	r := setupRAGRouter(&mockRAGService{})

	body := `{"messages":[],"collection_name":"docs"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rag/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty messages, got %d", rec.Code)
	}
}

func TestGenerate_InvalidJSON_Returns400(t *testing.T) {
	r := setupRAGRouter(&mockRAGService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/rag/generate", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGenerate_BlueprintError_Returns502(t *testing.T) {
	mock := &mockRAGService{generateErr: errors.New("blueprint unreachable")}
	r := setupRAGRouter(mock)

	body := `{"messages":[{"role":"user","content":"test"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/rag/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["error"] != "rag server unavailable" {
		t.Errorf("expected generic error, got %q", resp["error"])
	}
}

// --- collections tests ---

func TestListCollections_Success(t *testing.T) {
	mock := &mockRAGService{collections: []string{"contratos", "manuales"}}
	r := setupRAGRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/rag/collections", nil)
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string][]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp["collections"]) != 2 {
		t.Errorf("expected 2 collections, got %d", len(resp["collections"]))
	}
}

func TestListCollections_BlueprintDown_Returns502(t *testing.T) {
	mock := &mockRAGService{collectionsErr: errors.New("connection refused")}
	r := setupRAGRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/v1/rag/collections", nil)
	req.Header.Set("X-Tenant-Slug", "saldivia")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
}

// --- health tests ---

func TestHealth_BlueprintHealthy_Returns200(t *testing.T) {
	r := setupRAGRouter(&mockRAGService{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

func TestHealth_BlueprintDown_Returns503(t *testing.T) {
	mock := &mockRAGService{healthErr: errors.New("blueprint unreachable")}
	r := setupRAGRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
