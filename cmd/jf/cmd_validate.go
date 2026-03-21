package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
)

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	f, err := forest.FindForest(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return 1
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "✗ No forest.yml found (searched up from %s)\n", *dir)
		return 1
	}

	roots, err := forest.Discover(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Discovery failed: %s\n", err)
		return 1
	}

	issues := forest.Validate(roots, f)

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
