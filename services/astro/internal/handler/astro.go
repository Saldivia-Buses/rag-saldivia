package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/llm"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/tenant"

	astrocontext "github.com/Camionerou/rag-saldivia/services/astro/internal/context"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/natal"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/astro/internal/technique"
)

type Handler struct {
	db  *pgxpool.Pool
	llm llm.ChatClient
	q   *repository.Queries
}

func New(db *pgxpool.Pool, llmClient llm.ChatClient) *Handler {
	h := &Handler{db: db, llm: llmClient}
	if db != nil {
		h.q = repository.New(db)
	}
	return h
}

// --- Contact helpers ---

func tenantAndUser(r *http.Request) (pgtype.UUID, pgtype.UUID) {
	info, _ := tenant.FromContext(r.Context())
	uid := sdamw.UserIDFromContext(r.Context())
	var tid, uidPG pgtype.UUID
	tid.Scan(info.ID)
	uidPG.Scan(uid)
	return tid, uidPG
}

func (h *Handler) resolveContact(r *http.Request, contactName string) (*repository.Contact, error) {
	tid, uid := tenantAndUser(r)
	c, err := h.q.GetContactByName(r.Context(), repository.GetContactByNameParams{
		TenantID: tid, UserID: uid, Lower: contactName,
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func contactToChart(c *repository.Contact) (*natal.Chart, time.Time, error) {
	hour := 12.0
	if c.BirthTimeKnown && c.BirthTime.Valid {
		// pgtype.Time stores microseconds since midnight
		us := c.BirthTime.Microseconds
		hour = float64(us) / (3600 * 1e6)
	}
	bd := c.BirthDate.Time
	birthDate := time.Date(bd.Year(), bd.Month(), bd.Day(), 0, 0, 0, 0, time.UTC)

	chart, err := natal.BuildNatal(
		bd.Year(), int(bd.Month()), bd.Day(),
		hour, c.Lat, c.Lon, c.Alt, int(c.UtcOffset),
	)
	return chart, birthDate, err
}

// --- Technique request ---

type techniqueRequest struct {
	ContactName string `json:"contact_name"`
	Year        int    `json:"year"`
}

func (h *Handler) parseRequest(r *http.Request) (*techniqueRequest, error) {
	var req techniqueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if req.Year == 0 {
		req.Year = time.Now().Year()
	}
	return &req, nil
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// --- Technique endpoints ---

func (h *Handler) Natal(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, "contact not found", http.StatusNotFound)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		jsonError(w, "chart calculation failed", http.StatusInternalServerError)
		return
	}
	jsonOK(w, chart)
}

func (h *Handler) Transits(w http.ResponseWriter, r *http.Request)     { stub(w, "transits") }
func (h *Handler) Directions(w http.ResponseWriter, r *http.Request)   { stub(w, "directions") }
func (h *Handler) Progressions(w http.ResponseWriter, r *http.Request) { stub(w, "progressions") }
func (h *Handler) Returns(w http.ResponseWriter, r *http.Request)      { stub(w, "returns") }
func (h *Handler) FixedStars(w http.ResponseWriter, r *http.Request)   { stub(w, "fixed-stars") }

func (h *Handler) SolarArc(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, "contact not found", http.StatusNotFound)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		jsonError(w, "chart calculation failed", http.StatusInternalServerError)
		return
	}
	sa := technique.CalcSolarArcForYear(chart, req.Year)
	jsonOK(w, sa)
}

func (h *Handler) Profections(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, "contact not found", http.StatusNotFound)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		jsonError(w, "chart calculation failed", http.StatusInternalServerError)
		return
	}
	prof := technique.CalcProfection(chart, birthDate, req.Year)
	jsonOK(w, prof)
}

func (h *Handler) Firdaria(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, "contact not found", http.StatusNotFound)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		jsonError(w, "chart calculation failed", http.StatusInternalServerError)
		return
	}
	fird := technique.CalcFirdaria(birthDate, chart.Diurnal, req.Year)
	jsonOK(w, fird)
}

func (h *Handler) Brief(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, "contact not found", http.StatusNotFound)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		jsonError(w, "chart calculation failed", http.StatusInternalServerError)
		return
	}
	ctx, err := astrocontext.Build(chart, contact.Name, birthDate, req.Year)
	if err != nil {
		jsonError(w, "context build failed", http.StatusInternalServerError)
		return
	}
	jsonOK(w, ctx)
}

// --- SSE Query (full pipeline + LLM narration) ---

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

	var req struct {
		ContactName string `json:"contact_name"`
		Query       string `json:"query"`
		Year        int    `json:"year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sseError(w, flusher, "invalid request: "+err.Error())
		return
	}
	if req.Year == 0 {
		req.Year = time.Now().Year()
	}

	// 1. Resolve contact
	contact, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		sseError(w, flusher, "contact not found: "+req.ContactName)
		return
	}
	sseEvent(w, flusher, "contact_recognized", map[string]string{"name": contact.Name})

	// 2. Build chart + context
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		sseError(w, flusher, "chart calculation failed")
		return
	}
	ctx, err := astrocontext.Build(chart, contact.Name, birthDate, req.Year)
	if err != nil {
		sseError(w, flusher, "context build failed")
		return
	}
	sseEvent(w, flusher, "calc_context", map[string]string{"status": "complete", "brief_length": strconv.Itoa(len(ctx.Brief))})

	// 3. LLM narration (buffered + chunked SSE)
	if h.llm != nil {
		prompt := fmt.Sprintf("Eres un astrólogo profesional. Analiza el siguiente brief y responde la consulta del usuario.\n\n%s\n\nConsulta: %s", ctx.Brief, req.Query)
		response, err := h.llm.SimplePrompt(r.Context(), prompt, 0.7)
		if err != nil {
			sseError(w, flusher, "LLM error: "+err.Error())
		} else {
			// Stream response in ~50 char chunks
			for i := 0; i < len(response); i += 50 {
				end := i + 50
				if end > len(response) {
					end = len(response)
				}
				sseEvent(w, flusher, "token", map[string]string{"text": response[i:end]})
			}
		}
	} else {
		// No LLM — send brief directly
		sseEvent(w, flusher, "brief", map[string]string{"text": ctx.Brief})
	}

	sseEvent(w, flusher, "done", nil)
}

// --- Contact CRUD ---

func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	tid, uid := tenantAndUser(r)
	contacts, err := h.q.ListContacts(r.Context(), repository.ListContactsParams{
		TenantID: tid, UserID: uid, Limit: 100, Offset: 0,
	})
	if err != nil {
		jsonError(w, "list failed", http.StatusInternalServerError)
		return
	}
	jsonOK(w, contacts)
}

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	var req repository.CreateContactParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	tid, uid := tenantAndUser(r)
	req.TenantID = tid
	req.UserID = uid

	contact, err := h.q.CreateContact(r.Context(), req)
	if err != nil {
		jsonError(w, "create failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	jsonOK(w, contact)
}

// --- SSE helpers ---

func stub(w http.ResponseWriter, name string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"status": "not_implemented", "endpoint": name})
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
