package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/notification/internal/service"
)

// --- mock alert store ---

type mockAlertStore struct {
	alerts []service.InfraAlert
	err    error
}

func (m *mockAlertStore) SaveAlert(_ context.Context, alert service.InfraAlert) error {
	if m.err != nil {
		return m.err
	}
	m.alerts = append(m.alerts, alert)
	return nil
}

// --- mock mailer ---

type mockWebhookMailer struct {
	sent []mockEmail
	err  error
}

type mockEmail struct {
	to, subject, body string
}

func (m *mockWebhookMailer) Send(_ context.Context, to, subject, body string) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, mockEmail{to: to, subject: subject, body: body})
	return nil
}

// --- test payloads ---

const validAlertPayload = `{
  "version": "4",
  "groupKey": "{}:{alertname=\"HighErrorRate\"}",
  "status": "firing",
  "receiver": "critical",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighErrorRate",
        "severity": "critical",
        "service_name": "auth"
      },
      "annotations": {
        "summary": "High error rate on auth service",
        "description": "Error rate is above 5% for 5 minutes"
      },
      "startsAt": "2026-04-14T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "fingerprint": "abc123"
    }
  ]
}`

const warningAlertPayload = `{
  "version": "4",
  "groupKey": "{}:{alertname=\"HighLatency\"}",
  "status": "firing",
  "receiver": "default",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighLatency",
        "severity": "warning",
        "service_name": "chat"
      },
      "annotations": {
        "summary": "High latency on chat service"
      },
      "startsAt": "2026-04-14T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "fingerprint": "def456"
    }
  ]
}`

// --- tests ---

func TestAlertWebhook_ValidToken_PersistsAlert(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(store.alerts) != 1 {
		t.Fatalf("expected 1 alert persisted, got %d", len(store.alerts))
	}

	alert := store.alerts[0]
	if alert.AlertName != "HighErrorRate" {
		t.Errorf("expected alertname HighErrorRate, got %s", alert.AlertName)
	}
	if alert.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", alert.Severity)
	}
	if alert.Service != "auth" {
		t.Errorf("expected service auth, got %s", alert.Service)
	}
	if alert.Status != "firing" {
		t.Errorf("expected status firing, got %s", alert.Status)
	}
	if alert.Fingerprint != "abc123" {
		t.Errorf("expected fingerprint abc123, got %s", alert.Fingerprint)
	}
}

func TestAlertWebhook_CriticalSeverity_SendsEmail(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(mailer.sent) != 1 {
		t.Fatalf("expected 1 email sent for critical alert, got %d", len(mailer.sent))
	}
	if mailer.sent[0].to != "ops@sda.app" {
		t.Errorf("expected email to ops@sda.app, got %s", mailer.sent[0].to)
	}
}

func TestAlertWebhook_WarningSeverity_NoEmail(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(warningAlertPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(store.alerts) != 1 {
		t.Fatalf("expected 1 alert persisted, got %d", len(store.alerts))
	}

	if len(mailer.sent) != 0 {
		t.Errorf("expected 0 emails for warning severity, got %d", len(mailer.sent))
	}
}

func TestAlertWebhook_NoToken_Returns401(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	// No Authorization header
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(store.alerts) != 0 {
		t.Error("no alerts should be persisted when auth fails")
	}
}

func TestAlertWebhook_WrongToken_Returns401(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAlertWebhook_OversizedPayload_Returns413(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	// Create a payload larger than 1MB
	bigPayload := `{"version":"4","alerts":[{"labels":{"data":"` + strings.Repeat("x", 1<<20) + `"}}]}`

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(bigPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAlertWebhook_InvalidJSON_Returns400(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader("not json at all"))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAlertWebhook_StoreError_Returns500(t *testing.T) {
	store := &mockAlertStore{err: errStoreDown}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAlertWebhook_ResponseIsJSON(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
}

func TestAlertWebhook_MultipleAlerts_PersistsAll(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	payload := `{
		"version": "4",
		"status": "firing",
		"alerts": [
			{
				"status": "firing",
				"labels": {"alertname": "A", "severity": "critical"},
				"annotations": {"summary": "Alert A"},
				"startsAt": "2026-04-14T10:00:00Z",
				"fingerprint": "fp1"
			},
			{
				"status": "firing",
				"labels": {"alertname": "B", "severity": "warning"},
				"annotations": {"summary": "Alert B"},
				"startsAt": "2026-04-14T10:01:00Z",
				"fingerprint": "fp2"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(store.alerts) != 2 {
		t.Fatalf("expected 2 alerts persisted, got %d", len(store.alerts))
	}

	// Only the critical alert should trigger email
	if len(mailer.sent) != 1 {
		t.Fatalf("expected 1 email (critical only), got %d", len(mailer.sent))
	}
}

func TestAlertWebhook_MailerError_StillReturns200(t *testing.T) {
	store := &mockAlertStore{}
	mailer := &mockWebhookMailer{err: errSentinel("smtp connection refused")}
	h := NewAlertWebhook("test-secret", store, mailer, "ops@sda.app")

	// Critical alert triggers email, but mailer fails
	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/alert", strings.NewReader(validAlertPayload))
	req.Header.Set("Authorization", "Bearer test-secret")
	rec := httptest.NewRecorder()

	h.HandleAlertWebhook(rec, req)

	// Alert must still be persisted and response must be 200
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 even when mailer fails, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(store.alerts) != 1 {
		t.Fatalf("expected alert persisted despite mailer error, got %d", len(store.alerts))
	}
}

var errStoreDown = errSentinel("store unavailable")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
