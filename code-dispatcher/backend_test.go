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
		want := []string{"exec", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/tmp", "--json", "task"}
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

func TestBackendInvocationPolicy(t *testing.T) {
	t.Run("codex owns command args env filter and parser policy", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{
			"CODE_DISPATCHER_TIMEOUT": "10",
			"OPENAI_API_KEY":          "secret",
		})
		t.Cleanup(resetRuntimeSettingsForTest)

		cfg := &Config{Mode: "new", WorkDir: "/repo", Backend: "codex"}
		invocation := CodexBackend{}.BuildInvocation(cfg, "task")

		if invocation.BackendName != "codex" {
			t.Fatalf("BackendName = %q", invocation.BackendName)
		}
		if invocation.Command != "codex" {
			t.Fatalf("Command = %q", invocation.Command)
		}
		wantArgs := []string{"exec", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/repo", "--json", "task"}
		if !reflect.DeepEqual(invocation.Args, wantArgs) {
			t.Fatalf("Args = %v, want %v", invocation.Args, wantArgs)
		}
		if invocation.Env["OPENAI_API_KEY"] != "secret" {
			t.Fatalf("Env = %v, want OPENAI_API_KEY", invocation.Env)
		}
		if _, ok := invocation.Env["CODE_DISPATCHER_TIMEOUT"]; ok {
			t.Fatalf("Env should not include dispatcher control keys: %v", invocation.Env)
		}
		if invocation.WorkDir != "" {
			t.Fatalf("Codex WorkDir = %q, want empty because -C carries workdir", invocation.WorkDir)
		}
		if invocation.ParseStream == nil {
			t.Fatalf("Codex invocation missing stream parser")
		}
	})

	t.Run("claude owns command args env unset workdir and parser policy", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{
			"CODE_DISPATCHER_TIMEOUT": "10",
			"ANTHROPIC_API_KEY":       "secret",
		})
		t.Cleanup(resetRuntimeSettingsForTest)

		cfg := &Config{Mode: "new", WorkDir: "/repo", Backend: "claude"}
		invocation := ClaudeBackend{}.BuildInvocation(cfg, "task")

		if invocation.BackendName != "claude" {
			t.Fatalf("BackendName = %q", invocation.BackendName)
		}
		if invocation.Command != "claude" {
			t.Fatalf("Command = %q", invocation.Command)
		}
		wantArgs := []string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose", "task"}
		if !reflect.DeepEqual(invocation.Args, wantArgs) {
			t.Fatalf("Args = %v, want %v", invocation.Args, wantArgs)
		}
		if invocation.Env["ANTHROPIC_API_KEY"] != "secret" {
			t.Fatalf("Env = %v, want ANTHROPIC_API_KEY", invocation.Env)
		}
		if !reflect.DeepEqual(invocation.UnsetEnvKeys, []string{"CLAUDECODE"}) {
			t.Fatalf("UnsetEnvKeys = %v", invocation.UnsetEnvKeys)
		}
		if invocation.WorkDir != "/repo" {
			t.Fatalf("WorkDir = %q, want /repo", invocation.WorkDir)
		}
		if invocation.ParseStream == nil {
			t.Fatalf("Claude invocation missing stream parser")
		}
	})

	t.Run("claude resume keeps explicit workdir", func(t *testing.T) {
		// Claude stores sessions per project directory; resume only finds the
		// conversation when the process runs in the original task workdir.
		cfg := &Config{Mode: "resume", SessionID: "sid", WorkDir: "/repo", Backend: "claude"}
		invocation := ClaudeBackend{}.BuildInvocation(cfg, "task")
		if invocation.WorkDir != "/repo" {
			t.Fatalf("resume WorkDir = %q, want /repo", invocation.WorkDir)
		}
	})

	t.Run("claude resume without workdir stays empty", func(t *testing.T) {
		cfg := &Config{Mode: "resume", SessionID: "sid", WorkDir: "  ", Backend: "claude"}
		invocation := ClaudeBackend{}.BuildInvocation(cfg, "task")
		if invocation.WorkDir != "" {
			t.Fatalf("resume WorkDir = %q, want empty", invocation.WorkDir)
		}
	})

	t.Run("legacy claude resume keeps explicit workdir", func(t *testing.T) {
		cfg := &Config{Mode: "resume", SessionID: "sid", WorkDir: "/repo", Backend: "claude"}
		invocation := legacyBackendInvocation(cfg, "claude", []string{"-p"})
		if invocation.WorkDir != "/repo" {
			t.Fatalf("legacy resume WorkDir = %q, want /repo", invocation.WorkDir)
		}
	})

	t.Run("legacy path inherits registry backend policy", func(t *testing.T) {
		cfg := &Config{Mode: "new", WorkDir: "/repo", Backend: "claude"}
		invocation := legacyBackendInvocation(cfg, "fake-claude", []string{"-p", "injected"})
		if invocation.Command != "fake-claude" {
			t.Fatalf("Command = %q, want injected fake-claude", invocation.Command)
		}
		if !reflect.DeepEqual(invocation.Args, []string{"-p", "injected"}) {
			t.Fatalf("Args = %v, want injected args", invocation.Args)
		}
		if !reflect.DeepEqual(invocation.UnsetEnvKeys, []string{"CLAUDECODE"}) {
			t.Fatalf("UnsetEnvKeys = %v, want CLAUDECODE from backend policy", invocation.UnsetEnvKeys)
		}
		if invocation.WorkDir != "/repo" {
			t.Fatalf("WorkDir = %q, want /repo from backend policy", invocation.WorkDir)
		}
		if invocation.ParseStream == nil {
			t.Fatalf("legacy invocation missing stream parser")
		}
	})

	t.Run("legacy unknown backend falls back to generic invocation", func(t *testing.T) {
		cfg := &Config{Mode: "new", WorkDir: "/repo", Backend: "mystery"}
		invocation := legacyBackendInvocation(cfg, "mystery-cmd", []string{"-x"})
		if invocation.BackendName != "mystery" {
			t.Fatalf("BackendName = %q, want mystery", invocation.BackendName)
		}
		if invocation.Command != "mystery-cmd" {
			t.Fatalf("Command = %q, want mystery-cmd", invocation.Command)
		}
		if !reflect.DeepEqual(invocation.Args, []string{"-x"}) {
			t.Fatalf("Args = %v, want [-x]", invocation.Args)
		}
		if len(invocation.UnsetEnvKeys) != 0 || invocation.WorkDir != "" {
			t.Fatalf("generic invocation should carry no backend policy, got %+v", invocation)
		}
		if invocation.ParseStream == nil {
			t.Fatalf("generic invocation missing stream parser")
		}
	})
}

func TestRuntimeInjectedEnvForInvocation(t *testing.T) {
	t.Run("returns nil when no runtime settings", func(t *testing.T) {
		setRuntimeSettingsForTest(map[string]string{})
		t.Cleanup(resetRuntimeSettingsForTest)
		if got := runtimeInjectedEnvForInvocation(); len(got) != 0 {
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

		got := runtimeInjectedEnvForInvocation()
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
	want := []string{"exec", "-m", "o3", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/tmp", "--json", "task"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
