package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

const postMessageTerminateDelay = 1 * time.Second
const forceKillWaitTimeout = 5 * time.Second

type executionLifecycleRequest struct {
	Context      context.Context
	Invocation   BackendInvocation
	TaskText     string
	UseStdin     bool
	TimeoutSec   int
	LogPath      string
	StdoutLogger *logWriter
	StderrLogger *logWriter
	StderrOutput io.Writer
	LogInfoFn    func(string)
	LogWarnFn    func(string)
	LogErrorFn   func(string)
}

type executionLifecycleResult struct {
	Message  string
	ThreadID string
	ExitCode int
	Error    string
}

func runExecutionLifecycle(req executionLifecycleRequest) executionLifecycleResult {
	commandName := req.Invocation.Command
	backendArgs := req.Invocation.Args
	stderrBuf := &tailBuffer{limit: stderrCaptureLimit}

	ctx := req.Context
	ctx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutSec)*time.Second)
	defer cancel()
	notifyCtx := signalNotifyCtxFn
	if notifyCtx == nil {
		notifyCtx = signal.NotifyContext
	}
	ctx, stop := notifyCtx(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	attachStderr := func(msg string) string {
		return fmt.Sprintf("%s; stderr: %s", msg, stderrBuf.String())
	}

	cmd := newCommandRunner(ctx, commandName, backendArgs...)

	if len(req.Invocation.Env) > 0 {
		cmd.SetEnv(req.Invocation.Env)
	}
	if len(req.Invocation.UnsetEnvKeys) > 0 {
		cmd.UnsetEnv(req.Invocation.UnsetEnvKeys)
	}

	if req.Invocation.WorkDir != "" {
		cmd.SetDir(req.Invocation.WorkDir)
	}

	stderrWriters := []io.Writer{stderrBuf}
	if req.StderrLogger != nil {
		stderrWriters = append(stderrWriters, req.StderrLogger)
	}

	if req.StderrOutput != nil {
		stderrWriters = append([]io.Writer{req.StderrOutput}, stderrWriters...)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		req.LogErrorFn("Failed to create stderr pipe: " + err.Error())
		return executionLifecycleResult{
			ExitCode: 1,
			Error:    attachStderr("failed to create stderr pipe: " + err.Error()),
		}
	}

	var stdinPipe io.WriteCloser
	if req.UseStdin {
		stdinPipe, err = cmd.StdinPipe()
		if err != nil {
			req.LogErrorFn("Failed to create stdin pipe: " + err.Error())
			closeWithReason(stderr, "stdin-pipe-failed")
			return executionLifecycleResult{
				ExitCode: 1,
				Error:    attachStderr("failed to create stdin pipe: " + err.Error()),
			}
		}
	}

	stderrDone := make(chan error, 1)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		req.LogErrorFn("Failed to create stdout pipe: " + err.Error())
		closeWithReason(stderr, "stdout-pipe-failed")
		if stdinPipe != nil {
			_ = stdinPipe.Close()
		}
		return executionLifecycleResult{
			ExitCode: 1,
			Error:    attachStderr("failed to create stdout pipe: " + err.Error()),
		}
	}

	stdoutReader := io.Reader(stdout)
	if req.StdoutLogger != nil {
		stdoutReader = io.TeeReader(stdout, req.StdoutLogger)
	}

	messageSeen := make(chan struct{}, 1)
	completeSeen := make(chan struct{}, 1)
	parseCh := make(chan parseResult, 1)
	go func() {
		msg, tid := req.Invocation.ParseStream(stdoutReader, req.LogWarnFn, req.LogInfoFn, func() {
			select {
			case messageSeen <- struct{}{}:
			default:
			}
		}, func() {
			select {
			case completeSeen <- struct{}{}:
			default:
			}
		})
		select {
		case completeSeen <- struct{}{}:
		default:
		}
		parseCh <- parseResult{message: msg, threadID: tid}
	}()

	req.LogInfoFn(fmt.Sprintf("Starting %s with args: %s...", commandName, strings.Join(backendArgs[:min(5, len(backendArgs))], " ")))

	if err := cmd.Start(); err != nil {
		closeWithReason(stdout, "start-failed")
		closeWithReason(stderr, "start-failed")
		if stdinPipe != nil {
			_ = stdinPipe.Close()
		}
		if strings.Contains(err.Error(), "executable file not found") {
			msg := fmt.Sprintf("%s command not found in PATH", commandName)
			req.LogErrorFn(msg)
			return executionLifecycleResult{
				ExitCode: 127,
				Error:    attachStderr(msg),
			}
		}
		req.LogErrorFn("Failed to start " + commandName + ": " + err.Error())
		return executionLifecycleResult{
			ExitCode: 1,
			Error:    attachStderr("failed to start " + commandName + ": " + err.Error()),
		}
	}

	req.LogInfoFn(fmt.Sprintf("Starting %s with PID: %d", commandName, cmd.Process().Pid()))
	if req.LogPath != "" {
		req.LogInfoFn(fmt.Sprintf("Log capturing to: %s", req.LogPath))
	}

	go func() {
		_, copyErr := io.Copy(io.MultiWriter(stderrWriters...), stderr)
		stderrDone <- copyErr
	}()

	if req.UseStdin && stdinPipe != nil {
		req.LogInfoFn(fmt.Sprintf("Writing %d chars to stdin...", len(req.TaskText)))
		go func(data string) {
			defer stdinPipe.Close()
			_, _ = io.WriteString(stdinPipe, data)
		}(req.TaskText)
		req.LogInfoFn("Stdin closed")
	}

	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	var (
		waitErr              error
		forceKillTimer       *forceKillTimer
		ctxCancelled         bool
		messageTimer         *time.Timer
		messageTimerCh       <-chan time.Time
		forcedAfterComplete  bool
		terminated           bool
		messageSeenObserved  bool
		completeSeenObserved bool
	)

waitLoop:
	for {
		select {
		case err := <-waitCh:
			waitErr = err
			break waitLoop
		case <-ctx.Done():
			ctxCancelled = true
			req.LogErrorFn(cancelReason(commandName, ctx))
			if !terminated {
				if timer := terminateCommandFn(cmd); timer != nil {
					forceKillTimer = timer
					terminated = true
				}
			}
			for {
				select {
				case err := <-waitCh:
					waitErr = err
					break waitLoop
				case <-time.After(forceKillWaitTimeout):
					if proc := cmd.Process(); proc != nil {
						_ = sendKillSignal(proc)
					}
				}
			}
		case <-messageTimerCh:
			forcedAfterComplete = true
			messageTimerCh = nil
			if !terminated {
				req.LogWarnFn(fmt.Sprintf("%s output parsed; terminating lingering backend", commandName))
				if timer := terminateCommandFn(cmd); timer != nil {
					forceKillTimer = timer
					terminated = true
				}
			}
			closeWithReason(stdout, "terminate")
			closeWithReason(stderr, "terminate")
			for {
				select {
				case err := <-waitCh:
					waitErr = err
					break waitLoop
				case <-time.After(forceKillWaitTimeout):
					if proc := cmd.Process(); proc != nil {
						_ = sendKillSignal(proc)
					}
				}
			}
		case <-completeSeen:
			completeSeenObserved = true
			if messageTimer != nil {
				continue
			}
			messageTimer = time.NewTimer(postMessageTerminateDelay)
			messageTimerCh = messageTimer.C
		case <-messageSeen:
			messageSeenObserved = true
		}
	}

	if messageTimer != nil {
		if !messageTimer.Stop() {
			select {
			case <-messageTimer.C:
			default:
			}
		}
	}

	if forceKillTimer != nil {
		forceKillTimer.Stop()
	}

	var parsed parseResult
	switch {
	case ctxCancelled:
		closeWithReason(stdout, stdoutCloseReasonCtx)
		parsed = <-parseCh
	case messageSeenObserved || completeSeenObserved:
		closeWithReason(stdout, stdoutCloseReasonWait)
		parsed = <-parseCh
	default:
		drainTimer := time.NewTimer(stdoutDrainTimeout)
		defer drainTimer.Stop()

		select {
		case parsed = <-parseCh:
			closeWithReason(stdout, stdoutCloseReasonWait)
		case <-messageSeen:
			messageSeenObserved = true
			closeWithReason(stdout, stdoutCloseReasonWait)
			parsed = <-parseCh
		case <-completeSeen:
			completeSeenObserved = true
			closeWithReason(stdout, stdoutCloseReasonWait)
			parsed = <-parseCh
		case <-drainTimer.C:
			closeWithReason(stdout, stdoutCloseReasonDrain)
			parsed = <-parseCh
		}
	}

	// Grace period for stderr: in the normal case the backend's exit closed the
	// write end and the copy goroutine has already hit EOF, so the first case
	// fires immediately. The wait only elapses when something still holds the
	// write end open (e.g. a spawned grandchild inheriting stderr) or trailing
	// bytes are in flight under load; waiting up to stdoutDrainTimeout keeps
	// that tail available to attachStderr instead of truncating it, then the
	// force-close bounds the cost.
	select {
	case <-stderrDone:
		closeWithReason(stderr, stdoutCloseReasonWait)
	case <-time.After(stdoutDrainTimeout):
		closeWithReason(stderr, stdoutCloseReasonDrain)
		<-stderrDone
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		res := executionLifecycleResult{ThreadID: parsed.threadID}
		if errors.Is(ctxErr, context.DeadlineExceeded) {
			res.ExitCode = 124
			res.Error = attachStderr(fmt.Sprintf("%s execution timeout", commandName))
			return res
		}
		res.ExitCode = 130
		res.Error = attachStderr("execution cancelled")
		return res
	}

	if waitErr != nil {
		if forcedAfterComplete && parsed.message != "" {
			req.LogWarnFn(fmt.Sprintf("%s terminated after delivering output", commandName))
		} else {
			res := executionLifecycleResult{ThreadID: parsed.threadID, Message: parsed.message}
			if exitErr, ok := waitErr.(*exec.ExitError); ok {
				code := exitErr.ExitCode()
				req.LogErrorFn(fmt.Sprintf("%s exited with status %d", commandName, code))
				res.ExitCode = code
				if parsed.message != "" {
					res.Error = attachStderr(parsed.message)
				} else {
					res.Error = attachStderr(fmt.Sprintf("%s exited with status %d", commandName, code))
				}
				return res
			}
			req.LogErrorFn(commandName + " error: " + waitErr.Error())
			res.ExitCode = 1
			if parsed.message != "" {
				res.Error = attachStderr(parsed.message)
			} else {
				res.Error = attachStderr(commandName + " error: " + waitErr.Error())
			}
			return res
		}
	}

	if parsed.message == "" {
		emptyMessageError := fmt.Sprintf("%s completed without agent_message output", commandName)
		req.LogErrorFn(emptyMessageError)
		return executionLifecycleResult{
			ThreadID: parsed.threadID,
			ExitCode: 1,
			Error:    attachStderr(emptyMessageError),
		}
	}

	if req.StdoutLogger != nil {
		req.StdoutLogger.Flush()
	}
	if req.StderrLogger != nil {
		req.StderrLogger.Flush()
	}

	return executionLifecycleResult{
		Message:  parsed.message,
		ThreadID: parsed.threadID,
		ExitCode: 0,
	}
}

func cancelReason(commandName string, ctx context.Context) string {
	if ctx == nil {
		return "Context cancelled"
	}

	if commandName == "" {
		commandName = defaultBackendName
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Sprintf("%s execution timeout", commandName)
	}

	return fmt.Sprintf("Execution cancelled, terminating %s process", commandName)
}

type stdoutReasonCloser interface {
	CloseWithReason(string) error
}

func closeWithReason(rc io.ReadCloser, reason string) {
	if rc == nil {
		return
	}
	if c, ok := rc.(stdoutReasonCloser); ok {
		_ = c.CloseWithReason(reason)
		return
	}
	_ = rc.Close()
}

type forceKillTimer struct {
	timer   *time.Timer
	done    chan struct{}
	stopped atomic.Bool
	drained atomic.Bool
}

func (t *forceKillTimer) Stop() {
	if t == nil || t.timer == nil {
		return
	}
	if !t.stopped.CompareAndSwap(false, true) {
		return
	}
	if !t.timer.Stop() {
		<-t.done
		t.drained.Store(true)
	}
}

func terminateCommand(cmd commandRunner) *forceKillTimer {
	if cmd == nil {
		return nil
	}
	proc := cmd.Process()
	if proc == nil {
		return nil
	}

	_ = sendTermSignal(proc)

	done := make(chan struct{}, 1)
	timer := time.AfterFunc(time.Duration(forceKillDelay.Load())*time.Second, func() {
		if p := cmd.Process(); p != nil {
			_ = sendKillSignal(p)
		}
		close(done)
	})

	return &forceKillTimer{timer: timer, done: done}
}
