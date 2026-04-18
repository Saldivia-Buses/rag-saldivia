package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadModuleTools(t *testing.T) {
	// Create temp modules dir with a test manifest
	dir := t.TempDir()
	modDir := filepath.Join(dir, "testmod")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifest := `
module: testmod
name: Test Module
tools:
  - id: test_search
    service: test
    method: Search
    protocol: grpc
    type: read
    capability: authed
    description: "A test tool"
    parameters:
      type: object
      properties:
        query:
          type: string
      required:
        - query
  - id: test_action
    service: test
    method: DoAction
    protocol: grpc
    type: action
    capability: erp.entities.write
    requires_confirmation: true
    description: "A dangerous test action"
    parameters:
      type: object
      properties:
        target:
          type: string
`
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	// Load with module enabled
	defs, err := LoadModuleTools(dir, map[string]bool{"testmod": true}, map[string]string{"test": "http://test:8000"})
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(defs))
	}
	if defs[0].Name != "test_search" {
		t.Fatalf("expected test_search, got %s", defs[0].Name)
	}
	if defs[1].RequiresConfirmation != true {
		t.Fatal("expected test_action to require confirmation")
	}
}

func TestLoadModuleTools_DisabledModule(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "disabled")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte("module: disabled\nname: X\ntools: []"), 0644); err != nil {
		t.Fatal(err)
	}

	defs, err := LoadModuleTools(dir, map[string]bool{}, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected 0 tools for disabled module, got %d", len(defs))
	}
}

func TestLoadModuleTools_NoDir(t *testing.T) {
	defs, err := LoadModuleTools("/nonexistent", map[string]bool{}, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if defs != nil {
		t.Fatal("expected nil for nonexistent dir")
	}
}

func TestLoadModuleTools_InvalidYAML_Skips(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "badmod")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Invalid YAML — should be skipped, not fatal
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(":::invalid yaml:::"), 0644); err != nil {
		t.Fatal(err)
	}

	// Module is enabled but YAML is invalid — should be skipped gracefully
	defs, err := LoadModuleTools(dir, map[string]bool{"badmod": true}, map[string]string{})
	if err != nil {
		t.Fatalf("expected no error for invalid YAML, got: %v", err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected 0 tools from invalid YAML, got %d", len(defs))
	}
}

func TestLoadModuleTools_ToolMissingServiceURL_Skipped(t *testing.T) {
	// A tool whose service has no URL configured must be skipped, not fatal.
	dir := t.TempDir()
	modDir := filepath.Join(dir, "mymod")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := `
module: mymod
name: My Module
tools:
  - id: known_tool
    service: known
    method: Do
    protocol: grpc
    type: read
    capability: authed
    description: "known service"
  - id: unknown_tool
    service: unknown_svc
    method: Do
    protocol: grpc
    type: read
    capability: authed
    description: "missing service URL"
`
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	defs, err := LoadModuleTools(
		dir,
		map[string]bool{"mymod": true},
		map[string]string{"known": "http://known:8000"}, // unknown_svc has no URL
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 tool (unknown skipped), got %d", len(defs))
	}
	if defs[0].Name != "known_tool" {
		t.Errorf("expected known_tool, got %s", defs[0].Name)
	}
}

// TestLoadModuleTools_MissingCapability_Skipped verifies the fail-closed
// behaviour required by ADR 027 Phase 0 item 4: a tool YAML without a
// `capability:` field is dropped at load time (with an ERROR log) and
// never appears in the registry. The LLM never sees such a tool.
func TestLoadModuleTools_MissingCapability_Skipped(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "capless")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `
module: capless
name: Cap-less Module
tools:
  - id: fine_tool
    service: svc
    method: Do
    protocol: grpc
    type: read
    capability: authed
    description: "has capability"
  - id: cap_missing
    service: svc
    method: Do
    protocol: grpc
    type: read
    description: "no capability declared"
  - id: cap_blank
    service: svc
    method: Do
    protocol: grpc
    type: read
    capability: "   "
    description: "whitespace-only capability is also rejected"
`
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	defs, err := LoadModuleTools(
		dir,
		map[string]bool{"capless": true},
		map[string]string{"svc": "http://svc:9000"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected exactly 1 tool (fine_tool), got %d: %+v", len(defs), defs)
	}
	if defs[0].Name != "fine_tool" {
		t.Fatalf("expected fine_tool loaded, got %q", defs[0].Name)
	}
	if defs[0].Capability != "authed" {
		t.Fatalf("expected capability %q, got %q", "authed", defs[0].Capability)
	}
}

func TestLoadModuleTools_EmptyToolsList(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "emptymod")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte("module: emptymod\nname: Empty\ntools: []\n"), 0644); err != nil {
		t.Fatal(err)
	}

	defs, err := LoadModuleTools(dir, map[string]bool{"emptymod": true}, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(defs))
	}
}

// --- ParseEnabledModules ---

func TestParseEnabledModules_Empty(t *testing.T) {
	m := ParseEnabledModules("")
	if !m["fleet"] || !m["bigbrother"] || !m["erp"] {
		t.Errorf("empty string should enable all known modules, got: %v", m)
	}
}

func TestParseEnabledModules_All(t *testing.T) {
	m := ParseEnabledModules("all")
	if !m["fleet"] || !m["bigbrother"] || !m["erp"] {
		t.Errorf("\"all\" should enable all known modules, got: %v", m)
	}
}

func TestParseEnabledModules_AllCaseInsensitive(t *testing.T) {
	m := ParseEnabledModules("ALL")
	if !m["fleet"] || !m["erp"] {
		t.Errorf("\"ALL\" should enable all known modules, got: %v", m)
	}
}

func TestParseEnabledModules_None(t *testing.T) {
	m := ParseEnabledModules("none")
	if len(m) != 0 {
		t.Errorf("\"none\" should return empty map, got: %v", m)
	}
}

func TestParseEnabledModules_Subset(t *testing.T) {
	m := ParseEnabledModules("fleet,erp")
	if !m["fleet"] || !m["erp"] {
		t.Errorf("expected fleet and erp enabled, got: %v", m)
	}
	if m["bigbrother"] {
		t.Errorf("bigbrother should not be enabled, got: %v", m)
	}
	if len(m) != 2 {
		t.Errorf("expected exactly 2 entries, got %d: %v", len(m), m)
	}
}

func TestParseEnabledModules_Single(t *testing.T) {
	m := ParseEnabledModules("fleet")
	if !m["fleet"] {
		t.Errorf("expected fleet enabled, got: %v", m)
	}
	if len(m) != 1 {
		t.Errorf("expected exactly 1 entry, got %d: %v", len(m), m)
	}
}

func TestParseEnabledModules_WhitespaceIgnored(t *testing.T) {
	m := ParseEnabledModules(" fleet , erp ")
	if !m["fleet"] || !m["erp"] {
		t.Errorf("expected fleet and erp enabled (with whitespace trimmed), got: %v", m)
	}
}

// --- resolveEndpoint ---

func TestResolveEndpoint_GRPCProtocol(t *testing.T) {
	method, url := resolveEndpoint("http://svc:8000", ManifestTool{
		Protocol: "grpc",
		Method:   "Search",
	})
	if method != "POST" {
		t.Errorf("expected POST for gRPC, got %s", method)
	}
	if url != "http://svc:8000/Search" {
		t.Errorf("expected http://svc:8000/Search, got %s", url)
	}
}

func TestResolveEndpoint_HTTPProtocol_WithVerb(t *testing.T) {
	method, url := resolveEndpoint("http://svc:8000", ManifestTool{
		Protocol: "http",
		Endpoint: "POST /v1/fleet/vehicles",
	})
	if method != "POST" {
		t.Errorf("expected POST, got %s", method)
	}
	if url != "http://svc:8000/v1/fleet/vehicles" {
		t.Errorf("expected http://svc:8000/v1/fleet/vehicles, got %s", url)
	}
}

func TestResolveEndpoint_HTTPProtocol_GetVerb(t *testing.T) {
	method, url := resolveEndpoint("http://svc:8000", ManifestTool{
		Protocol: "http",
		Endpoint: "GET /v1/fleet/vehicles",
	})
	if method != "GET" {
		t.Errorf("expected GET, got %s", method)
	}
	if url != "http://svc:8000/v1/fleet/vehicles" {
		t.Errorf("expected http://svc:8000/v1/fleet/vehicles, got %s", url)
	}
}

func TestResolveEndpoint_HTTPProtocol_PathOnly(t *testing.T) {
	// Endpoint without verb defaults to POST
	method, url := resolveEndpoint("http://svc:8000", ManifestTool{
		Protocol: "http",
		Endpoint: "/v1/fleet/refuel",
	})
	if method != "POST" {
		t.Errorf("expected POST for path-only endpoint, got %s", method)
	}
	if url != "http://svc:8000/v1/fleet/refuel" {
		t.Errorf("expected http://svc:8000/v1/fleet/refuel, got %s", url)
	}
}

func TestResolveEndpoint_HTTPProtocol_FallbackToMethod(t *testing.T) {
	// HTTP protocol with no Endpoint field falls back to Method field
	method, url := resolveEndpoint("http://svc:8000", ManifestTool{
		Protocol: "http",
		Method:   "DoSomething",
	})
	if method != "POST" {
		t.Errorf("expected POST, got %s", method)
	}
	if url != "http://svc:8000/DoSomething" {
		t.Errorf("expected http://svc:8000/DoSomething, got %s", url)
	}
}

func TestResolveEndpoint_UnknownProtocol_DefaultsToGRPCBehavior(t *testing.T) {
	// Unknown/empty protocol falls through to default (same as gRPC)
	method, url := resolveEndpoint("http://svc:8000", ManifestTool{
		Protocol: "",
		Method:   "RunQuery",
	})
	if method != "POST" {
		t.Errorf("expected POST, got %s", method)
	}
	if url != "http://svc:8000/RunQuery" {
		t.Errorf("expected http://svc:8000/RunQuery, got %s", url)
	}
}

func TestLoadModuleTools_HTTPProtocolTool(t *testing.T) {
	// Verify HTTP protocol tools get their endpoint resolved correctly
	dir := t.TempDir()
	modDir := filepath.Join(dir, "fleetmod")
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatal(err)
	}
	manifest := `
module: fleetmod
name: Fleet Module
tools:
  - id: list_vehicles
    service: fleet
    endpoint: "POST /v1/fleet/vehicles"
    protocol: http
    type: read
    capability: authed
    description: "List vehicles"
`
	if err := os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	defs, err := LoadModuleTools(
		dir,
		map[string]bool{"fleetmod": true},
		map[string]string{"fleet": "http://fleet:9000"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(defs))
	}
	if defs[0].Method != "POST" {
		t.Errorf("expected POST method, got %s", defs[0].Method)
	}
	if defs[0].Endpoint != "http://fleet:9000/v1/fleet/vehicles" {
		t.Errorf("expected http://fleet:9000/v1/fleet/vehicles, got %s", defs[0].Endpoint)
	}
}
