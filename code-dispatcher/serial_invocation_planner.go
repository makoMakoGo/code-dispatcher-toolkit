package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type serialInvocationPlan struct {
	Invocation *BackendInvocation
	TaskSpec   TaskSpec
	Reasons    []string
}

func planSerialInvocation(cfg *Config, backend Backend) (*serialInvocationPlan, error) {
	taskText, piped, err := resolveSerialTaskText(cfg)
	if err != nil {
		return nil, err
	}

	promptFile := defaultPromptFileForBackend(cfg.Backend)
	if promptFile != "" {
		prompt, err := readAgentPromptFile(promptFile)
		if err != nil {
			if !os.IsNotExist(err) {
				logWarn("Failed to read prompt file: " + err.Error())
			}
		} else if strings.TrimSpace(prompt) != "" {
			taskText = wrapTaskWithAgentPrompt(prompt, taskText)
		}
	}

	useStdin := cfg.ExplicitStdin || shouldUseStdin(taskText, piped)
	targetArg := taskText
	if useStdin {
		targetArg = "-"
	}

	invocation := buildBackendInvocation(backend, cfg, targetArg)
	taskSpec := TaskSpec{
		Task:              taskText,
		WorkDir:           cfg.WorkDir,
		Mode:              cfg.Mode,
		SessionID:         cfg.SessionID,
		Backend:           cfg.Backend,
		UseStdin:          useStdin,
		plannedInvocation: &invocation,
	}

	return &serialInvocationPlan{
		Invocation: taskSpec.plannedInvocation,
		TaskSpec:   taskSpec,
		Reasons:    serialStdinReasons(taskText, piped, cfg.ExplicitStdin),
	}, nil
}

func resolveSerialTaskText(cfg *Config) (string, bool, error) {
	if cfg.ExplicitStdin {
		data, err := io.ReadAll(stdinReader)
		if err != nil {
			return "", false, fmt.Errorf("failed to read stdin: %w", err)
		}
		taskText := string(data)
		if taskText == "" {
			return "", false, fmt.Errorf("explicit stdin mode requires task input from stdin")
		}
		return taskText, !isTerminal(), nil
	}

	pipedTask, err := readPipedTask()
	if err != nil {
		return "", false, fmt.Errorf("failed to read piped stdin: %w", err)
	}
	if pipedTask != "" {
		return pipedTask, true, nil
	}
	return cfg.Task, false, nil
}

func serialStdinReasons(taskText string, piped bool, explicitStdin bool) []string {
	var reasons []string
	if piped {
		reasons = append(reasons, "piped input")
	}
	if explicitStdin {
		reasons = append(reasons, "explicit \"-\"")
	}
	if strings.Contains(taskText, "\n") {
		reasons = append(reasons, "newline")
	}
	if strings.Contains(taskText, "\\") {
		reasons = append(reasons, "backslash")
	}
	if strings.Contains(taskText, "\"") {
		reasons = append(reasons, "double-quote")
	}
	if strings.Contains(taskText, "'") {
		reasons = append(reasons, "single-quote")
	}
	if strings.Contains(taskText, "`") {
		reasons = append(reasons, "backtick")
	}
	if strings.Contains(taskText, "$") {
		reasons = append(reasons, "dollar")
	}
	if len(taskText) > 800 {
		reasons = append(reasons, "length>800")
	}
	return reasons
}
