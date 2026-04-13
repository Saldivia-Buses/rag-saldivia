package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/httperr"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/pkg/remote"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/inventory"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/scanner"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/service"
)

// Devices handles all BigBrother HTTP endpoints.
type Devices struct {
	db          *pgxpool.Pool
	nc          *nats.Conn
	audit       *audit.Writer
	tenantSlug  string
	control     *Control
	credentials *Credentials
	remoteSvc   *service.RemoteService
	documenter  *inventory.Documenter
	scanLoop    *scanner.Loop
}

// NewDevices creates a new BigBrother handler.
func NewDevices(db *pgxpool.Pool, nc *nats.Conn, auditWriter *audit.Writer, tenantSlug string) *Devices {
	plcSvc := service.NewPLCService(db, nc, auditWriter, tenantSlug)
	// CredentialService and RemoteService are initialized later when Encryptor is available.
	// For now they are nil — endpoints will return 501 until configured.
	return &Devices{
		db:         db,
		nc:         nc,
		audit:      auditWriter,
		tenantSlug: tenantSlug,
		control:    NewControl(plcSvc),
		documenter: inventory.NewDocumenter(db, tenantSlug),
	}
}

// SetCredentialService configures the credential service (requires Encryptor).
func (h *Devices) SetCredentialService(credSvc *service.CredentialService) {
	h.credentials = NewCredentials(credSvc)
	h.remoteSvc = service.NewRemoteService(h.db, credSvc, h.audit, h.tenantSlug)
}

// SetScanLoop configures the scanner loop reference for scan control endpoints.
func (h *Devices) SetScanLoop(loop *scanner.Loop) {
	h.scanLoop = loop
}

// Routes returns the chi router for BigBrother endpoints.
// RBAC enforced per endpoint group via RequirePermission middleware.
func (h *Devices) Routes() chi.Router {
	r := chi.NewRouter()

	// Read endpoints — bigbrother.read
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("bigbrother.read"))
		r.Get("/devices", h.ListDevices)
		r.Get("/devices/{id}", h.GetDevice)
		r.Get("/topology", h.GetTopology)
		r.Get("/events", h.ListEvents)
		r.Get("/stats", h.GetStats)
	})

	// PLC read — bigbrother.plc.read
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("bigbrother.plc.read"))
		r.Get("/devices/{id}/registers", h.control.ListRegisters)
	})

	// PLC write — bigbrother.plc.write
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("bigbrother.plc.write"))
		r.Post("/devices/{id}/registers/{addr}", h.control.WriteRegister)
	})

	// Remote exec — bigbrother.exec
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("bigbrother.exec"))
		r.Post("/devices/{id}/exec", h.ExecCommand)
	})

	// Admin endpoints — bigbrother.admin
	r.Group(func(r chi.Router) {
		r.Use(sdamw.RequirePermission("bigbrother.admin"))
		r.Post("/devices/{id}/registers/{addr}/approve", h.control.ApproveWrite)
		r.Post("/credentials", h.StoreCredential)
		r.Get("/credentials", h.ListCredentials)
		r.Delete("/credentials/{id}", h.DeleteCredential)
		r.Post("/scan", h.TriggerScan)
		r.Post("/scan/mode", h.SetScanMode)
	})

	return r
}

// ExecCommand handles remote command execution.
func (h *Devices) ExecCommand(w http.ResponseWriter, r *http.Request) {
	if h.remoteSvc == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "remote exec not configured", http.StatusNotImplemented))
		return
	}

	deviceID := chi.URLParam(r, "id")
	r.Body = http.MaxBytesReader(w, r.Body, 1024)

	var body struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid body"))
		return
	}

	userID := r.Header.Get("X-User-ID")
	result, err := h.remoteSvc.Exec(r.Context(), service.ExecRequest{
		DeviceID:  deviceID,
		Command:   remote.CommandType(body.Command),
		UserID:    userID,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		httperr.WriteError(w, r, httperr.Wrap(err, httperr.CodeInvalidInput, "command execution failed", http.StatusBadRequest))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// StoreCredential handles credential creation.
func (h *Devices) StoreCredential(w http.ResponseWriter, r *http.Request) {
	if h.credentials == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "credentials not configured", http.StatusNotImplemented))
		return
	}
	h.credentials.Store(w, r)
}

// ListCredentials handles credential listing (metadata only).
func (h *Devices) ListCredentials(w http.ResponseWriter, r *http.Request) {
	if h.credentials == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "credentials not configured", http.StatusNotImplemented))
		return
	}
	h.credentials.List(w, r)
}

// DeleteCredential handles credential deletion.
func (h *Devices) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	if h.credentials == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "credentials not configured", http.StatusNotImplemented))
		return
	}
	h.credentials.Delete(w, r)
}

// ListDevices returns all devices with optional filters.
func (h *Devices) ListDevices(w http.ResponseWriter, r *http.Request) {
	deviceType := r.URL.Query().Get("device_type")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	query := `SELECT id, ip, mac, hostname, vendor, device_type, os, model, location, status, first_seen, last_seen
		 FROM bb_devices WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)`
	args := []any{h.tenantSlug}
	argN := 2

	if deviceType != "" {
		query += fmt.Sprintf(` AND device_type = $%d`, argN)
		args = append(args, deviceType)
		argN++
	}
	if status != "" {
		query += fmt.Sprintf(` AND status = $%d`, argN)
		args = append(args, status)
		argN++
	}
	if search != "" {
		query += fmt.Sprintf(` AND (hostname ILIKE $%d OR ip::TEXT ILIKE $%d OR vendor ILIKE $%d OR model ILIKE $%d)`, argN, argN, argN, argN)
		args = append(args, "%"+search+"%")
		argN++
	}
	query += ` ORDER BY last_seen DESC LIMIT 100`

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	defer rows.Close()

	type device struct {
		ID         string  `json:"id"`
		IP         string  `json:"ip"`
		MAC        *string `json:"mac,omitempty"`
		Hostname   *string `json:"hostname,omitempty"`
		Vendor     *string `json:"vendor,omitempty"`
		DeviceType string  `json:"device_type"`
		OS         *string `json:"os,omitempty"`
		Model      *string `json:"model,omitempty"`
		Location   *string `json:"location,omitempty"`
		Status     string  `json:"status"`
		FirstSeen  string  `json:"first_seen"`
		LastSeen   string  `json:"last_seen"`
	}

	var devices []device
	for rows.Next() {
		var d device
		if err := rows.Scan(&d.ID, &d.IP, &d.MAC, &d.Hostname, &d.Vendor, &d.DeviceType,
			&d.OS, &d.Model, &d.Location, &d.Status, &d.FirstSeen, &d.LastSeen); err != nil {
			slog.Error("scan device failed", "error", err)
			continue
		}
		devices = append(devices, d)
	}

	if devices == nil {
		devices = []device{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"devices": devices,
		"total":   len(devices),
	})
}

// GetDevice returns full device documentation.
func (h *Devices) GetDevice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httperr.WriteError(w, r, httperr.InvalidInput("id is required"))
		return
	}

	doc, err := h.documenter.GenerateDoc(r.Context(), id)
	if err != nil {
		httperr.WriteError(w, r, httperr.NotFound("device"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// GetTopology returns the network map.
func (h *Devices) GetTopology(w http.ResponseWriter, r *http.Request) {
	entries, err := h.documenter.GetTopology(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"devices": entries,
		"total":   len(entries),
	})
}

// ListEvents returns the event timeline.
func (h *Devices) ListEvents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		`SELECT id, device_id, event_type, details, created_at FROM bb_events
		 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		 ORDER BY created_at DESC LIMIT 100`, h.tenantSlug)
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}
	defer rows.Close()

	type event struct {
		ID        string `json:"id"`
		DeviceID  *string `json:"device_id,omitempty"`
		EventType string `json:"event_type"`
		Details   []byte `json:"details"`
		CreatedAt string `json:"created_at"`
	}

	var events []event
	for rows.Next() {
		var e event
		rows.Scan(&e.ID, &e.DeviceID, &e.EventType, &e.Details, &e.CreatedAt)
		events = append(events, e)
	}
	if events == nil {
		events = []event{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"events": events,
		"total":  len(events),
	})
}

// GetStats returns network summary stats.
func (h *Devices) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.documenter.GetStats(r.Context())
	if err != nil {
		httperr.WriteError(w, r, httperr.Internal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// TriggerScan triggers an immediate manual scan cycle.
func (h *Devices) TriggerScan(w http.ResponseWriter, r *http.Request) {
	if h.scanLoop == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "scanner not configured", http.StatusNotImplemented))
		return
	}
	h.scanLoop.Trigger()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "scan_triggered",
		"mode":   string(h.scanLoop.Mode()),
	})
}

// SetScanMode changes the scan mode at runtime.
func (h *Devices) SetScanMode(w http.ResponseWriter, r *http.Request) {
	if h.scanLoop == nil {
		httperr.WriteError(w, r, httperr.Wrap(nil, httperr.CodeInternal, "scanner not configured", http.StatusNotImplemented))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.WriteError(w, r, httperr.InvalidInput("invalid body"))
		return
	}

	mode := scanner.ScanMode(body.Mode)
	switch mode {
	case scanner.ModePassive, scanner.ModeActive, scanner.ModeFull:
		h.scanLoop.SetMode(mode)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "mode_changed",
			"mode":   body.Mode,
		})
	default:
		httperr.WriteError(w, r, httperr.InvalidInput("invalid mode, must be: passive, active, full"))
	}
}
