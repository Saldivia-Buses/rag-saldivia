package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

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

func (m *mockChatService) ListSessions(_ context.Context, userID string) ([]service.Session, error) {
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

func (m *mockChatService) AddMessage(_ context.Context, sessionID, userID, role, content string, sources, metadata []byte) (*service.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	msg := service.Message{
		ID: "m-new", SessionID: sessionID, Role: role, Content: content,
		CreatedAt: time.Now(),
	}
	return &msg, nil
}

func (m *mockChatService) GetMessages(_ context.Context, sessionID string) ([]service.Message, error) {
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
	return req
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// --- tests ---

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

	for _, role := range []string{"user", "assistant", "system"} {
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
