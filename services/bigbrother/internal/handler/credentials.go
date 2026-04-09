package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/service"
)

// Credentials handles credential management endpoints.
type Credentials struct {
	credSvc *service.CredentialService
}

// NewCredentials creates a credential handler.
func NewCredentials(credSvc *service.CredentialService) *Credentials {
	return &Credentials{credSvc: credSvc}
}

// Store creates or rotates a credential for a device.
func (h *Credentials) Store(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10) // 64KB max

	var body struct {
		DeviceID       string `json:"device_id"`
		CredType       string `json:"cred_type"`
		Secret         string `json:"secret"` // plaintext credential
		KeyFingerprint string `json:"key_fingerprint,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	if body.DeviceID == "" || body.CredType == "" || body.Secret == "" {
		http.Error(w, `{"error":"device_id, cred_type, and secret are required"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	meta, err := h.credSvc.Store(r.Context(), service.StoreRequest{
		DeviceID:       body.DeviceID,
		CredType:       body.CredType,
		Plaintext:      []byte(body.Secret),
		KeyFingerprint: body.KeyFingerprint,
		UserID:         userID,
		IP:             r.RemoteAddr,
	})
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(meta)
}

// List returns credential metadata (never plaintext).
func (h *Credentials) List(w http.ResponseWriter, r *http.Request) {
	creds, err := h.credSvc.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"credentials": creds})
}

// Delete removes a credential.
func (h *Credentials) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")

	if err := h.credSvc.Delete(r.Context(), id, userID, r.RemoteAddr); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
