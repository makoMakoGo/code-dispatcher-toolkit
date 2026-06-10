package main

import (
	"path/filepath"
	"testing"
)

func TestRuntimeAssetManifestBackendsMatchRegistry(t *testing.T) {
	names := runtimeBackendNames()
	if len(names) != len(backendRegistry) {
		t.Fatalf("manifest backends = %v, registry has %d entries", names, len(backendRegistry))
	}

	for _, name := range names {
		if _, ok := backendRegistry[name]; !ok {
			t.Fatalf("manifest backend %q missing from backendRegistry", name)
		}
	}
}

func TestRuntimeAssetManifestSupportedBackendNamesText(t *testing.T) {
	if got := supportedBackendNamesText(); got != "codex, claude" {
		t.Fatalf("supportedBackendNamesText() = %q", got)
	}
}

func TestRuntimeAssetManifestPromptLookup(t *testing.T) {
	promptBase := createPromptBaseForTest(t)

	for _, backend := range runtimeAssets.Backends {
		got := defaultPromptFileForBackend(backend.Name)
		want := filepath.Join(promptBase, backend.PromptFile)
		if got != want {
			t.Fatalf("defaultPromptFileForBackend(%q) = %q, want %q", backend.Name, got, want)
		}
	}

	if got := defaultPromptFileForBackend("unknown"); got != "" {
		t.Fatalf("defaultPromptFileForBackend(unknown) = %q, want empty", got)
	}
}

func TestRuntimeAssetManifestInstallTargets(t *testing.T) {
	if runtimeAssets.RuntimeConfig.Template != "templates/runtime-config.env" {
		t.Fatalf("runtime config template = %q", runtimeAssets.RuntimeConfig.Template)
	}
	if runtimeAssets.RuntimeConfig.InstallPath != ".env" {
		t.Fatalf("runtime config install path = %q", runtimeAssets.RuntimeConfig.InstallPath)
	}
	if runtimeAssets.Binary.InstallDir != "bin" {
		t.Fatalf("binary install dir = %q", runtimeAssets.Binary.InstallDir)
	}
	if len(runtimeAssets.Binary.ReleaseAssets) != 3 {
		t.Fatalf("release asset count = %d, want 3", len(runtimeAssets.Binary.ReleaseAssets))
	}
}
