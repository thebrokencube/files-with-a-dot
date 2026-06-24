package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/lint"
)

func runLint(args []string) int {
	fs := dendrik.NewFlagSet("dendrik lint")
	jsonFlag := fs.BoolLong("json", "JSON output")
	plainFlag := fs.BoolLong("plain", "Undecorated text output (no color, no JSON)")
	strictFlag := fs.BoolLong("strict", "Promote warnings to errors")
	explainFlag := fs.StringLong("explain", "", "Show rationale for a check ID")
	noColor := fs.BoolLong("no-color", "Disable color output")

	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	// --explain mode
	if *explainFlag != "" {
		return handleExplain(*explainFlag, *jsonFlag, *noColor)
	}

	// Positional arg: tool directory path
	remaining := fs.GetArgs()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: dendrik lint <path> [--json] [--plain] [--strict]")
		return dendrik.ExitUserError
	}

	toolDir, err := filepath.Abs(remaining[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return dendrik.ExitUserError
	}

	// Gather (I/O) then run the pure core.
	data, err := lint.GatherToolData(toolDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering tool data: %v\n", err)
		return dendrik.ExitExternalErr
	}
	results := lint.Run(data, lint.Options{Strict: *strictFlag})

	// Output
	out := dendrik.NewOutput(*jsonFlag, *plainFlag, *noColor)

	if out.IsJSON() {
		type jsonOutput struct {
			Tool     string        `json:"tool"`
			Errors   int           `json:"errors"`
			Warnings int           `json:"warnings"`
			Results  []lint.Result `json:"results"`
		}
		errors, warnings := countSeverities(results)
		fmt.Print(string(out.MustResult(jsonOutput{
			Tool:     data.ToolName,
			Errors:   errors,
			Warnings: warnings,
			Results:  results,
		})))
		if errors > 0 {
			return dendrik.ExitUserError
		}
		return dendrik.ExitOK
	}

	// Human output
	errors, warnings := countSeverities(results)
	if len(results) == 0 {
		fmt.Println(out.Success("All %d checks passed for %s", len(conventions.Contract), data.ToolName))
		return dendrik.ExitOK
	}

	for _, r := range results {
		icon := out.Pal.Yellow + "W" + out.Pal.Reset
		if r.Severity == conventions.SeverityError {
			icon = out.Pal.Red + "E" + out.Pal.Reset
		}
		loc := r.File
		if r.Line > 0 {
			loc = fmt.Sprintf("%s:%d", r.File, r.Line)
		}
		fmt.Printf("  %s [%s] %s", icon, r.CheckID, r.Message)
		if loc != "" {
			fmt.Printf(" (%s)", loc)
		}
		fmt.Println()
		if r.Remediation != "" {
			fmt.Printf("    %s%s%s\n", out.Pal.Dim, r.Remediation, out.Pal.Reset)
		}
	}

	fmt.Printf("\n%s: %d error(s), %d warning(s)\n", data.ToolName, errors, warnings)
	if errors > 0 {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

func handleExplain(checkID string, jsonFlag, noColor bool) int {
	entry := conventions.LookupCheck(checkID)
	if entry == nil {
		fmt.Fprintf(os.Stderr, "Unknown check: %s\n", checkID)
		return dendrik.ExitUserError
	}

	out := dendrik.NewOutput(jsonFlag, false, noColor)
	if out.IsJSON() {
		fmt.Print(string(out.MustResult(entry)))
		return dendrik.ExitOK
	}

	fmt.Printf("%s%s%s [%s] %s\n", out.Pal.Bold, entry.ID, out.Pal.Reset, entry.Severity, entry.Summary)
	fmt.Printf("\n%sRationale:%s %s\n", out.Pal.Bold, out.Pal.Reset, entry.Rationale)
	fmt.Printf("\n%sRemediation:%s %s\n", out.Pal.Bold, out.Pal.Reset, entry.Remediation)
	return dendrik.ExitOK
}

func countSeverities(results []lint.Result) (errors, warnings int) {
	for _, r := range results {
		if r.Severity == conventions.SeverityError {
			errors++
		} else {
			warnings++
		}
	}
	return
}
