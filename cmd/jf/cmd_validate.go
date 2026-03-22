package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
)

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool("json", false, "Output as JSON")

	if err := parseFlags(fs, args); err != nil {
		return 1
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
		output.Result(result)
		if !result.Valid {
			return 1
		}
		return 0
	}

	if len(issues) == 0 {
		fmt.Printf("✓ Forest valid (%d nodes)\n", len(forest.Flatten(roots)))
		return 0
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
		return 1
	}
	return 0
}
