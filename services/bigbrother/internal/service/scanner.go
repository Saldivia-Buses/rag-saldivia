package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/fingerprint"
	"github.com/Camionerou/rag-saldivia/services/bigbrother/internal/scanner"
)

// maxNewDevicesPerScan prevents MAC spoofing table bloat.
const maxNewDevicesPerScan = 50

// Scanner orchestrates scan results persistence and event publishing.
type Scanner struct {
	db         *pgxpool.Pool
	nc         *nats.Conn
	tenantSlug string
}

// NewScanner creates a scanner service.
func NewScanner(db *pgxpool.Pool, nc *nats.Conn, tenantSlug string) *Scanner {
	return &Scanner{db: db, nc: nc, tenantSlug: tenantSlug}
}

// ProcessResults compares scan results with DB state and persists changes.
// This is the callback for scanner.Loop.
func (s *Scanner) ProcessResults(ctx context.Context, devices []scanner.Device) {
	start := time.Now()
	var newCount, updatedCount, offlineCount int

	for i, dev := range devices {
		if i >= maxNewDevicesPerScan+len(devices) {
			break // safety cap (only applies to truly new devices below)
		}

		// Enrich with vendor
		if dev.Vendor == "" {
			dev.Vendor = fingerprint.LookupVendor(dev.MAC)
		}

		// Upsert device
		isNew, err := s.upsertDevice(ctx, dev)
		if err != nil {
			slog.Error("upsert device failed", "ip", dev.IP, "error", err)
			continue
		}

		if isNew {
			newCount++
			if newCount > maxNewDevicesPerScan {
				slog.Warn("max new devices per scan reached, skipping rest", "max", maxNewDevicesPerScan)
				break
			}
			s.publishEvent("device.discovered", map[string]any{
				"ip":     dev.IP.String(),
				"mac":    dev.MAC.String(),
				"vendor": dev.Vendor,
			})
		} else {
			updatedCount++
		}
	}

	// Mark devices not seen in this scan as offline
	offlineCount, _ = s.markOffline(ctx, devices)

	duration := time.Since(start)
	slog.Info("scan results processed",
		"new", newCount, "updated", updatedCount, "offline", offlineCount,
		"duration", duration)

	s.publishEvent("scan.completed", map[string]any{
		"total":       len(devices),
		"new":         newCount,
		"online":      updatedCount + newCount,
		"offline":     offlineCount,
		"duration_ms": duration.Milliseconds(),
	})
}

func (s *Scanner) upsertDevice(ctx context.Context, dev scanner.Device) (isNew bool, err error) {
	var id string
	// Try upsert by MAC first (primary identifier)
	if dev.MAC != nil {
		err = s.db.QueryRow(ctx,
			`INSERT INTO bb_devices (tenant_id, ip, mac, hostname, vendor, status, last_seen)
			 VALUES (
				(SELECT id FROM tenants WHERE slug = $1 LIMIT 1),
				$2, $3, $4, $5, 'online', now()
			 )
			 ON CONFLICT (tenant_id, mac) WHERE mac IS NOT NULL
			 DO UPDATE SET ip = EXCLUDED.ip, hostname = COALESCE(EXCLUDED.hostname, bb_devices.hostname),
				vendor = COALESCE(EXCLUDED.vendor, bb_devices.vendor),
				status = 'online', last_seen = now()
			 RETURNING id, (xmax = 0) AS is_new`,
			s.tenantSlug, dev.IP.String(), dev.MAC.String(), nilIfEmpty(dev.Hostname), nilIfEmpty(dev.Vendor),
		).Scan(&id, &isNew)
	} else {
		// No MAC — upsert by IP
		err = s.db.QueryRow(ctx,
			`INSERT INTO bb_devices (tenant_id, ip, hostname, vendor, status, last_seen)
			 VALUES (
				(SELECT id FROM tenants WHERE slug = $1 LIMIT 1),
				$2, $3, $4, 'online', now()
			 )
			 ON CONFLICT (tenant_id, ip)
			 DO UPDATE SET hostname = COALESCE(EXCLUDED.hostname, bb_devices.hostname),
				vendor = COALESCE(EXCLUDED.vendor, bb_devices.vendor),
				status = 'online', last_seen = now()
			 RETURNING id, (xmax = 0) AS is_new`,
			s.tenantSlug, dev.IP.String(), nilIfEmpty(dev.Hostname), nilIfEmpty(dev.Vendor),
		).Scan(&id, &isNew)
	}
	return isNew, err
}

func (s *Scanner) markOffline(ctx context.Context, seen []scanner.Device) (int, error) {
	if len(seen) == 0 {
		return 0, nil
	}

	// Build list of seen MACs
	seenMACs := make([]string, 0, len(seen))
	for _, d := range seen {
		if d.MAC != nil {
			seenMACs = append(seenMACs, d.MAC.String())
		}
	}

	if len(seenMACs) == 0 {
		return 0, nil
	}

	// Mark devices not seen as offline (only if they were online)
	tag, err := s.db.Exec(ctx,
		`UPDATE bb_devices SET status = 'offline', last_seen = now()
		 WHERE tenant_id = (SELECT id FROM tenants WHERE slug = $1 LIMIT 1)
		   AND status = 'online'
		   AND mac IS NOT NULL
		   AND mac::TEXT NOT IN (SELECT unnest($2::TEXT[]))`,
		s.tenantSlug, seenMACs,
	)
	if err != nil {
		return 0, fmt.Errorf("mark offline: %w", err)
	}

	count := int(tag.RowsAffected())
	if count > 0 {
		slog.Info("devices marked offline", "count", count)
	}
	return count, nil
}

func (s *Scanner) publishEvent(eventType string, details map[string]any) {
	subject := fmt.Sprintf("tenant.%s.bigbrother.%s", s.tenantSlug, eventType)
	data, err := json.Marshal(details)
	if err != nil {
		slog.Error("marshal NATS event failed", "error", err)
		return
	}
	if err := s.nc.Publish(subject, data); err != nil {
		slog.Error("publish NATS event failed", "subject", subject, "error", err)
	}
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
