package server

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

// Set via -ldflags:
//
//	go build -ldflags "-X github.com/Camionerou/rag-saldivia/pkg/server.Version=0.1.0
//	  -X github.com/Camionerou/rag-saldivia/pkg/server.GitSHA=$(git rev-parse --short HEAD)
//	  -X github.com/Camionerou/rag-saldivia/pkg/server.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version   = "dev"
	GitSHA    = ""
	BuildTime = ""
)

func init() {
	// Go 1.18+ embeds VCS info automatically — use as fallback when ldflags not set
	if GitSHA == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					if len(s.Value) > 7 {
						GitSHA = s.Value[:7]
					} else {
						GitSHA = s.Value
					}
				case "vcs.time":
					if BuildTime == "" {
						BuildTime = s.Value
					}
				}
			}
		}
	}
	if GitSHA == "" {
		GitSHA = "unknown"
	}
	if BuildTime == "" {
		BuildTime = "unknown"
	}
}

// ReadVersionFile reads a VERSION file and returns its content as a trimmed string.
// Returns "dev" if the file doesn't exist or is empty. Useful as fallback
// when ldflags aren't set (e.g., `go run` during development).
func ReadVersionFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return Version // fall back to ldflags value or "dev"
	}
	v := strings.TrimSpace(string(data))
	if v == "" {
		return "dev"
	}
	return v
}

// BuildInfo returns build information as a map keyed by service name.
func BuildInfo(serviceName string) map[string]string {
	return map[string]string{
		"service":    serviceName,
		"version":    Version,
		"git_sha":    GitSHA,
		"build_time": BuildTime,
		"go_version": runtime.Version(),
	}
}

// BuildInfoHandler returns an http.HandlerFunc that responds with build info
// as JSON. Register as: r.Get("/v1/info", BuildInfoHandler("sda-app")). The
// payload is marshalled once at call time and reused for every request.
func BuildInfoHandler(serviceName string) http.HandlerFunc {
	payload, _ := json.Marshal(BuildInfo(serviceName))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}
}
