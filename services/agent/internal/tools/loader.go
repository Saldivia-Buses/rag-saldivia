package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ModuleManifest represents a module's tools.yaml file.
type ModuleManifest struct {
	Module string         `yaml:"module"`
	Name   string         `yaml:"name"`
	Tools  []ManifestTool `yaml:"tools"`
}

// ManifestTool is a tool definition from a YAML manifest.
type ManifestTool struct {
	ID                   string         `yaml:"id"`
	Service              string         `yaml:"service"`
	Method               string         `yaml:"method"`   // gRPC method name (protocol: grpc)
	Endpoint             string         `yaml:"endpoint"` // HTTP "VERB /path" (protocol: http)
	Protocol             string         `yaml:"protocol"` // "grpc" or "http"
	Type                 string         `yaml:"type"`
	RequiresConfirmation bool           `yaml:"requires_confirmation"`
	Description          string         `yaml:"description"`
	Parameters           map[string]any `yaml:"parameters"`
}

// LoadModuleTools reads tool manifests from a modules directory and returns
// tool definitions for all enabled modules. enabledModules is the set of
// module IDs enabled for the current tenant.
func LoadModuleTools(modulesDir string, enabledModules map[string]bool, serviceURLs map[string]string) ([]Definition, error) {
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no modules dir is fine
		}
		return nil, fmt.Errorf("read modules dir: %w", err)
	}

	var defs []Definition

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(modulesDir, entry.Name(), "tools.yaml")
		manifest, err := loadManifest(manifestPath)
		if err != nil {
			slog.Warn("skip module manifest", "path", manifestPath, "error", err)
			continue
		}

		if !enabledModules[manifest.Module] {
			slog.Debug("skip disabled module", "module", manifest.Module)
			continue
		}

		for _, t := range manifest.Tools {
			baseURL, ok := serviceURLs[t.Service]
			if !ok {
				slog.Warn("skip tool, no service URL", "tool", t.ID, "service", t.Service)
				continue
			}

			params, _ := json.Marshal(t.Parameters)

			// Resolve endpoint URL and HTTP method based on protocol
			httpMethod, fullURL := resolveEndpoint(baseURL, t)

			defs = append(defs, Definition{
				Name:                 t.ID,
				Service:              t.Service,
				Endpoint:             fullURL,
				Method:               httpMethod,
				Type:                 t.Type,
				RequiresConfirmation: t.RequiresConfirmation,
				Description:          t.Description,
				Parameters:           params,
			})
		}

		slog.Info("loaded module tools", "module", manifest.Module, "tools", len(manifest.Tools))
	}

	return defs, nil
}

// resolveEndpoint determines the HTTP method and full URL for a tool.
// Supports two protocols:
//   - grpc: Method field is a gRPC method name → POST baseURL/method
//   - http:  Endpoint field is "VERB /path" → VERB baseURL+path
func resolveEndpoint(baseURL string, t ManifestTool) (httpMethod, fullURL string) {
	switch t.Protocol {
	case "http":
		// Endpoint format: "POST /v1/astro/natal" or "GET /v1/astro/contacts"
		if t.Endpoint != "" {
			parts := strings.SplitN(t.Endpoint, " ", 2)
			if len(parts) == 2 {
				return strings.ToUpper(parts[0]), baseURL + parts[1]
			}
			// Endpoint is just a path without verb → default to POST
			return "POST", baseURL + t.Endpoint
		}
		// Fallback to method field
		return "POST", baseURL + "/" + t.Method
	default:
		// gRPC or unspecified: original behavior
		return "POST", baseURL + "/" + t.Method
	}
}

func loadManifest(path string) (*ModuleManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m ModuleManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &m, nil
}
