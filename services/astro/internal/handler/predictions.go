package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/repository"
)

// --- Prediction Tracking ---

type createPredictionRequest struct {
	SessionID   string   `json:"session_id"`
	ContactID   string   `json:"contact_id"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	DateFrom    string   `json:"date_from"` // ISO date
	DateTo      string   `json:"date_to"`
	Techniques  []string `json:"techniques"`
}

func (h *Handler) CreatePrediction(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req createPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ContactID == "" || req.Description == "" || req.DateFrom == "" || req.DateTo == "" {
		jsonError(w, "contact_id, description, date_from, date_to are required", http.StatusBadRequest)
		return
	}
	validCategories := map[string]bool{"timing": true, "event": true, "financial": true, "relational": true, "health": true, "general": true}
	if req.Category == "" {
		req.Category = "general"
	}
	if !validCategories[req.Category] {
		jsonError(w, "invalid category", http.StatusBadRequest)
		return
	}

	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var sessionID, contactID pgtype.UUID
	if req.SessionID != "" {
		if err := sessionID.Scan(req.SessionID); err != nil {
			jsonError(w, "invalid session_id", http.StatusBadRequest)
			return
		}
	}
	if err := contactID.Scan(req.ContactID); err != nil {
		jsonError(w, "invalid contact_id", http.StatusBadRequest)
		return
	}

	var dateFrom, dateTo pgtype.Date
	if err := dateFrom.Scan(req.DateFrom); err != nil {
		jsonError(w, "invalid date_from (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	if err := dateTo.Scan(req.DateTo); err != nil {
		jsonError(w, "invalid date_to (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	pred, err := h.q.CreatePrediction(r.Context(), repository.CreatePredictionParams{
		TenantID:    tid,
		UserID:      uid,
		SessionID:   sessionID,
		ContactID:   contactID,
		Category:    req.Category,
		Description: req.Description,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Techniques:  req.Techniques,
	})
	if err != nil {
		serverError(w, r, "create prediction failed", err)
		return
	}
	if h.auditor != nil {
		h.auditor.Write(r.Context(), audit.Entry{
			UserID:   sdamw.UserIDFromContext(r.Context()),
			Action:   "astro.prediction.create",
			Resource: req.Description[:min(50, len(req.Description))],
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(pred); err != nil {
		slog.Error("encode prediction response", "error", err)
	}
}

func (h *Handler) ListPredictions(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	limit := int32(50)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	preds, err := h.q.ListPredictions(r.Context(), repository.ListPredictionsParams{
		TenantID: tid, UserID: uid, Limit: limit, Offset: offset,
	})
	if err != nil {
		serverError(w, r, "list predictions failed", err)
		return
	}
	jsonOK(w, preds)
}

type verifyPredictionRequest struct {
	Outcome string `json:"outcome"` // correct, incorrect, partial
	Notes   string `json:"notes"`
}

func (h *Handler) VerifyPrediction(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req verifyPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	validOutcomes := map[string]bool{"correct": true, "incorrect": true, "partial": true}
	if !validOutcomes[req.Outcome] {
		jsonError(w, "outcome must be correct, incorrect, or partial", http.StatusBadRequest)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	var predID pgtype.UUID
	if err := predID.Scan(idStr); err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}

	var notes pgtype.Text
	if req.Notes != "" {
		if err := notes.Scan(req.Notes); err != nil {
			jsonError(w, "invalid notes", http.StatusBadRequest)
			return
		}
	}

	pred, err := h.q.VerifyPrediction(r.Context(), repository.VerifyPredictionParams{
		TenantID:     tid,
		UserID:       uid,
		ID:           predID,
		Outcome:      pgtype.Text{String: req.Outcome, Valid: true},
		OutcomeNotes: notes,
	})
	if err != nil {
		serverError(w, r, "verify prediction failed", err)
		return
	}
	if h.auditor != nil {
		h.auditor.Write(r.Context(), audit.Entry{
			UserID:   sdamw.UserIDFromContext(r.Context()),
			Action:   "astro.prediction.verify",
			Resource: idStr,
		})
	}
	jsonOK(w, pred)
}

func (h *Handler) PredictionStats(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	stats, err := h.q.PredictionStats(r.Context(), repository.PredictionStatsParams{
		TenantID: tid, UserID: uid,
	})
	if err != nil {
		serverError(w, r, "prediction stats failed", err)
		return
	}
	jsonOK(w, stats)
}

// min is builtin in Go 1.21+ — removed custom definition (D4 fix)
