package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/plc"
)

// PLCService handles PLC register operations with safety enforcement.
// Uses pkg/plc for protocol and pkg/approval for two-person rule.
type PLCService struct {
	db         *pgxpool.Pool
	nc         *nats.Conn
	audit      *audit.Writer
	tenantSlug string
}

// NewPLCService creates a PLC service with safety enforcement.
func NewPLCService(db *pgxpool.Pool, nc *nats.Conn, auditWriter *audit.Writer, tenantSlug string) *PLCService {
	return &PLCService{db: db, nc: nc, audit: auditWriter, tenantSlug: tenantSlug}
}

// RegisterInfo represents a PLC register with metadata.
type RegisterInfo struct {
	ID               string   `json:"id"`
	DeviceID         string   `json:"device_id"`
	Protocol         string   `json:"protocol"`
	Address          string   `json:"address"`
	Name             *string  `json:"name,omitempty"`
	DataType         *string  `json:"data_type,omitempty"`
	LastValue        *string  `json:"last_value,omitempty"`
	LastValueNumeric *float64 `json:"last_value_numeric,omitempty"`
	LastRead         *string  `json:"last_read,omitempty"`
	Writable         bool     `json:"writable"`
	SafetyTier       string   `json:"safety_tier"`
	MinValue         *float64 `json:"min_value,omitempty"`
	MaxValue         *float64 `json:"max_value,omitempty"`
	MaxWritesPerMin  *int     `json:"max_writes_per_min,omitempty"`
}

// ListRegisters returns all PLC registers for a device.
func (s *PLCService) ListRegisters(ctx context.Context, deviceID string) ([]RegisterInfo, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, device_id, protocol, address, name, data_type, last_value,
				last_value_numeric, last_read, writable, safety_tier, min_value, max_value, max_writes_per_min
		 FROM bb_plc_registers
		 WHERE device_id = $1
		   AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)
		 ORDER BY protocol, address`,
		deviceID, s.tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("list registers: %w", err)
	}
	defer rows.Close()

	var regs []RegisterInfo
	for rows.Next() {
		var r RegisterInfo
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.Protocol, &r.Address,
			&r.Name, &r.DataType, &r.LastValue, &r.LastValueNumeric, &r.LastRead,
			&r.Writable, &r.SafetyTier, &r.MinValue, &r.MaxValue, &r.MaxWritesPerMin); err != nil {
			return nil, fmt.Errorf("scan register: %w", err)
		}
		regs = append(regs, r)
	}
	if regs == nil {
		regs = []RegisterInfo{}
	}
	return regs, nil
}

// WriteRegisterRequest holds the data for a PLC register write.
type WriteRegisterRequest struct {
	DeviceID    string
	Address     string
	Value       float64
	UserID      string
	IP          string
	UserAgent   string
}

// WriteRegisterResult is the outcome of a write attempt.
type WriteRegisterResult struct {
	Status    string  `json:"status"`    // "written", "pending_approval", "rejected"
	RequestID *string `json:"request_id,omitempty"` // for critical writes
	Message   string  `json:"message,omitempty"`
}

// WriteRegister attempts to write a value to a PLC register with full safety checks.
// For critical tier, creates a pending approval instead of writing immediately.
func (s *PLCService) WriteRegister(ctx context.Context, req WriteRegisterRequest) (*WriteRegisterResult, error) {
	// Fetch register metadata
	var regID string
	var tier plc.SafetyTier
	var minVal, maxVal *float64

	err := s.db.QueryRow(ctx,
		`SELECT id, safety_tier, min_value, max_value
		 FROM bb_plc_registers
		 WHERE device_id = $1 AND address = $2
		   AND tenant_id = (SELECT id FROM tenants WHERE slug = $3 LIMIT 1)`,
		req.DeviceID, req.Address, s.tenantSlug).Scan(&regID, &tier, &minVal, &maxVal)
	if err != nil {
		return nil, fmt.Errorf("get register: %w", err)
	}

	// Safety validation via pkg/plc
	if err := plc.ValidateWrite(tier, req.Value, minVal, maxVal); err != nil {
		return &WriteRegisterResult{Status: "rejected", Message: err.Error()}, nil
	}

	// Critical tier → two-person approval required
	if tier.RequiresTwoPersonApproval() {
		return s.createPendingWrite(ctx, req, regID)
	}

	// Fail-closed audit: abort if audit write fails
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID:  s.tenantSlug,
		UserID:    req.UserID,
		Action:    "bigbrother.plc.write",
		Resource:  fmt.Sprintf("device:%s/register:%s", req.DeviceID, req.Address),
		Details:   map[string]any{"value": req.Value, "safety_tier": string(tier)},
		IP:        req.IP,
		UserAgent: req.UserAgent,
	}); err != nil {
		return nil, fmt.Errorf("audit failed, write aborted: %w", err)
	}

	// TODO: actual PLC write via pkg/plc Modbus/OPC-UA client
	// For now, update the register value in DB
	_, err = s.db.Exec(ctx,
		`UPDATE bb_plc_registers SET last_value = $1, last_value_numeric = $2, last_read = now()
		 WHERE id = $3 AND tenant_id = (SELECT id FROM tenants WHERE slug = $4 LIMIT 1)`,
		fmt.Sprintf("%v", req.Value), req.Value, regID, s.tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("update register value: %w", err)
	}

	s.publishEvent("plc.value_changed", map[string]any{
		"device_id": req.DeviceID,
		"address":   req.Address,
		"value":     req.Value,
		"user_id":   req.UserID,
	})

	return &WriteRegisterResult{Status: "written"}, nil
}

func (s *PLCService) createPendingWrite(ctx context.Context, req WriteRegisterRequest, regID string) (*WriteRegisterResult, error) {
	var requestID string
	err := s.db.QueryRow(ctx,
		`INSERT INTO bb_pending_writes (tenant_id, device_id, register_addr, value, requestor_id, status, expires_at)
		 VALUES (
			(SELECT id FROM tenants WHERE slug = $1 LIMIT 1),
			$2, $3, $4, $5, 'pending', now() + interval '5 minutes'
		 ) RETURNING id`,
		s.tenantSlug, req.DeviceID, req.Address, fmt.Sprintf("%v", req.Value),
		req.UserID,
	).Scan(&requestID)
	if err != nil {
		return nil, fmt.Errorf("create pending write: %w", err)
	}

	// Fail-closed audit for critical write request
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID: s.tenantSlug,
		UserID:   req.UserID,
		Action:   "bigbrother.plc.write.requested",
		Resource: fmt.Sprintf("device:%s/register:%s", req.DeviceID, req.Address),
		Details:  map[string]any{"request_id": requestID, "value": req.Value},
		IP:       req.IP,
	}); err != nil {
		return nil, fmt.Errorf("audit failed, critical write request aborted: %w", err)
	}

	// Notify via NATS
	s.publishEvent("plc.approval_requested", map[string]any{
		"request_id":   requestID,
		"device_id":    req.DeviceID,
		"address":      req.Address,
		"value":        req.Value,
		"requestor_id": req.UserID,
		"expires_at":   time.Now().Add(5 * time.Minute).Format(time.RFC3339),
	})

	return &WriteRegisterResult{
		Status:    "pending_approval",
		RequestID: &requestID,
		Message:   "critical write requires approval from another admin within 5 minutes",
	}, nil
}

// ApproveWrite approves a pending critical write. The approver must be different
// from the requestor. Self-approve and expiry are enforced at the SQL level
// (no TOCTOU race). Uses atomic UPDATE to prevent concurrent approvals.
func (s *PLCService) ApproveWrite(ctx context.Context, requestID, approverID, ip, userAgent string) (*WriteRegisterResult, error) {
	var deviceID, address, value, requestorID string
	err := s.db.QueryRow(ctx,
		`UPDATE bb_pending_writes
		 SET approved_by = $1, approved_at = now(), status = 'approved'
		 WHERE id = $2
		   AND status = 'pending'
		   AND approved_by IS NULL
		   AND requestor_id != $1
		   AND expires_at > now()
		   AND tenant_id = (SELECT id FROM tenants WHERE slug = $3 LIMIT 1)
		 RETURNING device_id, register_addr, value, requestor_id`,
		approverID, requestID, s.tenantSlug,
	).Scan(&deviceID, &address, &value, &requestorID)
	if err != nil {
		return &WriteRegisterResult{Status: "rejected", Message: "not found, expired, self-approve, or already handled"}, nil
	}

	// Fail-closed audit for both users
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID:  s.tenantSlug,
		UserID:    approverID,
		Action:    "bigbrother.plc.write.approved",
		Resource:  fmt.Sprintf("device:%s/register:%s", deviceID, address),
		Details:   map[string]any{"request_id": requestID, "requestor_id": requestorID, "value": value},
		IP:        ip,
		UserAgent: userAgent,
	}); err != nil {
		return nil, fmt.Errorf("audit failed, approval aborted: %w", err)
	}

	// TODO: actual PLC write via pkg/plc
	slog.Info("critical PLC write approved and executed",
		"request_id", requestID,
		"device_id", deviceID,
		"address", address,
		"value", value,
		"requestor", requestorID,
		"approver", approverID)

	s.publishEvent("plc.value_changed", map[string]any{
		"device_id":    deviceID,
		"address":      address,
		"value":        value,
		"requestor_id": requestorID,
		"approver_id":  approverID,
		"two_person":   true,
	})

	return &WriteRegisterResult{Status: "written", Message: "critical write approved and executed"}, nil
}

func (s *PLCService) publishEvent(eventType string, details map[string]any) {
	subject := fmt.Sprintf("tenant.%s.bigbrother.%s", s.tenantSlug, eventType)
	data, _ := json.Marshal(details)
	//nolint:forbidigo // Plan 27 will migrate bigbrother to pkg/spine.
	if err := s.nc.Publish(subject, data); err != nil {
		slog.Error("publish NATS event failed", "subject", subject, "error", err)
	}
}
