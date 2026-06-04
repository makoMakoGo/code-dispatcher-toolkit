package main

import "strings"

// Backend defines the contract for invoking different AI CLI backends.
// Each backend is responsible for supplying the executable command and
// building the argument list based on the dispatcher config.
type Backend interface {
	Name() string
	BuildArgs(cfg *Config, targetArg string) []string
	Command() string
}

type CodexBackend struct{}

func (CodexBackend) Name() string { return "codex" }
func (CodexBackend) Command() string {
	return "codex"
}
func (CodexBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildCodexArgs(cfg, targetArg)
}

type ClaudeBackend struct{}

func (ClaudeBackend) Name() string { return "claude" }
func (ClaudeBackend) Command() string {
	return "claude"
}
func (ClaudeBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildClaudeArgs(cfg, targetArg)
}

func runtimeEnvForBackend(backendName string) map[string]string {
	env := runtimeInjectedEnv()
	if len(env) == 0 {
		return nil
	}

	return env
}

func runtimeUnsetEnvKeysForBackend(backendName string) []string {
	if normalizeBackendName(backendName) != "claude" {
		return nil
	}
	return []string{"CLAUDECODE"}
}

func normalizeBackendName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func resolveBackendModel(backendName string) string {
	key := "CODE_DISPATCHER_" + strings.ToUpper(normalizeBackendName(backendName)) + "_MODEL"
	val, ok := lookupRuntimeSetting(key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(val)
}

func buildClaudeArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		return nil
	}
	args := []string{"-p", "--dangerously-skip-permissions"}

	if cfg.Mode == "resume" {
		if cfg.SessionID != "" {
			// Claude CLI uses -r <session_id> for resume.
			args = append(args, "-r", cfg.SessionID)
		}
	}
	// Note: claude CLI doesn't support -C flag; workdir set via cmd.Dir

	args = append(args, "--output-format", "stream-json", "--verbose", targetArg)

	return args
}
