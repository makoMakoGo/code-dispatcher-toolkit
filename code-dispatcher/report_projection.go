package main

import (
	"fmt"
	"strings"
)

const reportProjectionSummaryMaxLen = 150

type TaskReportProjection struct {
	Found          bool
	Coverage       string
	CoverageNum    float64
	CoverageTarget float64
	FilesChanged   []string
	KeyOutput      string
	TestsPassed    int
	TestsFailed    int
}

type ProjectedTaskResult struct {
	Result TaskResult
	Report TaskReportProjection
}

func projectTaskResults(results []TaskResult) []ProjectedTaskResult {
	projected := make([]ProjectedTaskResult, 0, len(results))
	for _, result := range results {
		projected = append(projected, projectTaskResult(result, defaultCoverageTarget))
	}
	return projected
}

func projectTaskResult(result TaskResult, coverageTarget float64) ProjectedTaskResult {
	projection := TaskReportProjection{CoverageTarget: coverageTarget}
	if report, found := extractStructuredReport(result.Message, reportProjectionSummaryMaxLen); found {
		projection.Found = true
		projection.Coverage = report.Coverage
		projection.CoverageNum = extractCoverageNum(report.Coverage)
		projection.FilesChanged = report.FilesChanged
		projection.TestsPassed = report.TestsPassed
		projection.TestsFailed = report.TestsFailed
		projection.KeyOutput = report.KeyOutput
	}
	return ProjectedTaskResult{Result: result, Report: projection}
}

func generateFinalOutput(results []TaskResult) string {
	return generateFinalOutputWithMode(results, true)
}

func generateFinalOutputWithMode(results []TaskResult, summaryOnly bool) string {
	return generateProjectedFinalOutputWithMode(projectTaskResults(results), summaryOnly)
}

func generateProjectedFinalOutput(results []ProjectedTaskResult) string {
	return generateProjectedFinalOutputWithMode(results, true)
}

// generateProjectedFinalOutputWithMode generates output based on mode.
// summaryOnly=true: structured report - every token has value.
// summaryOnly=false: full output with complete messages.
func generateProjectedFinalOutputWithMode(results []ProjectedTaskResult, summaryOnly bool) string {
	var sb strings.Builder
	successSymbol, warningSymbol, failedSymbol := getStatusSymbols()

	reportCoverageTarget := projectedCoverageTarget(results)

	success := 0
	failed := 0
	belowTarget := 0
	for _, projected := range results {
		res := projected.Result
		report := projected.Report
		if res.ExitCode == 0 && res.Error == "" {
			success++
			target := report.CoverageTarget
			if target <= 0 {
				target = reportCoverageTarget
			}
			if report.Coverage != "" && target > 0 && report.CoverageNum < target {
				belowTarget++
			}
		} else {
			failed++
		}
	}

	if summaryOnly {
		sb.WriteString("=== Execution Report ===\n")
		sb.WriteString(fmt.Sprintf("%d tasks | %d passed | %d failed", len(results), success, failed))
		if belowTarget > 0 {
			sb.WriteString(fmt.Sprintf(" | %d below %.0f%%", belowTarget, reportCoverageTarget))
		}
		sb.WriteString("\n\n")

		sb.WriteString("## Task Results\n")

		for _, projected := range results {
			res := projected.Result
			report := projected.Report
			taskID := sanitizeOutput(res.TaskID)
			coverage := sanitizeOutput(report.Coverage)
			keyOutput := sanitizeOutput(report.KeyOutput)
			logPath := sanitizeOutput(res.LogPath)
			filesChanged := sanitizeOutput(strings.Join(report.FilesChanged, ", "))

			target := report.CoverageTarget
			if target <= 0 {
				target = reportCoverageTarget
			}

			isSuccess := res.ExitCode == 0 && res.Error == ""
			isBelowTarget := isSuccess && coverage != "" && target > 0 && report.CoverageNum < target

			if isSuccess && !isBelowTarget {
				sb.WriteString(fmt.Sprintf("\n### %s %s", taskID, successSymbol))
				if coverage != "" {
					sb.WriteString(fmt.Sprintf(" %s", coverage))
				}
				sb.WriteString("\n")

				if keyOutput != "" {
					sb.WriteString(fmt.Sprintf("Did: %s\n", keyOutput))
				}
				if len(report.FilesChanged) > 0 {
					sb.WriteString(fmt.Sprintf("Files: %s\n", filesChanged))
				}
				if report.TestsPassed > 0 {
					sb.WriteString(fmt.Sprintf("Tests: %d passed\n", report.TestsPassed))
				}
				if logPath != "" {
					sb.WriteString(fmt.Sprintf("Log: %s\n", logPath))
				}

			} else if isSuccess && isBelowTarget {
				sb.WriteString(fmt.Sprintf("\n### %s %s %s (below %.0f%%)\n", taskID, warningSymbol, coverage, target))

				if keyOutput != "" {
					sb.WriteString(fmt.Sprintf("Did: %s\n", keyOutput))
				}
				if len(report.FilesChanged) > 0 {
					sb.WriteString(fmt.Sprintf("Files: %s\n", filesChanged))
				}
				if report.TestsPassed > 0 {
					sb.WriteString(fmt.Sprintf("Tests: %d passed\n", report.TestsPassed))
				}
				gap := sanitizeOutput(extractCoverageGap(res.Message))
				if gap != "" {
					sb.WriteString(fmt.Sprintf("Gap: %s\n", gap))
				}
				if logPath != "" {
					sb.WriteString(fmt.Sprintf("Log: %s\n", logPath))
				}

			} else {
				sb.WriteString(fmt.Sprintf("\n### %s %s FAILED\n", taskID, failedSymbol))
				sb.WriteString(fmt.Sprintf("Exit code: %d\n", res.ExitCode))
				if errText := sanitizeOutput(res.Error); errText != "" {
					sb.WriteString(fmt.Sprintf("Error: %s\n", errText))
				}
				detail := sanitizeOutput(extractErrorDetail(res.Message, 300))
				if detail != "" {
					sb.WriteString(fmt.Sprintf("Detail: %s\n", detail))
				}
				if logPath != "" {
					sb.WriteString(fmt.Sprintf("Log: %s\n", logPath))
				}
			}
		}

		sb.WriteString("\n## Summary\n")
		sb.WriteString(fmt.Sprintf("- %d/%d completed successfully\n", success, len(results)))

		if belowTarget > 0 || failed > 0 {
			var needFix []string
			var needCoverage []string
			for _, projected := range results {
				res := projected.Result
				report := projected.Report
				if res.ExitCode != 0 || res.Error != "" {
					taskID := sanitizeOutput(res.TaskID)
					reason := sanitizeOutput(res.Error)
					if reason == "" && res.ExitCode != 0 {
						reason = fmt.Sprintf("exit code %d", res.ExitCode)
					}
					reason = safeTruncate(reason, 50)
					needFix = append(needFix, fmt.Sprintf("%s (%s)", taskID, reason))
					continue
				}

				target := report.CoverageTarget
				if target <= 0 {
					target = reportCoverageTarget
				}
				if report.Coverage != "" && target > 0 && report.CoverageNum < target {
					needCoverage = append(needCoverage, sanitizeOutput(res.TaskID))
				}
			}
			if len(needFix) > 0 {
				sb.WriteString(fmt.Sprintf("- Fix: %s\n", strings.Join(needFix, ", ")))
			}
			if len(needCoverage) > 0 {
				sb.WriteString(fmt.Sprintf("- Coverage: %s\n", strings.Join(needCoverage, ", ")))
			}
		}

	} else {
		sb.WriteString("=== Parallel Execution Summary ===\n")
		sb.WriteString(fmt.Sprintf("Total: %d | Success: %d | Failed: %d\n\n", len(results), success, failed))

		for _, projected := range results {
			res := projected.Result
			report := projected.Report
			taskID := sanitizeOutput(res.TaskID)
			sb.WriteString(fmt.Sprintf("--- Task: %s ---\n", taskID))
			if res.Error != "" {
				sb.WriteString(fmt.Sprintf("Status: FAILED (exit code %d)\nError: %s\n", res.ExitCode, sanitizeOutput(res.Error)))
			} else if res.ExitCode != 0 {
				sb.WriteString(fmt.Sprintf("Status: FAILED (exit code %d)\n", res.ExitCode))
			} else {
				sb.WriteString("Status: SUCCESS\n")
			}
			if report.Coverage != "" {
				sb.WriteString(fmt.Sprintf("Coverage: %s\n", sanitizeOutput(report.Coverage)))
			}
			if res.SessionID != "" {
				sb.WriteString(fmt.Sprintf("Session: %s\n", sanitizeOutput(res.SessionID)))
			}
			if res.LogPath != "" {
				logPath := sanitizeOutput(res.LogPath)
				if res.sharedLog {
					sb.WriteString(fmt.Sprintf("Log: %s (shared)\n", logPath))
				} else {
					sb.WriteString(fmt.Sprintf("Log: %s\n", logPath))
				}
			}
			if res.Message != "" {
				message := sanitizeOutput(res.Message)
				if message != "" {
					sb.WriteString(fmt.Sprintf("\n%s\n", message))
				}
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func projectedCoverageTarget(results []ProjectedTaskResult) float64 {
	for _, projected := range results {
		if projected.Report.CoverageTarget > 0 {
			return projected.Report.CoverageTarget
		}
	}
	return defaultCoverageTarget
}
