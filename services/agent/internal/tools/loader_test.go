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
