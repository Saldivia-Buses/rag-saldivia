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

	"github.com/Camionerou/rag-saldivia/services/notification/internal/service"
)

// --- mock ---

type mockNotificationService struct {
	notifications []service.Notification
	prefs         *service.Preferences
	unreadCount   int
	markedCount   int64
	err           error
	sendCalled    bool
	sendReq       service.SendRequest
	sendErr       error
}

func (m *mockNotificationService) List(_ context.Context, userID string, unreadOnly bool, limit int) ([]service.Notification, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.notifications, nil
}

func (m *mockNotificationService) UnreadCount(_ context.Context, userID string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.unreadCount, nil
}

func (m *mockNotificationService) MarkRead(_ context.Context, notifID, userID string) error {
	if m.err != nil {
		return m.err
	}
	for _, n := range m.notifications {
		if n.ID == notifID && n.UserID == userID {
			return nil
		}
	}
	return service.ErrNotificationNotFound
}

func (m *mockNotificationService) MarkAllRead(_ context.Context, userID string) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.markedCount, nil
}

func (m *mockNotificationService) GetPreferences(_ context.Context, userID string) (*service.Preferences, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.prefs != nil {
		return m.prefs, nil
	}
	return &service.Preferences{EmailEnabled: true, InAppEnabled: true, MutedTypes: []string{}}, nil
}

func (m *mockNotificationService) UpdatePreferences(_ context.Context, userID string, emailEnabled, inAppEnabled bool, quietStart, quietEnd *string, mutedTypes []string) (*service.Preferences, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &service.Preferences{
		EmailEnabled: emailEnabled,
		InAppEnabled: inAppEnabled,
		MutedTypes:   mutedTypes,
	}, nil
}

func (m *mockNotificationService) Send(_ context.Context, req service.SendRequest) error {
	m.sendCalled = true
	m.sendReq = req
	if m.sendErr != nil {
		return m.sendErr
	}
	return nil
}

// --- helpers ---

func setupNotifRouter(mock *mockNotificationService) *chi.Mux {
	h := NewNotification(mock)
	r := chi.NewRouter()
	r.Mount("/v1/notifications", h.Routes())
	return r
}

func withUser(req *http.Request, userID string) *http.Request {
	req.Header.Set("X-User-ID", userID)
	return req
}

// --- tests ---

func TestList_MissingUserID_Returns401(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/notifications", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestList_ReturnsNotifications(t *testing.T) {
	mock := &mockNotificationService{
		notifications: []service.Notification{
			{ID: "n-1", UserID: "u-1", Title: "Test", CreatedAt: time.Now()},
		},
	}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUnreadCount_Returns_Count(t *testing.T) {
	mock := &mockNotificationService{unreadCount: 5}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications/count", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]int
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["count"] != 5 {
		t.Errorf("expected count 5, got %d", resp["count"])
	}
}

func TestMarkRead_Owner_Success(t *testing.T) {
	mock := &mockNotificationService{
		notifications: []service.Notification{
			{ID: "n-1", UserID: "u-1"},
		},
	}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodPatch, "/v1/notifications/n-1/read", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestMarkRead_NotFound_Returns404(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	req := withUser(httptest.NewRequest(http.MethodPatch, "/v1/notifications/nonexistent/read", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMarkAllRead_Returns_Count(t *testing.T) {
	mock := &mockNotificationService{markedCount: 3}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/read-all", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]int64
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["marked"] != 3 {
		t.Errorf("expected marked 3, got %d", resp["marked"])
	}
}

func TestGetPreferences_Returns_Defaults(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications/preferences", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var prefs service.Preferences
	json.NewDecoder(rec.Body).Decode(&prefs)
	if !prefs.EmailEnabled {
		t.Error("expected email_enabled true by default")
	}
}

func TestUpdatePreferences_Success(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	body := `{"email_enabled":false,"in_app_enabled":true,"muted_types":["chat.new_message"]}`
	req := withUser(httptest.NewRequest(http.MethodPut, "/v1/notifications/preferences", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var prefs service.Preferences
	json.NewDecoder(rec.Body).Decode(&prefs)
	if prefs.EmailEnabled {
		t.Error("expected email_enabled false")
	}
}

func TestUpdatePreferences_InvalidJSON_Returns400(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	req := withUser(httptest.NewRequest(http.MethodPut, "/v1/notifications/preferences", strings.NewReader("not json")), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestList_ServiceError_Returns500_GenericMessage(t *testing.T) {
	mock := &mockNotificationService{err: errors.New("database down")}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications", nil), "u-1")
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

func TestUnreadCount_ServiceError_Returns500(t *testing.T) {
	mock := &mockNotificationService{err: errors.New("db error")}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications/count", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMarkAllRead_ServiceError_Returns500(t *testing.T) {
	mock := &mockNotificationService{err: errors.New("db error")}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/read-all", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestGetPreferences_ServiceError_Returns500(t *testing.T) {
	mock := &mockNotificationService{err: errors.New("db error")}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications/preferences", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestUpdatePreferences_ServiceError_Returns500(t *testing.T) {
	mock := &mockNotificationService{err: errors.New("db error")}
	r := setupNotifRouter(mock)

	body := `{"email_enabled":true,"in_app_enabled":true}`
	req := withUser(httptest.NewRequest(http.MethodPut, "/v1/notifications/preferences", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestMarkRead_ServiceError_Returns500(t *testing.T) {
	mock := &mockNotificationService{err: errors.New("db error")}
	// Remove notifications so ErrNotificationNotFound won't fire — the general
	// error path should return 500.
	// But the mock MarkRead logic checks the err field first, so this is fine.
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodPatch, "/v1/notifications/n-1/read", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", rec.Code)
	}
}

// ── User isolation ────────────────────────────────────────────────────────────

func TestMarkRead_OtherUsersNotification_Returns404(t *testing.T) {
	// n-owner belongs to u-owner. u-attacker must get 404, not 403,
	// to avoid leaking whether the notification ID exists.
	mock := &mockNotificationService{
		notifications: []service.Notification{
			{ID: "n-owner", UserID: "u-owner"},
		},
	}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodPatch, "/v1/notifications/n-owner/read", nil), "u-attacker")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for other user's notification, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestList_OnlyCurrentUserNotifications(t *testing.T) {
	// The service layer filters by userID — the handler must pass the header value
	// through correctly. Verify the response is 200 and the mock receives the right
	// userID (indirectly: mock returns whatever is in m.notifications for any call,
	// so we just verify the handler passes X-User-ID through without error).
	mock := &mockNotificationService{
		notifications: []service.Notification{
			{ID: "n-1", UserID: "u-alice", Title: "Hello"},
		},
	}
	r := setupNotifRouter(mock)

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications", nil), "u-alice")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var notifications []service.Notification
	if err := json.NewDecoder(rec.Body).Decode(&notifications); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(notifications) != 1 {
		t.Errorf("expected 1 notification for u-alice, got %d", len(notifications))
	}
}

// ── Query params ─────────────────────────────────────────────────────────────

func TestList_UnreadOnlyFilter_PassedToService(t *testing.T) {
	// Handler must accept unread=true without error.
	r := setupNotifRouter(&mockNotificationService{})

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications?unread=true", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with unread=true, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestList_LimitParam_PassedToService(t *testing.T) {
	// Handler must accept limit=10 without error.
	r := setupNotifRouter(&mockNotificationService{})

	req := withUser(httptest.NewRequest(http.MethodGet, "/v1/notifications?limit=10", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with limit=10, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ── Preferences validation ────────────────────────────────────────────────────

func TestUpdatePreferences_MissingUserID_Returns401(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	body := `{"email_enabled":true,"in_app_enabled":true}`
	req := httptest.NewRequest(http.MethodPut, "/v1/notifications/preferences", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-User-ID header — requireUserID middleware must reject.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMarkAllRead_MissingUserID_Returns401(t *testing.T) {
	r := setupNotifRouter(&mockNotificationService{})

	req := httptest.NewRequest(http.MethodPost, "/v1/notifications/read-all", nil)
	// No X-User-ID header.
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

// ── Response shape ────────────────────────────────────────────────────────────

func TestMarkAllRead_ReturnsMarkedField(t *testing.T) {
	mock := &mockNotificationService{markedCount: 0}
	r := setupNotifRouter(mock)

	// Even when zero notifications are marked, the "marked" field must be present.
	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/read-all", nil), "u-1")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["marked"]; !ok {
		t.Error("response must contain 'marked' field even when count is 0")
	}
}

// ── Send endpoint ────────────────────────────────────────────────────────────

func TestSend_Email_Success(t *testing.T) {
	mock := &mockNotificationService{}
	r := setupNotifRouter(mock)

	body := `{"type":"email","to":"enzo@saldivia.com","subject":"Test Subject","body":"Test body content"}`
	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/send", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !mock.sendCalled {
		t.Fatal("expected Send to be called")
	}
	if mock.sendReq.Type != "email" {
		t.Errorf("expected type email, got %s", mock.sendReq.Type)
	}
	if mock.sendReq.To != "enzo@saldivia.com" {
		t.Errorf("expected to enzo@saldivia.com, got %s", mock.sendReq.To)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "sent" {
		t.Errorf("expected status sent, got %s", resp["status"])
	}
}

func TestSend_InApp_Success(t *testing.T) {
	mock := &mockNotificationService{}
	r := setupNotifRouter(mock)

	body := `{"type":"in_app","to":"user-123","subject":"New notification","body":"Hello"}`
	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/send", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if mock.sendReq.Type != "in_app" {
		t.Errorf("expected type in_app, got %s", mock.sendReq.Type)
	}
}

func TestSend_MissingFields_Returns400(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing type", `{"to":"a@b.com","subject":"s"}`},
		{"missing to", `{"type":"email","subject":"s"}`},
		{"missing subject", `{"type":"email","to":"a@b.com"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNotificationService{}
			r := setupNotifRouter(mock)

			req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/send", strings.NewReader(tt.body)), "u-1")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			if mock.sendCalled {
				t.Error("Send should not be called for invalid input")
			}
		})
	}
}

func TestSend_InvalidType_Returns400(t *testing.T) {
	mock := &mockNotificationService{}
	r := setupNotifRouter(mock)

	body := `{"type":"sms","to":"a@b.com","subject":"Test"}`
	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/send", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSend_MissingUserID_Returns401(t *testing.T) {
	mock := &mockNotificationService{}
	r := setupNotifRouter(mock)

	body := `{"type":"email","to":"a@b.com","subject":"Test","body":"Body"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/notifications/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSend_ServiceError_Returns500(t *testing.T) {
	mock := &mockNotificationService{sendErr: errors.New("smtp down")}
	r := setupNotifRouter(mock)

	body := `{"type":"email","to":"a@b.com","subject":"Test","body":"Body"}`
	req := withUser(httptest.NewRequest(http.MethodPost, "/v1/notifications/send", strings.NewReader(body)), "u-1")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}
