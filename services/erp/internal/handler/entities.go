package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/service"
)

// EntitiesService is the interface the Entities handler depends on.
type EntitiesService interface {
	List(ctx context.Context, tenantID, entityType, search string, activeOnly bool, limit, offset int) ([]repository.ListEntitiesRow, error)
	Count(ctx context.Context, tenantID, entityType string, activeOnly bool) (int32, error)
	Get(ctx context.Context, id pgtype.UUID, tenantID string) (*service.EntityDetail, error)
	Create(ctx context.Context, req service.CreateEntityRequest) (repository.CreateEntityRow, error)
	Update(ctx context.Context, req service.UpdateEntityRequest) (repository.UpdateEntityRow, error)
	Delete(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
	AddContact(ctx context.Context, tenantID string, entityID pgtype.UUID, contactType, label, value, userID, ip string, metadata []byte) (repository.ErpEntityContact, error)
	AddNote(ctx context.Context, tenantID string, entityID pgtype.UUID, userID, noteType, body, ip string) (repository.ErpEntityNote, error)
	AddDocument(ctx context.Context, tenantID string, entityID pgtype.UUID, name, docType, fileKey, userID, ip string) (repository.ErpEntityDocument, error)
}

// hashTaxID returns a SHA-256 hex hash of a tax ID for searchable storage.
func hashTaxID(taxID string) string {
	h := sha256.Sum256([]byte(taxID))
	return hex.EncodeToString(h[:])
}

// Entities handles entity endpoints (employees, customers, suppliers).
type Entities struct {
	svc EntitiesService
}

// NewEntities creates an entities handler.
func NewEntities(svc EntitiesService) *Entities {
	return &Entities{svc: svc}
}

// Routes returns the chi router for entity endpoints.
func (h *Entities) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.entities.read"))
		r.Get("/", h.List)
		r.Get("/{id}", h.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.entities.write"))
		r.Post("/", h.Create)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Post("/{id}/contacts", h.AddContact)
		r.Post("/{id}/notes", h.AddNote)
		r.Post("/{id}/documents", h.AddDocument)
	})

	return r
}

// List returns paginated entities.
func (h *Entities) List(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	entityType := r.URL.Query().Get("type")
	if entityType == "" {
		http.Error(w, `{"error":"type query param required (employee, customer, supplier)"}`, http.StatusBadRequest)
		return
	}

	search := r.URL.Query().Get("search")
	activeOnly := r.URL.Query().Get("active") != "false"
	p := pagination.Parse(r)

	entities, err := h.svc.List(r.Context(), slug, entityType, search, activeOnly, p.Limit(), p.Offset())
	if err != nil {
		slog.Error("list entities failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	count, err := h.svc.Count(r.Context(), slug, entityType, activeOnly)
	if err != nil {
		slog.Error("count entities failed", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"entities":  entities,
		"total":     count,
		"page":      p.Page,
		"page_size": p.PageSize,
	})
}

// Get returns an entity with contacts, documents, notes, relations.
func (h *Entities) Get(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	detail, err := h.svc.Get(r.Context(), id, slug)
	if err != nil {
		slog.Error("get entity failed", "error", err)
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// Create creates a new entity.
func (h *Entities) Create(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	var body struct {
		Type     string  `json:"type"`
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		TaxID    *string `json:"tax_id,omitempty"`
		Email    *string `json:"email,omitempty"`
		Phone    *string `json:"phone,omitempty"`
		Address  *json.RawMessage `json:"address,omitempty"`
		Metadata *json.RawMessage `json:"metadata,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	// Hash tax_id for searchable storage. Envelope encryption for encrypted_tax_id
	// will be wired when pkg/crypto KEK is provisioned per-tenant.
	var taxIDHash *string
	if body.TaxID != nil && *body.TaxID != "" {
		hash := hashTaxID(*body.TaxID)
		taxIDHash = &hash
	}

	var addr, meta []byte
	if body.Address != nil {
		addr = []byte(*body.Address)
	}
	if body.Metadata != nil {
		meta = []byte(*body.Metadata)
	}

	entity, err := h.svc.Create(r.Context(), service.CreateEntityRequest{
		TenantID:  slug,
		Type:      body.Type,
		Code:      body.Code,
		Name:      body.Name,
		TaxIDHash: taxIDHash,
		Email:     body.Email,
		Phone:     body.Phone,
		Address:   addr,
		Metadata:  meta,
		UserID:    r.Header.Get("X-User-ID"),
		IP:        r.RemoteAddr,
	})
	if err != nil {
		slog.Error("create entity failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entity)
}

// Update updates an existing entity.
func (h *Entities) Update(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Code     string  `json:"code"`
		Name     string  `json:"name"`
		TaxID    *string `json:"tax_id,omitempty"`
		Email    *string `json:"email,omitempty"`
		Phone    *string `json:"phone,omitempty"`
		Address  *json.RawMessage `json:"address,omitempty"`
		Metadata *json.RawMessage `json:"metadata,omitempty"`
		Active   *bool   `json:"active,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	active := true
	if body.Active != nil {
		active = *body.Active
	}

	var taxIDHash *string
	if body.TaxID != nil && *body.TaxID != "" {
		hash := hashTaxID(*body.TaxID)
		taxIDHash = &hash
	}

	var addr, meta []byte
	if body.Address != nil {
		addr = []byte(*body.Address)
	}
	if body.Metadata != nil {
		meta = []byte(*body.Metadata)
	}

	entity, err := h.svc.Update(r.Context(), service.UpdateEntityRequest{
		ID:        id,
		TenantID:  slug,
		Code:      body.Code,
		Name:      body.Name,
		TaxIDHash: taxIDHash,
		Email:     body.Email,
		Phone:     body.Phone,
		Address:   addr,
		Metadata:  meta,
		Active:    active,
		UserID:    r.Header.Get("X-User-ID"),
		IP:        r.RemoteAddr,
	})
	if err != nil {
		slog.Error("update entity failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// Delete soft-deletes an entity.
func (h *Entities) Delete(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)

	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	if err := h.svc.Delete(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddContact adds a contact to an entity.
func (h *Entities) AddContact(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	entityID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Type  string `json:"type"`
		Label string `json:"label"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	contact, err := h.svc.AddContact(r.Context(), slug, entityID, body.Type, body.Label, body.Value, r.Header.Get("X-User-ID"), r.RemoteAddr, nil)
	if err != nil {
		slog.Error("add contact failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contact)
}

// AddNote adds a note to an entity.
func (h *Entities) AddNote(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	entityID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Type string `json:"type"`
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if body.Type == "" {
		body.Type = "note" // default after decode so client can override
	}

	note, err := h.svc.AddNote(r.Context(), slug, entityID, r.Header.Get("X-User-ID"), body.Type, body.Body, r.RemoteAddr)
	if err != nil {
		slog.Error("add note failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

// AddDocument registers a document for an entity.
func (h *Entities) AddDocument(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	entityID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Name    string `json:"name"`
		DocType string `json:"doc_type"`
		FileKey string `json:"file_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.svc.AddDocument(r.Context(), slug, entityID, body.Name, body.DocType, body.FileKey, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		slog.Error("add document failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}
