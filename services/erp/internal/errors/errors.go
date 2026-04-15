// Package errors provides structured error types and HTTP response helpers for the ERP service.
package errors

import (
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"net/http"
)

// Code is a machine-readable error code string.
type Code string

const (
	CodeInternal     Code = "internal_error"
	CodeNotFound     Code = "not_found"
	CodeInvalidInput Code = "invalid_input"
	CodeUnauthorized Code = "unauthorized"
	CodeConflict     Code = "conflict"
	CodeForbidden    Code = "forbidden"
)

// ERPError is a structured error with HTTP status, code, and message.
type ERPError struct {
	StatusCode int    // HTTP status code
	Code       Code   // machine-readable code
	Message    string // human-readable message (safe to expose)
	Cause      error  // underlying error (NOT exposed to client)
}

func (e ERPError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e ERPError) Unwrap() error { return e.Cause }

// Constructor helpers

func Internal(cause error) ERPError {
	return ERPError{StatusCode: http.StatusInternalServerError, Code: CodeInternal, Message: "internal error", Cause: cause}
}

func NotFound(resource string) ERPError {
	return ERPError{StatusCode: http.StatusNotFound, Code: CodeNotFound, Message: resource + " not found"}
}

func InvalidInput(msg string) ERPError {
	return ERPError{StatusCode: http.StatusBadRequest, Code: CodeInvalidInput, Message: msg}
}

func InvalidID(id string) ERPError {
	return ERPError{StatusCode: http.StatusBadRequest, Code: CodeInvalidInput, Message: "invalid id: " + id}
}

func Conflict(msg string) ERPError {
	return ERPError{StatusCode: http.StatusConflict, Code: CodeConflict, Message: msg}
}

// Wrap wraps an existing error into an ERPError.
func Wrap(cause error, code Code, msg string, status int) ERPError {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return ERPError{StatusCode: status, Code: code, Message: msg, Cause: cause}
}

// errorResponse is the JSON body sent to clients.
type errorResponse struct {
	Error string `json:"error"`
	Code  Code   `json:"code,omitempty"`
}

// WriteError writes a structured JSON error response and logs appropriately.
// It accepts ERPError or any error (fallback to internal).
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var erpErr ERPError
	if stderrors.As(err, &erpErr) {
		// Log based on severity
		if erpErr.StatusCode >= 500 {
			slog.Error("server error", "code", erpErr.Code, "error", erpErr.Cause, "path", r.URL.Path)
		} else if erpErr.StatusCode >= 400 {
			slog.Warn("client error", "code", erpErr.Code, "message", erpErr.Message, "path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(erpErr.StatusCode)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: erpErr.Message, Code: erpErr.Code})
		return
	}
	// Unknown error type — treat as internal
	slog.Error("unhandled error", "error", err, "path", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "internal error", Code: CodeInternal})
}
