package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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
	Method               string         `yaml:"method"`
	Protocol             string         `yaml:"protocol"`
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

			defs = append(defs, Definition{
				Name:                 t.ID,
				Service:              t.Service,
				Endpoint:             baseURL + "/" + t.Method, // simplified — real impl uses gRPC
				Method:               "POST",
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
