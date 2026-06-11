package main

import (
	"encoding/json"
	"fmt"
	"strings"

	_ "embed"
)

//go:embed runtime_assets.json
var runtimeAssetsJSON []byte

type runtimeAssetManifest struct {
	RuntimeConfig   runtimeConfigAsset          `json:"runtime_config"`
	Binary          runtimeBinaryAsset          `json:"binary"`
	Backends        []runtimeBackendAsset       `json:"backends"`
	LegacyUninstall runtimeLegacyUninstallAsset `json:"legacy_uninstall"`
}

type runtimeConfigAsset struct {
	Template    string `json:"template"`
	InstallPath string `json:"install_path"`
}

type runtimeBinaryAsset struct {
	InstallDir    string                `json:"install_dir"`
	NameByOS      map[string]string     `json:"name_by_os"`
	ReleaseAssets []runtimeReleaseAsset `json:"release_assets"`
}

type runtimeReleaseAsset struct {
	System   string   `json:"system"`
	Machines []string `json:"machines"`
	Asset    string   `json:"asset"`
}

type runtimeBackendAsset struct {
	Name           string `json:"name"`
	PromptTemplate string `json:"prompt_template"`
	PromptFile     string `json:"prompt_file"`
}

type runtimeLegacyUninstallAsset struct {
	PromptFiles []string `json:"prompt_files"`
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
	if strings.TrimSpace(manifest.RuntimeConfig.Template) == "" {
		return fmt.Errorf("runtime asset manifest missing runtime_config.template")
	}
	if strings.TrimSpace(manifest.RuntimeConfig.InstallPath) == "" {
		return fmt.Errorf("runtime asset manifest missing runtime_config.install_path")
	}
	if strings.TrimSpace(manifest.Binary.InstallDir) == "" {
		return fmt.Errorf("runtime asset manifest missing binary.install_dir")
	}
	if strings.TrimSpace(manifest.Binary.NameByOS["default"]) == "" {
		return fmt.Errorf("runtime asset manifest missing binary.name_by_os.default")
	}
	if strings.TrimSpace(manifest.Binary.NameByOS["nt"]) == "" {
		return fmt.Errorf("runtime asset manifest missing binary.name_by_os.nt")
	}
	if len(manifest.Binary.ReleaseAssets) == 0 {
		return fmt.Errorf("runtime asset manifest missing binary.release_assets")
	}
	for i, asset := range manifest.Binary.ReleaseAssets {
		if strings.TrimSpace(asset.System) == "" {
			return fmt.Errorf("runtime asset manifest release_assets[%d] missing system", i)
		}
		if len(asset.Machines) == 0 {
			return fmt.Errorf("runtime asset manifest release_assets[%d] missing machines", i)
		}
		if strings.TrimSpace(asset.Asset) == "" {
			return fmt.Errorf("runtime asset manifest release_assets[%d] missing asset", i)
		}
	}

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
		if strings.TrimSpace(backend.PromptTemplate) == "" {
			return fmt.Errorf("runtime asset manifest backend %q missing prompt_template", name)
		}
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
