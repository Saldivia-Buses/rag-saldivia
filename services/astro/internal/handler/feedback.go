package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/repository"
)

// --- Feedback (thumbs up/down) ---

type feedbackRequest struct {
	MessageID string `json:"message_id"`
	Thumbs    string `json:"thumbs"` // "up" or "down"
	Comment   string `json:"comment"`
}

func (h *Handler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req feedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.MessageID == "" || (req.Thumbs != "up" && req.Thumbs != "down") {
		jsonError(w, "message_id and thumbs (up/down) are required", http.StatusBadRequest)
		return
	}

	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var msgID pgtype.UUID
	if err := msgID.Scan(req.MessageID); err != nil {
		jsonError(w, "invalid message_id", http.StatusBadRequest)
		return
	}

	var comment pgtype.Text
	if req.Comment != "" {
		comment.Scan(req.Comment)
	}

	fb, err := h.q.CreateFeedback(r.Context(), repository.CreateFeedbackParams{
		TenantID:  tid,
		MessageID: msgID,
		UserID:    uid,
		Thumbs:    req.Thumbs,
		Comment:   comment,
	})
	if err != nil {
		serverError(w, r, "submit feedback failed", err)
		return
	}
	if h.auditor != nil {
		h.auditor.Write(r.Context(), audit.Entry{
			UserID:   sdamw.UserIDFromContext(r.Context()),
			Action:   "astro.feedback." + req.Thumbs,
			Resource: req.MessageID,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fb)
}
