package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// WorkshopService is the interface the Workshop handler depends on.
type WorkshopService interface {
	ListCustomerVehicles(ctx context.Context, tenantID, ownerFilter string, limit, offset int) ([]repository.ListCustomerVehiclesRow, error)
	GetCustomerVehicle(ctx context.Context, id pgtype.UUID, tenantID string) (repository.ErpCustomerVehicle, error)
	CreateCustomerVehicle(ctx context.Context, p repository.CreateCustomerVehicleParams, userID, ip string) (repository.ErpCustomerVehicle, error)
	UpdateCustomerVehicle(ctx context.Context, p repository.UpdateCustomerVehicleParams, userID, ip string) error
	ListIncidentTypes(ctx context.Context, tenantID string) ([]repository.ErpVehicleIncidentType, error)
	CreateIncidentType(ctx context.Context, tenantID, name, userID, ip string) (repository.ErpVehicleIncidentType, error)
	ListVehicleIncidents(ctx context.Context, tenantID, vehicleFilter, statusFilter string, limit, offset int) ([]repository.ListVehicleIncidentsRow, error)
	CreateVehicleIncident(ctx context.Context, p repository.CreateVehicleIncidentParams, userID, ip string) (repository.ErpVehicleIncident, error)
	ResolveVehicleIncident(ctx context.Context, id pgtype.UUID, tenantID, userID, ip string) error
	GetKPIs(ctx context.Context, tenantID string) (repository.GetWorkshopKPIsRow, error)
}

type Workshop struct{ svc WorkshopService }

func NewWorkshop(svc WorkshopService) *Workshop { return &Workshop{svc: svc} }

func (h *Workshop) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.maintenance.read"))
		r.Get("/vehicles", h.ListVehicles)
		r.Get("/vehicles/{id}", h.GetVehicle)
		r.Get("/incident-types", h.ListIncidentTypes)
		r.Get("/incidents", h.ListIncidents)
		r.Get("/kpis", h.GetKPIs)
	})
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.maintenance.write"))
		r.Post("/vehicles", h.CreateVehicle)
		r.Patch("/vehicles/{id}", h.UpdateVehicle)
		r.Post("/incident-types", h.CreateIncidentType)
		r.Post("/incidents", h.CreateIncident)
		r.Patch("/incidents/{id}/resolve", h.ResolveIncident)
	})
	return r
}

// ─── Vehicles ──────────────────────────────────────────────────────────────────

func (h *Workshop) ListVehicles(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	vehicles, err := h.svc.ListCustomerVehicles(r.Context(), tenantSlug(r), r.URL.Query().Get("owner_id"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"vehicles": vehicles})
}

func (h *Workshop) GetVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	v, err := h.svc.GetCustomerVehicle(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("vehicle"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Workshop) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Plate               string  `json:"plate"`
		ChassisSerial       string  `json:"chassis_serial"`
		BodySerial          string  `json:"body_serial"`
		Brand               string  `json:"brand"`
		Color               string  `json:"color"`
		Destination         string  `json:"destination"`
		Observations        string  `json:"observations"`
		FuelType            string  `json:"fuel_type"`
		OwnerID             *string `json:"owner_id"`
		DriverID            *string `json:"driver_id"`
		ManufacturingUnitID *string `json:"manufacturing_unit_id"`
		InternalNumber      *int32  `json:"internal_number"`
		ModelYear           *int32  `json:"model_year"`
		SeatingCapacity     int32   `json:"seating_capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Brand == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("brand is required"))
		return
	}
	validFuel := map[string]bool{"diesel": true, "gasolina": true, "gnc": true, "electric": true, "hybrid": true}
	if body.FuelType == "" {
		body.FuelType = "diesel"
	}
	if !validFuel[body.FuelType] {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid fuel_type (diesel, gasolina, gnc, electric, hybrid)"))
		return
	}
	var intNum pgtype.Int4
	if body.InternalNumber != nil {
		intNum = pgtype.Int4{Int32: *body.InternalNumber, Valid: true}
	}
	var modelYear pgtype.Int4
	if body.ModelYear != nil {
		modelYear = pgtype.Int4{Int32: *body.ModelYear, Valid: true}
	}
	v, err := h.svc.CreateCustomerVehicle(r.Context(), repository.CreateCustomerVehicleParams{
		TenantID:            slug,
		OwnerID:             optUUID(body.OwnerID),
		DriverID:            optUUID(body.DriverID),
		ManufacturingUnitID: optUUID(body.ManufacturingUnitID),
		Plate:               body.Plate,
		ChassisSerial:       body.ChassisSerial,
		BodySerial:          body.BodySerial,
		InternalNumber:      intNum,
		Brand:               body.Brand,
		ModelYear:           modelYear,
		SeatingCapacity:     body.SeatingCapacity,
		FuelType:            body.FuelType,
		Color:               body.Color,
		Destination:         body.Destination,
		Observations:        body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Workshop) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		Plate           string `json:"plate"`
		ChassisSerial   string `json:"chassis_serial"`
		BodySerial      string `json:"body_serial"`
		Brand           string `json:"brand"`
		Color           string `json:"color"`
		Destination     string `json:"destination"`
		Observations    string `json:"observations"`
		FuelType        string `json:"fuel_type"`
		ModelYear       *int32 `json:"model_year"`
		SeatingCapacity int32  `json:"seating_capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.FuelType == "" {
		body.FuelType = "diesel"
	}
	validFuel := map[string]bool{"diesel": true, "gasolina": true, "gnc": true, "electric": true, "hybrid": true}
	if !validFuel[body.FuelType] {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid fuel_type (diesel, gasolina, gnc, electric, hybrid)"))
		return
	}
	var modelYear pgtype.Int4
	if body.ModelYear != nil {
		modelYear = pgtype.Int4{Int32: *body.ModelYear, Valid: true}
	}
	if err := h.svc.UpdateCustomerVehicle(r.Context(), repository.UpdateCustomerVehicleParams{
		ID:              id,
		TenantID:        slug,
		Plate:           body.Plate,
		ChassisSerial:   body.ChassisSerial,
		BodySerial:      body.BodySerial,
		Brand:           body.Brand,
		ModelYear:       modelYear,
		SeatingCapacity: body.SeatingCapacity,
		FuelType:        body.FuelType,
		Color:           body.Color,
		Destination:     body.Destination,
		Observations:    body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "vehicle not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("vehicle"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Incident Types ────────────────────────────────────────────────────────────

func (h *Workshop) ListIncidentTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.svc.ListIncidentTypes(r.Context(), tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"incident_types": types})
}

func (h *Workshop) CreateIncidentType(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Name == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("name is required"))
		return
	}
	t, err := h.svc.CreateIncidentType(r.Context(), slug, body.Name, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(t)
}

// ─── Incidents ─────────────────────────────────────────────────────────────────

func (h *Workshop) ListIncidents(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	q := r.URL.Query()
	incidents, err := h.svc.ListVehicleIncidents(r.Context(), tenantSlug(r), q.Get("vehicle_id"), q.Get("status"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"incidents": incidents})
}

func (h *Workshop) CreateIncident(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		VehicleID      string  `json:"vehicle_id"`
		IncidentDate   string  `json:"incident_date"`
		Location       string  `json:"location"`
		Responsible    string  `json:"responsible"`
		Notes          string  `json:"notes"`
		IncidentTypeID *string `json:"incident_type_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.VehicleID == "" || body.IncidentDate == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("vehicle_id and incident_date are required"))
		return
	}
	vehicleID, err := parseUUID(body.VehicleID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid vehicle_id"))
		return
	}
	inc, err := h.svc.CreateVehicleIncident(r.Context(), repository.CreateVehicleIncidentParams{
		TenantID:       slug,
		VehicleID:      vehicleID,
		IncidentTypeID: optUUID(body.IncidentTypeID),
		IncidentDate:   pgDate(body.IncidentDate),
		Location:       body.Location,
		Responsible:    body.Responsible,
		Notes:          body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(inc)
}

func (h *Workshop) ResolveIncident(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	if err := h.svc.ResolveVehicleIncident(r.Context(), id, slug, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "incident not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("incident"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── KPIs ──────────────────────────────────────────────────────────────────────

func (h *Workshop) GetKPIs(w http.ResponseWriter, r *http.Request) {
	kpis, err := h.svc.GetKPIs(r.Context(), tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(kpis)
}
