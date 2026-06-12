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

func TestValidateRuntimeAssetManifestRejectsInvalidBackends(t *testing.T) {
	cases := []struct {
		name     string
		manifest runtimeAssetManifest
		wantErr  string
	}{
		{
			name:     "empty backends",
			manifest: runtimeAssetManifest{},
			wantErr:  "runtime asset manifest missing backends",
		},
		{
			name: "missing name",
			manifest: runtimeAssetManifest{Backends: []runtimeBackendAsset{
				{Name: "  ", PromptFile: "codex-prompt.md"},
			}},
			wantErr: "runtime asset manifest backends[0] missing name",
		},
		{
			name: "duplicate backend",
			manifest: runtimeAssetManifest{Backends: []runtimeBackendAsset{
				{Name: "codex", PromptFile: "codex-prompt.md"},
				{Name: "Codex", PromptFile: "other.md"},
			}},
			wantErr: `runtime asset manifest duplicate backend "codex"`,
		},
		{
			name: "missing prompt file",
			manifest: runtimeAssetManifest{Backends: []runtimeBackendAsset{
				{Name: "codex", PromptFile: " "},
			}},
			wantErr: `runtime asset manifest backend "codex" missing prompt_file`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRuntimeAssetManifest(tc.manifest)
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("validateRuntimeAssetManifest error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
