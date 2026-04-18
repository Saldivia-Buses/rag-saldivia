package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildInfo_ContainsExpectedFields(t *testing.T) {
	info := BuildInfo("sda-test")

	expected := map[string]string{
		"service":    "sda-test",
		"version":    Version,
		"git_sha":    GitSHA,
		"build_time": BuildTime,
		"go_version": runtime.Version(),
	}

	for key, want := range expected {
		got, ok := info[key]
		if !ok {
			t.Errorf("BuildInfo() missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("BuildInfo()[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestBuildInfo_ServiceNamePassedThrough(t *testing.T) {
	tests := []struct {
		name    string
		service string
	}{
		{name: "app service", service: "sda-app"},
		{name: "erp service", service: "sda-erp"},
		{name: "empty name", service: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := BuildInfo(tt.service)
			if info["service"] != tt.service {
				t.Errorf("BuildInfo(%q)[service] = %q, want %q", tt.service, info["service"], tt.service)
			}
		})
	}
}

func TestBuildInfo_NonEmptyDefaults(t *testing.T) {
	info := BuildInfo("test")

	// GitSHA and BuildTime are set in init() — they should never be empty
	if info["git_sha"] == "" {
		t.Error("git_sha should not be empty after init()")
	}
	if info["build_time"] == "" {
		t.Error("build_time should not be empty after init()")
	}
	if info["go_version"] == "" {
		t.Error("go_version should not be empty")
	}
}

func TestBuildInfoHandler_Returns200WithJSON(t *testing.T) {
	h := BuildInfoHandler("sda-app")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["service"] != "sda-app" {
		t.Errorf("body[service] = %q, want %q", body["service"], "sda-app")
	}
	if body["go_version"] != runtime.Version() {
		t.Errorf("body[go_version] = %q, want %q", body["go_version"], runtime.Version())
	}
}

func TestBuildInfoHandler_ConsistentAcrossCalls(t *testing.T) {
	h := BuildInfoHandler("sda-erp")

	// Call twice — the payload is cached at creation time
	for i := range 2 {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/v1/info", nil)
		h.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("call %d: expected 200, got %d", i, w.Code)
		}

		var body map[string]string
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("call %d: decode error: %v", i, err)
		}
		if body["service"] != "sda-erp" {
			t.Errorf("call %d: service = %q, want sda-erp", i, body["service"])
		}
	}
}

func TestReadVersionFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func(t *testing.T) string // returns path
		expected string
	}{
		{
			name: "valid version file",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(tmpDir, "VERSION_valid")
				if err := os.WriteFile(p, []byte("1.2.3\n"), 0644); err != nil {
					t.Fatalf("write VERSION_valid: %v", err)
				}
				return p
			},
			expected: "1.2.3",
		},
		{
			name: "version with whitespace",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(tmpDir, "VERSION_ws")
				if err := os.WriteFile(p, []byte("  2.0.0-rc1  \n"), 0644); err != nil {
					t.Fatalf("write VERSION_ws: %v", err)
				}
				return p
			},
			expected: "2.0.0-rc1",
		},
		{
			name: "empty file returns dev",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(tmpDir, "VERSION_empty")
				if err := os.WriteFile(p, []byte(""), 0644); err != nil {
					t.Fatalf("write VERSION_empty: %v", err)
				}
				return p
			},
			expected: "dev",
		},
		{
			name: "whitespace-only file returns dev",
			setup: func(t *testing.T) string {
				t.Helper()
				p := filepath.Join(tmpDir, "VERSION_ws_only")
				if err := os.WriteFile(p, []byte("   \n  \n"), 0644); err != nil {
					t.Fatalf("write VERSION_ws_only: %v", err)
				}
				return p
			},
			expected: "dev",
		},
		{
			name: "missing file returns Version fallback",
			setup: func(_ *testing.T) string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			expected: Version, // "dev" or whatever ldflags set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			got := ReadVersionFile(path)
			if got != tt.expected {
				t.Errorf("ReadVersionFile(%q) = %q, want %q", path, got, tt.expected)
			}
		})
	}
}
