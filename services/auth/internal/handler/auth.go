// Package handler implements HTTP handlers for the auth service.
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Camionerou/rag-saldivia/services/auth/internal/service"
)

// Auth handles HTTP requests for authentication.
type Auth struct {
	authSvc *service.Auth
}

// NewAuth creates auth HTTP handlers.
func NewAuth(authSvc *service.Auth) *Auth {
	return &Auth{authSvc: authSvc}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Login handles POST /v1/auth/login
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	// Limit request body to 1MB to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email and password are required"})
		return
	}

	tokens, err := h.authSvc.Login(r.Context(), service.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		IP:        r.RemoteAddr, // chi's RealIP middleware already rewrites this
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid email or password"})
		case errors.Is(err, service.ErrAccountLocked):
			writeJSON(w, http.StatusTooManyRequests, errorResponse{Error: "too many attempts, try again later"})
		default:
			reqID := middleware.GetReqID(r.Context())
			slog.Error("login failed", "error", err, "request_id", reqID)
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// Health handles GET /health
func (h *Auth) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "auth"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
