package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	"github.com/Camionerou/rag-saldivia/pkg/remote"
)

// RemoteService handles remote command execution with allowlist enforcement.
// Uses pkg/remote for transport, adds policy (allowlist, audit, sanitization).
type RemoteService struct {
	db         *pgxpool.Pool
	credSvc    *CredentialService
	audit      *audit.Writer
	tenantSlug string
}

// NewRemoteService creates a remote execution service.
func NewRemoteService(db *pgxpool.Pool, credSvc *CredentialService, auditWriter *audit.Writer, tenantSlug string) *RemoteService {
	return &RemoteService{db: db, credSvc: credSvc, audit: auditWriter, tenantSlug: tenantSlug}
}

// ExecRequest holds the data for a remote command execution.
type ExecRequest struct {
	DeviceID  string
	Command   remote.CommandType
	UserID    string
	IP        string
	UserAgent string
}

// ExecResponse holds the result of a remote execution.
type ExecResponse struct {
	Output   string `json:"output"`
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command"`
}

// Exec runs a predefined command on a remote device.
// Only enum-based commands are allowed — no custom strings.
func (s *RemoteService) Exec(ctx context.Context, req ExecRequest) (*ExecResponse, error) {
	// Validate command is in allowlist
	if !remote.IsValidCommand(req.Command) {
		return nil, fmt.Errorf("invalid command: %q — only predefined commands allowed", req.Command)
	}

	// Get device info to determine OS type
	var ip, deviceType string
	err := s.db.QueryRow(ctx,
		`SELECT ip::TEXT, device_type FROM bb_devices
		 WHERE id = $1 AND tenant_id = (SELECT id FROM tenants WHERE slug = $2 LIMIT 1)`,
		req.DeviceID, s.tenantSlug).Scan(&ip, &deviceType)
	if err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	isWindows := deviceType == "workstation" // heuristic: refine with OS field later
	cmdStr, ok := remote.ResolveCommand(req.Command, isWindows)
	if !ok {
		return nil, fmt.Errorf("resolve command: %q", req.Command)
	}

	// Fail-closed audit before execution
	if err := s.audit.WriteStrict(ctx, audit.Entry{
		TenantID:  s.tenantSlug,
		UserID:    req.UserID,
		Action:    "bigbrother.exec",
		Resource:  fmt.Sprintf("device:%s", req.DeviceID),
		Details:   map[string]any{"command": string(req.Command), "resolved": cmdStr},
		IP:        req.IP,
		UserAgent: req.UserAgent,
	}); err != nil {
		return nil, fmt.Errorf("audit failed, exec aborted: %w", err)
	}

	// TODO: get credential from credential service, establish SSH/WinRM connection,
	// execute command, sanitize output. For now, return a placeholder.
	output := fmt.Sprintf("[stub] would execute %q on %s (%s)", cmdStr, ip, deviceType)
	output = remote.SanitizeOutput(output)

	// Record event
	s.db.Exec(ctx,
		`INSERT INTO bb_events (tenant_id, device_id, event_type, details)
		 VALUES (
			(SELECT id FROM tenants WHERE slug = $1 LIMIT 1),
			$2, 'exec_completed',
			$3::jsonb
		 )`,
		s.tenantSlug, req.DeviceID,
		fmt.Sprintf(`{"command":"%s","exit_code":0,"user_id":"%s"}`, req.Command, req.UserID))

	return &ExecResponse{
		Output:   output,
		ExitCode: 0,
		Command:  string(req.Command),
	}, nil
}
