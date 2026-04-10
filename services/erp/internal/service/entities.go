package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/traces"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// Entities handles entity business logic (employees, customers, suppliers).
type Entities struct {
	repo      *repository.Queries
	audit     *audit.Writer
	publisher *traces.Publisher
}

// NewEntities creates an entities service.
func NewEntities(repo *repository.Queries, auditWriter *audit.Writer, publisher *traces.Publisher) *Entities {
	return &Entities{repo: repo, audit: auditWriter, publisher: publisher}
}

// List returns paginated entities filtered by type.
func (s *Entities) List(ctx context.Context, tenantID, entityType, search string, activeOnly bool, limit, offset int) ([]repository.ListEntitiesRow, error) {
	return s.repo.ListEntities(ctx, repository.ListEntitiesParams{
		TenantID:   tenantID,
		Type:       entityType,
		ActiveOnly: activeOnly,
		Search:     search,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
}

// Count returns total entity count for a type.
func (s *Entities) Count(ctx context.Context, tenantID, entityType string, activeOnly bool) (int32, error) {
	return s.repo.CountEntities(ctx, repository.CountEntitiesParams{
		TenantID:   tenantID,
		Type:       entityType,
		ActiveOnly: activeOnly,
	})
}

// Get returns an entity with all related data.
func (s *Entities) Get(ctx context.Context, id pgtype.UUID, tenantID string) (*EntityDetail, error) {
	entity, err := s.repo.GetEntity(ctx, repository.GetEntityParams{ID: id, TenantID: tenantID})
	if err != nil {
		return nil, fmt.Errorf("get entity: %w", err)
	}

	contacts, err := s.repo.ListEntityContacts(ctx, repository.ListEntityContactsParams{
		EntityID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}

	documents, err := s.repo.ListEntityDocuments(ctx, repository.ListEntityDocumentsParams{
		EntityID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	notes, err := s.repo.ListEntityNotes(ctx, repository.ListEntityNotesParams{
		EntityID: id, TenantID: tenantID, Limit: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}

	relations, err := s.repo.ListEntityRelations(ctx, repository.ListEntityRelationsParams{
		FromID: id, TenantID: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("list relations: %w", err)
	}

	return &EntityDetail{
		Entity:    entity,
		Contacts:  contacts,
		Documents: documents,
		Notes:     notes,
		Relations: relations,
	}, nil
}

// EntityDetail bundles an entity with all its related data.
type EntityDetail struct {
	Entity    repository.GetEntityRow        `json:"entity"`
	Contacts  []repository.ErpEntityContact  `json:"contacts"`
	Documents []repository.ErpEntityDocument `json:"documents"`
	Notes     []repository.ErpEntityNote     `json:"notes"`
	Relations []repository.ErpEntityRelation `json:"relations"`
}

// CreateEntityRequest holds data for creating an entity.
type CreateEntityRequest struct {
	TenantID       string
	Type           string
	Code           string
	Name           string
	EncryptedTaxID []byte
	TaxIDHash      *string
	Email          *string
	Phone          *string
	Address        []byte
	Metadata       []byte
	UserID         string
	IP             string
}

// Create creates a new entity.
func (s *Entities) Create(ctx context.Context, req CreateEntityRequest) (repository.CreateEntityRow, error) {
	if req.Type == "" || req.Code == "" || req.Name == "" {
		return repository.CreateEntityRow{}, fmt.Errorf("type, code, and name are required")
	}

	if req.Address == nil {
		req.Address = []byte(`{}`)
	}
	if req.Metadata == nil {
		req.Metadata = []byte(`{}`)
	}

	entity, err := s.repo.CreateEntity(ctx, repository.CreateEntityParams{
		TenantID:       req.TenantID,
		Type:           req.Type,
		Code:           req.Code,
		Name:           req.Name,
		EncryptedTaxID: req.EncryptedTaxID,
		TaxIDHash:      pgTextPtr(req.TaxIDHash),
		Email:          pgTextPtr(req.Email),
		Phone:          pgTextPtr(req.Phone),
		Address:        req.Address,
		Metadata:       req.Metadata,
	})
	if err != nil {
		return repository.CreateEntityRow{}, fmt.Errorf("create entity: %w", err)
	}

	idStr := uuidStr(entity.ID)

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.entity.created",
		Resource: idStr,
		Details:  map[string]any{"type": req.Type, "code": req.Code},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_entities", map[string]any{
		"action":    "created",
		"entity_id": idStr,
		"type":      req.Type,
	})

	slog.Info("entity created", "id", idStr, "type", req.Type, "code", req.Code)
	return entity, nil
}

// UpdateEntityRequest holds data for updating an entity.
type UpdateEntityRequest struct {
	ID             pgtype.UUID
	TenantID       string
	Code           string
	Name           string
	EncryptedTaxID []byte
	TaxIDHash      *string
	Email          *string
	Phone          *string
	Address        []byte
	Metadata       []byte
	Active         bool
	UserID         string
	IP             string
}

// Update updates an existing entity.
func (s *Entities) Update(ctx context.Context, req UpdateEntityRequest) (repository.UpdateEntityRow, error) {
	if req.Code == "" || req.Name == "" {
		return repository.UpdateEntityRow{}, fmt.Errorf("code and name are required")
	}
	if req.Address == nil {
		req.Address = []byte(`{}`)
	}
	if req.Metadata == nil {
		req.Metadata = []byte(`{}`)
	}

	entity, err := s.repo.UpdateEntity(ctx, repository.UpdateEntityParams{
		ID:             req.ID,
		TenantID:       req.TenantID,
		Code:           req.Code,
		Name:           req.Name,
		EncryptedTaxID: req.EncryptedTaxID,
		TaxIDHash:      pgTextPtr(req.TaxIDHash),
		Email:          pgTextPtr(req.Email),
		Phone:          pgTextPtr(req.Phone),
		Address:        req.Address,
		Metadata:       req.Metadata,
		Active:         req.Active,
	})
	if err != nil {
		return repository.UpdateEntityRow{}, fmt.Errorf("update entity: %w", err)
	}

	idStr := uuidStr(entity.ID)

	s.audit.Write(ctx, audit.Entry{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Action:   "erp.entity.updated",
		Resource: idStr,
		Details:  map[string]any{"code": req.Code},
		IP:       req.IP,
	})

	s.publisher.Broadcast(req.TenantID, "erp_entities", map[string]any{
		"action":    "updated",
		"entity_id": idStr,
	})

	return entity, nil
}

// Delete soft-deletes an entity.
func (s *Entities) Delete(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error {
	if err := s.repo.SoftDeleteEntity(ctx, repository.SoftDeleteEntityParams{
		ID: id, TenantID: tenantID,
	}); err != nil {
		return fmt.Errorf("delete entity: %w", err)
	}

	idStr := uuidStr(id)
	s.audit.Write(ctx, audit.Entry{
		TenantID: tenantID, UserID: userID,
		Action: "erp.entity.deleted", Resource: idStr, IP: ip,
	})
	s.publisher.Broadcast(tenantID, "erp_entities", map[string]any{
		"action": "deleted", "entity_id": idStr,
	})
	return nil
}

// AddContact adds a contact to an entity.
func (s *Entities) AddContact(ctx context.Context, tenantID string, entityID pgtype.UUID, contactType, label, value string, metadata []byte) (repository.ErpEntityContact, error) {
	if metadata == nil {
		metadata = []byte(`{}`)
	}
	return s.repo.CreateEntityContact(ctx, repository.CreateEntityContactParams{
		TenantID: tenantID, EntityID: entityID,
		Type: contactType, Label: label, Value: value, Metadata: metadata,
	})
}

// AddNote adds a note to an entity.
func (s *Entities) AddNote(ctx context.Context, tenantID string, entityID pgtype.UUID, userID, noteType, body string) (repository.ErpEntityNote, error) {
	return s.repo.CreateEntityNote(ctx, repository.CreateEntityNoteParams{
		TenantID: tenantID, EntityID: entityID,
		UserID: userID, Type: noteType, Body: body,
	})
}

// AddDocument registers a document for an entity.
func (s *Entities) AddDocument(ctx context.Context, tenantID string, entityID pgtype.UUID, name, docType, fileKey string) (repository.ErpEntityDocument, error) {
	return s.repo.CreateEntityDocument(ctx, repository.CreateEntityDocumentParams{
		TenantID: tenantID, EntityID: entityID,
		Name: name, DocType: docType, FileKey: fileKey,
	})
}

func pgTextPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}
