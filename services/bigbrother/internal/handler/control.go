package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/service"
)

// Control handles PLC and exec-related endpoints.
type Control struct {
	plcSvc *service.PLCService
}

// NewControl creates a new control handler.
func NewControl(plcSvc *service.PLCService) *Control {
	return &Control{plcSvc: plcSvc}
}

// ListRegisters returns PLC registers for a device.
func (h *Control) ListRegisters(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	if deviceID == "" {
		http.Error(w, `{"error":"device id required"}`, http.StatusBadRequest)
		return
	}

	regs, err := h.plcSvc.ListRegisters(r.Context(), deviceID)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"registers": regs})
}

// WriteRegister handles PLC register write with safety checks.
func (h *Control) WriteRegister(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	address := chi.URLParam(r, "addr")

	r.Body = http.MaxBytesReader(w, r.Body, 1024) // 1KB max

	var body struct {
		Value float64 `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	// Extract user info from JWT claims (set by auth middleware)
	userID := r.Header.Get("X-User-ID")

	result, err := h.plcSvc.WriteRegister(r.Context(), service.WriteRegisterRequest{
		DeviceID:  deviceID,
		Address:   address,
		Value:     body.Value,
		UserID:    userID,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	switch result.Status {
	case "pending_approval":
		status = http.StatusAccepted
	case "rejected":
		status = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(result)
}

// ApproveWrite handles two-person approval for critical PLC writes.
func (h *Control) ApproveWrite(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")
	if requestID == "" {
		http.Error(w, `{"error":"request_id required"}`, http.StatusBadRequest)
		return
	}

	approverID := r.Header.Get("X-User-ID")

	result, err := h.plcSvc.ApproveWrite(r.Context(), requestID, approverID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if result.Status == "rejected" {
		status = http.StatusForbidden
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(result)
}
