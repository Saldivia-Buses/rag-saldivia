// Package httperr provides structured error types and HTTP response helpers
// shared across all SDA services.
package httperr

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

// Error is a structured error with HTTP status, code, and message.
type Error struct {
	StatusCode int    // HTTP status code
	Code       Code   // machine-readable code
	Message    string // human-readable message (safe to expose)
	Cause      error  // underlying error (NOT exposed to client)
}

func (e Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e Error) Unwrap() error { return e.Cause }

// Constructor helpers

// Internal returns a 500 error wrapping an underlying cause.
func Internal(cause error) Error {
	return Error{StatusCode: http.StatusInternalServerError, Code: CodeInternal, Message: "internal error", Cause: cause}
}

// NotFound returns a 404 error for the named resource.
func NotFound(resource string) Error {
	return Error{StatusCode: http.StatusNotFound, Code: CodeNotFound, Message: resource + " not found"}
}

// InvalidInput returns a 400 error with the provided message.
func InvalidInput(msg string) Error {
	return Error{StatusCode: http.StatusBadRequest, Code: CodeInvalidInput, Message: msg}
}

// InvalidID returns a 400 error for an invalid identifier.
func InvalidID(id string) Error {
	return Error{StatusCode: http.StatusBadRequest, Code: CodeInvalidInput, Message: "invalid id: " + id}
}

// Conflict returns a 409 error with the provided message.
func Conflict(msg string) Error {
	return Error{StatusCode: http.StatusConflict, Code: CodeConflict, Message: msg}
}

// Unauthorized returns a 401 error with the provided message.
func Unauthorized(msg string) Error {
	return Error{StatusCode: http.StatusUnauthorized, Code: CodeUnauthorized, Message: msg}
}

// Forbidden returns a 403 error with the provided message.
func Forbidden(msg string) Error {
	return Error{StatusCode: http.StatusForbidden, Code: CodeForbidden, Message: msg}
}

// Wrap wraps an existing error into a structured Error.
func Wrap(cause error, code Code, msg string, status int) Error {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	return Error{StatusCode: status, Code: code, Message: msg, Cause: cause}
}

// errorResponse is the JSON body sent to clients.
type errorResponse struct {
	Error string `json:"error"`
	Code  Code   `json:"code,omitempty"`
}

// WriteError writes a structured JSON error response and logs appropriately.
// It accepts Error or any error (fallback to 500 internal).
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var e Error
	if stderrors.As(err, &e) {
		// Log based on severity
		if e.StatusCode >= 500 {
			slog.Error("server error", "code", e.Code, "error", e.Cause, "path", r.URL.Path)
		} else if e.StatusCode >= 400 {
			slog.Warn("client error", "code", e.Code, "message", e.Message, "path", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(e.StatusCode)
		_ = json.NewEncoder(w).Encode(errorResponse{Error: e.Message, Code: e.Code})
		return
	}
	// Unknown error type — treat as internal
	slog.Error("unhandled error", "error", err, "path", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "internal error", Code: CodeInternal})
}
