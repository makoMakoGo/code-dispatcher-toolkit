package main

import (
	"encoding/json"
	"fmt"
	"strings"

	_ "embed"
)

//go:embed runtime_assets.json
var runtimeAssetsJSON []byte

// runtimeAssetManifest models only the manifest sections the dispatcher
// binary consumes at runtime: the backend list. Installer-only sections
// (runtime_config, binary, legacy_uninstall) are owned and validated by
// runtime_assets.py and test_runtime_assets.py.
type runtimeAssetManifest struct {
	Backends []runtimeBackendAsset `json:"backends"`
}

type runtimeBackendAsset struct {
	Name       string `json:"name"`
	PromptFile string `json:"prompt_file"`
}

var runtimeAssets = mustLoadRuntimeAssetManifest(runtimeAssetsJSON)

func mustLoadRuntimeAssetManifest(data []byte) runtimeAssetManifest {
	manifest, err := loadRuntimeAssetManifest(data)
	if err != nil {
		panic(err)
	}
	return manifest
}

func loadRuntimeAssetManifest(data []byte) (runtimeAssetManifest, error) {
	var manifest runtimeAssetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("load runtime asset manifest: %w", err)
	}
	if err := validateRuntimeAssetManifest(manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func validateRuntimeAssetManifest(manifest runtimeAssetManifest) error {
	if len(manifest.Backends) == 0 {
		return fmt.Errorf("runtime asset manifest missing backends")
	}
	seen := make(map[string]struct{}, len(manifest.Backends))
	for i, backend := range manifest.Backends {
		name := normalizeBackendName(backend.Name)
		if name == "" {
			return fmt.Errorf("runtime asset manifest backends[%d] missing name", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("runtime asset manifest duplicate backend %q", name)
		}
		seen[name] = struct{}{}
		if strings.TrimSpace(backend.PromptFile) == "" {
			return fmt.Errorf("runtime asset manifest backend %q missing prompt_file", name)
		}
	}
	return nil
}

func runtimePromptFileForBackend(backendName string) (string, bool) {
	name := normalizeBackendName(backendName)
	for _, backend := range runtimeAssets.Backends {
		if normalizeBackendName(backend.Name) == name {
			return backend.PromptFile, true
		}
	}
	return "", false
}

func runtimeBackendNames() []string {
	names := make([]string, 0, len(runtimeAssets.Backends))
	for _, backend := range runtimeAssets.Backends {
		names = append(names, normalizeBackendName(backend.Name))
	}
	return names
}

func supportedBackendNamesText() string {
	return strings.Join(runtimeBackendNames(), ", ")
}
