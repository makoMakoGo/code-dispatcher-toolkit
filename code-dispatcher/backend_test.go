package main

import (
	"reflect"
	"testing"
)

func TestClaudeBuildArgs_ModesAndPermissions(t *testing.T) {
	backend := ClaudeBackend{}

	t.Run("new mode always includes skip-permissions", func(t *testing.T) {
		cfg := &Config{Mode: "new", WorkDir: "/repo"}
		got := backend.BuildArgs(cfg, "todo")
		want := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "todo"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("stdin mode", func(t *testing.T) {
		cfg := &Config{Mode: "new", WorkDir: "/repo"}
		got := backend.BuildArgs(cfg, "-")
		want := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "-"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("resume mode includes session id", func(t *testing.T) {
		cfg := &Config{Mode: "resume", SessionID: "sid-123", WorkDir: "/ignored"}
		got := backend.BuildArgs(cfg, "resume-task")
		want := []string{"-p", "--dangerously-skip-permissions", "-r", "sid-123", "--output-format", "stream-json", "--verbose", "resume-task"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("resume mode without session still returns base flags", func(t *testing.T) {
		cfg := &Config{Mode: "resume", WorkDir: "/ignored"}
		got := backend.BuildArgs(cfg, "follow-up")
		want := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "follow-up"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("nil config returns nil", func(t *testing.T) {
		if backend.BuildArgs(nil, "ignored") != nil {
			t.Fatalf("nil config should return nil args")
		}
	})
}

func TestVariousBackendsBuildArgs(t *testing.T) {
	setRuntimeSettingsForTest(map[string]string{})
	t.Cleanup(resetRuntimeSettingsForTest)

	t.Run("codex build args always includes bypass flag", func(t *testing.T) {
		backend := CodexBackend{}
		cfg := &Config{Mode: "new", WorkDir: "/tmp"}
		got := backend.BuildArgs(cfg, "task")
		want := []string{"e", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/tmp", "--json", "task"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

}

func TestClaudeBuildArgs_BackendMetadata(t *testing.T) {
	tests := []struct {
		backend Backend
		name    string
		command string
	}{
		{backend: CodexBackend{}, name: "codex", command: "codex"},
		{backend: ClaudeBackend{}, name: "claude", command: "claude"},
	}

	for _, tt := range tests {
		if got := tt.backend.Name(); got != tt.name {
			t.Fatalf("Name() = %s, want %s", got, tt.name)
		}
		if got := tt.backend.Command(); got != tt.command {
			t.Fatalf("Command() = %s, want %s", got, tt.command)
		}
	}
}

func TestRuntimeEnvForBackend(t *testing.T) {
	t.Run("returns nil when no runtime settings", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := runtimeEnvForBackend("claude"); len(got) != 0 {
			t.Fatalf("got %v, want nil/empty", got)
		}
	})

	t.Run("filters dispatcher control keys", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{
			"CODE_DISPATCHER_SKIP_PERMISSIONS": "false",
			"CODE_DISPATCHER_TIMEOUT":          "7200",
			"ANTHROPIC_API_KEY":                "secret",
			"FOO":                              "bar",
		})
		t.Cleanup(resetRuntimeSettingsForTest)

		got := runtimeEnvForBackend("claude")
		if got["ANTHROPIC_API_KEY"] != "secret" || got["FOO"] != "bar" {
			t.Fatalf("got %v, want ANTHROPIC_API_KEY/FOO", got)
		}
		if _, ok := got["CODE_DISPATCHER_TIMEOUT"]; ok {
			t.Fatalf("got %v, control key CODE_DISPATCHER_TIMEOUT should be filtered", got)
		}
		if _, ok := got["CODE_DISPATCHER_SKIP_PERMISSIONS"]; ok {
			t.Fatalf("got %v, control key CODE_DISPATCHER_SKIP_PERMISSIONS should be filtered", got)
		}
	})

	t.Run("claude unsets nested session markers", func(t *testing.T) {
		got := runtimeUnsetEnvKeysForBackend("claude")
		want := []string{"CLAUDECODE"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("non claude backend does not unset nested markers", func(t *testing.T) {
		if got := runtimeUnsetEnvKeysForBackend("codex"); len(got) != 0 {
			t.Fatalf("got %v, want nil/empty", got)
		}
	})
}

func TestResolveBackendModel(t *testing.T) {
	t.Run("returns empty when not set", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := resolveBackendModel("claude"); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("codex model from env", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{"CODE_DISPATCHER_CODEX_MODEL": "o3"})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := resolveBackendModel("codex"); got != "o3" {
			t.Fatalf("got %q, want o3", got)
		}
	})

	t.Run("codex model whitespace trimmed", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{"CODE_DISPATCHER_CODEX_MODEL": "  o4-mini  "})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := resolveBackendModel("codex"); got != "o4-mini" {
			t.Fatalf("got %q, want o4-mini", got)
		}
	})

	t.Run("whitespace-only treated as empty", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{"CODE_DISPATCHER_CODEX_MODEL": "   "})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := resolveBackendModel("codex"); got != "" {
			t.Fatalf("got %q, want empty for whitespace-only value", got)
		}
	})
}

func TestCodexBuildArgs_WithModel(t *testing.T) {
	setRuntimeSettingsForTest(map[string]string{"CODE_DISPATCHER_CODEX_MODEL": "o3"})
	t.Cleanup(resetRuntimeSettingsForTest)

	cfg := &Config{Mode: "new", WorkDir: "/tmp"}
	got := buildCodexArgs(cfg, "task")
	want := []string{"e", "-m", "o3", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/tmp", "--json", "task"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
