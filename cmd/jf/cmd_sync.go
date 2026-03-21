package main

import (
	"flag"
	"fmt"
)

func runSync(args []string) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")
	force := fs.Bool("force", false, "Push as plain text if marklassian conversion fails")
	failFast := fs.Bool("fail-fast", false, "Stop on first error")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	fmt.Println("── Push ──")
	pushCode := pushForest(*dir, nil, "", *force, *failFast)

	fmt.Println()
	fmt.Println("── Pull ──")
	pullCode := pullForest(*dir, nil, *failFast)

	if pushCode != 0 || pullCode != 0 {
		return 1
	}
	return 0
}
