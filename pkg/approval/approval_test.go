package approval

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPendingActionIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{"future", time.Now().Add(5 * time.Minute), false},
		{"past", time.Now().Add(-1 * time.Minute), true},
		{"now", time.Now().Add(-1 * time.Millisecond), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &PendingAction{ExpiresAt: tt.expiresAt}
			if got := a.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateRequestValidate(t *testing.T) {
	valid := CreateRequest{
		TenantID:    uuid.New(),
		ResourceID:  uuid.New(),
		Action:      "plc_write",
		RequestorID: uuid.New(),
		TTL:         5 * time.Minute,
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("valid request failed: %v", err)
	}

	tests := []struct {
		name   string
		modify func(r *CreateRequest)
	}{
		{"nil tenant", func(r *CreateRequest) { r.TenantID = uuid.Nil }},
		{"nil resource", func(r *CreateRequest) { r.ResourceID = uuid.Nil }},
		{"empty action", func(r *CreateRequest) { r.Action = "" }},
		{"nil requestor", func(r *CreateRequest) { r.RequestorID = uuid.Nil }},
		{"zero TTL", func(r *CreateRequest) { r.TTL = 0 }},
		{"negative TTL", func(r *CreateRequest) { r.TTL = -1 * time.Minute }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := valid
			tt.modify(&r)
			if err := r.Validate(); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestStatusConstants(t *testing.T) {
	// Verify status values match what the DB CHECK constraint expects
	if StatusPending != "pending" {
		t.Error("StatusPending mismatch")
	}
	if StatusApproved != "approved" {
		t.Error("StatusApproved mismatch")
	}
	if StatusExpired != "expired" {
		t.Error("StatusExpired mismatch")
	}
	if StatusRejected != "rejected" {
		t.Error("StatusRejected mismatch")
	}
}
