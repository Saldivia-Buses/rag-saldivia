package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/llm"
)

// Handler serves astro HTTP endpoints.
type Handler struct {
	db  *pgxpool.Pool
	llm llm.ChatClient
}

// New creates a handler with optional DB pool and LLM client.
func New(db *pgxpool.Pool, llmClient llm.ChatClient) *Handler {
	return &Handler{db: db, llm: llmClient}
}

// stub returns 501 for endpoints not yet implemented.
func stub(w http.ResponseWriter, name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"status": "not_implemented", "endpoint": name})
}

func (h *Handler) Natal(w http.ResponseWriter, r *http.Request)        { stub(w, "natal") }
func (h *Handler) Transits(w http.ResponseWriter, r *http.Request)     { stub(w, "transits") }
func (h *Handler) SolarArc(w http.ResponseWriter, r *http.Request)     { stub(w, "solar-arc") }
func (h *Handler) Directions(w http.ResponseWriter, r *http.Request)   { stub(w, "directions") }
func (h *Handler) Progressions(w http.ResponseWriter, r *http.Request) { stub(w, "progressions") }
func (h *Handler) Returns(w http.ResponseWriter, r *http.Request)      { stub(w, "returns") }
func (h *Handler) Profections(w http.ResponseWriter, r *http.Request)  { stub(w, "profections") }
func (h *Handler) Firdaria(w http.ResponseWriter, r *http.Request)     { stub(w, "firdaria") }
func (h *Handler) FixedStars(w http.ResponseWriter, r *http.Request)   { stub(w, "fixed-stars") }
func (h *Handler) Brief(w http.ResponseWriter, r *http.Request)        { stub(w, "brief") }
func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) { stub(w, "list-contacts") }
func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	stub(w, "create-contact")
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sseEvent(w, flusher, "status", map[string]string{"message": "not_implemented"})
	sseEvent(w, flusher, "done", nil)
}

func sseEvent(w http.ResponseWriter, f http.Flusher, event string, data any) {
	if data == nil {
		fmt.Fprintf(w, "event: %s\ndata: {}\n\n", event)
	} else {
		b, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
	}
	f.Flush()
}

func sseError(w http.ResponseWriter, f http.Flusher, msg string) {
	slog.Error("astro query error", "error", msg)
	sseEvent(w, f, "error", map[string]string{"message": msg})
}
