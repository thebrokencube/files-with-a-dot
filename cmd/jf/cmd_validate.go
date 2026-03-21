package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
	"os"
)

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool("json", false, "Output as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	f, err := forest.FindForest(*dir)
	if err != nil {
		if *jsonOut {
			output.Error(err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		}
		return 1
	}
	if f == nil {
		if *jsonOut {
			output.Error("No forest.yml found", *dir)
		} else {
			fmt.Fprintf(os.Stderr, "✗ No forest.yml found (searched up from %s)\n", *dir)
		}
		return 1
	}

	roots, err := forest.Discover(f)
	if err != nil {
		if *jsonOut {
			output.Error("Discovery failed", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "✗ Discovery failed: %s\n", err)
		}
		return 1
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
