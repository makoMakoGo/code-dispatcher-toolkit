package main

import (
	"fmt"
	"io"
	"strings"
)

type parallelRunOutcome struct {
	Handled  bool
	ExitCode int
	Stdout   string
	Stderr   string
}

func runParallelInvocation(dispatcherName string, args []string) parallelRunOutcome {
	if !hasParallelFlag(args) {
		return parallelRunOutcome{}
	}

	backendName := ""
	backendSpecified := false
	fullOutput := false
	var extras []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--parallel":
			continue
		case arg == "--full-output":
			fullOutput = true
		case arg == "--backend":
			if i+1 >= len(args) {
				return parallelErrorOutcome("--backend flag requires a value")
			}
			value := strings.TrimSpace(args[i+1])
			if value == "" {
				return parallelErrorOutcome("--backend flag requires a value")
			}
			backendName = value
			backendSpecified = true
			i++
		case strings.HasPrefix(arg, "--backend="):
			value := strings.TrimSpace(strings.TrimPrefix(arg, "--backend="))
			if value == "" {
				return parallelErrorOutcome("--backend flag requires a value")
			}
			backendName = value
			backendSpecified = true
		default:
			extras = append(extras, arg)
		}
	}

	if len(extras) > 0 {
		var stderr strings.Builder
		fmt.Fprintln(&stderr, "ERROR: --parallel reads its task configuration from stdin; only --backend and --full-output are allowed.")
		fmt.Fprintln(&stderr, "Usage examples:")
		fmt.Fprintf(&stderr, "  %s --parallel --backend codex < tasks.txt\n", dispatcherName)
		fmt.Fprintf(&stderr, "  echo '...' | %s --parallel --backend claude\n", dispatcherName)
		return parallelRunOutcome{Handled: true, ExitCode: 1, Stderr: stderr.String()}
	}

	if !backendSpecified {
		var stderr strings.Builder
		fmt.Fprintf(&stderr, "ERROR: --backend is required in --parallel mode (supported: %s)\n", supportedBackendNamesText())
		fmt.Fprintln(&stderr, "Usage examples:")
		fmt.Fprintf(&stderr, "  %s --parallel --backend codex < tasks.txt\n", dispatcherName)
		fmt.Fprintf(&stderr, "  %s --parallel --backend claude <<'EOF'\n", dispatcherName)
		return parallelRunOutcome{Handled: true, ExitCode: 1, Stderr: stderr.String()}
	}

	backend, err := selectBackendFn(backendName)
	if err != nil {
		return parallelErrorOutcome(fmt.Sprintf("selecting backend %q: %v", backendName, err))
	}
	backendName = backend.Name()

	data, err := io.ReadAll(stdinReader)
	if err != nil {
		return parallelErrorOutcome(fmt.Sprintf("failed to read stdin: %v", err))
	}

	cfg, err := parseParallelConfig(data)
	if err != nil {
		return parallelErrorOutcome(err.Error())
	}

	cfg.GlobalBackend = backendName
	for i := range cfg.Tasks {
		if strings.TrimSpace(cfg.Tasks[i].Backend) == "" {
			cfg.Tasks[i].Backend = backendName
		}
	}

	timeoutSec := resolveTimeout()
	layers, err := topologicalSort(cfg.Tasks)
	if err != nil {
		return parallelErrorOutcome(err.Error())
	}

	results := executeConcurrent(layers, timeoutSec)
	output := generateFinalOutputWithMode(results, !fullOutput)

	return parallelRunOutcome{
		Handled:  true,
		ExitCode: parallelExitCode(results),
		Stdout:   output + "\n",
	}
}

func hasParallelFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--parallel" {
			return true
		}
	}
	return false
}

func parallelErrorOutcome(msg string) parallelRunOutcome {
	return parallelRunOutcome{
		Handled:  true,
		ExitCode: 1,
		Stderr:   fmt.Sprintf("ERROR: %s\n", msg),
	}
}

func parallelExitCode(results []TaskResult) int {
	exitCode := 0
	for _, res := range results {
		if res.ExitCode != 0 {
			exitCode = res.ExitCode
		}
	}
	return exitCode
}
