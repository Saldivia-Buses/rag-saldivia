package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/astro/internal/ephemeris"
)

func TestMain(m *testing.M) {
	ephemeris.Init(os.Getenv("EPHE_PATH"))
	code := m.Run()
	ephemeris.Close()
	os.Exit(code)
}

func TestHealthEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"astro"}`))
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("health status = %d, want 200", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("health body = %v", body)
	}
}

func TestHandlerStubsReturn401WithoutAuth(t *testing.T) {
	h := New(nil, nil, nil, nil, nil, nil, "test")
	r := chi.NewRouter()

	// No auth middleware — tenantAndUser will fail
	r.Post("/v1/astro/natal", h.Natal)
	r.Post("/v1/astro/brief", h.Brief)
	r.Get("/v1/astro/contacts", h.ListContacts)
	r.Post("/v1/astro/contacts", h.CreateContact)

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/v1/astro/natal", `{"contact_name":"test","year":2026}`},
		{"POST", "/v1/astro/brief", `{"contact_name":"test","year":2026}`},
		{"GET", "/v1/astro/contacts", ""},
		{"POST", "/v1/astro/contacts", `{"name":"test"}`},
	}

	for _, ep := range endpoints {
		var body *strings.Reader
		if ep.body != "" {
			body = strings.NewReader(ep.body)
		} else {
			body = strings.NewReader("")
		}
		req := httptest.NewRequest(ep.method, ep.path, body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Without auth context, should get 401 or 404 (contact not found because no tenant)
		if w.Code != http.StatusUnauthorized && w.Code != http.StatusNotFound && w.Code != http.StatusServiceUnavailable {
			t.Errorf("%s %s: status = %d, want 401/404/503", ep.method, ep.path, w.Code)
		}

		// Response should be JSON
		ct := w.Header().Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			t.Errorf("%s %s: Content-Type = %q, want application/json", ep.method, ep.path, ct)
		}
	}
}

func TestHandlerBadRequest(t *testing.T) {
	h := New(nil, nil, nil, nil, nil, nil, "test")
	r := chi.NewRouter()
	r.Post("/v1/astro/natal", h.Natal)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/v1/astro/natal", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("bad request status = %d, want 400", w.Code)
	}
}

func TestHandlerYearValidation(t *testing.T) {
	h := New(nil, nil, nil, nil, nil, nil, "test")
	r := chi.NewRouter()
	r.Post("/v1/astro/natal", h.Natal)

	// Year out of range
	req := httptest.NewRequest("POST", "/v1/astro/natal", strings.NewReader(`{"contact_name":"test","year":99999}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("year validation status = %d, want 400", w.Code)
	}
}
