package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/service"
)

// Devices handles all BigBrother HTTP endpoints.
type Devices struct {
	db          *pgxpool.Pool
	nc          *nats.Conn
	audit       *audit.Writer
	tenantSlug  string
	control     *Control
}

// NewDevices creates a new BigBrother handler.
func NewDevices(db *pgxpool.Pool, nc *nats.Conn, auditWriter *audit.Writer, tenantSlug string) *Devices {
	plcSvc := service.NewPLCService(db, nc, auditWriter, tenantSlug)
	return &Devices{
		db:         db,
		nc:         nc,
		audit:      auditWriter,
		tenantSlug: tenantSlug,
		control:    NewControl(plcSvc),
	}
}

// Routes returns the chi router for BigBrother endpoints.
func (h *Devices) Routes() chi.Router {
	r := chi.NewRouter()

	// Device endpoints
	r.Get("/devices", h.ListDevices)
	r.Get("/devices/{id}", h.GetDevice)

	// PLC register endpoints
	r.Get("/devices/{id}/registers", h.control.ListRegisters)
	r.Post("/devices/{id}/registers/{addr}", h.control.WriteRegister)
	r.Post("/devices/{id}/registers/{addr}/approve", h.control.ApproveWrite)

	// Events + stats
	r.Get("/events", h.ListEvents)
	r.Get("/stats", h.GetStats)

	return r
}

// ListDevices returns all devices with optional filters.
func (h *Devices) ListDevices(w http.ResponseWriter, r *http.Request) {
	deviceType := r.URL.Query().Get("device_type")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	// TODO: implement with sqlc repository
	_ = deviceType
	_ = status
	_ = search

	rows, err := h.db.Query(r.Context(),
		`SELECT id, ip, mac, hostname, vendor, device_type, os, model, location, status, first_seen, last_seen
		 FROM bb_devices WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		 ORDER BY last_seen DESC LIMIT 100`, h.tenantSlug)
	if err != nil {
		slog.Error("list devices failed", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
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

// GetDevice returns a single device by ID.
func (h *Devices) GetDevice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"id is required"}`, http.StatusBadRequest)
		return
	}

	// TODO: implement full device detail with ports, capabilities, PLC registers
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":     id,
		"status": "stub — implement with sqlc repository",
	})
}

// ListEvents returns the event timeline.
func (h *Devices) ListEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"events": []any{},
		"total":  0,
	})
}

// GetStats returns network summary stats.
func (h *Devices) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"total_devices": 0,
		"online":        0,
		"offline":       0,
		"by_type":       map[string]int{},
		"last_scan":     nil,
	})
}
