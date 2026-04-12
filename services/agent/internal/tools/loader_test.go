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
	os.MkdirAll(modDir, 0755)

	manifest := `
module: testmod
name: Test Module
tools:
  - id: test_search
    service: test
    method: Search
    protocol: grpc
    type: read
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
    requires_confirmation: true
    description: "A dangerous test action"
    parameters:
      type: object
      properties:
        target:
          type: string
`
	os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0644)

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
	os.MkdirAll(modDir, 0755)
	os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte("module: disabled\nname: X\ntools: []"), 0644)

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
	os.MkdirAll(modDir, 0755)
	// Invalid YAML — should be skipped, not fatal
	os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(":::invalid yaml:::"), 0644)

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
	os.MkdirAll(modDir, 0755)
	manifest := `
module: mymod
name: My Module
tools:
  - id: known_tool
    service: known
    method: Do
    protocol: grpc
    type: read
    description: "known service"
  - id: unknown_tool
    service: unknown_svc
    method: Do
    protocol: grpc
    type: read
    description: "missing service URL"
`
	os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0644)

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

func TestLoadModuleTools_EmptyToolsList(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "emptymod")
	os.MkdirAll(modDir, 0755)
	os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte("module: emptymod\nname: Empty\ntools: []\n"), 0644)

	defs, err := LoadModuleTools(dir, map[string]bool{"emptymod": true}, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(defs))
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
		Endpoint: "POST /v1/astro/natal",
	})
	if method != "POST" {
		t.Errorf("expected POST, got %s", method)
	}
	if url != "http://svc:8000/v1/astro/natal" {
		t.Errorf("expected http://svc:8000/v1/astro/natal, got %s", url)
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
		Endpoint: "/v1/astro/transits",
	})
	if method != "POST" {
		t.Errorf("expected POST for path-only endpoint, got %s", method)
	}
	if url != "http://svc:8000/v1/astro/transits" {
		t.Errorf("expected http://svc:8000/v1/astro/transits, got %s", url)
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
	modDir := filepath.Join(dir, "astromod")
	os.MkdirAll(modDir, 0755)
	manifest := `
module: astromod
name: Astro Module
tools:
  - id: natal_chart
    service: astro
    endpoint: "POST /v1/astro/natal"
    protocol: http
    type: read
    description: "Build natal chart"
`
	os.WriteFile(filepath.Join(modDir, "tools.yaml"), []byte(manifest), 0644)

	defs, err := LoadModuleTools(
		dir,
		map[string]bool{"astromod": true},
		map[string]string{"astro": "http://astro:9000"},
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
	if defs[0].Endpoint != "http://astro:9000/v1/astro/natal" {
		t.Errorf("expected http://astro:9000/v1/astro/natal, got %s", defs[0].Endpoint)
	}
}
