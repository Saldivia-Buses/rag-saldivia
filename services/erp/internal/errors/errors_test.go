package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	stderrors "errors"
	"testing"

	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
)

func TestInternal_HasCorrectStatus(t *testing.T) {
	cause := stderrors.New("db connection failed")
	err := erperrors.Internal(cause)
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", err.StatusCode)
	}
	if err.Code != erperrors.CodeInternal {
		t.Errorf("expected code %q, got %q", erperrors.CodeInternal, err.Code)
	}
}

func TestNotFound_HasCorrectStatus(t *testing.T) {
	err := erperrors.NotFound("nonconformity")
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", err.StatusCode)
	}
	if err.Code != erperrors.CodeNotFound {
		t.Errorf("expected code %q, got %q", erperrors.CodeNotFound, err.Code)
	}
	if err.Message != "nonconformity not found" {
		t.Errorf("expected message to contain resource name, got %q", err.Message)
	}
}

func TestInvalidInput_HasCorrectStatus(t *testing.T) {
	err := erperrors.InvalidInput("field is required")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Code != erperrors.CodeInvalidInput {
		t.Errorf("expected code %q, got %q", erperrors.CodeInvalidInput, err.Code)
	}
}

func TestInvalidID_HasCorrectStatus(t *testing.T) {
	err := erperrors.InvalidID("bad-uuid")
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Code != erperrors.CodeInvalidInput {
		t.Errorf("expected code %q, got %q", erperrors.CodeInvalidInput, err.Code)
	}
	if err.Message != "invalid id: bad-uuid" {
		t.Errorf("unexpected message %q", err.Message)
	}
}

func TestConflict_HasCorrectStatus(t *testing.T) {
	err := erperrors.Conflict("record already exists")
	if err.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", err.StatusCode)
	}
	if err.Code != erperrors.CodeConflict {
		t.Errorf("expected code %q, got %q", erperrors.CodeConflict, err.Code)
	}
}

func TestWriteError_ERPError_WritesCorrectJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	erperrors.WriteError(w, r, erperrors.NotFound("audit"))

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] != "audit not found" {
		t.Errorf("expected error message %q, got %q", "audit not found", body["error"])
	}
	if body["code"] != string(erperrors.CodeNotFound) {
		t.Errorf("expected code %q, got %q", erperrors.CodeNotFound, body["code"])
	}
}

func TestWriteError_ERPError_WritesCorrectStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	erperrors.WriteError(w, r, erperrors.InvalidInput("invalid severity"))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected HTTP status 400, got %d", w.Code)
	}
}

func TestWriteError_GenericError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)

	erperrors.WriteError(w, r, stderrors.New("some unexpected error"))

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

func TestERPError_Unwrap(t *testing.T) {
	cause := stderrors.New("original cause")
	err := erperrors.Internal(cause)
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() should return the original cause")
	}
}
