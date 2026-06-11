package main

import (
	"reflect"
	"strings"
	"testing"
)

func structuredReportMessageForTest(coverage, files, tests, summary string) string {
	var sb strings.Builder
	sb.WriteString("---CODE-DISPATCHER-REPORT---\n")
	if coverage != "" {
		sb.WriteString("Coverage: " + coverage + "\n")
	}
	if files != "" {
		sb.WriteString("Files: " + files + "\n")
	}
	if tests != "" {
		sb.WriteString("Tests: " + tests + "\n")
	}
	if summary != "" {
		sb.WriteString("Summary: " + summary + "\n")
	}
	sb.WriteString("---END-CODE-DISPATCHER-REPORT---")
	return sb.String()
}

func TestProjectTaskResultUsesExplicitReportOnly(t *testing.T) {
	result := TaskResult{
		TaskID:  "task-1",
		Message: "noise fallback.go 12%\n" + structuredReportMessageForTest("92%", "structured.go", "7 passed, 0 failed", "Structured report used"),
	}

	projected := projectTaskResult(result)

	if !projected.Report.Found {
		t.Fatalf("expected report projection to be found")
	}
	if projected.Report.Coverage != "92%" {
		t.Fatalf("Coverage = %q", projected.Report.Coverage)
	}
	if projected.Report.CoverageNum != 92 {
		t.Fatalf("CoverageNum = %v", projected.Report.CoverageNum)
	}
	if !reflect.DeepEqual(projected.Report.FilesChanged, []string{"structured.go"}) {
		t.Fatalf("FilesChanged = %#v", projected.Report.FilesChanged)
	}
	if projected.Report.TestsPassed != 7 || projected.Report.TestsFailed != 0 {
		t.Fatalf("tests = %d passed, %d failed", projected.Report.TestsPassed, projected.Report.TestsFailed)
	}
	if projected.Report.KeyOutput != "Structured report used" {
		t.Fatalf("KeyOutput = %q", projected.Report.KeyOutput)
	}
}

func TestProjectTaskResultDoesNotGuessWithoutReport(t *testing.T) {
	result := TaskResult{
		TaskID:  "task-1",
		Message: "Summary: Should not be extracted\nCoverage: 88%\nFiles: guessed.go\nTests: 4 passed, 0 failed",
	}

	projected := projectTaskResult(result)

	if projected.Report.Found {
		t.Fatalf("unexpected report projection: %#v", projected.Report)
	}
	if projected.Report.Coverage != "" || len(projected.Report.FilesChanged) != 0 || projected.Report.KeyOutput != "" {
		t.Fatalf("projection invented structured facts: %#v", projected.Report)
	}
}

func TestGenerateFinalOutputDoesNotInventReportFieldsWithoutStructuredReport(t *testing.T) {
	summary := generateFinalOutput([]TaskResult{{
		TaskID:   "plain",
		ExitCode: 0,
		Message:  "done\nCoverage: 88%\nFiles: guessed.go\nTests: 4 passed, 0 failed",
	}})

	for _, want := range []string{"1 tasks | 1 passed | 0 failed", "### plain"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q:\n%s", want, summary)
		}
	}
	for _, unwanted := range []string{"Did:", "Files:", "Tests:", "Coverage:"} {
		if strings.Contains(summary, unwanted) {
			t.Fatalf("summary invented %q without structured report:\n%s", unwanted, summary)
		}
	}
}

func TestGenerateProjectedFinalOutputWithMode(t *testing.T) {
	results := []ProjectedTaskResult{
		projectTaskResult(TaskResult{TaskID: "ok", ExitCode: 0, Message: structuredReportMessageForTest("92%", "a.go", "3 passed, 0 failed", "done")}),
		projectTaskResult(TaskResult{TaskID: "warn", ExitCode: 0, Message: structuredReportMessageForTest("80%", "b.go", "2 passed, 0 failed", "needs coverage")}),
		{Result: TaskResult{TaskID: "bad", ExitCode: 2, Error: "boom", Message: "panic: bad"}},
	}

	summary := generateProjectedFinalOutput(results)
	for _, want := range []string{"3 tasks | 2 passed | 1 failed | 1 below 90%", "### ok", "Did: done", "Files: a.go", "Tests: 3 passed", "### warn", "80% (below 90%)", "### bad"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary missing %q:\n%s", want, summary)
		}
	}

	full := generateProjectedFinalOutputWithMode(results, false)
	for _, want := range []string{"=== Parallel Execution Summary ===", "Coverage: 92%", "Coverage: 80%", "panic: bad"} {
		if !strings.Contains(full, want) {
			t.Fatalf("full output missing %q:\n%s", want, full)
		}
	}
}
