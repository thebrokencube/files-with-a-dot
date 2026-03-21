package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
)

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
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

	if len(roots) == 0 {
		fmt.Println("No jira: nodes found.")
		return 0
	}

	all := forest.Flatten(roots)

	for _, n := range all {
		sync := "push"
		if n.Sync == "pull" {
			sync = "pull"
		}
		fmt.Printf("%-12s %-5s %s\n", n.Key, sync, n.File)
	}
	return 0
}
