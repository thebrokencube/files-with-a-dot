package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"jf/internal/setup"
	"os"
)

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ContinueOnError)
	checkOnly := fs.Bool("check", false, "Non-interactive check only")
	jsonOut := fs.Bool("json", false, "Output as JSON (with --check)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if *checkOnly {
		return setupCheck(*jsonOut)
	}

	return setupInteractive()
}

func setupCheck(jsonOut bool) int {
	results, ok := setup.CheckAll(setup.DefaultChecker)

	if jsonOut {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
		if ok {
			return 0
		}
		return 1
	}

	for _, r := range results {
		switch r.Status {
		case "ok":
			fmt.Printf("✓ %-12s %s\n", r.Name, r.Detail)
		case "missing":
			fmt.Printf("✗ %-12s %s\n", r.Name, r.Detail)
			if r.Fix != "" {
				fmt.Printf("  Fix: %s\n", r.Fix)
			}
		case "outdated":
			fmt.Printf("⚠ %-12s %s\n", r.Name, r.Detail)
			if r.Fix != "" {
				fmt.Printf("  Fix: %s\n", r.Fix)
			}
		}
	}

	if ok {
		fmt.Println("\nAll checks passed.")
		return 0
	}

	fmt.Println("\nSome checks failed. Fix the issues above and re-run: jf setup --check")
	return 1
}

func setupInteractive() int {
	fmt.Println("jf setup — checking prerequisites...")
	fmt.Println()

	results, ok := setup.CheckAll(setup.DefaultChecker)

	for _, r := range results {
		switch r.Status {
		case "ok":
			fmt.Printf("✓ %-12s %s\n", r.Name, r.Detail)
		default:
			fmt.Printf("✗ %-12s %s\n", r.Name, r.Detail)
			if r.Fix != "" {
				fmt.Printf("  → %s\n", r.Fix)
			}
		}
	}

	if ok {
		fmt.Println("\nAll prerequisites met. You're ready to use jf.")
		return 0
	}

	fmt.Fprintf(os.Stderr, "\nInstall missing prerequisites, then re-run: jf setup\n")
	return 1
}
