package main

import (
	"strings"
	"testing"
)

func TestRunParallelInvocationIgnoresNonParallelArgs(t *testing.T) {
	defer resetTestHooks()

	outcome := runParallelInvocation("code-dispatcher", []string{"--backend", "codex", "task"})
	if outcome.Handled {
		t.Fatalf("expected non-parallel args to be ignored, got %+v", outcome)
	}
}

func TestRunParallelInvocationSummaryOutput(t *testing.T) {
	defer resetTestHooks()

	stdinReader = strings.NewReader(`---TASK---
id: a
---CONTENT---
do a`)
	runParallelTaskFn = func(task TaskSpec, timeout int) TaskResult {
		if task.Backend != "codex" {
			t.Fatalf("task backend = %q, want codex", task.Backend)
		}
		return TaskResult{TaskID: task.ID, ExitCode: 0, Message: "done"}
	}

	outcome := runParallelInvocation("code-dispatcher", []string{"--parallel", "--backend", "codex"})
	if !outcome.Handled || outcome.ExitCode != 0 {
		t.Fatalf("outcome = %+v", outcome)
	}
	if !strings.Contains(outcome.Stdout, "=== Execution Report ===") || !strings.Contains(outcome.Stdout, "a") {
		t.Fatalf("stdout = %q, want summary report for task a", outcome.Stdout)
	}
	if outcome.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", outcome.Stderr)
	}
}

func TestRunParallelInvocationFullOutputAndExitCode(t *testing.T) {
	defer resetTestHooks()

	stdinReader = strings.NewReader(`---TASK---
id: a
---CONTENT---
do a
---TASK---
id: b
---CONTENT---
do b`)
	runParallelTaskFn = func(task TaskSpec, timeout int) TaskResult {
		if task.ID == "b" {
			return TaskResult{TaskID: task.ID, ExitCode: 7, Error: "failed"}
		}
		return TaskResult{TaskID: task.ID, ExitCode: 0, Message: "done"}
	}

	outcome := runParallelInvocation("code-dispatcher", []string{"--parallel", "--backend", "codex", "--full-output"})
	if !outcome.Handled || outcome.ExitCode != 7 {
		t.Fatalf("outcome = %+v", outcome)
	}
	if !strings.Contains(outcome.Stdout, "=== Parallel Execution Summary ===") {
		t.Fatalf("stdout = %q, want full output summary", outcome.Stdout)
	}
	if !strings.Contains(outcome.Stdout, "Status: FAILED (exit code 7)") {
		t.Fatalf("stdout = %q, want failed task status", outcome.Stdout)
	}
}

func TestRunParallelInvocationValidationError(t *testing.T) {
	defer resetTestHooks()

	stdinReader = strings.NewReader(`---TASK---
id: a
unknown: nope
---CONTENT---
do a`)

	outcome := runParallelInvocation("code-dispatcher", []string{"--parallel", "--backend", "codex"})
	if !outcome.Handled || outcome.ExitCode != 1 {
		t.Fatalf("outcome = %+v", outcome)
	}
	if !strings.Contains(outcome.Stderr, "unknown key: unknown") {
		t.Fatalf("stderr = %q, want unknown-key validation error", outcome.Stderr)
	}
}
