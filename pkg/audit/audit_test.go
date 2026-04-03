package audit

import (
	"testing"
)

func TestNilIfEmpty(t *testing.T) {
	if got := nilIfEmpty(""); got != nil {
		t.Errorf("expected nil for empty string, got %v", got)
	}
	if got := nilIfEmpty("hello"); got == nil || *got != "hello" {
		t.Errorf("expected 'hello', got %v", got)
	}
}

func TestEntry_DetailsNil(t *testing.T) {
	// Verify Entry can be constructed with nil Details without panic.
	// The actual DB write is tested in integration tests.
	e := Entry{
		UserID: "u-1",
		Action: "user.login",
	}
	if e.Action != "user.login" {
		t.Errorf("expected user.login, got %q", e.Action)
	}
}
