// Package inventory provides device documentation and change detection.
package inventory

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Documenter generates structured device documentation from DB data.
type Documenter struct {
	db         *pgxpool.Pool
	tenantSlug string
}

// NewDocumenter creates an inventory documenter.
func NewDocumenter(db *pgxpool.Pool, tenantSlug string) *Documenter {
	return &Documenter{db: db, tenantSlug: tenantSlug}
}

// DeviceDoc is a complete device documentation sheet.
type DeviceDoc struct {
	ID         string         `json:"id"`
	IP         string         `json:"ip"`
	MAC        *string        `json:"mac,omitempty"`
	Hostname   *string        `json:"hostname,omitempty"`
	Vendor     *string        `json:"vendor,omitempty"`
	DeviceType string         `json:"device_type"`
	OS         *string        `json:"os,omitempty"`
	Model      *string        `json:"model,omitempty"`
	Location   *string        `json:"location,omitempty"`
	Status     string         `json:"status"`
	FirstSeen  string         `json:"first_seen"`
	LastSeen   string         `json:"last_seen"`
	Ports      []PortDoc      `json:"ports"`
	Caps       []CapDoc       `json:"capabilities"`
	Registers  []RegisterDoc  `json:"plc_registers,omitempty"`
	ComputerInfo *ComputerDoc `json:"computer_info,omitempty"`
}

// PortDoc is a documented open port.
type PortDoc struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Service  string `json:"service,omitempty"`
	Version  string `json:"version,omitempty"`
	State    string `json:"state"`
}

// CapDoc is a documented device capability.
type CapDoc struct {
	Capability string `json:"capability"`
	VerifiedAt string `json:"verified_at,omitempty"`
}

// RegisterDoc is a documented PLC register.
type RegisterDoc struct {
	Protocol   string  `json:"protocol"`
	Address    string  `json:"address"`
	Name       string  `json:"name,omitempty"`
	Value      string  `json:"last_value,omitempty"`
	SafetyTier string  `json:"safety_tier"`
	Writable   bool    `json:"writable"`
}

// ComputerDoc is the computer info section.
type ComputerDoc struct {
	OSVersion   string  `json:"os_version,omitempty"`
	CPU         string  `json:"cpu,omitempty"`
	RAMGB       float64 `json:"ram_gb,omitempty"`
	DiskTotalGB float64 `json:"disk_total_gb,omitempty"`
	DiskFreeGB  float64 `json:"disk_free_gb,omitempty"`
	LastScan    string  `json:"last_scan,omitempty"`
}

// GenerateDoc creates a complete documentation sheet for a device.
func (d *Documenter) GenerateDoc(ctx context.Context, deviceID string) (*DeviceDoc, error) {
	doc := &DeviceDoc{}

	// Device base info
	err := d.db.QueryRow(ctx,
		`SELECT id, ip::TEXT, mac::TEXT, hostname, vendor, device_type, os, model, location, status, first_seen, last_seen
		 FROM bb_devices WHERE id = $1 AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)`,
		deviceID, d.tenantSlug,
	).Scan(&doc.ID, &doc.IP, &doc.MAC, &doc.Hostname, &doc.Vendor, &doc.DeviceType,
		&doc.OS, &doc.Model, &doc.Location, &doc.Status, &doc.FirstSeen, &doc.LastSeen)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	// Ports
	rows, err := d.db.Query(ctx,
		`SELECT port, protocol, service, version, state FROM bb_ports
		 WHERE device_id = $1 AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)
		 ORDER BY port`, deviceID, d.tenantSlug)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p PortDoc
			rows.Scan(&p.Port, &p.Protocol, &p.Service, &p.Version, &p.State)
			doc.Ports = append(doc.Ports, p)
		}
	}
	if doc.Ports == nil {
		doc.Ports = []PortDoc{}
	}

	// Capabilities
	rows2, err := d.db.Query(ctx,
		`SELECT capability, verified_at FROM bb_capabilities
		 WHERE device_id = $1 AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)`,
		deviceID, d.tenantSlug)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var c CapDoc
			rows2.Scan(&c.Capability, &c.VerifiedAt)
			doc.Caps = append(doc.Caps, c)
		}
	}
	if doc.Caps == nil {
		doc.Caps = []CapDoc{}
	}

	// PLC registers (if device is a PLC)
	if doc.DeviceType == "plc" {
		rows3, err := d.db.Query(ctx,
			`SELECT protocol, address, name, last_value, safety_tier, writable
			 FROM bb_plc_registers
			 WHERE device_id = $1 AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)
			 ORDER BY protocol, address`, deviceID, d.tenantSlug)
		if err == nil {
			defer rows3.Close()
			for rows3.Next() {
				var r RegisterDoc
				rows3.Scan(&r.Protocol, &r.Address, &r.Name, &r.Value, &r.SafetyTier, &r.Writable)
				doc.Registers = append(doc.Registers, r)
			}
		}
	}

	// Computer info
	ci := &ComputerDoc{}
	err = d.db.QueryRow(ctx,
		`SELECT os_version, cpu, ram_gb, disk_total_gb, disk_free_gb, last_scan
		 FROM bb_computer_info
		 WHERE device_id = $1 AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)`,
		deviceID, d.tenantSlug,
	).Scan(&ci.OSVersion, &ci.CPU, &ci.RAMGB, &ci.DiskTotalGB, &ci.DiskFreeGB, &ci.LastScan)
	if err == nil {
		doc.ComputerInfo = ci
	}

	return doc, nil
}

// TopologyEntry is a node in the network topology view.
type TopologyEntry struct {
	ID         string  `json:"id"`
	IP         string  `json:"ip"`
	MAC        *string `json:"mac,omitempty"`
	Hostname   *string `json:"hostname,omitempty"`
	DeviceType string  `json:"device_type"`
	Status     string  `json:"status"`
	Vendor     *string `json:"vendor,omitempty"`
}

// GetTopology returns all devices for the network map view.
func (d *Documenter) GetTopology(ctx context.Context) ([]TopologyEntry, error) {
	rows, err := d.db.Query(ctx,
		`SELECT id, ip::TEXT, mac::TEXT, hostname, device_type, status, vendor
		 FROM bb_devices
		 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		 ORDER BY device_type, ip`, d.tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("get topology: %w", err)
	}
	defer rows.Close()

	var entries []TopologyEntry
	for rows.Next() {
		var e TopologyEntry
		rows.Scan(&e.ID, &e.IP, &e.MAC, &e.Hostname, &e.DeviceType, &e.Status, &e.Vendor)
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []TopologyEntry{}
	}
	return entries, nil
}

// NetworkStats holds aggregate network statistics.
type NetworkStats struct {
	TotalDevices int            `json:"total_devices"`
	Online       int            `json:"online"`
	Offline      int            `json:"offline"`
	ByType       map[string]int `json:"by_type"`
	LastScan     *string        `json:"last_scan"`
}

// GetStats returns aggregate network statistics.
func (d *Documenter) GetStats(ctx context.Context) (*NetworkStats, error) {
	stats := &NetworkStats{ByType: make(map[string]int)}

	// Count by status
	rows, err := d.db.Query(ctx,
		`SELECT status, COUNT(*) FROM bb_devices
		 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		 GROUP BY status`, d.tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		rows.Scan(&status, &count)
		stats.TotalDevices += count
		if status == "online" {
			stats.Online = count
		} else if status == "offline" {
			stats.Offline = count
		}
	}

	// Count by type
	rows2, err := d.db.Query(ctx,
		`SELECT device_type, COUNT(*) FROM bb_devices
		 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		 GROUP BY device_type`, d.tenantSlug)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var dtype string
			var count int
			rows2.Scan(&dtype, &count)
			stats.ByType[dtype] = count
		}
	}

	// Last scan
	var lastScan *string
	d.db.QueryRow(ctx,
		`SELECT details->>'duration_ms' FROM bb_events
		 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		   AND event_type = 'scan_completed'
		 ORDER BY created_at DESC LIMIT 1`, d.tenantSlug).Scan(&lastScan)
	stats.LastScan = lastScan

	return stats, nil
}
