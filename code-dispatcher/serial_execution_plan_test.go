package main

import (
	"context"
	"io"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

type countingInvocationBackend struct {
	builds atomic.Int32
}

func (b *countingInvocationBackend) Name() string { return "codex" }
func (b *countingInvocationBackend) Command() string {
	return "planned-backend"
}
func (b *countingInvocationBackend) BuildArgs(_ *Config, targetArg string) []string {
	return []string{"--planned", targetArg}
}
func (b *countingInvocationBackend) BuildInvocation(cfg *Config, targetArg string) BackendInvocation {
	b.builds.Add(1)
	return BackendInvocation{
		BackendName: b.Name(),
		Command:     b.Command(),
		Args:        b.BuildArgs(cfg, targetArg),
		ParseStream: func(r io.Reader, _ func(string), _ func(string), onMessage func(), onComplete func()) (string, string) {
			data, _ := io.ReadAll(r)
			message := strings.TrimSpace(string(data))
			if message != "" && onMessage != nil {
				onMessage()
			}
			if onComplete != nil {
				onComplete()
			}
			return message, "planned-session"
		},
	}
}

func TestSerialPlanExecutesTheInvocationItDisplays(t *testing.T) {
	resetTestHooks()
	t.Cleanup(resetTestHooks)
	stdinReader = strings.NewReader("")
	isTerminalFn = func() bool { return true }

	backend := &countingInvocationBackend{}
	cfg := &Config{Mode: "new", Task: "task", WorkDir: ".", Backend: "codex"}
	plan, err := planSerialInvocation(cfg, backend)
	if err != nil {
		t.Fatalf("planSerialInvocation() error = %v", err)
	}
	if plan.Invocation == nil || plan.Invocation != plan.TaskSpec.plannedInvocation {
		t.Fatalf("serial task does not retain the exact planned invocation")
	}

	fake := newFakeCmd(fakeCmdConfig{
		StdoutPlan: []fakeStdoutEvent{{Data: "done\n"}},
		PID:        77,
	})
	var startedCommand string
	var startedArgs []string
	newCommandRunner = func(_ context.Context, name string, args ...string) commandRunner {
		startedCommand = name
		startedArgs = append([]string(nil), args...)
		return fake
	}

	result := runTaskWithContext(context.Background(), plan.TaskSpec, nil, true, 2)
	if result.ExitCode != 0 || result.Message != "done" || result.SessionID != "planned-session" {
		t.Fatalf("runTaskWithContext() result = %+v", result)
	}
	if got := backend.builds.Load(); got != 1 {
		t.Fatalf("BuildInvocation called %d times, want exactly once during planning", got)
	}
	if startedCommand != plan.Invocation.Command {
		t.Fatalf("started command = %q, displayed command = %q", startedCommand, plan.Invocation.Command)
	}
	if !reflect.DeepEqual(startedArgs, plan.Invocation.Args) {
		t.Fatalf("started args = %v, displayed args = %v", startedArgs, plan.Invocation.Args)
	}
}

func TestResolveTaskInvocationRejectsUnknownBackend(t *testing.T) {
	resetTestHooks()
	t.Cleanup(resetTestHooks)

	cfg := &Config{Mode: "new", Task: "task", WorkDir: ".", Backend: "unknown"}
	_, err := resolveTaskInvocation(TaskSpec{Backend: "unknown"}, cfg, nil, "task")
	if err == nil || !strings.Contains(err.Error(), `unsupported backend "unknown"`) {
		t.Fatalf("resolveTaskInvocation() error = %v, want unsupported backend", err)
	}
}
