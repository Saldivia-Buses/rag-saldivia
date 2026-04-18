package tools

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/Camionerou/rag-saldivia/pkg/audit"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
)

// testDefs returns a set of tool definitions for testing.
//
// Capability is mandatory since ADR 027 Phase 0 item 4. read_tool is open to
// any authed caller; dangerous_action requires a specific RBAC permission —
// we treat it as "ingest.write" so tests exercise both the authed-sentinel
// path and the permissioned path.
func testDefs() []Definition {
	return []Definition{
		{
			Name:                 "read_tool",
			Service:              "search",
			Endpoint:             "http://127.0.0.1:1/v1/search", // port 1: instant connection refused
			Method:               "POST",
			Type:                 "read",
			Capability:           CapabilityAuthed,
			RequiresConfirmation: false,
			Description:          "Search documents",
		},
		{
			Name:                 "dangerous_action",
			Service:              "ingest",
			Endpoint:             "http://127.0.0.1:1/v1/delete", // port 1: instant connection refused
			Method:               "DELETE",
			Type:                 "action",
			Capability:           "ingest.write",
			RequiresConfirmation: true,
			Description:          "Delete all documents in a collection",
		},
	}
}

// authedCtx returns a context pre-populated as if the Auth middleware had
// accepted a JWT for a user with the given role and permissions. Tests use
// this to exercise the RBAC path inside Execute / ExecuteConfirmed without
// spinning up a real HTTP handler stack.
func authedCtx(role string, perms ...string) context.Context {
	ctx := context.Background()
	ctx = sdamw.WithUserID(ctx, "user-test")
	ctx = sdamw.WithRole(ctx, role)
	ctx = sdamw.WithPermissions(ctx, perms)
	return ctx
}

func TestExecute_UnknownTool(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(testDefs())
	ctx := authedCtx("user", "ingest.write")

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
	ctx := authedCtx("user", "ingest.write")
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
	ctx := authedCtx("user")

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
	ctx := authedCtx("user", "ingest.write")

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
	ctx := authedCtx("user")

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
	ctx := authedCtx("user", "ingest.write")

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
	ctx := authedCtx("user")

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
	ctx := authedCtx("user", "ingest.write")

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
	ctx := authedCtx("user")

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

// --- ADR 027 Phase 0 item 4: RBAC dispatch gate ---------------------------

// recordingAudit captures audit entries so tests can assert the dispatch /
// denial records are written with the right action, user, and capability.
// Implements audit.Logger.
type recordingAudit struct {
	entries []audit.Entry
}

func (r *recordingAudit) Write(_ context.Context, e audit.Entry) {
	r.entries = append(r.entries, e)
}

// permDeniedDefs are fixture definitions that cover the full capability
// matrix: authed-only, specific permission, admin-gated wildcard, and
// an undeclared cap (which should be defensively denied even though the
// loader would have skipped it at load time).
func permDeniedDefs() []Definition {
	return []Definition{
		{Name: "open_read", Type: "read", Capability: CapabilityAuthed,
			Endpoint: "http://127.0.0.1:1/open", Method: "GET"},
		{Name: "erp_read", Type: "read", Capability: "erp.invoicing.read",
			Endpoint: "http://127.0.0.1:1/erp", Method: "GET"},
		{Name: "erp_write", Type: "action", Capability: "erp.invoicing.write",
			RequiresConfirmation: true,
			Endpoint:             "http://127.0.0.1:1/erp", Method: "POST"},
		{Name: "no_cap_defensive", Type: "read", Capability: "",
			Endpoint: "http://127.0.0.1:1/x", Method: "GET"},
	}
}

// TestExecute_Denies_WhenCapabilityMissingFromPerms verifies that a tool
// whose capability is NOT in the user's permission list returns a denied
// result instead of dispatching.
func TestExecute_Denies_WhenCapabilityMissingFromPerms(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("user", "erp.invoicing.read") // has read, not write

	got, err := exec.Execute(ctx, "tok", "erp_write", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status != "denied" {
		t.Fatalf("status = %q, want denied", got.Status)
	}
	if got.Error == "" {
		t.Fatal("denied result must carry an error message for the LLM")
	}
}

// TestExecute_Admin_BypassesCapabilityCheck mirrors pkg/middleware
// RequirePermission admin bypass: the admin role can dispatch any tool
// regardless of its capability string.
func TestExecute_Admin_BypassesCapabilityCheck(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("admin") // no perms, role=admin

	got, err := exec.Execute(ctx, "tok", "erp_read", nil)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status == "denied" {
		t.Fatalf("admin must not be denied; got %q %q", got.Status, got.Error)
	}
}

// TestExecute_Authed_Sentinel_AllowsAnyAuthedUser verifies the "authed"
// capability: the caller only needs a user identity, no extra perms.
func TestExecute_Authed_Sentinel_AllowsAnyAuthedUser(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("viewer") // no erp perms, not admin

	got, err := exec.Execute(ctx, "tok", "open_read", nil)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status == "denied" {
		t.Fatalf("authed tool must not deny an authed viewer; got %q %q", got.Status, got.Error)
	}
}

// TestExecute_Unauth_DeniesEvenAuthedSentinel guards the contract that the
// "authed" sentinel still requires an actual authenticated identity.
// An empty context (no JWT middleware run) must be rejected.
func TestExecute_Unauth_DeniesEvenAuthedSentinel(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := context.Background() // no user id, no role, no perms

	got, err := exec.Execute(ctx, "tok", "open_read", nil)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status != "denied" {
		t.Fatalf("status = %q, want denied", got.Status)
	}
}

// TestExecute_EmptyCapability_DefensivelyDenied is the belt to the loader's
// suspenders: if a Definition somehow reaches the executor with an empty
// capability (hand-constructed in a test / migration bug / stale cache),
// it must be rejected rather than silently accepted.
func TestExecute_EmptyCapability_DefensivelyDenied(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("admin") // even admin cannot bypass a missing capability

	got, err := exec.Execute(ctx, "tok", "no_cap_defensive", nil)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status != "denied" {
		t.Fatalf("status = %q, want denied", got.Status)
	}
}

// TestExecute_Wildcard_SatisfiesCapability verifies pkg/middleware wildcard
// matching flows through the executor: a user with "erp.*" dispatches
// erp.invoicing.read.
func TestExecute_Wildcard_SatisfiesCapability(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("user", "erp.*")

	got, err := exec.Execute(ctx, "tok", "erp_read", nil)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got.Status == "denied" {
		t.Fatalf("wildcard erp.* must satisfy erp.invoicing.read; got %q %q", got.Status, got.Error)
	}
}

// TestExecuteConfirmed_Denies_WhenCapabilityMissing guards the second-stage
// gate: a user who loses a permission between Execute (pending_confirmation)
// and ExecuteConfirmed must not be able to complete the write. Matches the
// skill contract "perm check is at dispatch, never skippable by omission".
func TestExecuteConfirmed_Denies_WhenCapabilityMissing(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("user", "erp.invoicing.read") // read only

	got, err := exec.ExecuteConfirmed(ctx, "tok", "erp_write", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("execute_confirmed: %v", err)
	}
	if got.Status != "denied" {
		t.Fatalf("status = %q, want denied", got.Status)
	}
}

// TestExecute_AuditLogger_Records verifies the audit entries: one per
// dispatch and one per denial, with the right action, user, and
// capability details.
func TestExecute_AuditLogger_Records(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	rec := &recordingAudit{}
	exec.SetAuditLogger(rec)

	// 1. Denied dispatch — user without the cap.
	denyCtx := authedCtx("user", "erp.invoicing.read")
	if _, err := exec.Execute(denyCtx, "tok", "erp_write", nil); err != nil {
		t.Fatalf("execute: %v", err)
	}

	// 2. Allowed dispatch — user with the cap. Tool requires confirmation,
	// so ExecuteConfirmed drives the actual dispatch path.
	allowCtx := authedCtx("user", "erp.invoicing.write")
	if _, err := exec.ExecuteConfirmed(allowCtx, "tok", "erp_write", nil); err != nil {
		t.Fatalf("execute_confirmed: %v", err)
	}

	if len(rec.entries) != 2 {
		t.Fatalf("want 2 audit entries, got %d (%+v)", len(rec.entries), rec.entries)
	}
	if rec.entries[0].Action != "agent.tool.denied" {
		t.Errorf("entry[0].Action = %q, want agent.tool.denied", rec.entries[0].Action)
	}
	if rec.entries[0].Resource != "erp_write" {
		t.Errorf("entry[0].Resource = %q, want erp_write", rec.entries[0].Resource)
	}
	if cap, _ := rec.entries[0].Details["capability"].(string); cap != "erp.invoicing.write" {
		t.Errorf("entry[0].capability = %q, want erp.invoicing.write", cap)
	}
	if reason, _ := rec.entries[0].Details["reason"].(string); reason != "capability_missing" {
		t.Errorf("entry[0].reason = %q, want capability_missing", reason)
	}
	if rec.entries[1].Action != "agent.tool.dispatch" {
		t.Errorf("entry[1].Action = %q, want agent.tool.dispatch", rec.entries[1].Action)
	}
	if confirmed, _ := rec.entries[1].Details["confirmed"].(bool); !confirmed {
		t.Errorf("entry[1].confirmed flag not set on ExecuteConfirmed path")
	}
	if rec.entries[0].UserID != "user-test" || rec.entries[1].UserID != "user-test" {
		t.Errorf("entries must carry user-test; got %q / %q", rec.entries[0].UserID, rec.entries[1].UserID)
	}
}

// TestExecute_MultiToolIndependence verifies that a denied call does not
// affect a separate allowed call. The agent loop runs multiple tool calls
// per turn; one denial must not poison the rest.
func TestExecute_MultiToolIndependence(t *testing.T) {
	t.Parallel()
	exec := NewExecutor(permDeniedDefs())
	ctx := authedCtx("user", "erp.invoicing.read") // read only

	// read_tool path — allowed (authed sentinel).
	readRes, err := exec.Execute(ctx, "tok", "open_read", nil)
	if err != nil {
		t.Fatalf("open_read: %v", err)
	}
	// read_tool will actually attempt HTTP and get a timeout. Either
	// success or timeout is acceptable here — the point is that it was
	// NOT denied.
	if readRes.Status == "denied" {
		t.Fatalf("open_read was denied unexpectedly: %q", readRes.Error)
	}

	// erp_write — denied (no write perm).
	writeRes, err := exec.Execute(ctx, "tok", "erp_write", nil)
	if err != nil {
		t.Fatalf("erp_write: %v", err)
	}
	if writeRes.Status != "denied" {
		t.Fatalf("erp_write status = %q, want denied", writeRes.Status)
	}

	// erp_read — allowed by the exact perm.
	allowedRes, err := exec.Execute(ctx, "tok", "erp_read", nil)
	if err != nil {
		t.Fatalf("erp_read: %v", err)
	}
	if allowedRes.Status == "denied" {
		t.Fatalf("erp_read was denied despite erp.invoicing.read perm: %q", allowedRes.Error)
	}
}
