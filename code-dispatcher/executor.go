package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// commandRunner abstracts exec.Cmd for testability
type commandRunner interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	StdinPipe() (io.WriteCloser, error)
	SetStderr(io.Writer)
	SetDir(string)
	SetEnv(env map[string]string)
	UnsetEnv(keys []string)
	Process() processHandle
}

// processHandle abstracts os.Process for testability
type processHandle interface {
	Pid() int
	Kill() error
	Signal(os.Signal) error
}

// realCmd implements commandRunner using exec.Cmd
type realCmd struct {
	cmd          *exec.Cmd
	stdoutWriter *os.File
	stderrWriter *os.File
}

func (r *realCmd) Start() error {
	if r.cmd == nil {
		return errors.New("command is nil")
	}
	err := r.cmd.Start()
	if r.stdoutWriter != nil {
		_ = r.stdoutWriter.Close()
		r.stdoutWriter = nil
	}
	if r.stderrWriter != nil {
		_ = r.stderrWriter.Close()
		r.stderrWriter = nil
	}
	return err
}

func (r *realCmd) Wait() error {
	if r.cmd == nil {
		return errors.New("command is nil")
	}
	return r.cmd.Wait()
}

func (r *realCmd) StdoutPipe() (io.ReadCloser, error) {
	if r.cmd == nil {
		return nil, errors.New("command is nil")
	}
	if r.cmd.Stdout != nil {
		return nil, errors.New("stdout already set")
	}
	if r.cmd.Process != nil {
		return nil, errors.New("stdout pipe after process started")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	r.cmd.Stdout = writer
	r.stdoutWriter = writer
	return reader, nil
}

func (r *realCmd) StderrPipe() (io.ReadCloser, error) {
	if r.cmd == nil {
		return nil, errors.New("command is nil")
	}
	if r.cmd.Stderr != nil {
		return nil, errors.New("stderr already set")
	}
	if r.cmd.Process != nil {
		return nil, errors.New("stderr pipe after process started")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	r.cmd.Stderr = writer
	r.stderrWriter = writer
	return reader, nil
}

func (r *realCmd) StdinPipe() (io.WriteCloser, error) {
	if r.cmd == nil {
		return nil, errors.New("command is nil")
	}
	return r.cmd.StdinPipe()
}

func (r *realCmd) SetStderr(w io.Writer) {
	if r.cmd != nil {
		r.cmd.Stderr = w
	}
}

func (r *realCmd) SetDir(dir string) {
	if r.cmd != nil {
		r.cmd.Dir = dir
	}
}

func (r *realCmd) SetEnv(env map[string]string) {
	if r == nil || r.cmd == nil || len(env) == 0 {
		return
	}

	merged := mergedCommandEnvMap(r.cmd.Env)
	for k, v := range env {
		if strings.TrimSpace(k) == "" {
			continue
		}
		merged[k] = v
	}

	r.cmd.Env = envMapToList(merged)
}

func (r *realCmd) UnsetEnv(keys []string) {
	if r == nil || r.cmd == nil || len(keys) == 0 {
		return
	}

	merged := mergedCommandEnvMap(r.cmd.Env)
	for _, key := range keys {
		deleteEnvKey(merged, key)
	}
	r.cmd.Env = envMapToList(merged)
}

func mergedCommandEnvMap(cmdEnv []string) map[string]string {
	if len(cmdEnv) == 0 {
		merged := make(map[string]string, len(os.Environ()))
		appendEnvList(merged, os.Environ())
		return merged
	}

	merged := make(map[string]string, len(cmdEnv))
	appendEnvList(merged, cmdEnv)
	return merged
}

func appendEnvList(dst map[string]string, envList []string) {
	for _, kv := range envList {
		if kv == "" {
			continue
		}
		idx := strings.IndexByte(kv, '=')
		if idx <= 0 {
			continue
		}
		dst[kv[:idx]] = kv[idx+1:]
	}
}

func envMapToList(envMap map[string]string) []string {
	keys := make([]string, 0, len(envMap))
	for k := range envMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+envMap[k])
	}
	return out
}

func deleteEnvKey(envMap map[string]string, key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	for existing := range envMap {
		if strings.EqualFold(existing, key) {
			delete(envMap, existing)
		}
	}
}

func (r *realCmd) Process() processHandle {
	if r == nil || r.cmd == nil || r.cmd.Process == nil {
		return nil
	}
	return &realProcess{proc: r.cmd.Process}
}

// realProcess implements processHandle using os.Process
type realProcess struct {
	proc *os.Process
}

func (p *realProcess) Pid() int {
	if p == nil || p.proc == nil {
		return 0
	}
	return p.proc.Pid
}

func (p *realProcess) Kill() error {
	if p == nil || p.proc == nil {
		return nil
	}
	return p.proc.Kill()
}

func (p *realProcess) Signal(sig os.Signal) error {
	if p == nil || p.proc == nil {
		return nil
	}
	return p.proc.Signal(sig)
}

// newCommandRunner creates a new commandRunner (test hook injection point)
var newCommandRunner = func(ctx context.Context, name string, args ...string) commandRunner {
	cmd := commandContext(ctx, name, args...)
	prepareCommandForSignals(cmd)
	return &realCmd{cmd: cmd}
}

type parseResult struct {
	message  string
	threadID string
}

type taskLoggerContextKey struct{}

func withTaskLogger(ctx context.Context, logger *Logger) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, taskLoggerContextKey{}, logger)
}

func taskLoggerFromContext(ctx context.Context) *Logger {
	logger, _ := effectiveContext(ctx).Value(taskLoggerContextKey{}).(*Logger)
	return logger
}

type taskLoggerHandle struct {
	logger  *Logger
	path    string
	shared  bool
	closeFn func()
}

func newTaskLoggerHandle(taskID string) taskLoggerHandle {
	taskLogger, err := NewLoggerWithSuffix(taskID)
	if err == nil {
		return taskLoggerHandle{
			logger:  taskLogger,
			path:    taskLogger.Path(),
			closeFn: func() { _ = taskLogger.Close() },
		}
	}

	msg := fmt.Sprintf("Failed to create task logger for %s: %v, using main logger", taskID, err)
	mainLogger := activeLogger()
	if mainLogger != nil {
		logWarn(msg)
		return taskLoggerHandle{
			logger: mainLogger,
			path:   mainLogger.Path(),
			shared: true,
		}
	}

	fmt.Fprintln(os.Stderr, msg)
	return taskLoggerHandle{}
}

// defaultRunParallelTaskFn is the default implementation of runParallelTaskFn (exposed for test reset)
func defaultRunParallelTaskFn(task TaskSpec, timeout int) TaskResult {
	if task.WorkDir == "" {
		task.WorkDir = defaultWorkdir
	}
	if task.Mode == "" {
		task.Mode = "new"
	}

	backendName := task.Backend
	if backendName == "" {
		backendName = defaultBackendName
	}

	promptFile := defaultPromptFileForBackend(backendName)
	if promptFile != "" {
		prompt, err := readAgentPromptFile(promptFile)
		if err != nil {
			if !os.IsNotExist(err) {
				logWarn("Failed to read default prompt file: " + err.Error())
			}
		} else if strings.TrimSpace(prompt) != "" {
			task.Task = wrapTaskWithAgentPrompt(prompt, task.Task)
		}
	}
	backend, err := selectBackendFn(backendName)
	if err != nil {
		return TaskResult{TaskID: task.ID, ExitCode: 1, Error: err.Error()}
	}
	task.Backend = backend.Name()
	task.UseStdin = task.UseStdin || shouldUseStdin(task.Task, false)

	parentCtx := context.Background()
	if task.Context != nil {
		parentCtx = task.Context
		task.Context = nil
	}
	return runTaskWithContext(parentCtx, task, backend, nil, false, true, timeout)
}

var runParallelTaskFn = defaultRunParallelTaskFn

func topologicalSort(tasks []TaskSpec) ([][]TaskSpec, error) {
	idToTask := make(map[string]TaskSpec, len(tasks))
	indegree := make(map[string]int, len(tasks))
	adj := make(map[string][]string, len(tasks))

	for _, task := range tasks {
		idToTask[task.ID] = task
		indegree[task.ID] = 0
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if _, ok := idToTask[dep]; !ok {
				return nil, fmt.Errorf("dependency %q not found for task %q", dep, task.ID)
			}
			indegree[task.ID]++
			adj[dep] = append(adj[dep], task.ID)
		}
	}

	queue := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if indegree[task.ID] == 0 {
			queue = append(queue, task.ID)
		}
	}

	layers := make([][]TaskSpec, 0)
	processed := 0

	for len(queue) > 0 {
		current := queue
		queue = nil
		layer := make([]TaskSpec, len(current))
		for i, id := range current {
			layer[i] = idToTask[id]
			processed++
		}
		layers = append(layers, layer)

		next := make([]string, 0)
		for _, id := range current {
			for _, neighbor := range adj[id] {
				indegree[neighbor]--
				if indegree[neighbor] == 0 {
					next = append(next, neighbor)
				}
			}
		}
		queue = append(queue, next...)
	}

	if processed != len(tasks) {
		cycleIDs := make([]string, 0)
		for id, deg := range indegree {
			if deg > 0 {
				cycleIDs = append(cycleIDs, id)
			}
		}
		sort.Strings(cycleIDs)
		return nil, fmt.Errorf("cycle detected involving tasks: %s", strings.Join(cycleIDs, ","))
	}

	return layers, nil
}

func executeConcurrent(layers [][]TaskSpec, timeout int) []TaskResult {
	maxWorkers := resolveMaxParallelWorkers()
	return executeConcurrentWithContext(context.Background(), layers, timeout, maxWorkers)
}

func executeConcurrentWithContext(parentCtx context.Context, layers [][]TaskSpec, timeout int, maxWorkers int) []TaskResult {
	totalTasks := 0
	for _, layer := range layers {
		totalTasks += len(layer)
	}

	results := make([]TaskResult, 0, totalTasks)
	failed := make(map[string]TaskResult, totalTasks)
	resultsCh := make(chan TaskResult, totalTasks)

	var startPrintMu sync.Mutex
	bannerPrinted := false

	printTaskStart := func(taskID, logPath string, shared bool) {
		if logPath == "" {
			return
		}
		startPrintMu.Lock()
		if !bannerPrinted {
			fmt.Fprintln(os.Stderr, "=== Starting Parallel Execution ===")
			bannerPrinted = true
		}
		label := "Log"
		if shared {
			label = "Log (shared)"
		}
		fmt.Fprintf(os.Stderr, "Task %s: %s: %s\n", taskID, label, logPath)
		startPrintMu.Unlock()
	}

	ctx := parentCtx
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workerLimit := maxWorkers
	if workerLimit < 0 {
		workerLimit = 0
	}

	var sem chan struct{}
	if workerLimit > 0 {
		sem = make(chan struct{}, workerLimit)
	}

	logConcurrencyPlanning(workerLimit, totalTasks)

	acquireSlot := func() bool {
		if sem == nil {
			return true
		}
		select {
		case sem <- struct{}{}:
			return true
		case <-ctx.Done():
			return false
		}
	}

	releaseSlot := func() {
		if sem == nil {
			return
		}
		select {
		case <-sem:
		default:
		}
	}

	var activeWorkers int64

	for _, layer := range layers {
		var wg sync.WaitGroup
		executed := 0

		for _, task := range layer {
			if skip, reason := shouldSkipTask(task, failed); skip {
				res := TaskResult{TaskID: task.ID, ExitCode: 1, Error: reason}
				results = append(results, res)
				failed[task.ID] = res
				continue
			}

			if ctx.Err() != nil {
				res := cancelledTaskResult(task.ID, ctx)
				results = append(results, res)
				failed[task.ID] = res
				continue
			}

			executed++
			wg.Add(1)
			go func(ts TaskSpec) {
				defer wg.Done()
				var taskLogPath string
				handle := taskLoggerHandle{}
				defer func() {
					if r := recover(); r != nil {
						resultsCh <- TaskResult{TaskID: ts.ID, ExitCode: 1, Error: fmt.Sprintf("panic: %v", r), LogPath: taskLogPath, sharedLog: handle.shared}
					}
				}()

				if !acquireSlot() {
					resultsCh <- cancelledTaskResult(ts.ID, ctx)
					return
				}
				defer releaseSlot()

				current := atomic.AddInt64(&activeWorkers, 1)
				logConcurrencyState("start", ts.ID, int(current), workerLimit)
				defer func() {
					after := atomic.AddInt64(&activeWorkers, -1)
					logConcurrencyState("done", ts.ID, int(after), workerLimit)
				}()

				handle = newTaskLoggerHandle(ts.ID)
				taskLogPath = handle.path
				if handle.closeFn != nil {
					defer handle.closeFn()
				}

				taskCtx := ctx
				if handle.logger != nil {
					taskCtx = withTaskLogger(ctx, handle.logger)
				}
				ts.Context = taskCtx

				printTaskStart(ts.ID, taskLogPath, handle.shared)

				res := runParallelTaskFn(ts, timeout)
				if taskLogPath != "" {
					if res.LogPath == "" || (handle.shared && handle.logger != nil && res.LogPath == handle.logger.Path()) {
						res.LogPath = taskLogPath
					}
				}
				// 只有当最终的 LogPath 确实是共享 logger 的路径时才标记为 shared
				if handle.shared && handle.logger != nil && res.LogPath == handle.logger.Path() {
					res.sharedLog = true
				}
				resultsCh <- res
			}(task)
		}

		wg.Wait()

		for i := 0; i < executed; i++ {
			res := <-resultsCh
			results = append(results, res)
			if res.ExitCode != 0 || res.Error != "" {
				failed[res.TaskID] = res
			}
		}
	}

	return results
}

func cancelledTaskResult(taskID string, ctx context.Context) TaskResult {
	exitCode := 130
	msg := "execution cancelled"
	if ctx != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		exitCode = 124
		msg = "execution timeout"
	}
	return TaskResult{TaskID: taskID, ExitCode: exitCode, Error: msg}
}

func shouldSkipTask(task TaskSpec, failed map[string]TaskResult) (bool, string) {
	if len(task.Dependencies) == 0 {
		return false, ""
	}

	var blocked []string
	for _, dep := range task.Dependencies {
		if _, ok := failed[dep]; ok {
			blocked = append(blocked, dep)
		}
	}

	if len(blocked) == 0 {
		return false, ""
	}

	return true, fmt.Sprintf("skipped due to failed dependencies: %s", strings.Join(blocked, ","))
}

// getStatusSymbols returns status symbols based on ASCII mode.
func getStatusSymbols() (success, warning, failed string) {
	if parseBoolFlag(getEnv("CODE_DISPATCHER_ASCII_MODE", ""), false) {
		return "PASS", "WARN", "FAIL"
	}
	return "✓", "⚠️", "✗"
}

func runTask(taskSpec TaskSpec, silent bool, timeoutSec int) TaskResult {
	return runTaskWithContext(context.Background(), taskSpec, nil, nil, false, silent, timeoutSec)
}

func runBackendProcess(parentCtx context.Context, backendArgs []string, taskText string, useStdin bool, timeoutSec int) (message, threadID string, exitCode int) {
	res := runTaskWithContext(parentCtx, TaskSpec{Task: taskText, WorkDir: defaultWorkdir, Mode: "new", UseStdin: useStdin}, nil, backendArgs, true, false, timeoutSec)
	return res.Message, res.SessionID, res.ExitCode
}

func runTaskWithContext(parentCtx context.Context, taskSpec TaskSpec, backend Backend, customArgs []string, useCustomArgs bool, silent bool, timeoutSec int) TaskResult {
	parentCtx = effectiveTaskContext(parentCtx, taskSpec.Context)

	result := TaskResult{TaskID: taskSpec.ID}
	injectedLogger := taskLoggerFromContext(parentCtx)
	logger := injectedLogger

	cfg := &Config{
		Mode:      taskSpec.Mode,
		Task:      taskSpec.Task,
		SessionID: taskSpec.SessionID,
		WorkDir:   taskSpec.WorkDir,
		Backend:   defaultBackendName,
	}

	commandName := backendCommand
	argsBuilder := buildArgsFn
	if backend != nil {
		cfg.Backend = backend.Name()
	} else if taskSpec.Backend != "" {
		cfg.Backend = taskSpec.Backend
	} else if commandName != "" {
		cfg.Backend = commandName
	}

	if cfg.Mode == "" {
		cfg.Mode = "new"
	}
	if cfg.WorkDir == "" {
		cfg.WorkDir = defaultWorkdir
	}

	if cfg.Mode == "resume" && strings.TrimSpace(cfg.SessionID) == "" {
		result.ExitCode = 1
		result.Error = "resume mode requires non-empty session_id"
		return result
	}

	useStdin := taskSpec.UseStdin
	targetArg := taskSpec.Task
	if useStdin {
		targetArg = "-"
	}

	invocation := BackendInvocation{}
	switch {
	case useCustomArgs:
		invocation = legacyBackendInvocation(cfg, commandName, customArgs)
	case backend != nil:
		invocation = backend.BuildInvocation(cfg, targetArg)
	default:
		invocation = legacyBackendInvocation(cfg, commandName, argsBuilder(cfg, targetArg))
	}
	if invocation.BackendName != "" {
		cfg.Backend = invocation.BackendName
	}

	prefixMsg := func(msg string) string {
		if taskSpec.ID == "" {
			return msg
		}
		return fmt.Sprintf("[Task: %s] %s", taskSpec.ID, msg)
	}

	var logInfoFn func(string)
	var logWarnFn func(string)
	var logErrorFn func(string)

	if silent {
		// Silent mode: only persist to file when available; avoid stderr noise.
		logInfoFn = func(msg string) {
			if logger != nil {
				logger.Info(prefixMsg(msg))
			}
		}
		logWarnFn = func(msg string) {
			if logger != nil {
				logger.Warn(prefixMsg(msg))
			}
		}
		logErrorFn = func(msg string) {
			if logger != nil {
				logger.Error(prefixMsg(msg))
			}
		}
	} else {
		logInfoFn = func(msg string) { logInfo(prefixMsg(msg)) }
		logWarnFn = func(msg string) { logWarn(prefixMsg(msg)) }
		logErrorFn = func(msg string) { logError(prefixMsg(msg)) }
	}

	var stdoutLogger *logWriter
	var stderrLogger *logWriter

	var tempLogger *Logger
	if logger == nil && silent && activeLogger() == nil {
		if l, err := NewLogger(); err == nil {
			setLogger(l)
			tempLogger = l
			logger = l
		}
	}
	defer func() {
		if tempLogger != nil {
			_ = closeLogger()
		}
	}()
	defer func() {
		if result.LogPath != "" || logger == nil {
			return
		}
		result.LogPath = logger.Path()
	}()
	if logger == nil {
		logger = activeLogger()
	}
	if logger != nil {
		result.LogPath = logger.Path()
	}

	if !silent {
		// Note: Empty prefix ensures backend output is logged as-is without any dispatcher format.
		// This preserves the original stdout/stderr content from codex/claude backends.
		// Trade-off: Reduces distinguishability between stdout/stderr in logs, but maintains
		// output fidelity which is critical for debugging backend-specific issues.
		stdoutLogger = newLogWriter("", logLineLimit)
		stderrLogger = newLogWriter("", logLineLimit)
	}

	var stderrOutput io.Writer
	if !silent {
		stderrOutput = os.Stderr
	}

	lifecycleResult := runExecutionLifecycle(executionLifecycleRequest{
		Context:      parentCtx,
		Invocation:   invocation,
		TaskText:     taskSpec.Task,
		UseStdin:     useStdin,
		TimeoutSec:   timeoutSec,
		LogPath:      result.LogPath,
		StdoutLogger: stdoutLogger,
		StderrLogger: stderrLogger,
		StderrOutput: stderrOutput,
		LogInfoFn:    logInfoFn,
		LogWarnFn:    logWarnFn,
		LogErrorFn:   logErrorFn,
	})

	result.ExitCode = lifecycleResult.ExitCode
	result.Message = lifecycleResult.Message
	result.SessionID = lifecycleResult.ThreadID
	result.Error = lifecycleResult.Error
	if result.LogPath == "" && injectedLogger != nil {
		result.LogPath = injectedLogger.Path()
	}

	return result
}

func effectiveTaskContext(parentCtx, taskCtx context.Context) context.Context {
	if taskCtx != nil {
		return taskCtx
	}
	if parentCtx != nil {
		return parentCtx
	}
	return context.Background()
}

func effectiveContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
