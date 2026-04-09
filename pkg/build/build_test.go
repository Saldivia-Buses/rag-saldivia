package build

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInfo_ContainsExpectedFields(t *testing.T) {
	info := Info("sda-test")

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
			t.Errorf("Info() missing key %q", key)
			continue
		}
		if got != want {
			t.Errorf("Info()[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestInfo_ServiceNamePassedThrough(t *testing.T) {
	tests := []struct {
		name    string
		service string
	}{
		{name: "auth service", service: "sda-auth"},
		{name: "chat service", service: "sda-chat"},
		{name: "empty name", service: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := Info(tt.service)
			if info["service"] != tt.service {
				t.Errorf("Info(%q)[service] = %q, want %q", tt.service, info["service"], tt.service)
			}
		})
	}
}

func TestInfo_NonEmptyDefaults(t *testing.T) {
	info := Info("test")

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

func TestHandler_Returns200WithJSON(t *testing.T) {
	h := Handler("sda-auth")

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

	if body["service"] != "sda-auth" {
		t.Errorf("body[service] = %q, want %q", body["service"], "sda-auth")
	}
	if body["go_version"] != runtime.Version() {
		t.Errorf("body[go_version] = %q, want %q", body["go_version"], runtime.Version())
	}
}

func TestHandler_ConsistentAcrossCalls(t *testing.T) {
	h := Handler("sda-chat")

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
		if body["service"] != "sda-chat" {
			t.Errorf("call %d: service = %q, want sda-chat", i, body["service"])
		}
	}
}

func TestReadVersionFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string // returns path
		expected string
	}{
		{
			name: "valid version file",
			setup: func() string {
				p := filepath.Join(tmpDir, "VERSION_valid")
				os.WriteFile(p, []byte("1.2.3\n"), 0644)
				return p
			},
			expected: "1.2.3",
		},
		{
			name: "version with whitespace",
			setup: func() string {
				p := filepath.Join(tmpDir, "VERSION_ws")
				os.WriteFile(p, []byte("  2.0.0-rc1  \n"), 0644)
				return p
			},
			expected: "2.0.0-rc1",
		},
		{
			name: "empty file returns dev",
			setup: func() string {
				p := filepath.Join(tmpDir, "VERSION_empty")
				os.WriteFile(p, []byte(""), 0644)
				return p
			},
			expected: "dev",
		},
		{
			name: "whitespace-only file returns dev",
			setup: func() string {
				p := filepath.Join(tmpDir, "VERSION_ws_only")
				os.WriteFile(p, []byte("   \n  \n"), 0644)
				return p
			},
			expected: "dev",
		},
		{
			name: "missing file returns Version fallback",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			expected: Version, // "dev" or whatever ldflags set
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			got := ReadVersionFile(path)
			if got != tt.expected {
				t.Errorf("ReadVersionFile(%q) = %q, want %q", path, got, tt.expected)
			}
		})
	}
}
