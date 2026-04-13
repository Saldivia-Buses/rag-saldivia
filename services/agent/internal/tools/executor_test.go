package tools

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
)

// testDefs returns a set of tool definitions for testing.
func testDefs() []Definition {
	return []Definition{
		{
			Name:                 "read_tool",
			Service:              "search",
			Endpoint:             "http://127.0.0.1:1/v1/search", // port 1: instant connection refused
			Method:               "POST",
			Type:                 "read",
			RequiresConfirmation: false,
			Description:          "Search documents",
		},
		{
			Name:                 "dangerous_action",
			Service:              "ingest",
			Endpoint:             "http://127.0.0.1:1/v1/delete", // port 1: instant connection refused
			Method:               "DELETE",
			Type:                 "action",
			RequiresConfirmation: true,
			Description:          "Delete all documents in a collection",
		},
	}
}

func TestExecute_UnknownTool(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.Execute(ctx, "tok", "nonexistent", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Status != "error" {
		t.Fatalf("expected status %q, got %q", "error", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestExecute_RequiresConfirmation(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()
	params := json.RawMessage(`{"collection":"test"}`)

	result, err := exec.Execute(ctx, "tok", "dangerous_action", params)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Status != "pending_confirmation" {
		t.Fatalf("expected status %q, got %q", "pending_confirmation", result.Status)
	}
	if !result.RequiresConfirmation {
		t.Fatal("expected RequiresConfirmation to be true")
	}
	if result.ActionPlan == "" {
		t.Fatal("expected non-empty ActionPlan")
	}
	// The original params should be returned in Data for the confirmation flow
	if string(result.Data) != string(params) {
		t.Fatalf("expected params echoed back, got %q", string(result.Data))
	}
}

func TestExecute_NormalTool_AttemptsHTTP(t *testing.T) {
	// A normal (non-confirmation) tool should try to make the HTTP call.
	// With no real server running, it will fail with a connection error
	// but the status should be "timeout" (client error path), not "pending_confirmation".
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.Execute(ctx, "tok", "read_tool", json.RawMessage(`{"q":"test"}`))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// Should NOT be pending_confirmation — it should attempt execution
	if result.Status == "pending_confirmation" {
		t.Fatal("read_tool should not require confirmation")
	}
	// With no server, expect timeout (connection refused falls into client.Do error path)
	if result.Status != "timeout" {
		t.Fatalf("expected status %q with no server, got %q", "timeout", result.Status)
	}
}

func TestExecuteConfirmed_UnknownTool(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.ExecuteConfirmed(ctx, "tok", "nonexistent", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Status != "error" {
		t.Fatalf("expected status %q, got %q", "error", result.Status)
	}
}

func TestExecuteConfirmed_ToolWithoutConfirmation_Rejected(t *testing.T) {
	// Security: calling ExecuteConfirmed on a tool that doesn't require
	// confirmation must be rejected. This prevents bypassing the normal
	// Execute flow by going directly to ExecuteConfirmed.
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.ExecuteConfirmed(ctx, "tok", "read_tool", json.RawMessage(`{"q":"test"}`))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.Status != "error" {
		t.Fatalf("expected status %q, got %q", "error", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected error message about confirmation not required")
	}
}

func TestExecuteConfirmed_ToolWithConfirmation_Proceeds(t *testing.T) {
	// A tool that requires confirmation should proceed when called via
	// ExecuteConfirmed (it attempts the HTTP call).
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.ExecuteConfirmed(ctx, "tok", "dangerous_action", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// Should attempt execution (not error about confirmation).
	// With no server, will get timeout from connection refused.
	if result.Status == "error" && result.Error != "" {
		// Make sure it's not the "does not require confirmation" error
		if result.Error == `tool "dangerous_action" does not require confirmation` {
			t.Fatal("should not reject a tool that genuinely requires confirmation")
		}
	}
	// With no server, expect timeout
	if result.Status != "timeout" {
		t.Fatalf("expected status %q with no server, got %q", "timeout", result.Status)
	}
}

func TestListTools(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())

	names := exec.ListTools()
	if len(names) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(names))
	}

	slices.Sort(names)
	if names[0] != "dangerous_action" || names[1] != "read_tool" {
		t.Fatalf("unexpected tool names: %v", names)
	}
}

func TestListTools_Empty(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(nil)

	names := exec.ListTools()
	if len(names) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(names))
	}
}

func TestGetDefinition_Found(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())

	def, ok := exec.GetDefinition("read_tool")
	if !ok {
		t.Fatal("expected to find read_tool")
	}
	if def.Name != "read_tool" {
		t.Fatalf("expected name %q, got %q", "read_tool", def.Name)
	}
	if def.Service != "search" {
		t.Fatalf("expected service %q, got %q", "search", def.Service)
	}
	if def.Type != "read" {
		t.Fatalf("expected type %q, got %q", "read", def.Type)
	}
}

func TestGetDefinition_NotFound(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())

	_, ok := exec.GetDefinition("does_not_exist")
	if ok {
		t.Fatal("expected not found for nonexistent tool")
	}
}

// TestExecute_EmptyToolName_ReturnsErrorResult verifies that an empty string
// tool name is treated as an unknown tool and returns a graceful error result,
// not a panic or nil pointer.
func TestExecute_EmptyToolName_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.Execute(ctx, "tok", "", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != "error" {
		t.Fatalf("expected status %q for empty tool name, got %q", "error", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected non-empty error message")
	}
}

// TestExecuteConfirmed_EmptyToolName_ReturnsErrorResult mirrors the Execute check
// for the confirmed execution path.
func TestExecuteConfirmed_EmptyToolName_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	result, err := exec.ExecuteConfirmed(ctx, "tok", "", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != "error" {
		t.Fatalf("expected status %q for empty tool name, got %q", "error", result.Status)
	}
}

// TestExecute_NilParams_DoesNotPanic verifies that nil params are forwarded
// without panicking (the HTTP body becomes an empty reader).
func TestExecute_NilParams_DoesNotPanic(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := context.Background()

	// nil params should not panic — just attempt execution with empty body
	result, err := exec.Execute(ctx, "tok", "read_tool", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// With no server, should timeout (connection refused), not panic
	if result.Status == "pending_confirmation" {
		t.Fatal("read_tool should not require confirmation")
	}
}

func TestNewExecutor_DuplicateNames_LastWins(t *testing.T) {
	// If two definitions have the same name, the last one wins (map semantics).
	t.Parallel()
	defs := []Definition{
		{Name: "dup", Service: "first", Type: "read"},
		{Name: "dup", Service: "second", Type: "action"},
	}
	exec := NewExecutor(defs)

	def, ok := exec.GetDefinition("dup")
	if !ok {
		t.Fatal("expected to find dup")
	}
	if def.Service != "second" {
		t.Fatalf("expected last-wins, got service %q", def.Service)
	}
	if len(exec.ListTools()) != 1 {
		t.Fatalf("expected 1 tool after dedup, got %d", len(exec.ListTools()))
	}
}
