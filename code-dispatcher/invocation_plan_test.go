package main

import (
	"context"
	"strings"
	"testing"
)

type countingBackend struct {
	builds int
}

func (b *countingBackend) Name() string { return "counting" }

func (b *countingBackend) BuildInvocation(_ *Config, targetArg string) BackendInvocation {
	b.builds++
	return BackendInvocation{
		BackendName: b.Name(),
		Command:     "counting",
		Args:        []string{targetArg},
		ParseStream: parseJSONStreamInternal,
	}
}

func TestRunTaskExecutesPreparedInvocationWithoutReplanning(t *testing.T) {
	defer resetTestHooks()

	backend := &countingBackend{}
	task, invocation, err := planTaskInvocation(TaskSpec{
		Task:    "task",
		WorkDir: defaultWorkdir,
		Mode:    "new",
	}, backend)
	if err != nil {
		t.Fatalf("planTaskInvocation() error = %v", err)
	}
	if backend.builds != 1 {
		t.Fatalf("BuildInvocation calls = %d, want 1", backend.builds)
	}
	if invocation.Command != "counting" {
		t.Fatalf("Command = %q, want counting", invocation.Command)
	}

	fake := newFakeCmd(fakeCmdConfig{
		StdoutPlan: []fakeStdoutEvent{{Data: `{"type":"item.completed","item":{"type":"agent_message","text":"done"}}` + "\n"}},
	})
	newCommandRunner = func(context.Context, string, ...string) commandRunner { return fake }

	result := runTask(task, false, 5)
	if result.ExitCode != 0 || result.Message != "done" {
		t.Fatalf("runTask() result = %+v", result)
	}
	if backend.builds != 1 {
		t.Fatalf("BuildInvocation calls after execution = %d, want 1", backend.builds)
	}
}

func TestRunTaskRejectsUnknownBackendWithoutGenericFallback(t *testing.T) {
	defer resetTestHooks()

	started := false
	newCommandRunner = func(context.Context, string, ...string) commandRunner {
		started = true
		return newFakeCmd(fakeCmdConfig{})
	}

	result := runTask(TaskSpec{Task: "task", Backend: "mystery"}, false, 5)
	if result.ExitCode != 1 || !strings.Contains(result.Error, `unsupported backend "mystery"`) {
		t.Fatalf("runTask() result = %+v", result)
	}
	if started {
		t.Fatal("command runner started for an unsupported backend")
	}
}
