package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestPlanSerialInvocationInlineCodexPreview(t *testing.T) {
	defer resetTestHooks()
	useEmptyPromptHome(t)
	stdinReader = strings.NewReader("")
	isTerminalFn = func() bool { return true }

	cfg := &Config{Mode: "new", Task: "do work", WorkDir: "/repo", Backend: "codex", Timeout: 9}
	plan, err := planSerialInvocation(cfg, CodexBackend{})
	if err != nil {
		t.Fatalf("planSerialInvocation: %v", err)
	}

	if plan.TaskSpec.Task != "do work" || plan.TaskSpec.UseStdin {
		t.Fatalf("TaskSpec = %+v", plan.TaskSpec)
	}
	if plan.TaskSpec.Backend != "codex" {
		t.Fatalf("TaskSpec.Backend = %q, want codex", plan.TaskSpec.Backend)
	}
	wantArgs := []string{"exec", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "-C", "/repo", "--json", "do work"}
	if plan.Invocation.Command != "codex" || !reflect.DeepEqual(plan.Invocation.Args, wantArgs) {
		t.Fatalf("invocation = %s %v, want codex %v", plan.Invocation.Command, plan.Invocation.Args, wantArgs)
	}
}

func TestPlanSerialInvocationExplicitStdin(t *testing.T) {
	defer resetTestHooks()
	useEmptyPromptHome(t)
	stdinReader = strings.NewReader("stdin task")
	isTerminalFn = func() bool { return true }

	cfg := &Config{Mode: "new", Task: "-", WorkDir: "/repo", Backend: "codex", ExplicitStdin: true}
	plan, err := planSerialInvocation(cfg, CodexBackend{})
	if err != nil {
		t.Fatalf("planSerialInvocation: %v", err)
	}

	if plan.TaskSpec.Task != "stdin task" || !plan.TaskSpec.UseStdin {
		t.Fatalf("TaskSpec = %+v", plan.TaskSpec)
	}
	if !hasSerialReason(plan.Reasons, `explicit "-"`) {
		t.Fatalf("Reasons = %v, want explicit stdin reason", plan.Reasons)
	}
	if got := plan.Invocation.Args[len(plan.Invocation.Args)-1]; got != "-" {
		t.Fatalf("last arg = %q, want stdin target", got)
	}
}

func TestPlanSerialInvocationPipedInput(t *testing.T) {
	defer resetTestHooks()
	useEmptyPromptHome(t)
	stdinReader = strings.NewReader("piped task")
	isTerminalFn = func() bool { return false }

	cfg := &Config{Mode: "new", Task: "ignored", WorkDir: "/repo", Backend: "codex"}
	plan, err := planSerialInvocation(cfg, CodexBackend{})
	if err != nil {
		t.Fatalf("planSerialInvocation: %v", err)
	}

	if plan.TaskSpec.Task != "piped task" || !plan.TaskSpec.UseStdin {
		t.Fatalf("TaskSpec = %+v", plan.TaskSpec)
	}
	if !hasSerialReason(plan.Reasons, "piped input") {
		t.Fatalf("Reasons = %v, want piped input reason", plan.Reasons)
	}
	if got := plan.Invocation.Args[len(plan.Invocation.Args)-1]; got != "-" {
		t.Fatalf("last arg = %q, want stdin target", got)
	}
}

func TestPlanSerialInvocationResumePreviewUsesBackendPolicy(t *testing.T) {
	defer resetTestHooks()
	useEmptyPromptHome(t)
	stdinReader = strings.NewReader("")
	isTerminalFn = func() bool { return true }

	cfg := &Config{Mode: "resume", SessionID: "sid", Task: "continue", WorkDir: "/repo", Backend: "codex"}
	plan, err := planSerialInvocation(cfg, CodexBackend{})
	if err != nil {
		t.Fatalf("planSerialInvocation: %v", err)
	}

	wantArgs := []string{"exec", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "--json", "resume", "sid", "continue"}
	if !reflect.DeepEqual(plan.Invocation.Args, wantArgs) {
		t.Fatalf("args = %v, want %v", plan.Invocation.Args, wantArgs)
	}
	if plan.TaskSpec.Mode != "resume" || plan.TaskSpec.SessionID != "sid" {
		t.Fatalf("TaskSpec = %+v", plan.TaskSpec)
	}
}

func TestPlanSerialInvocationInjectsDefaultPrompt(t *testing.T) {
	defer resetTestHooks()
	stdinReader = strings.NewReader("")
	isTerminalFn = func() bool { return true }

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	promptPath := filepath.Join(home, ".code-dispatcher", "prompts", "codex-prompt.md")
	if err := os.MkdirAll(filepath.Dir(promptPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(promptPath, []byte("PROMPT\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg := &Config{Mode: "new", Task: "do", WorkDir: "/repo", Backend: "codex"}
	plan, err := planSerialInvocation(cfg, CodexBackend{})
	if err != nil {
		t.Fatalf("planSerialInvocation: %v", err)
	}

	wantTask := "<agent-prompt>\nPROMPT\n</agent-prompt>\n\ndo"
	if plan.TaskSpec.Task != wantTask {
		t.Fatalf("task = %q, want %q", plan.TaskSpec.Task, wantTask)
	}
	if !plan.TaskSpec.UseStdin {
		t.Fatalf("prompt-wrapped task should use stdin")
	}
	if got := plan.Invocation.Args[len(plan.Invocation.Args)-1]; got != "-" {
		t.Fatalf("last arg = %q, want stdin target", got)
	}
}

func hasSerialReason(reasons []string, want string) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}

func useEmptyPromptHome(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
}
