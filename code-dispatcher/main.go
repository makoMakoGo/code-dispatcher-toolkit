package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultWorkdir        = "."
	defaultTimeout        = 7200 // seconds (2 hours)
	defaultCoverageTarget = 90.0
	logLineLimit          = 1000
	stdinSpecialChars     = "\n\\\"'`$"
	stderrCaptureLimit    = 64 * 1024
	defaultBackendName    = "codex"
	defaultBackendCommand = "codex"

	// stdout close reasons
	stdoutCloseReasonWait  = "wait-done"
	stdoutCloseReasonDrain = "drain-timeout"
	stdoutCloseReasonCtx   = "context-cancel"
	stdoutDrainTimeout     = 100 * time.Millisecond
)

// Test hooks for dependency injection
var (
	stdinReader    io.Reader = os.Stdin
	isTerminalFn             = defaultIsTerminal
	backendCommand           = defaultBackendCommand
	cleanupHook    func()
	loggerPtr      atomic.Pointer[Logger]

	buildArgsFn        = buildCodexArgs
	selectBackendFn    = selectBackend
	commandContext     = exec.CommandContext
	cleanupLogsFn      = cleanupOldLogs
	signalNotifyFn     = signal.Notify
	signalStopFn       = signal.Stop
	signalNotifyCtxFn  = signal.NotifyContext
	terminateCommandFn = terminateCommand
	defaultBuildArgsFn = buildCodexArgs
	runTaskFn          = runTask
	exitFn             = os.Exit
)

var forceKillDelay atomic.Int32

func init() {
	forceKillDelay.Store(5) // seconds - default value
}

func runStartupCleanup() {
	if cleanupLogsFn == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logWarn(fmt.Sprintf("cleanupOldLogs panic: %v", r))
		}
	}()
	if _, err := cleanupLogsFn(); err != nil {
		logWarn(fmt.Sprintf("cleanupOldLogs error: %v", err))
	}
}

func runCleanupMode() int {
	if cleanupLogsFn == nil {
		fmt.Fprintln(os.Stderr, "Cleanup failed: log cleanup function not configured")
		return 1
	}

	stats, err := cleanupLogsFn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cleanup failed: %v\n", err)
		return 1
	}

	fmt.Println("Cleanup completed")
	fmt.Printf("Files scanned: %d\n", stats.Scanned)
	fmt.Printf("Files deleted: %d\n", stats.Deleted)
	if len(stats.DeletedFiles) > 0 {
		for _, f := range stats.DeletedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}
	fmt.Printf("Files kept: %d\n", stats.Kept)
	if len(stats.KeptFiles) > 0 {
		for _, f := range stats.KeptFiles {
			fmt.Printf("  - %s\n", f)
		}
	}
	if stats.Errors > 0 {
		fmt.Printf("Deletion errors: %d\n", stats.Errors)
	}
	return 0
}

func main() {
	exitCode := run()
	exitFn(exitCode)
}

// run is the main logic, returns exit code for testability
func run() (exitCode int) {
	name := currentDispatcherName()
	// Handle --help first (no logger needed)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			printHelp()
			return 0
		case "--cleanup":
			return runCleanupMode()
		}
	}

	// Initialize logger for all other commands
	logger, err := NewLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to initialize logger: %v\n", err)
		return 1
	}
	setLogger(logger)

	defer func() {
		logger := activeLogger()
		if logger != nil {
			logger.Flush()
		}
		if err := closeLogger(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to close logger: %v\n", err)
		}
		// On failure, display recent errors and report log cleanup status.
		if logger != nil {
			var recentErrors []string
			if exitCode != 0 {
				recentErrors = logger.ExtractRecentErrors(10)
			}
			removeErr := logger.RemoveLogFile()
			if len(recentErrors) > 0 {
				fmt.Fprintln(os.Stderr, "\n=== Recent Errors ===")
				for _, entry := range recentErrors {
					fmt.Fprintln(os.Stderr, entry)
				}
				if removeErr != nil && !os.IsNotExist(removeErr) {
					fmt.Fprintf(os.Stderr, "Log file: %s\n", logger.Path())
				} else {
					fmt.Fprintf(os.Stderr, "Log file: %s (deleted)\n", logger.Path())
				}
			}
			if removeErr != nil && !os.IsNotExist(removeErr) {
				fmt.Fprintf(os.Stderr, "WARN: failed to remove log file %s: %v\n", logger.Path(), removeErr)
			}
		}
	}()
	defer runCleanupHook()

	// Clean up stale logs from previous runs.
	runStartupCleanup()

	// Handle remaining commands
	if len(os.Args) > 1 {
		outcome := runParallelInvocation(name, os.Args[1:])
		if outcome.Handled {
			if outcome.Stderr != "" {
				fmt.Fprint(os.Stderr, outcome.Stderr)
			}
			if outcome.Stdout != "" {
				fmt.Print(outcome.Stdout)
			}
			return outcome.ExitCode
		}
	}

	logInfo("Script started")

	cfg, err := parseArgs()
	if err != nil {
		logError(err.Error())
		return 1
	}
	logInfo(fmt.Sprintf("Parsed args: mode=%s, task_len=%d, backend=%s", cfg.Mode, len(cfg.Task), cfg.Backend))

	backend, err := selectBackendFn(cfg.Backend)
	if err != nil {
		logError(err.Error())
		return 1
	}
	cfg.Backend = backend.Name()

	cmdInjected := backendCommand != defaultBackendCommand
	argsInjected := buildArgsFn != nil && reflect.ValueOf(buildArgsFn).Pointer() != reflect.ValueOf(defaultBuildArgsFn).Pointer()

	// Wire selected backend into runtime hooks for the rest of the execution,
	// but preserve any injected test hooks for the default backend.
	if backend.Name() != defaultBackendName || !cmdInjected {
		backendCommand = backend.Command()
	}
	if backend.Name() != defaultBackendName || !argsInjected {
		buildArgsFn = backend.BuildArgs
	}
	logInfo(fmt.Sprintf("Selected backend: %s", backend.Name()))

	timeoutSec := resolveTimeout()
	logInfo(fmt.Sprintf("Timeout: %ds", timeoutSec))
	cfg.Timeout = timeoutSec

	if cfg.ExplicitStdin {
		logInfo("Explicit stdin mode: reading task from stdin")
	}

	plan, err := planSerialInvocation(cfg, backend)
	if err != nil {
		logError(err.Error())
		return 1
	}

	// Print startup information to stderr
	fmt.Fprintf(os.Stderr, "[%s]\n", name)
	fmt.Fprintf(os.Stderr, "  Backend: %s\n", cfg.Backend)
	fmt.Fprintf(os.Stderr, "  Command: %s %s\n", plan.Invocation.Command, strings.Join(plan.Invocation.Args, " "))
	fmt.Fprintf(os.Stderr, "  PID: %d\n", os.Getpid())
	fmt.Fprintf(os.Stderr, "  Log: %s\n", logger.Path())

	if plan.TaskSpec.UseStdin && len(plan.Reasons) > 0 {
		logWarn(fmt.Sprintf("Using stdin mode for task due to: %s", strings.Join(plan.Reasons, ", ")))
	}

	logInfo(fmt.Sprintf("%s running...", cfg.Backend))

	result := runTaskFn(plan.TaskSpec, false, cfg.Timeout)

	if result.ExitCode != 0 {
		return result.ExitCode
	}

	fmt.Println(result.Message)
	if result.SessionID != "" {
		fmt.Printf("\n---\nSESSION_ID: %s\n", result.SessionID)
	}

	return 0
}

func readAgentPromptFile(path string) (string, error) {
	raw := strings.TrimSpace(path)
	if raw == "" {
		return "", nil
	}

	expanded := raw
	if raw == "~" || strings.HasPrefix(raw, "~/") || strings.HasPrefix(raw, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if raw == "~" {
			expanded = home
		} else {
			expanded = home + raw[1:]
		}
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	absPath = filepath.Clean(absPath)

	allowedDir := filepath.Clean(resolvePromptBaseDir())
	if allowedDir == "" {
		return "", fmt.Errorf("failed to resolve prompt base dir for prompt file validation")
	}

	allowedAbs, err := filepath.Abs(allowedDir)
	if err == nil {
		allowedDir = filepath.Clean(allowedAbs)
	}

	isWithinDir := func(path, dir string) bool {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return false
		}
		rel = filepath.Clean(rel)
		if rel == "." {
			return true
		}
		if rel == ".." {
			return false
		}
		prefix := ".." + string(os.PathSeparator)
		return !strings.HasPrefix(rel, prefix)
	}

	if !isWithinDir(absPath, allowedDir) {
		logWarn(fmt.Sprintf("Refusing to read prompt file outside %s: %s", allowedDir, absPath))
		return "", fmt.Errorf("prompt file must be under %s", allowedDir)
	}

	resolvedPath, errPath := filepath.EvalSymlinks(absPath)
	resolvedBase, errBase := filepath.EvalSymlinks(allowedDir)
	if errPath == nil && errBase == nil {
		resolvedPath = filepath.Clean(resolvedPath)
		resolvedBase = filepath.Clean(resolvedBase)
		if !isWithinDir(resolvedPath, resolvedBase) {
			logWarn(fmt.Sprintf("Refusing to read prompt file outside %s (resolved): %s", resolvedBase, resolvedPath))
			return "", fmt.Errorf("prompt file must be under %s", resolvedBase)
		}
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

func wrapTaskWithAgentPrompt(prompt string, task string) string {
	return "<agent-prompt>\n" + prompt + "\n</agent-prompt>\n\n" + task
}

func setLogger(l *Logger) {
	loggerPtr.Store(l)
}

func closeLogger() error {
	logger := loggerPtr.Swap(nil)
	if logger == nil {
		return nil
	}
	return logger.Close()
}

func activeLogger() *Logger {
	return loggerPtr.Load()
}

func logInfo(msg string) {
	if logger := activeLogger(); logger != nil {
		logger.Info(msg)
	}
}

func logWarn(msg string) {
	if logger := activeLogger(); logger != nil {
		logger.Warn(msg)
	}
}

func logError(msg string) {
	if logger := activeLogger(); logger != nil {
		logger.Error(msg)
	}
}

func runCleanupHook() {
	if logger := activeLogger(); logger != nil {
		logger.Flush()
	}
	if cleanupHook != nil {
		cleanupHook()
	}
}

func printHelp() {
	name := currentDispatcherName()
	help := fmt.Sprintf(`%[1]s - Go dispatcher for AI CLI backends

Usage:
	%[1]s --backend <backend> "task" [workdir]
	%[1]s --backend <backend> - [workdir]              Read task from stdin
	%[1]s --backend <backend> resume <session_id> "task"
	%[1]s --backend <backend> resume <session_id> -     Read follow-up task from stdin
	%[1]s --parallel --backend <backend>               Run tasks in parallel (config from stdin)
	%[1]s --parallel --backend <backend> --full-output Run tasks in parallel with full output (legacy)
	%[1]s --help

Supported backends:
	codex | claude

Common mistakes:
	--resume is invalid; use: resume <session_id> <task>
	resume and new mode both require: --backend <backend>
	resume should not append [workdir]; it follows backend session context

Parallel mode examples:
	%[1]s --parallel --backend codex < tasks.txt
	echo '...' | %[1]s --parallel --backend claude

		Prompt Injection (default-on):
			Prompt file path: ~/.code-dispatcher/prompts/<backend>-prompt.md
		    Backends: codex | claude
		    Empty/missing prompt files behave like no injection.

	Runtime Config:
	    ~/.code-dispatcher/.env (single source of truth)
	    Supported keys include: CODE_DISPATCHER_TIMEOUT, CODE_DISPATCHER_ASCII_MODE,
	    CODE_DISPATCHER_MAX_PARALLEL_WORKERS, CODE_DISPATCHER_LOGGER_CLOSE_TIMEOUT_MS

Exit Codes:
    0    Success
    1    General error (missing args, no output)
    124  Timeout
    127  backend command not found
    130  Interrupted (Ctrl+C)
    *    Passthrough from backend process`, name)
	fmt.Println(help)
}
