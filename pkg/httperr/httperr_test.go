package httperr_test

import (
	stderrors "errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
)

func TestInternal_HasCorrectStatus(t *testing.T) {
	cause := stderrors.New("db connection failed")
	err := httperr.Internal(cause)
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeInternal {
		t.Errorf("expected code %q, got %q", httperr.CodeInternal, err.Code)
	}
}

func TestNotFound_HasCorrectStatus(t *testing.T) {
	err := httperr.NotFound("nonconformity")
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeNotFound {
		t.Errorf("expected code %q, got %q", httperr.CodeNotFound, err.Code)
	}
	if err.Message != "nonconformity not found" {
		t.Errorf("expected message to contain resource name, got %q", err.Message)
	}
}

func TestInvalidInput_HasCorrectStatus(t *testing.T) {
	err := httperr.InvalidInput("field is required")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeInvalidInput {
		t.Errorf("expected code %q, got %q", httperr.CodeInvalidInput, err.Code)
	}
}

func TestInvalidID_HasCorrectStatus(t *testing.T) {
	err := httperr.InvalidID("bad-uuid")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeInvalidInput {
		t.Errorf("expected code %q, got %q", httperr.CodeInvalidInput, err.Code)
	}
	if err.Message != "invalid id: bad-uuid" {
		t.Errorf("unexpected message %q", err.Message)
	}
}

func TestConflict_HasCorrectStatus(t *testing.T) {
	err := httperr.Conflict("record already exists")
	if err.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeConflict {
		t.Errorf("expected code %q, got %q", httperr.CodeConflict, err.Code)
	}
}

func TestUnauthorized_HasCorrectStatus(t *testing.T) {
	err := httperr.Unauthorized("not authenticated")
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeUnauthorized {
		t.Errorf("expected code %q, got %q", httperr.CodeUnauthorized, err.Code)
	}
}

func TestForbidden_HasCorrectStatus(t *testing.T) {
	err := httperr.Forbidden("access denied")
	if err.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeForbidden {
		t.Errorf("expected code %q, got %q", httperr.CodeForbidden, err.Code)
	}
}

func TestWriteError_Error_WritesCorrectJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	httperr.WriteError(w, r, httperr.NotFound("audit"))

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] != "audit not found" {
		t.Errorf("expected error message %q, got %q", "audit not found", body["error"])
	}
	if body["code"] != string(httperr.CodeNotFound) {
		t.Errorf("expected code %q, got %q", httperr.CodeNotFound, body["code"])
	}
}

func TestWriteError_Error_WritesCorrectStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	httperr.WriteError(w, r, httperr.InvalidInput("invalid severity"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected HTTP status 400, got %d", w.Code)
	}
}

func TestWriteError_GenericError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	httperr.WriteError(w, r, stderrors.New("some unexpected error"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected HTTP status 500, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] != "internal error" {
		t.Errorf("expected %q, got %q", "internal error", body["error"])
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := stderrors.New("original cause")
	err := httperr.Internal(cause)
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() should return the original cause")
	}
}

func TestWrap_CustomStatus(t *testing.T) {
	cause := stderrors.New("db timeout")
	err := httperr.Wrap(cause, httperr.CodeInternal, "service unavailable", http.StatusServiceUnavailable)
	if err.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", err.StatusCode)
	}
	if err.Code != httperr.CodeInternal {
		t.Errorf("expected code %q, got %q", httperr.CodeInternal, err.Code)
	}
}

func TestError_ErrorString_WithCause(t *testing.T) {
	cause := stderrors.New("root cause")
	err := httperr.Internal(cause)
	want := "internal error: root cause"
	if err.Error() != want {
		t.Errorf("expected %q, got %q", want, err.Error())
	}
}

func TestError_ErrorString_WithoutCause(t *testing.T) {
	err := httperr.NotFound("user")
	want := "user not found"
	if err.Error() != want {
		t.Errorf("expected %q, got %q", want, err.Error())
	}
}
