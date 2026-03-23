package main

import (
	"encoding/json"
	"fmt"
	"jf/internal/setup"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSetup(args []string) int {
	fs := dendrik.NewFlagSet("setup")
	checkOnly := fs.Bool('c', "check", "Non-interactive check only")
	jsonOut := fs.Bool('j', "json", "Output as JSON (with --check)")

	if err := dendrik.Parse(fs, args); err != nil {
		return dendrik.ExitUserError
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
			return dendrik.ExitOK
		}
		return dendrik.ExitUserError
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
		return dendrik.ExitOK
	}

	fmt.Println("\nSome checks failed. Fix the issues above and re-run: jf setup --check")
	return dendrik.ExitUserError
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
		return dendrik.ExitOK
	}

	fmt.Fprintf(os.Stderr, "\nInstall missing prerequisites, then re-run: jf setup\n")
	return dendrik.ExitUserError
}
