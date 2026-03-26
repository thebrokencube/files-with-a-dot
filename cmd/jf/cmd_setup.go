package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/setup"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSetup(args []string) int {
	fs := dendrik.NewFlagSet("setup")
	checkOnly := fs.Bool('c', "check", "Non-interactive check only")
	jsonOut := fs.Bool('j', "json", "Output as JSON (with --check)")
	discover := fs.BoolLong("discover", "Discover and save Jira site from acli auth")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *discover {
		return setupDiscover()
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

func setupDiscover() int {
	fmt.Println("Discovering Jira site from acli auth...")

	// Call acli jira auth status to get the site
	out, err := exec.Command("acli", "jira", "auth", "status").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to get auth status: %s\n", err)
		fmt.Fprintln(os.Stderr, "  Make sure you're authenticated: acli auth login")
		return dendrik.ExitUserError
	}

	// Parse "Site: gustohq.atlassian.net" from output
	var siteName string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Site:") {
			site := strings.TrimPrefix(line, "Site:")
			site = strings.TrimSpace(site)
			// Extract name from "gustohq.atlassian.net" -> "gustohq"
			siteName = strings.TrimSuffix(site, ".atlassian.net")
			break
		}
	}

	if siteName == "" {
		fmt.Fprintln(os.Stderr, "✗ Could not find site in auth status")
		return dendrik.ExitUserError
	}

	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to load config: %s\n", err)
		return dendrik.ExitUserError
	}

	// Update config
	cfg.Site = siteName

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to save config: %s\n", err)
		return dendrik.ExitUserError
	}

	fmt.Printf("✓ Discovered site: %s\n", siteName)
	fmt.Printf("✓ Updated ~/.jf.yml\n")
	fmt.Printf("\n  Browse URLs: https://%s.atlassian.net/browse/<KEY>\n", siteName)

	return dendrik.ExitOK
}
