package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgconn"
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

func tenantAndUser(r *http.Request) (pgtype.UUID, pgtype.UUID, error) {
	info, err := tenant.FromContext(r.Context())
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, fmt.Errorf("no tenant in context")
	}
	uid := sdamw.UserIDFromContext(r.Context())
	if uid == "" {
		return pgtype.UUID{}, pgtype.UUID{}, fmt.Errorf("no user in context")
	}
	var tid, uidPG pgtype.UUID
	tid.Scan(info.ID)
	uidPG.Scan(uid)
	return tid, uidPG, nil
}

func (h *Handler) resolveContact(r *http.Request, contactName string) (*repository.Contact, int, error) {
	if contactName == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("contact_name is required")
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}
	c, err := h.q.GetContactByName(r.Context(), repository.GetContactByNameParams{
		TenantID: tid, UserID: uid, Lower: contactName,
	})
	if err != nil {
		return nil, http.StatusNotFound, fmt.Errorf("contact not found")
	}
	return &c, 0, nil
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

// maxBodySize limits request body to 1MB.
const maxBodySize = 1 << 20

type techniqueRequest struct {
	ContactName string `json:"contact_name"`
	Year        int    `json:"year"`
}

func (h *Handler) parseRequest(w http.ResponseWriter, r *http.Request) (*techniqueRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req techniqueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if req.Year == 0 {
		req.Year = time.Now().Year()
	}
	if req.Year < -5000 || req.Year > 5000 {
		return nil, fmt.Errorf("year out of range")
	}
	return &req, nil
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// serverError logs a 500-class error with request context, then sends a generic message.
func serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	slog.Error(msg, "error", err, "request_id", middleware.GetReqID(r.Context()))
	jsonError(w, msg, http.StatusInternalServerError)
}

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// --- Technique endpoints ---

func (h *Handler) Natal(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	jsonOK(w, chart)
}

func (h *Handler) Transits(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	jsonOK(w, technique.CalcTransits(chart, req.Year))
}

func (h *Handler) Directions(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	midYear := time.Date(req.Year, 7, 1, 0, 0, 0, 0, time.UTC)
	age := midYear.Sub(birthDate).Hours() / (24 * 365.25)
	jsonOK(w, technique.FindDirections(chart, age, 2.0))
}

func (h *Handler) Progressions(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	prog, err := technique.CalcProgressions(chart, req.Year)
	if err != nil {
		serverError(w, r, "progressions failed", err)
		return
	}
	jsonOK(w, prog)
}

func (h *Handler) Returns(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	sr, err := technique.CalcSolarReturnAtBirthplace(chart, req.Year)
	if err != nil {
		serverError(w, r, "solar return failed", err)
		return
	}
	jsonOK(w, sr)
}

func (h *Handler) FixedStars(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	jsonOK(w, technique.FindFixedStarConjunctions(chart))
}

func (h *Handler) SolarArc(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, _, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	sa := technique.CalcSolarArcForYear(chart, req.Year)
	jsonOK(w, sa)
}

func (h *Handler) Profections(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	prof := technique.CalcProfection(chart, birthDate, req.Year)
	jsonOK(w, prof)
}

func (h *Handler) Firdaria(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	fird := technique.CalcFirdaria(birthDate, chart.Diurnal, req.Year)
	jsonOK(w, fird)
}

func (h *Handler) Brief(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseRequest(w, r)
	if err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	contact, code, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		jsonError(w, err.Error(), code)
		return
	}
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		serverError(w, r, "chart calculation failed", err)
		return
	}
	ctx, err := astrocontext.Build(chart, contact.Name, birthDate, req.Year)
	if err != nil {
		serverError(w, r, "context build failed", err)
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

	// Parse body BEFORE setting SSE headers (D6)
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req struct {
		ContactName string `json:"contact_name"`
		Query       string `json:"query"`
		Year        int    `json:"year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Year == 0 {
		req.Year = time.Now().Year()
	}
	if req.Year < -5000 || req.Year > 5000 {
		jsonError(w, "year out of range", http.StatusBadRequest)
		return
	}
	if req.ContactName == "" {
		jsonError(w, "contact_name is required", http.StatusBadRequest)
		return
	}
	if len(req.Query) > 2000 {
		jsonError(w, "query too long (max 2000 chars)", http.StatusBadRequest)
		return
	}
	req.Query = sanitizeQuery(req.Query)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// 1. Resolve contact
	contact, _, err := h.resolveContact(r, req.ContactName)
	if err != nil {
		sseError(w, flusher, r, "contact not found")
		return
	}
	sseEvent(w, flusher, "contact_recognized", map[string]string{"name": contact.Name})

	// 2. Build chart + context
	chart, birthDate, err := contactToChart(contact)
	if err != nil {
		sseError(w, flusher, r, "chart calculation failed")
		return
	}
	ctx, err := astrocontext.Build(chart, contact.Name, birthDate, req.Year)
	if err != nil {
		sseError(w, flusher, r, "context build failed")
		return
	}
	sseEvent(w, flusher, "calc_context", map[string]string{"status": "complete", "brief_length": strconv.Itoa(len(ctx.Brief))})

	// 3. LLM narration (buffered + chunked SSE)
	if h.llm != nil {
		prompt := fmt.Sprintf("Eres un astrólogo profesional. Analiza el siguiente brief y responde la consulta del usuario.\n\n%s\n\nConsulta: %s", ctx.Brief, req.Query)
		response, err := h.llm.SimplePrompt(r.Context(), prompt, 0.7)
		if err != nil {
			slog.Error("llm call failed", "error", err, "request_id", middleware.GetReqID(r.Context()))
			sseError(w, flusher, r, "narration unavailable")
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
	contacts, err := h.q.ListContacts(r.Context(), repository.ListContactsParams{
		TenantID: tid, UserID: uid, Limit: limit, Offset: offset,
	})
	if err != nil {
		serverError(w, r, "list failed", err)
		return
	}
	jsonOK(w, contacts)
}

// createContactRequest is the public-facing struct for CreateContact.
// Deliberately excludes TenantID/UserID — those come from JWT context.
type createContactRequest struct {
	Name           string      `json:"name"`
	BirthDate      pgtype.Date `json:"birth_date"`
	BirthTime      pgtype.Time `json:"birth_time"`
	BirthTimeKnown bool        `json:"birth_time_known"`
	City           string      `json:"city"`
	Nation         string      `json:"nation"`
	Lat            float64     `json:"lat"`
	Lon            float64     `json:"lon"`
	Alt            float64     `json:"alt"`
	UtcOffset      int32       `json:"utc_offset"`
	Relationship   pgtype.Text `json:"relationship"`
	Notes          pgtype.Text `json:"notes"`
	Kind           string      `json:"kind"`
}

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	if h.q == nil {
		jsonError(w, "database not configured", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req createContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	if !req.BirthDate.Valid || req.BirthDate.Time.IsZero() {
		jsonError(w, "birth_date is required", http.StatusBadRequest)
		return
	}
	if req.BirthDate.Time.Year() < -5000 || req.BirthDate.Time.Year() > 5000 {
		jsonError(w, "birth_date out of ephemeris range", http.StatusBadRequest)
		return
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lon < -180 || req.Lon > 180 {
		jsonError(w, "invalid coordinates", http.StatusBadRequest)
		return
	}
	validKinds := map[string]bool{"persona": true, "empresa": true}
	if req.Kind != "" && !validKinds[req.Kind] {
		jsonError(w, "kind must be 'persona' or 'empresa'", http.StatusBadRequest)
		return
	}
	tid, uid, err := tenantAndUser(r)
	if err != nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	params := repository.CreateContactParams{
		TenantID:       tid,
		UserID:         uid,
		Name:           req.Name,
		BirthDate:      req.BirthDate,
		BirthTime:      req.BirthTime,
		BirthTimeKnown: req.BirthTimeKnown,
		City:           req.City,
		Nation:         req.Nation,
		Lat:            req.Lat,
		Lon:            req.Lon,
		Alt:            req.Alt,
		UtcOffset:      req.UtcOffset,
		Relationship:   req.Relationship,
		Notes:          req.Notes,
		Kind:           req.Kind,
	}
	contact, err := h.q.CreateContact(r.Context(), params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			jsonError(w, "contact with this name already exists", http.StatusConflict)
		} else {
			serverError(w, r, "create failed", err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

// --- SSE helpers ---

func sseEvent(w http.ResponseWriter, f http.Flusher, event string, data any) {
	if data == nil {
		fmt.Fprintf(w, "event: %s\ndata: {}\n\n", event)
	} else {
		b, err := json.Marshal(data)
		if err != nil {
			slog.Error("sse marshal failed", "error", err, "event", event)
			b = []byte(`{"error":"marshal failed"}`)
		}
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
	}
	f.Flush()
}

func sseError(w http.ResponseWriter, f http.Flusher, r *http.Request, msg string) {
	slog.Error("astro query error", "error", msg, "request_id", middleware.GetReqID(r.Context()))
	sseEvent(w, f, "error", map[string]string{"message": msg})
}

// sanitizeQuery strips control characters and trims the query string.
func sanitizeQuery(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' {
			return -1
		}
		return r
	}, s)
}
