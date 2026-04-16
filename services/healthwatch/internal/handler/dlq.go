package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
)

// DLQ provides admin endpoints for dead-letter event management.
type DLQ struct {
	pool *pgxpool.Pool
	nc   *nats.Conn
}

// NewDLQ creates a DLQ handler. Requires platform DB pool + NATS for replay.
func NewDLQ(pool *pgxpool.Pool, nc *nats.Conn) *DLQ {
	return &DLQ{pool: pool, nc: nc}
}

// Routes mounts DLQ admin endpoints under /v1/admin/dlq.
// RBAC enforced per-route via RequirePermission middleware.
func (h *DLQ) Routes(r chi.Router) {
	r.Route("/v1/admin/dlq", func(r chi.Router) {
		r.With(sdamw.RequirePermission("admin.dlq.read")).Get("/", h.list)
		r.With(sdamw.RequirePermission("admin.dlq.read")).Get("/{id}", h.get)
		r.With(sdamw.RequirePermission("admin.dlq.replay")).Post("/{id}/replay", h.replay)
		r.With(sdamw.RequirePermission("admin.dlq.drop")).Post("/{id}/drop", h.drop)
	})
}

// DeadEvent is the JSON shape returned by list/get.
type DeadEvent struct {
	ID              string          `json:"id"`
	OriginalSubject string          `json:"original_subject"`
	OriginalStream  string          `json:"original_stream"`
	ConsumerName    string          `json:"consumer_name"`
	TenantID        *string         `json:"tenant_id"`
	EventType       *string         `json:"event_type"`
	DeliveryCount   int             `json:"delivery_count"`
	LastError       string          `json:"last_error"`
	DeadAt          time.Time       `json:"dead_at"`
	Envelope        json.RawMessage `json:"envelope,omitempty"`
	ReplayCount     int             `json:"replay_count"`
	LastReplayedAt  *time.Time      `json:"last_replayed_at"`
	DroppedAt       *time.Time      `json:"dropped_at"`
}

func (h *DLQ) list(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	consumerFilter := r.URL.Query().Get("consumer")
	tenantFilter := r.URL.Query().Get("tenant")

	query := `SELECT id, original_subject, original_stream, consumer_name,
		tenant_id, event_type, delivery_count, last_error, dead_at,
		replay_count, last_replayed_at, dropped_at
		FROM dead_events WHERE dropped_at IS NULL`
	args := []any{}
	argN := 1

	if consumerFilter != "" {
		query += ` AND consumer_name = $` + strconv.Itoa(argN)
		args = append(args, consumerFilter)
		argN++
	}
	if tenantFilter != "" {
		query += ` AND tenant_id = $` + strconv.Itoa(argN)
		args = append(args, tenantFilter)
		argN++
	}

	query += ` ORDER BY dead_at DESC LIMIT $` + strconv.Itoa(argN) + ` OFFSET $` + strconv.Itoa(argN+1)
	args = append(args, limit, offset)

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	defer rows.Close()

	events := []DeadEvent{}
	for rows.Next() {
		var e DeadEvent
		if err := rows.Scan(&e.ID, &e.OriginalSubject, &e.OriginalStream,
			&e.ConsumerName, &e.TenantID, &e.EventType, &e.DeliveryCount,
			&e.LastError, &e.DeadAt, &e.ReplayCount, &e.LastReplayedAt,
			&e.DroppedAt); err != nil {
			httperr.WriteError(w, r, httperr.Internal(err))
			return
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"events": events, "count": len(events)})
}

func (h *DLQ) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var e DeadEvent
	err := h.pool.QueryRow(r.Context(),
		`SELECT id, original_subject, original_stream, consumer_name,
			tenant_id, event_type, delivery_count, last_error, dead_at,
			envelope, replay_count, last_replayed_at, dropped_at
		 FROM dead_events WHERE id = $1`, id,
	).Scan(&e.ID, &e.OriginalSubject, &e.OriginalStream, &e.ConsumerName,
		&e.TenantID, &e.EventType, &e.DeliveryCount, &e.LastError, &e.DeadAt,
		&e.Envelope, &e.ReplayCount, &e.LastReplayedAt, &e.DroppedAt)
	if err == pgx.ErrNoRows {
		httperr.WriteError(w, r, httperr.NotFound("dead event"))
		return
	}
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(e)
}

func (h *DLQ) replay(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := sdamw.UserIDFromContext(r.Context())

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	defer func() { _ = tx.Rollback(r.Context()) }()

	var originalSubject string
	var envelope json.RawMessage
	var droppedAt *time.Time
	err = tx.QueryRow(r.Context(),
		`SELECT original_subject, envelope, dropped_at
		 FROM dead_events WHERE id = $1 FOR UPDATE`, id,
	).Scan(&originalSubject, &envelope, &droppedAt)
	if err == pgx.ErrNoRows {
		httperr.WriteError(w, r, httperr.NotFound("dead event"))
		return
	}
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	if droppedAt != nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, "GONE", "dead event already dropped", http.StatusGone))
		return
	}

	newEventID, _ := uuid.NewV7()

	msg := &nats.Msg{Subject: originalSubject, Data: envelope, Header: nats.Header{}}
	if err := h.nc.PublishMsg(msg); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	_, err = tx.Exec(r.Context(),
		`INSERT INTO dead_events_replays (dead_event_id, replayed_by_user_id, new_event_id, status)
		 VALUES ($1, $2, $3, 'pending')`,
		id, userID, newEventID)
	if err != nil {
		slog.Error("record replay failed", "id", id, "error", err)
	}

	_, err = tx.Exec(r.Context(),
		`UPDATE dead_events SET replay_count = replay_count + 1, last_replayed_at = now()
		 WHERE id = $1`, id)
	if err != nil {
		slog.Error("update replay count failed", "id", id, "error", err)
	}

	if err := tx.Commit(r.Context()); err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"replayed":     true,
		"new_event_id": newEventID,
		"subject":      originalSubject,
	})
}

func (h *DLQ) drop(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.pool.Exec(r.Context(),
		`UPDATE dead_events SET dropped_at = now() WHERE id = $1 AND dropped_at IS NULL`, id)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	if result.RowsAffected() == 0 {
		httperr.WriteError(w, r, httperr.NotFound("dead event not found or already dropped"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"dropped": true, "id": id})
}
