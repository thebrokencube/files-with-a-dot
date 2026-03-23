package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/output"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runValidate(args []string) int {
	fs := dendrik.NewFlagSet("validate")
	dir := fs.String('d', "dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool('j', "json", "Output as JSON")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	f, roots, code := loadForestOrFail(*dir, *jsonOut)
	if code != 0 {
		return code
	}

	issues := forest.Validate(roots, f)

	if *jsonOut {
		result := output.ValidateResult{
			Valid: true,
			Nodes: len(forest.Flatten(roots)),
		}
		for _, iss := range issues {
			result.Issues = append(result.Issues, output.ValidateIssue{
				Level:   iss.Level,
				Message: iss.String(),
			})
			if iss.Level == "error" {
				result.Valid = false
			}
		}
		dendrik.WriteResult(os.Stdout, result)
		if !result.Valid {
			return dendrik.ExitUserError
		}
		return dendrik.ExitOK
	}

	if len(issues) == 0 {
		fmt.Printf("✓ Forest valid (%d nodes)\n", len(forest.Flatten(roots)))
		return dendrik.ExitOK
	}

	errors := 0
	warnings := 0
	for _, iss := range issues {
		fmt.Println(iss.String())
		if iss.Level == "error" {
			errors++
		} else {
			warnings++
		}
	}

	fmt.Printf("\n%d error(s), %d warning(s)\n", errors, warnings)

	if errors > 0 {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}
