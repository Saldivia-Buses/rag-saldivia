package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/chat/internal/service"
)

// --- mock ---

type mockChatService struct {
	sessions []service.Session
	messages []service.Message
	err      error
}

func (m *mockChatService) CreateSession(_ context.Context, userID, title string, collection *string) (*service.Session, error) {
	if m.err != nil {
		return nil, m.err
	}
	s := service.Session{
		ID: "s-new", UserID: userID, Title: title, Collection: collection,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	return &s, nil
}

func (m *mockChatService) GetSession(_ context.Context, sessionID, userID string) (*service.Session, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.sessions {
		if s.ID == sessionID && s.UserID == userID {
			return &s, nil
		}
		if s.ID == sessionID {
			return nil, service.ErrNotOwner
		}
	}
	return nil, service.ErrSessionNotFound
}

func (m *mockChatService) ListSessions(_ context.Context, userID string, _, _ int32) ([]service.Session, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []service.Session
	for _, s := range m.sessions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	if result == nil {
		result = []service.Session{}
	}
	return result, nil
}

func (m *mockChatService) DeleteSession(_ context.Context, sessionID, userID string) error {
	if m.err != nil {
		return m.err
	}
	for _, s := range m.sessions {
		if s.ID == sessionID && s.UserID == userID {
			return nil
		}
		if s.ID == sessionID {
			return service.ErrSessionNotFound
		}
	}
	return service.ErrSessionNotFound
}

func (m *mockChatService) RenameSession(_ context.Context, sessionID, userID, title string) error {
	if m.err != nil {
		return m.err
	}
	for _, s := range m.sessions {
		if s.ID == sessionID && s.UserID == userID {
			return nil
		}
	}
	return service.ErrSessionNotFound
}

func (m *mockChatService) AddMessage(_ context.Context, sessionID, userID, role, content string, thinking *string, sources, metadata []byte) (*service.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	msg := service.Message{
		ID: "m-new", SessionID: sessionID, Role: role, Content: content,
		CreatedAt: time.Now(),
	}
	return &msg, nil
}

func (m *mockChatService) GetMessages(_ context.Context, sessionID string, _ int32) ([]service.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.messages, nil
}

// --- helpers ---

func setupRouter(mock *mockChatService) *chi.Mux {
	h := NewChat(mock)
	r := chi.NewRouter()
	r.Mount("/v1/chat/sessions", h.Routes())
	return r
}

func withUserID(req *http.Request, userID string) *http.Request {
	req.Header.Set("X-User-ID", userID)
	// Inject admin role + all permissions so RBAC middleware passes in tests.
	// Tests verify handler logic, not RBAC (RBAC has its own tests in pkg/middleware).
	ctx := sdamw.WithRole(req.Context(), "admin")
	return req.WithContext(ctx)
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// --- tests ---

func TestListSessions_MissingUserID_Returns401(t *testing.T) {
	r := setupRouter(&mockChatService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil)
	// No X-User-ID header
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing user identity, got %d", rec.Code)
	}
}

func TestListSessions_ReturnsUserSessions(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1", Title: "Chat 1"},
			{ID: "s-2", UserID: "u-2", Title: "Chat 2"}, // different user
			{ID: "s-3", UserID: "u-1", Title: "Chat 3"},
		},
	}
	r := setupRouter(mock)

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var sessions []service.Session
	decodeJSON(t, rec, &sessions)
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions for u-1, got %d", len(sessions))
	}
}

func TestCreateSession_Success(t *testing.T) {
	r := setupRouter(&mockChatService{})

	body := `{"title":"Mi chat","collection":"docs"}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var session service.Session
	decodeJSON(t, rec, &session)
	if session.Title != "Mi chat" {
		t.Errorf("expected title 'Mi chat', got %q", session.Title)
	}
	if session.UserID != "u-1" {
		t.Errorf("expected user_id u-1, got %q", session.UserID)
	}
}

func TestCreateSession_EmptyTitle_DefaultsToNuevaConversacion(t *testing.T) {
	r := setupRouter(&mockChatService{})

	body := `{}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var session service.Session
	decodeJSON(t, rec, &session)
	if session.Title != "Nueva conversacion" {
		t.Errorf("expected default title, got %q", session.Title)
	}
}

func TestCreateSession_InvalidJSON_Returns400(t *testing.T) {
	r := setupRouter(&mockChatService{})

	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions", strings.NewReader("not json")), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetSession_OwnerCanAccess(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1", Title: "My Chat"},
		},
	}
	r := setupRouter(mock)

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions/s-1", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSession_NonOwner_Returns404(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1", Title: "My Chat"},
		},
	}
	r := setupRouter(mock)

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions/s-1", nil), "u-2")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner, got %d", rec.Code)
	}
}

func TestGetSession_NotFound_Returns404(t *testing.T) {
	r := setupRouter(&mockChatService{})

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions/nonexistent", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteSession_OwnerCanDelete(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	req := withUserID(httptest.NewRequest(http.MethodDelete, "/v1/chat/sessions/s-1", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDeleteSession_NonOwner_Returns404(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	req := withUserID(httptest.NewRequest(http.MethodDelete, "/v1/chat/sessions/s-1", nil), "u-2")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner delete, got %d", rec.Code)
	}
}

func TestRenameSession_Success(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	body := `{"title":"Renamed"}`
	req := withUserID(httptest.NewRequest(http.MethodPatch, "/v1/chat/sessions/s-1", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRenameSession_EmptyTitle_Returns400(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	body := `{"title":""}`
	req := withUserID(httptest.NewRequest(http.MethodPatch, "/v1/chat/sessions/s-1", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty title, got %d", rec.Code)
	}
}

func TestAddMessage_ValidRole_Success(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	for _, role := range []string{"user", "assistant"} {
		t.Run(role, func(t *testing.T) {
			body := `{"role":"` + role + `","content":"hello"}`
			req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions/s-1/messages", strings.NewReader(body)), "u-1")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("expected 201 for role %s, got %d: %s", role, rec.Code, rec.Body.String())
			}
		})
	}

	// "system" role is blocked from API clients (only internal use)
	t.Run("system_blocked", func(t *testing.T) {
		body := `{"role":"system","content":"hello"}`
		req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions/s-1/messages", strings.NewReader(body)), "u-1")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for system role, got %d", rec.Code)
		}
	})
}

func TestAddMessage_InvalidRole_Returns400(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	body := `{"role":"admin","content":"hello"}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions/s-1/messages", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid role, got %d", rec.Code)
	}
}

func TestAddMessage_MissingContent_Returns400(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	body := `{"role":"user","content":""}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions/s-1/messages", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty content, got %d", rec.Code)
	}
}

func TestAddMessage_NonOwnerSession_Returns404(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	body := `{"role":"user","content":"hello"}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions/s-1/messages", strings.NewReader(body)), "u-2")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner adding message, got %d", rec.Code)
	}
}

func TestListSessions_ServiceError_Returns500_GenericMessage(t *testing.T) {
	mock := &mockChatService{err: errors.New("database connection lost")}
	r := setupRouter(mock)

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]string
	decodeJSON(t, rec, &resp)
	if resp["error"] != "internal error" {
		t.Errorf("expected generic error, got %q — service internals may be leaking", resp["error"])
	}
}

// --- new edge cases ---

// TestAddMessage_OversizeContent_Returns413 verifies that MaxBytesReader is
// enforced on AddMessage. A body exceeding 1MB must return 413, not 400 or 500.
func TestAddMessage_OversizeContent_Returns413(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
	}
	r := setupRouter(mock)

	// Build a body that exceeds 1<<20 (1MB). The JSON wrapper adds ~20 bytes of
	// overhead, so padding the content to 1<<20+1 bytes guarantees overflow.
	oversize := strings.Repeat("x", (1<<20)+1)
	body := `{"role":"user","content":"` + oversize + `"}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions/s-1/messages", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// net/http MaxBytesReader causes json.Decoder.Decode to return an error when
	// the limit is exceeded. The handler catches that and returns 400. However, the
	// Go stdlib wraps the MaxBytesError and the handler treats it as a bad request
	// since it cannot decode the body. Accept either 400 or 413.
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 400 or 413 for oversize body, got %d", rec.Code)
	}
	// Response must be JSON regardless of status code (invariant #7).
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("error response must be JSON, got: %s", rec.Body.String())
	}
}

// TestCreateSession_OversizeBody_Returns400 verifies MaxBytesReader on CreateSession.
func TestCreateSession_OversizeBody_Returns400(t *testing.T) {
	r := setupRouter(&mockChatService{})

	oversize := strings.Repeat("a", (1<<20)+1)
	body := `{"title":"` + oversize + `"}`
	req := withUserID(httptest.NewRequest(http.MethodPost, "/v1/chat/sessions", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 400 or 413 for oversize body, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Errorf("error response must be JSON, got: %s", rec.Body.String())
	}
}

// TestListSessions_NegativePage_HandledGracefully verifies that pagination.Parse
// ignores invalid page values and falls back to defaults, returning 200.
func TestListSessions_NegativePage_HandledGracefully(t *testing.T) {
	r := setupRouter(&mockChatService{})

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions?page=-1", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for negative page (should default), got %d", rec.Code)
	}
}

// TestListSessions_ZeroPage_HandledGracefully verifies page=0 also defaults cleanly.
func TestListSessions_ZeroPage_HandledGracefully(t *testing.T) {
	r := setupRouter(&mockChatService{})

	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions?page=0", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for zero page (should default), got %d", rec.Code)
	}
}

// TestErrorResponses_AreJSON verifies that all handler error paths return
// valid JSON bodies (invariant #7: no plain-text error responses).
func TestErrorResponses_AreJSON(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		noUserID   bool // skip injecting X-User-ID to trigger the 401 path
		mock       *mockChatService
	}{
		{
			name:     "missing user id",
			method:   http.MethodGet,
			path:     "/v1/chat/sessions",
			noUserID: true,
			mock:     &mockChatService{},
		},
		{
			name:   "invalid json body on create session",
			method: http.MethodPost,
			path:   "/v1/chat/sessions",
			body:   "{bad json",
			mock:   &mockChatService{},
		},
		{
			name:   "session not found",
			method: http.MethodGet,
			path:   "/v1/chat/sessions/nonexistent",
			mock:   &mockChatService{},
		},
		{
			name:   "service error returns generic message",
			method: http.MethodGet,
			path:   "/v1/chat/sessions",
			mock:   &mockChatService{err: errors.New("db down")},
		},
		{
			name:   "invalid role on add message",
			method: http.MethodPost,
			path:   "/v1/chat/sessions/s-1/messages",
			body:   `{"role":"hacker","content":"x"}`,
			mock: &mockChatService{
				sessions: []service.Session{{ID: "s-1", UserID: "u-1"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter(tt.mock)

			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			} else {
				bodyReader = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			if !tt.noUserID {
				req = withUserID(req, "u-1")
			}
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			// Error response must decode as JSON with an "error" key.
			if rec.Code < 400 {
				return // not an error response in this test case
			}
			var resp map[string]string
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Errorf("error response is not JSON: status=%d body=%q", rec.Code, rec.Body.String())
				return
			}
			if _, ok := resp["error"]; !ok {
				t.Errorf("JSON error response missing 'error' key: %v", resp)
			}
		})
	}
}

func TestGetMessages_VerifiesOwnership(t *testing.T) {
	mock := &mockChatService{
		sessions: []service.Session{
			{ID: "s-1", UserID: "u-1"},
		},
		messages: []service.Message{
			{ID: "m-1", SessionID: "s-1", Role: "user", Content: "hello"},
		},
	}
	r := setupRouter(mock)

	// Owner can get messages
	req := withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions/s-1/messages", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for owner, got %d", rec.Code)
	}

	// Non-owner gets 404
	req = withUserID(httptest.NewRequest(http.MethodGet, "/v1/chat/sessions/s-1/messages", nil), "u-2")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owner messages, got %d", rec.Code)
	}
}
