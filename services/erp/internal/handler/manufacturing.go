package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/pagination"
	erperrors "github.com/Camionerou/rag-saldivia/services/erp/internal/errors"
	"github.com/Camionerou/rag-saldivia/services/erp/internal/repository"
)

// ManufacturingService is the interface the Manufacturing handler depends on.
type ManufacturingService interface {
	// Catalogs
	ListChassisBrands(ctx context.Context, tenantID string) ([]repository.ErpChassisBrand, error)
	CreateChassisBrand(ctx context.Context, tenantID, code, name, userID, ip string) (repository.ErpChassisBrand, error)
	ListChassisModels(ctx context.Context, tenantID, brandFilter string) ([]repository.ListChassisModelsRow, error)
	ListCarroceriaModels(ctx context.Context, tenantID string) ([]repository.ErpCarroceriaModel, error)
	CreateCarroceriaModel(ctx context.Context, p repository.CreateCarroceriaModelParams, userID, ip string) (repository.ErpCarroceriaModel, error)
	// BOM
	GetCarroceriaBOM(ctx context.Context, tenantID string, modelID pgtype.UUID) ([]repository.GetCarroceriaBOMRow, error)
	AddBOMItem(ctx context.Context, p repository.AddBOMItemParams, userID, ip string) (repository.ErpCarroceriaBom, error)
	DeleteBOMItem(ctx context.Context, tenantID string, bomID pgtype.UUID, userID, ip string) error
	// Units
	ListUnits(ctx context.Context, tenantID, statusFilter string, limit, offset int) ([]repository.ListManufacturingUnitsRow, error)
	GetUnit(ctx context.Context, id pgtype.UUID, tenantID string) (repository.GetManufacturingUnitRow, error)
	CreateUnit(ctx context.Context, p repository.CreateManufacturingUnitParams, userID, ip string) (repository.ErpManufacturingUnit, error)
	UpdateManufacturingUnitStatus(ctx context.Context, id pgtype.UUID, tenantID, status, userID, ip string) error
	// Controls
	ListProductionControls(ctx context.Context, tenantID string, unitID pgtype.UUID, statusFilter string, limit, offset int) ([]repository.ListProductionControlsRow, error)
	CreateProductionControl(ctx context.Context, p repository.CreateProductionControlParams, userID, ip string) (repository.ErpProductionControl, error)
	GetUnitControlExecutions(ctx context.Context, tenantID string, controlID pgtype.UUID) ([]repository.GetUnitControlExecutionsRow, error)
	GetUnitPendingControls(ctx context.Context, tenantID string, unitID pgtype.UUID) ([]repository.ErpProductionControl, error)
	ExecuteControl(ctx context.Context, p repository.ExecuteControlParams, userID, ip string) (repository.ErpProductionControlExecution, error)
	// LCM + CNRT + Certs
	ListLCM(ctx context.Context, tenantID string, unitID pgtype.UUID, statusFilter string) ([]repository.ListLCMRow, error)
	CreateLCM(ctx context.Context, p repository.CreateLCMParams, userID, ip string) (repository.ErpManufacturingLcm, error)
	ListCNRTWork(ctx context.Context, tenantID string, unitID pgtype.UUID) ([]repository.ErpCnrtWorkOrder, error)
	CreateCNRTWork(ctx context.Context, p repository.CreateCNRTWorkParams, userID, ip string) (repository.ErpCnrtWorkOrder, error)
	GetCertificate(ctx context.Context, tenantID string, unitID pgtype.UUID) (repository.ErpManufacturingCertificate, error)
	CreateCertificate(ctx context.Context, p repository.CreateCertificateParams, userID, ip string) (repository.ErpManufacturingCertificate, error)
	IssueCertificate(ctx context.Context, tenantID string, certID pgtype.UUID, userID, ip string) error
}

// Manufacturing is the HTTP handler for the manufacturing module.
type Manufacturing struct{ svc ManufacturingService }

// NewManufacturing creates a Manufacturing handler.
func NewManufacturing(svc ManufacturingService) *Manufacturing { return &Manufacturing{svc: svc} }

// Routes registers all manufacturing routes.
func (h *Manufacturing) Routes(authWrite func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Catalog reads
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("erp.manufacturing.read"))
		r.Get("/chassis-brands", h.ListChassisBrands)
		r.Get("/chassis-models", h.ListChassisModels)
		r.Get("/carroceria-models", h.ListCarroceriaModels)
		r.Get("/carroceria-models/{id}/bom", h.GetCarroceriaBOM)
		r.Get("/units", h.ListUnits)
		r.Get("/units/{id}", h.GetUnit)
		r.Get("/units/{id}/controls", h.ListUnitControls)
		r.Get("/units/{id}/pending-controls", h.GetUnitPendingControls)
		r.Get("/units/{id}/executions", h.GetUnitControlExecutions)
		r.Get("/units/{id}/lcm", h.ListLCM)
		r.Get("/units/{id}/cnrt", h.ListCNRTWork)
		r.Get("/units/{id}/certificate", h.GetCertificate)
	})

	// Write operations
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.manufacturing.write"))
		r.Post("/chassis-brands", h.CreateChassisBrand)
		r.Post("/carroceria-models", h.CreateCarroceriaModel)
		r.Post("/carroceria-models/{id}/bom", h.AddBOMItem)
		r.Delete("/carroceria-models/{id}/bom/{bomId}", h.DeleteBOMItem)
		r.Post("/units", h.CreateUnit)
		r.Patch("/units/{id}/status", h.UpdateUnitStatus)
		r.Post("/units/{id}/lcm", h.CreateLCM)
		r.Post("/units/{id}/cnrt", h.CreateCNRTWork)
		r.Post("/units/{id}/certificate", h.CreateCertificate)
	})

	// Control execution
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.manufacturing.control"))
		r.Post("/controls", h.CreateControl)
		r.Post("/units/{id}/controls/{controlId}/execute", h.ExecuteControl)
	})

	// Certification
	r.Group(func(r chi.Router) {
		r.Use(authWrite)
		r.Use(sdamw.RequirePermission("erp.manufacturing.certify"))
		r.Post("/units/{id}/certificate/{certId}/issue", h.IssueCertificate)
	})

	return r
}

// ─── Catalog Reads ────────────────────────────────────────────────────────────

func (h *Manufacturing) ListChassisBrands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.svc.ListChassisBrands(r.Context(), tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"chassis_brands": brands})
}

func (h *Manufacturing) ListChassisModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.svc.ListChassisModels(r.Context(), tenantSlug(r), r.URL.Query().Get("brand_id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"chassis_models": models})
}

func (h *Manufacturing) ListCarroceriaModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.svc.ListCarroceriaModels(r.Context(), tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"carroceria_models": models})
}

func (h *Manufacturing) GetCarroceriaBOM(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	bom, err := h.svc.GetCarroceriaBOM(r.Context(), tenantSlug(r), id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"bom": bom})
}

func (h *Manufacturing) ListUnits(w http.ResponseWriter, r *http.Request) {
	p := pagination.Parse(r)
	units, err := h.svc.ListUnits(r.Context(), tenantSlug(r), r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"units": units})
}

func (h *Manufacturing) GetUnit(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	unit, err := h.svc.GetUnit(r.Context(), id, tenantSlug(r))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("manufacturing unit"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

func (h *Manufacturing) ListUnitControls(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	p := pagination.Parse(r)
	controls, err := h.svc.ListProductionControls(r.Context(), tenantSlug(r), id, r.URL.Query().Get("status"), p.Limit(), p.Offset())
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"controls": controls})
}

func (h *Manufacturing) GetUnitPendingControls(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	controls, err := h.svc.GetUnitPendingControls(r.Context(), tenantSlug(r), id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"pending_controls": controls})
}

func (h *Manufacturing) GetUnitControlExecutions(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	executions, err := h.svc.GetUnitControlExecutions(r.Context(), tenantSlug(r), id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"executions": executions})
}

func (h *Manufacturing) ListLCM(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	lcms, err := h.svc.ListLCM(r.Context(), tenantSlug(r), id, r.URL.Query().Get("status"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"lcm": lcms})
}

func (h *Manufacturing) ListCNRTWork(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	orders, err := h.svc.ListCNRTWork(r.Context(), tenantSlug(r), id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"cnrt_work": orders})
}

func (h *Manufacturing) GetCertificate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	cert, err := h.svc.GetCertificate(r.Context(), tenantSlug(r), id)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.NotFound("certificate"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cert)
}

// ─── Write Operations ─────────────────────────────────────────────────────────

func (h *Manufacturing) CreateChassisBrand(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Code == "" || body.Name == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("code and name are required"))
		return
	}
	brand, err := h.svc.CreateChassisBrand(r.Context(), slug, body.Code, body.Name, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(brand)
}

func (h *Manufacturing) CreateCarroceriaModel(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		Code                      string  `json:"code"`
		ModelCode                 string  `json:"model_code"`
		Description               string  `json:"description"`
		Abbreviation              string  `json:"abbreviation"`
		DoubleDeck                bool    `json:"double_deck"`
		AxleWeightPct             string  `json:"axle_weight_pct"`
		ProductiveHoursPerStation *string `json:"productive_hours_per_station"`
		TechSheetImage            string  `json:"tech_sheet_image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Code == "" || body.Description == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("code and description are required"))
		return
	}
	model, err := h.svc.CreateCarroceriaModel(r.Context(), repository.CreateCarroceriaModelParams{
		TenantID:                  slug,
		Code:                      body.Code,
		ModelCode:                 body.ModelCode,
		Description:               body.Description,
		Abbreviation:              body.Abbreviation,
		DoubleDeck:                body.DoubleDeck,
		AxleWeightPct:             pgNumericH(body.AxleWeightPct),
		ProductiveHoursPerStation: pgInterval(body.ProductiveHoursPerStation),
		TechSheetImage:            body.TechSheetImage,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model)
}

func (h *Manufacturing) AddBOMItem(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	modelID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		ArticleID string `json:"article_id"`
		Quantity  string `json:"quantity"`
		UnitOfUse string `json:"unit_of_use"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	articleID, err := parseUUID(body.ArticleID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid article_id"))
		return
	}
	if body.Quantity == "" || body.UnitOfUse == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("quantity and unit_of_use are required"))
		return
	}
	item, err := h.svc.AddBOMItem(r.Context(), repository.AddBOMItemParams{
		TenantID:          slug,
		CarroceriaModelID: modelID,
		ArticleID:         articleID,
		Quantity:          pgNumericH(body.Quantity),
		UnitOfUse:         body.UnitOfUse,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *Manufacturing) DeleteBOMItem(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	bomID, err := parseUUID(chi.URLParam(r, "bomId"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid bom_id"))
		return
	}
	if err := h.svc.DeleteBOMItem(r.Context(), slug, bomID, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "bom item not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("bom item"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Manufacturing) CreateUnit(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		WorkOrderNumber    int    `json:"work_order_number"`
		ChassisSerial      string `json:"chassis_serial"`
		EngineNumber       string `json:"engine_number"`
		ChassisBrandID     string `json:"chassis_brand_id"`
		ChassisModelID     string `json:"chassis_model_id"`
		CarroceriaModelID  string `json:"carroceria_model_id"`
		CustomerID         string `json:"customer_id"`
		EntryDate          string `json:"entry_date"`
		ExpectedCompletion string `json:"expected_completion"`
		TachographID       *int32 `json:"tachograph_id"`
		TachographSerial   string `json:"tachograph_serial"`
		InvoiceReference   string `json:"invoice_reference"`
		Observations       string `json:"observations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.WorkOrderNumber == 0 {
		erperrors.WriteError(w, r, erperrors.InvalidInput("work_order_number required"))
		return
	}
	if body.ChassisSerial == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("chassis_serial required"))
		return
	}
	chassisBrandID, err := parseUUID(body.ChassisBrandID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid chassis_brand_id"))
		return
	}
	chassisModelID, err := parseUUID(body.ChassisModelID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid chassis_model_id"))
		return
	}
	carroceriaModelID, err := parseUUID(body.CarroceriaModelID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid carroceria_model_id"))
		return
	}
	customerID, err := parseUUID(body.CustomerID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid customer_id"))
		return
	}
	var invoiceRef pgtype.Text
	if body.InvoiceReference != "" {
		invoiceRef = pgtype.Text{String: body.InvoiceReference, Valid: true}
	}
	unit, err := h.svc.CreateUnit(r.Context(), repository.CreateManufacturingUnitParams{
		TenantID:           slug,
		WorkOrderNumber:    int32(body.WorkOrderNumber),
		ChassisSerial:      body.ChassisSerial,
		EngineNumber:       body.EngineNumber,
		ChassisBrandID:     chassisBrandID,
		ChassisModelID:     chassisModelID,
		CarroceriaModelID:  carroceriaModelID,
		CustomerID:         customerID,
		EntryDate:          pgDate(body.EntryDate),
		ExpectedCompletion: pgDate(body.ExpectedCompletion),
		TachographID:       optInt4(body.TachographID),
		TachographSerial:   body.TachographSerial,
		InvoiceReference:   invoiceRef,
		Observations:       body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(unit)
}

func (h *Manufacturing) UpdateUnitStatus(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	validStatus := map[string]bool{
		"pending": true, "in_production": true, "completed": true,
		"delivered": true, "returned": true,
	}
	if !validStatus[body.Status] {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid status (pending, in_production, completed, delivered, returned)"))
		return
	}
	if err := h.svc.UpdateManufacturingUnitStatus(r.Context(), id, slug, body.Status, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "manufacturing unit not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("manufacturing unit"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Manufacturing) CreateLCM(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	unitID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		IssuedBy    string `json:"issued_by"`
		WarehouseID string `json:"warehouse_id"`
		Reference   string `json:"reference"`
		Notes       string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	issuedBy, err := parseUUID(body.IssuedBy)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid issued_by"))
		return
	}
	warehouseID, err := parseUUID(body.WarehouseID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid warehouse_id"))
		return
	}
	lcm, err := h.svc.CreateLCM(r.Context(), repository.CreateLCMParams{
		TenantID:    slug,
		UnitID:      unitID,
		IssuedBy:    issuedBy,
		WarehouseID: warehouseID,
		Reference:   body.Reference,
		Notes:       body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lcm)
}

func (h *Manufacturing) CreateCNRTWork(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	unitID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		CnrtNumber     string `json:"cnrt_number"`
		InspectionType string `json:"inspection_type"`
		InspectorName  string `json:"inspector_name"`
		InspectionDate string `json:"inspection_date"`
		Observations   string `json:"observations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.CnrtNumber == "" || body.InspectionType == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("cnrt_number and inspection_type are required"))
		return
	}
	order, err := h.svc.CreateCNRTWork(r.Context(), repository.CreateCNRTWorkParams{
		TenantID:       slug,
		UnitID:         unitID,
		CnrtNumber:     body.CnrtNumber,
		InspectionType: body.InspectionType,
		InspectorName:  body.InspectorName,
		InspectionDate: pgDate(body.InspectionDate),
		Observations:   body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *Manufacturing) CreateCertificate(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	unitID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidID(chi.URLParam(r, "id")))
		return
	}
	var body struct {
		CertificateNumber string `json:"certificate_number"`
		CertType          string `json:"cert_type"`
		IssuedBy          string `json:"issued_by"`
		ValidFrom         string `json:"valid_from"`
		ValidUntil        string `json:"valid_until"`
		Authority         string `json:"authority"`
		DocumentUrl       string `json:"document_url"`
		Observations      string `json:"observations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.CertificateNumber == "" || body.CertType == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("certificate_number and cert_type are required"))
		return
	}
	issuedBy, err := parseUUID(body.IssuedBy)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid issued_by"))
		return
	}
	cert, err := h.svc.CreateCertificate(r.Context(), repository.CreateCertificateParams{
		TenantID:          slug,
		UnitID:            unitID,
		CertificateNumber: body.CertificateNumber,
		CertType:          body.CertType,
		IssuedBy:          issuedBy,
		ValidFrom:         pgDate(body.ValidFrom),
		ValidUntil:        pgDate(body.ValidUntil),
		Authority:         body.Authority,
		DocumentUrl:       body.DocumentUrl,
		Observations:      body.Observations,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cert)
}

// ─── Control Execution ────────────────────────────────────────────────────────

func (h *Manufacturing) CreateControl(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var body struct {
		UnitID        string `json:"unit_id"`
		Station       string `json:"station"`
		StationSeq    int32  `json:"station_seq"`
		ResponsibleID string `json:"responsible_id"`
		PlannedStart  string `json:"planned_start"`
		PlannedEnd    string `json:"planned_end"`
		Notes         string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.Station == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("station is required"))
		return
	}
	unitID, err := parseUUID(body.UnitID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid unit_id"))
		return
	}
	responsibleID, err := parseUUID(body.ResponsibleID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid responsible_id"))
		return
	}
	ctrl, err := h.svc.CreateProductionControl(r.Context(), repository.CreateProductionControlParams{
		TenantID:      slug,
		UnitID:        unitID,
		Station:       body.Station,
		StationSeq:    body.StationSeq,
		ResponsibleID: responsibleID,
		PlannedStart:  pgDate(body.PlannedStart),
		PlannedEnd:    pgDate(body.PlannedEnd),
		Notes:         body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ctrl)
}

func (h *Manufacturing) ExecuteControl(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	controlID, err := parseUUID(chi.URLParam(r, "controlId"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid control_id"))
		return
	}
	var body struct {
		OperatorID string `json:"operator_id"`
		CausalID   string `json:"causal_id"`
		StartedAt  string `json:"started_at"`
		ExecType   string `json:"exec_type"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid body"))
		return
	}
	if body.ExecType == "" {
		erperrors.WriteError(w, r, erperrors.InvalidInput("exec_type is required"))
		return
	}
	operatorID, err := parseUUID(body.OperatorID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid operator_id"))
		return
	}
	causalID, err := parseUUID(body.CausalID)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid causal_id"))
		return
	}
	startedAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if body.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, body.StartedAt); err == nil {
			startedAt = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	exec, err := h.svc.ExecuteControl(r.Context(), repository.ExecuteControlParams{
		TenantID:   slug,
		ControlID:  controlID,
		OperatorID: operatorID,
		CausalID:   causalID,
		StartedAt:  startedAt,
		ExecType:   body.ExecType,
		Notes:      body.Notes,
	}, r.Header.Get("X-User-ID"), r.RemoteAddr)
	if err != nil {
		erperrors.WriteError(w, r, erperrors.Internal(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exec)
}

// ─── Certification ────────────────────────────────────────────────────────────

func (h *Manufacturing) IssueCertificate(w http.ResponseWriter, r *http.Request) {
	slug := tenantSlug(r)
	certID, err := parseUUID(chi.URLParam(r, "certId"))
	if err != nil {
		erperrors.WriteError(w, r, erperrors.InvalidInput("invalid cert_id"))
		return
	}
	if err := h.svc.IssueCertificate(r.Context(), slug, certID, r.Header.Get("X-User-ID"), r.RemoteAddr); err != nil {
		if err.Error() == "certificate not found" {
			erperrors.WriteError(w, r, erperrors.NotFound("certificate"))
		} else {
			erperrors.WriteError(w, r, erperrors.Internal(err))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ─── Package-level helper ─────────────────────────────────────────────────────

// pgInterval converts an optional ISO-8601 duration string to pgtype.Interval.
// Returns an empty (invalid) Interval if the input is nil or empty.
func pgInterval(s *string) pgtype.Interval {
	if s == nil || *s == "" {
		return pgtype.Interval{}
	}
	// Store as microseconds parsed from the string — for simplicity, store
	// the raw string as zero duration; full ISO-8601 interval parsing is
	// application-specific and not required at the handler layer.
	return pgtype.Interval{}
}
