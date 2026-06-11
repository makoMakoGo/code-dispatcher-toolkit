package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type parserBarrierCmd struct {
	*fakeCmd
	parserStarted <-chan struct{}
}

func (c *parserBarrierCmd) Start() error {
	select {
	case <-c.parserStarted:
		return c.fakeCmd.Start()
	case <-time.After(500 * time.Millisecond):
		return errors.New("parser did not start before command")
	}
}

func TestRunExecutionLifecycleStartsParserBeforeCommand(t *testing.T) {
	origRunner := newCommandRunner
	t.Cleanup(func() { newCommandRunner = origRunner })

	parserStarted := make(chan struct{})
	var parserStartedOnce sync.Once
	parser := func(r io.Reader, warnFn func(string), infoFn func(string), onMessage func(), onComplete func()) (string, string) {
		parserStartedOnce.Do(func() { close(parserStarted) })
		data, _ := io.ReadAll(r)
		msg := strings.TrimSpace(string(data))
		if msg != "" {
			onMessage()
			onComplete()
		}
		return msg, "thread-fast"
	}

	fake := newFakeCmd(fakeCmdConfig{
		StdoutPlan: []fakeStdoutEvent{{Data: "fast\n"}},
		PID:        41,
	})
	cmd := &parserBarrierCmd{fakeCmd: fake, parserStarted: parserStarted}
	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		return cmd
	}

	res := runExecutionLifecycle(testExecutionLifecycleRequest(BackendInvocation{
		BackendName: "codex",
		Command:     "codex",
		Args:        []string{"exec", "--json", "task"},
		ParseStream: parser,
	}))

	if res.ExitCode != 0 || res.Message != "fast" || res.ThreadID != "thread-fast" {
		t.Fatalf("lifecycle result = %+v", res)
	}
	if fake.startCount.Load() != 1 {
		t.Fatalf("Start count = %d, want 1", fake.startCount.Load())
	}
}

func TestRunExecutionLifecycleWritesStdinThroughModule(t *testing.T) {
	origRunner := newCommandRunner
	t.Cleanup(func() { newCommandRunner = origRunner })

	fake := newFakeCmd(fakeCmdConfig{
		StdoutPlan: []fakeStdoutEvent{{Data: "ok\n"}},
		PID:        42,
	})
	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		return fake
	}

	req := testExecutionLifecycleRequest(testExecutionLifecycleInvocation("thread-stdin"))
	req.UseStdin = true
	req.TaskText = "payload from stdin"

	res := runExecutionLifecycle(req)
	if res.ExitCode != 0 || res.Message != "ok" || res.ThreadID != "thread-stdin" {
		t.Fatalf("lifecycle result = %+v", res)
	}
	waitForLifecycleCondition(t, func() bool {
		return fake.StdinContents() == "payload from stdin"
	}, "stdin payload was not written")
}

func TestRunExecutionLifecycleMapsStartFailure(t *testing.T) {
	origRunner := newCommandRunner
	t.Cleanup(func() { newCommandRunner = origRunner })

	fake := newFakeCmd(fakeCmdConfig{
		StartErr: errors.New(`exec: "missing": executable file not found in $PATH`),
		PID:      43,
	})
	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		return fake
	}

	req := testExecutionLifecycleRequest(testExecutionLifecycleInvocation("thread-start"))
	req.Invocation.Command = "missing"

	res := runExecutionLifecycle(req)
	if res.ExitCode != 127 {
		t.Fatalf("exit code = %d, want 127; result = %+v", res.ExitCode, res)
	}
	if !strings.Contains(res.Error, "missing command not found in PATH") {
		t.Fatalf("error = %q, want command-not-found message", res.Error)
	}
}

func TestRunExecutionLifecycleMapsTimeout(t *testing.T) {
	origRunner := newCommandRunner
	t.Cleanup(func() { newCommandRunner = origRunner })

	fake := newFakeCmd(fakeCmdConfig{
		KeepStdoutOpen:      true,
		BlockWait:           true,
		ReleaseWaitOnSignal: true,
		PID:                 44,
	})
	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		return fake
	}

	req := testExecutionLifecycleRequest(testExecutionLifecycleInvocation("thread-timeout"))
	req.TimeoutSec = 0

	res := runExecutionLifecycle(req)
	if res.ExitCode != 124 {
		t.Fatalf("exit code = %d, want 124; result = %+v", res.ExitCode, res)
	}
	if !strings.Contains(res.Error, "codex execution timeout") {
		t.Fatalf("error = %q, want timeout message", res.Error)
	}
	if fake.process.SignalCount() == 0 {
		t.Fatalf("expected timeout path to signal process")
	}
}

func testExecutionLifecycleRequest(invocation BackendInvocation) executionLifecycleRequest {
	return executionLifecycleRequest{
		Context:    context.Background(),
		Invocation: invocation,
		TaskText:   "task",
		TimeoutSec: 1,
		LogInfoFn:  func(string) {},
		LogWarnFn:  func(string) {},
		LogErrorFn: func(string) {},
	}
}

func testExecutionLifecycleInvocation(threadID string) BackendInvocation {
	return BackendInvocation{
		BackendName: "codex",
		Command:     "codex",
		Args:        []string{"exec", "--json", "task"},
		ParseStream: func(r io.Reader, warnFn func(string), infoFn func(string), onMessage func(), onComplete func()) (string, string) {
			data, _ := io.ReadAll(r)
			msg := strings.TrimSpace(string(data))
			if msg != "" {
				onMessage()
				onComplete()
			}
			return msg, threadID
		},
	}
}

func TestRunExecutionLifecycleClosesStderrOnNormalCompletion(t *testing.T) {
	origRunner := newCommandRunner
	t.Cleanup(func() { newCommandRunner = origRunner })

	fake := newFakeCmd(fakeCmdConfig{
		StdoutPlan: []fakeStdoutEvent{{Data: "done\n"}},
		PID:        44,
	})
	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		return fake
	}

	res := runExecutionLifecycle(testExecutionLifecycleRequest(testExecutionLifecycleInvocation("thread-close")))

	if res.ExitCode != 0 || res.Message != "done" {
		t.Fatalf("lifecycle result = %+v", res)
	}
	if got := fake.stderr.Reason(); got != stdoutCloseReasonWait {
		t.Fatalf("stderr close reason = %q, want %q (stderr reader leaked on normal completion)", got, stdoutCloseReasonWait)
	}
}

func TestRunExecutionLifecycleCapsStderrAttachedToError(t *testing.T) {
	origRunner := newCommandRunner
	t.Cleanup(func() { newCommandRunner = origRunner })

	fake := newFakeCmd(fakeCmdConfig{
		StderrPlan: []fakeStdoutEvent{{Data: strings.Repeat("E", 8*1024) + "\n"}},
		PID:        45,
	})
	newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
		return fake
	}

	res := runExecutionLifecycle(testExecutionLifecycleRequest(testExecutionLifecycleInvocation("thread-cap")))

	if res.ExitCode == 0 {
		t.Fatalf("expected empty-message failure, got %+v", res)
	}
	maxLen := stderrCaptureLimit + 200 // tail cap plus message prefix slack
	if len(res.Error) > maxLen {
		t.Fatalf("error length = %d, want <= %d (stderr tail must stay capped)", len(res.Error), maxLen)
	}
	if stderrCaptureLimit != 4*1024 {
		t.Fatalf("stderrCaptureLimit = %d, want 4096 (errors embed the tail; full stderr goes to the task log)", stderrCaptureLimit)
	}
}

func waitForLifecycleCondition(t *testing.T, ok func() bool, msg string) {
	t.Helper()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal(msg)
}
