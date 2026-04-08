package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
)

// --- Session CRUD ---

type createSessionRequest struct {
	Title     string `json:"title"`
	ContactID string `json:"contact_id"`
}

func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	title := req.Title
	if title == "" {
		title = "Nueva consulta"
	}
	var contactID pgtype.UUID
	if req.ContactID != "" {
		contactID.Scan(req.ContactID)
	}

	session, err := h.q.CreateSession(r.Context(), repository.CreateSessionParams{
		TenantID:  tid,
		UserID:    uid,
		ContactID: contactID,
		Title:     title,
	})
	if err != nil {
		serverError(w, r, "create session failed", err)
		return
	}
	if h.auditor != nil {
		h.auditor.Write(r.Context(), audit.Entry{
			UserID:   sdamw.UserIDFromContext(r.Context()),
			Action:   "astro.session.create",
			Resource: session.ID.Bytes[:],
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
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
	sessions, err := h.q.ListSessions(r.Context(), repository.ListSessionsParams{
		TenantID: tid, UserID: uid, Limit: limit, Offset: offset,
	})
	if err != nil {
		serverError(w, r, "list sessions failed", err)
		return
	}
	jsonOK(w, sessions)
}

func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	var sessionID pgtype.UUID
	if err := sessionID.Scan(idStr); err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	session, err := h.q.GetSession(r.Context(), repository.GetSessionParams{
		TenantID: tid, UserID: uid, ID: sessionID,
	})
	if err != nil {
		jsonError(w, "session not found", http.StatusNotFound)
		return
	}
	jsonOK(w, session)
}

func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	var sessionID pgtype.UUID
	if err := sessionID.Scan(idStr); err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.q.DeleteSession(r.Context(), repository.DeleteSessionParams{
		TenantID: tid, UserID: uid, ID: sessionID,
	}); err != nil {
		serverError(w, r, "delete session failed", err)
		return
	}
	if h.auditor != nil {
		h.auditor.Write(r.Context(), audit.Entry{
			UserID:   sdamw.UserIDFromContext(r.Context()),
			Action:   "astro.session.delete",
			Resource: idStr,
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateSessionRequest struct {
	Title  *string `json:"title,omitempty"`
	Pinned *bool   `json:"pinned,omitempty"`
}

func (h *Handler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	var sessionID pgtype.UUID
	if err := sessionID.Scan(idStr); err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req updateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title != nil {
		h.q.UpdateSessionTitle(r.Context(), repository.UpdateSessionTitleParams{
			TenantID: tid, UserID: uid, ID: sessionID, Title: *req.Title,
		})
	}
	if req.Pinned != nil {
		h.q.UpdateSessionPinned(r.Context(), repository.UpdateSessionPinnedParams{
			TenantID: tid, UserID: uid, ID: sessionID, Pinned: *req.Pinned,
		})
	}
	if h.auditor != nil {
		h.auditor.Write(r.Context(), audit.Entry{
			UserID:   sdamw.UserIDFromContext(r.Context()),
			Action:   "astro.session.update",
			Resource: idStr,
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, _, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := chi.URLParam(r, "id")
	var sessionID pgtype.UUID
	if err := sessionID.Scan(idStr); err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	limit := int32(100)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = int32(n)
		}
	}
	messages, err := h.q.GetMessages(r.Context(), repository.GetMessagesParams{
		TenantID: tid, SessionID: sessionID, Limit: limit,
	})
	if err != nil {
		serverError(w, r, "get messages failed", err)
		return
	}
	jsonOK(w, messages)
}
