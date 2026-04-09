// Package approval provides a generic two-person approval pattern.
//
// Any operation that requires a second person to approve before execution
// can use this package. The pattern: User A requests an action, User B
// (different user, same permission level) approves it within a time window.
//
// The package defines types and interfaces — concrete storage implementations
// live in the consuming service (e.g., BigBrother uses bb_pending_writes).
//
// Example use cases:
//   - PLC critical writes (BigBrother)
//   - Tenant deletion (Platform)
//   - Data purge (Compliance)
package approval

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("pending action not found")
	ErrExpired       = errors.New("pending action expired")
	ErrSelfApprove   = errors.New("requestor cannot approve their own action")
	ErrAlreadyExists = errors.New("a pending action already exists for this resource")
	ErrAlreadyHandled = errors.New("action already approved, rejected, or expired")
)

// Status represents the lifecycle of a pending action.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusExpired  Status = "expired"
	StatusRejected Status = "rejected"
)

// PendingAction represents an operation awaiting a second person's approval.
type PendingAction struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ResourceID  uuid.UUID  // what is being acted upon
	ResourceKey string     // human-readable key (e.g., register address)
	Action      string     // "plc_write", "tenant_delete", etc.
	Payload     []byte     // action-specific data (JSON)
	RequestorID uuid.UUID
	ApprovedBy  *uuid.UUID // nil until approved
	ApprovedAt  *time.Time // nil until approved
	Status      Status
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

// IsExpired returns true if the action has passed its expiration time.
func (a *PendingAction) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// CreateRequest holds the data needed to create a new pending action.
type CreateRequest struct {
	TenantID    uuid.UUID
	ResourceID  uuid.UUID
	ResourceKey string
	Action      string
	Payload     []byte
	RequestorID uuid.UUID
	TTL         time.Duration // how long before expiration
}

// Store is the interface for persisting pending actions.
// Implementations must ensure atomic operations to prevent race conditions.
type Store interface {
	// Create inserts a new pending action. Returns ErrAlreadyExists if there
	// is already a pending action for the same resource (enforced by partial
	// unique index on the storage layer).
	Create(ctx context.Context, req CreateRequest) (*PendingAction, error)

	// Approve atomically approves a pending action. The implementation MUST:
	//   1. Verify requestor_id != approver_id (no self-approve)
	//   2. Verify status = 'pending' (not already handled)
	//   3. Verify not expired
	//   4. UPDATE ... SET approved_by=$1, status='approved'
	//      WHERE id=$2 AND status='pending' AND approved_by IS NULL RETURNING *
	// Only one concurrent Approve call can succeed (atomic).
	// Returns ErrSelfApprove, ErrExpired, ErrAlreadyHandled, or ErrNotFound.
	Approve(ctx context.Context, id, approverID uuid.UUID) (*PendingAction, error)

	// Reject marks a pending action as rejected.
	Reject(ctx context.Context, id, rejectorID uuid.UUID) error

	// Get retrieves a pending action by ID.
	Get(ctx context.Context, id uuid.UUID) (*PendingAction, error)

	// CleanExpired marks all expired pending actions as expired.
	// Returns the number of actions expired.
	CleanExpired(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// Validate checks a CreateRequest for basic validity.
func (r CreateRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}
	if r.ResourceID == uuid.Nil {
		return errors.New("resource_id is required")
	}
	if r.Action == "" {
		return errors.New("action is required")
	}
	if r.RequestorID == uuid.Nil {
		return errors.New("requestor_id is required")
	}
	if r.TTL <= 0 {
		return errors.New("TTL must be positive")
	}
	return nil
}
