package main

import (
	"io"
	"strings"
)

// Backend defines the contract for invoking different AI CLI backends.
// Each backend owns the policy for turning dispatcher config into a ready
// backend invocation.
type Backend interface {
	Name() string
	BuildArgs(cfg *Config, targetArg string) []string
	Command() string
	BuildInvocation(cfg *Config, targetArg string) BackendInvocation
}

type BackendStreamParser func(r io.Reader, warnFn func(string), infoFn func(string), onMessage func(), onComplete func()) (message, threadID string)

type BackendInvocation struct {
	BackendName          string
	Command              string
	Args                 []string
	Env                  map[string]string
	UnsetEnvKeys         []string
	WorkDir              string
	StderrFilterPatterns []string
	ParseStream          BackendStreamParser
}

type CodexBackend struct{}

func (CodexBackend) Name() string { return "codex" }
func (CodexBackend) Command() string {
	return "codex"
}
func (CodexBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildCodexArgs(cfg, targetArg)
}
func (b CodexBackend) BuildInvocation(cfg *Config, targetArg string) BackendInvocation {
	return BackendInvocation{
		BackendName:          b.Name(),
		Command:              b.Command(),
		Args:                 b.BuildArgs(cfg, targetArg),
		Env:                  runtimeInjectedEnvForInvocation(),
		StderrFilterPatterns: codexNoisePatterns,
		ParseStream:          parseStreamForBackend(b.Name()),
	}
}

type ClaudeBackend struct{}

func (ClaudeBackend) Name() string { return "claude" }
func (ClaudeBackend) Command() string {
	return "claude"
}
func (ClaudeBackend) BuildArgs(cfg *Config, targetArg string) []string {
	return buildClaudeArgs(cfg, targetArg)
}
func (b ClaudeBackend) BuildInvocation(cfg *Config, targetArg string) BackendInvocation {
	workDir := ""
	if cfg != nil && cfg.Mode != "resume" && strings.TrimSpace(cfg.WorkDir) != "" {
		workDir = cfg.WorkDir
	}
	return BackendInvocation{
		BackendName:  b.Name(),
		Command:      b.Command(),
		Args:         b.BuildArgs(cfg, targetArg),
		Env:          runtimeInjectedEnvForInvocation(),
		UnsetEnvKeys: []string{"CLAUDECODE"},
		WorkDir:      workDir,
		ParseStream:  parseStreamForBackend(b.Name()),
	}
}

func runtimeEnvForInvocation() map[string]string {
	return runtimeInjectedEnvForInvocation()
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

func runtimeInjectedEnvForInvocation() map[string]string {
	env := runtimeInjectedEnv()
	if len(env) == 0 {
		return nil
	}
	return env
}

func parseStreamForBackend(backendName string) BackendStreamParser {
	return func(r io.Reader, warnFn func(string), infoFn func(string), onMessage func(), onComplete func()) (string, string) {
		return parseBackendStreamInternal(r, backendName, warnFn, infoFn, onMessage, onComplete)
	}
}

func legacyBackendInvocation(cfg *Config, commandName string, args []string) BackendInvocation {
	backendName := ""
	if cfg != nil {
		backendName = cfg.Backend
	}
	name := normalizeBackendName(backendName)
	invocation := BackendInvocation{
		BackendName:  name,
		Command:      commandName,
		Args:         args,
		Env:          runtimeEnvForInvocation(),
		UnsetEnvKeys: runtimeUnsetEnvKeysForBackend(name),
		ParseStream:  parseStreamForBackend(name),
	}
	if name == "codex" {
		invocation.StderrFilterPatterns = codexNoisePatterns
	}
	if name == "claude" && cfg != nil && cfg.Mode != "resume" && strings.TrimSpace(cfg.WorkDir) != "" {
		invocation.WorkDir = cfg.WorkDir
	}
	return invocation
}

func buildCodexArgs(cfg *Config, targetArg string) []string {
	if cfg == nil {
		panic("buildCodexArgs: nil config")
	}

	var resumeSessionID string
	isResume := cfg.Mode == "resume"
	if isResume {
		resumeSessionID = strings.TrimSpace(cfg.SessionID)
		if resumeSessionID == "" {
			logError("invalid config: resume mode requires non-empty session_id")
			isResume = false
		}
	}

	args := []string{"exec"}

	if model := resolveBackendModel("codex"); model != "" {
		args = append(args, "-m", model)
	}

	args = append(args, "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check")

	if isResume {
		return append(args,
			"--json",
			"resume",
			resumeSessionID,
			targetArg,
		)
	}

	return append(args,
		"-C", cfg.WorkDir,
		"--json",
		targetArg,
	)
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
