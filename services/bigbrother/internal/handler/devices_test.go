package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/scanner"
)

// newBareDevices creates a Devices with all optional services nil.
// This is the state right after NewDevices before SetScanLoop/SetCredentialService.
func newBareDevices() *Devices {
	return &Devices{
		tenantSlug: "test-tenant",
	}
}

// --- SetScanMode ---

func TestSetScanMode_ValidModes(t *testing.T) {
	t.Parallel()

	modes := []string{"passive", "active", "full"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			h := newBareDevices()
			loop := scanner.NewLoop(&stubScanner{}, scanner.ModePassive, nil)
			h.SetScanLoop(loop)

			body := `{"mode":"` + mode + `"}`
			req := httptest.NewRequest(http.MethodPost, "/scan/mode", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.SetScanMode(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}

			var resp map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if resp["mode"] != mode {
				t.Fatalf("expected mode %q in response, got %v", mode, resp["mode"])
			}
			if resp["status"] != "mode_changed" {
				t.Fatalf("expected status mode_changed, got %v", resp["status"])
			}
		})
	}
}

func TestSetScanMode_InvalidMode(t *testing.T) {
	t.Parallel()
	h := newBareDevices()
	loop := scanner.NewLoop(&stubScanner{}, scanner.ModePassive, nil)
	h.SetScanLoop(loop)

	body := `{"mode":"turbo"}`
	req := httptest.NewRequest(http.MethodPost, "/scan/mode", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.SetScanMode(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetScanMode_NilScanLoop(t *testing.T) {
	t.Parallel()
	h := newBareDevices() // scanLoop is nil

	body := `{"mode":"passive"}`
	req := httptest.NewRequest(http.MethodPost, "/scan/mode", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.SetScanMode(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSetScanMode_InvalidBody(t *testing.T) {
	t.Parallel()
	h := newBareDevices()
	loop := scanner.NewLoop(&stubScanner{}, scanner.ModePassive, nil)
	h.SetScanLoop(loop)

	req := httptest.NewRequest(http.MethodPost, "/scan/mode", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.SetScanMode(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- TriggerScan ---

func TestTriggerScan_NilScanLoop(t *testing.T) {
	t.Parallel()
	h := newBareDevices()

	req := httptest.NewRequest(http.MethodPost, "/scan", nil)
	rec := httptest.NewRecorder()

	h.TriggerScan(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTriggerScan_WithScanLoop(t *testing.T) {
	t.Parallel()
	h := newBareDevices()
	loop := scanner.NewLoop(&stubScanner{}, scanner.ModeActive, nil)
	h.SetScanLoop(loop)

	req := httptest.NewRequest(http.MethodPost, "/scan", nil)
	rec := httptest.NewRecorder()

	h.TriggerScan(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "scan_triggered" {
		t.Fatalf("expected status scan_triggered, got %v", resp["status"])
	}
	if resp["mode"] != "active" {
		t.Fatalf("expected mode active, got %v", resp["mode"])
	}
}

// --- StoreCredential ---

func TestStoreCredential_NilCredentials(t *testing.T) {
	t.Parallel()
	h := newBareDevices()

	req := httptest.NewRequest(http.MethodPost, "/credentials", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.StoreCredential(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- ListCredentials ---

func TestListCredentials_NilCredentials(t *testing.T) {
	t.Parallel()
	h := newBareDevices()

	req := httptest.NewRequest(http.MethodGet, "/credentials", nil)
	rec := httptest.NewRecorder()

	h.ListCredentials(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- DeleteCredential ---

func TestDeleteCredential_NilCredentials(t *testing.T) {
	t.Parallel()
	h := newBareDevices()

	req := httptest.NewRequest(http.MethodDelete, "/credentials/abc", nil)
	rec := httptest.NewRecorder()

	h.DeleteCredential(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- ExecCommand ---

func TestExecCommand_NilRemoteSvc(t *testing.T) {
	t.Parallel()
	h := newBareDevices()

	body := `{"command":"reboot"}`
	req := httptest.NewRequest(http.MethodPost, "/devices/dev1/exec", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ExecCommand(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- helpers ---

// stubScanner is a minimal NetworkScanner for tests.
type stubScanner struct{}

func (s *stubScanner) Scan(_ context.Context) ([]scanner.Device, error) {
	return nil, nil
}
